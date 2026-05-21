package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	"github.com/topxeq/xxaitk/internal/output"
)

type SSHHandler struct{}

type SSHPayload struct {
	Host          string   `json:"host"`
	Port          int      `json:"port,omitempty"`
	User          string   `json:"user"`
	Password      string   `json:"password,omitempty"`
	Key           string   `json:"key,omitempty"`
	KeyPassphrase string   `json:"key_passphrase,omitempty"`
	KnownHosts    string   `json:"known_hosts,omitempty"`
	StrictHostKey bool     `json:"strict_host_key,omitempty"`
	Timeout       string   `json:"timeout,omitempty"`
	CmdTimeout    string   `json:"cmd_timeout,omitempty"`
	Action        string   `json:"action"`
	Cmd           string   `json:"cmd,omitempty"`
	CmdFile       string   `json:"cmd_file,omitempty"`
	LocalPath     string   `json:"local_path,omitempty"`
	RemotePath    string   `json:"remote_path,omitempty"`
	FileName      string   `json:"file_name,omitempty"`
	TargetPath    string   `json:"target_path,omitempty"`
	TempPath      string   `json:"temp_path,omitempty"`
	Mode          string   `json:"mode,omitempty"`
	Plan          string   `json:"plan,omitempty"`
	PlanJSON      string   `json:"plan_json,omitempty"`
	Direction     string   `json:"direction,omitempty"`
	Recursive     bool     `json:"recursive,omitempty"`
	Delete        bool     `json:"delete,omitempty"`
	DryRun        bool     `json:"dry_run,omitempty"`
	Conflict      string   `json:"conflict,omitempty"`
	Include       []string `json:"include,omitempty"`
	Exclude       []string `json:"exclude,omitempty"`
}

type SSHCommandResult struct {
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code"`
	TimedOut bool   `json:"timed_out"`
}

type sshConn struct {
	client *ssh.Client
	sftp   *sftp.Client
}

func (h *SSHHandler) Handle(data string, source string) *output.Response {
	start := time.Now()

	payload := &SSHPayload{}
	if err := json.Unmarshal([]byte(data), payload); err != nil {
		return output.Fail("ssh", source, "SSH_PARSE_ERROR",
			fmt.Sprintf("parse error: %v", err), "", start)
	}

	if payload.Host == "" || payload.User == "" {
		return output.Fail("ssh", source, "SSH_MISSING_PARAMS",
			"host and user are required", "", start)
	}
	if payload.Password == "" && payload.Key == "" {
		return output.Fail("ssh", source, "SSH_MISSING_AUTH",
			"password or key is required", "", start)
	}
	if payload.Port == 0 {
		payload.Port = 22
	}
	if payload.Action == "" {
		payload.Action = "cmd"
	}
	if payload.Direction == "" {
		payload.Direction = "push"
	}
	if payload.Conflict == "" {
		payload.Conflict = "fail_on_conflict"
	}

	timeout := 30 * time.Second
	if payload.Timeout != "" {
		if d, err := time.ParseDuration(payload.Timeout); err == nil {
			timeout = d
		}
	}

	var cmdTimeout time.Duration
	if payload.CmdTimeout != "" {
		if d, err := time.ParseDuration(payload.CmdTimeout); err == nil {
			cmdTimeout = d
		}
	}

	conn, err := dialSSH(payload, timeout)
	if err != nil {
		return output.Fail("ssh", source, "SSH_CONNECT_ERROR", err.Error(), "", start)
	}
	defer conn.close()

	switch payload.Action {
	case "cmd":
		return h.handleCmd(conn, payload, source, cmdTimeout, start)
	case "upload":
		return h.handleUpload(conn, payload, source, start)
	case "download":
		return h.handleDownload(conn, payload, source, start)
	case "upload_atomic":
		return h.handleUploadAtomic(conn, payload, source, start)
	case "mkdir":
		return h.handleMkdir(conn, payload, source, start)
	case "remove":
		return h.handleRemove(conn, payload, source, start)
	case "chmod":
		return h.handleChmod(conn, payload, source, start)
	case "move":
		return h.handleMove(conn, payload, source, start)
	case "deploy":
		return h.handleDeploy(conn, payload, source, cmdTimeout, start)
	case "sync":
		return h.handleSync(conn, payload, source, start)
	default:
		return output.Fail("ssh", source, "SSH_UNKNOWN_ACTION",
			fmt.Sprintf("unknown action: %s", payload.Action),
			"use: cmd, upload, download, upload_atomic, mkdir, remove, chmod, move, deploy, sync", start)
	}
}

func dialSSH(p *SSHPayload, timeout time.Duration) (*sshConn, error) {
	var authMethods []ssh.AuthMethod

	if p.Password != "" {
		authMethods = append(authMethods, ssh.Password(p.Password))
	}

	if p.Key != "" {
		keyData, err := os.ReadFile(expandHome(p.Key))
		if err != nil {
			return nil, fmt.Errorf("read key file: %w", err)
		}
		var signer ssh.Signer
		if p.KeyPassphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(keyData, []byte(p.KeyPassphrase))
		} else {
			signer, err = ssh.ParsePrivateKey(keyData)
		}
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	hostKeyCallback := ssh.InsecureIgnoreHostKey()
	if p.StrictHostKey {
		if p.KnownHosts == "" {
			return nil, fmt.Errorf("strict_host_key requires known_hosts path")
		}
		kh, err := knownhosts.New(expandHome(p.KnownHosts))
		if err != nil {
			return nil, fmt.Errorf("parse known_hosts: %w", err)
		}
		hostKeyCallback = kh
	}

	config := &ssh.ClientConfig{
		User:            p.User,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         timeout,
	}

	addr := net.JoinHostPort(p.Host, strconv.Itoa(p.Port))
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("ssh dial: %w", err)
	}

	return &sshConn{client: client}, nil
}

func (c *sshConn) sftpClient() (*sftp.Client, error) {
	if c.sftp != nil {
		return c.sftp, nil
	}
	s, err := sftp.NewClient(c.client)
	if err != nil {
		return nil, fmt.Errorf("create sftp client: %w", err)
	}
	c.sftp = s
	return s, nil
}

func (c *sshConn) close() {
	if c.sftp != nil {
		c.sftp.Close()
	}
	c.client.Close()
}

func (h *SSHHandler) handleCmd(conn *sshConn, p *SSHPayload, source string, cmdTimeout time.Duration, start time.Time) *output.Response {
	cmds, err := readCommands(p.Cmd, p.CmdFile)
	if err != nil {
		return output.Fail("ssh", source, "SSH_CMDFILE_ERROR", err.Error(), "", start)
	}
	if len(cmds) == 0 {
		return output.Fail("ssh", source, "SSH_NO_COMMAND", "no command specified", "", start)
	}

	results := make([]SSHCommandResult, 0, len(cmds))
	for _, cmd := range cmds {
		r, err := runRemoteCommand(conn.client, cmd, cmdTimeout)
		if err != nil {
			return output.Fail("ssh", source, "SSH_CMD_ERROR",
				fmt.Sprintf("command %q failed: %v", cmd, err), "", start)
		}
		results = append(results, *r)
	}

	return output.Success("ssh", source, map[string]interface{}{
		"action":  "cmd",
		"results": results,
	}, start)
}

func (h *SSHHandler) handleUpload(conn *sshConn, p *SSHPayload, source string, start time.Time) *output.Response {
	sftpClient, err := conn.sftpClient()
	if err != nil {
		return output.Fail("ssh", source, "SSH_SFTP_ERROR", err.Error(), "", start)
	}

	localPath := expandHome(p.LocalPath)
	remotePath := resolveRemoteUploadPath(localPath, p.RemotePath, p.FileName)

	n, err := uploadFile(sftpClient, localPath, remotePath)
	if err != nil {
		return output.Fail("ssh", source, "SSH_UPLOAD_ERROR", err.Error(), "", start)
	}

	return output.Success("ssh", source, map[string]interface{}{
		"action":      "upload",
		"local_path":  localPath,
		"remote_path": remotePath,
		"bytes":       n,
	}, start)
}

func (h *SSHHandler) handleDownload(conn *sshConn, p *SSHPayload, source string, start time.Time) *output.Response {
	sftpClient, err := conn.sftpClient()
	if err != nil {
		return output.Fail("ssh", source, "SSH_SFTP_ERROR", err.Error(), "", start)
	}

	localPath := expandHome(p.LocalPath)
	remotePath := p.RemotePath
	resolvedLocal := resolveLocalDownloadPath(localPath, remotePath, p.FileName)

	n, err := downloadFile(sftpClient, remotePath, resolvedLocal)
	if err != nil {
		return output.Fail("ssh", source, "SSH_DOWNLOAD_ERROR", err.Error(), "", start)
	}

	return output.Success("ssh", source, map[string]interface{}{
		"action":      "download",
		"remote_path": remotePath,
		"local_path":  resolvedLocal,
		"bytes":       n,
	}, start)
}

func (h *SSHHandler) handleUploadAtomic(conn *sshConn, p *SSHPayload, source string, start time.Time) *output.Response {
	sftpClient, err := conn.sftpClient()
	if err != nil {
		return output.Fail("ssh", source, "SSH_SFTP_ERROR", err.Error(), "", start)
	}

	localPath := expandHome(p.LocalPath)
	remotePath := p.RemotePath
	tempPath := p.TempPath
	if tempPath == "" {
		tempPath = remotePath + ".tmp"
	}

	n, err := uploadFile(sftpClient, localPath, tempPath)
	if err != nil {
		return output.Fail("ssh", source, "SSH_UPLOAD_ERROR", err.Error(), "", start)
	}

	if err := moveRemotePath(sftpClient, tempPath, remotePath); err != nil {
		return output.Fail("ssh", source, "SSH_ATOMIC_MOVE_ERROR", err.Error(), "", start)
	}

	return output.Success("ssh", source, map[string]interface{}{
		"action":      "upload_atomic",
		"local_path":  localPath,
		"remote_path": remotePath,
		"bytes":       n,
	}, start)
}

func (h *SSHHandler) handleMkdir(conn *sshConn, p *SSHPayload, source string, start time.Time) *output.Response {
	sftpClient, err := conn.sftpClient()
	if err != nil {
		return output.Fail("ssh", source, "SSH_SFTP_ERROR", err.Error(), "", start)
	}

	if p.RemotePath == "" {
		return output.Fail("ssh", source, "SSH_MISSING_PARAMS",
			"remote_path is required for mkdir", "", start)
	}

	if err := sftpClient.MkdirAll(p.RemotePath); err != nil {
		return output.Fail("ssh", source, "SSH_MKDIR_ERROR", err.Error(), "", start)
	}

	return output.Success("ssh", source, map[string]interface{}{
		"action":      "mkdir",
		"remote_path": p.RemotePath,
	}, start)
}

func (h *SSHHandler) handleRemove(conn *sshConn, p *SSHPayload, source string, start time.Time) *output.Response {
	sftpClient, err := conn.sftpClient()
	if err != nil {
		return output.Fail("ssh", source, "SSH_SFTP_ERROR", err.Error(), "", start)
	}

	if p.RemotePath == "" {
		return output.Fail("ssh", source, "SSH_MISSING_PARAMS",
			"remote_path is required for remove", "", start)
	}

	if err := removeRemotePath(sftpClient, p.RemotePath); err != nil {
		return output.Fail("ssh", source, "SSH_REMOVE_ERROR", err.Error(), "", start)
	}

	return output.Success("ssh", source, map[string]interface{}{
		"action":      "remove",
		"remote_path": p.RemotePath,
	}, start)
}

func (h *SSHHandler) handleChmod(conn *sshConn, p *SSHPayload, source string, start time.Time) *output.Response {
	sftpClient, err := conn.sftpClient()
	if err != nil {
		return output.Fail("ssh", source, "SSH_SFTP_ERROR", err.Error(), "", start)
	}

	mode := p.Mode
	if mode == "" {
		mode = p.FileName
	}
	if p.RemotePath == "" || mode == "" {
		return output.Fail("ssh", source, "SSH_MISSING_PARAMS",
			"remote_path and mode are required for chmod", "", start)
	}

	parsed, err := strconv.ParseUint(mode, 8, 32)
	if err != nil {
		return output.Fail("ssh", source, "SSH_CHMOD_ERROR",
			fmt.Sprintf("invalid mode %q: %v", mode, err), "", start)
	}

	if err := sftpClient.Chmod(p.RemotePath, os.FileMode(parsed)); err != nil {
		return output.Fail("ssh", source, "SSH_CHMOD_ERROR", err.Error(), "", start)
	}

	return output.Success("ssh", source, map[string]interface{}{
		"action":      "chmod",
		"remote_path": p.RemotePath,
		"mode":        mode,
	}, start)
}

func (h *SSHHandler) handleMove(conn *sshConn, p *SSHPayload, source string, start time.Time) *output.Response {
	sftpClient, err := conn.sftpClient()
	if err != nil {
		return output.Fail("ssh", source, "SSH_SFTP_ERROR", err.Error(), "", start)
	}

	targetPath := p.TargetPath
	if targetPath == "" {
		targetPath = p.LocalPath
	}
	if p.RemotePath == "" || targetPath == "" {
		return output.Fail("ssh", source, "SSH_MISSING_PARAMS",
			"remote_path and target_path are required for move", "", start)
	}

	if err := moveRemotePath(sftpClient, p.RemotePath, targetPath); err != nil {
		return output.Fail("ssh", source, "SSH_MOVE_ERROR", err.Error(), "", start)
	}

	return output.Success("ssh", source, map[string]interface{}{
		"action":      "move",
		"remote_path": p.RemotePath,
		"target_path": targetPath,
	}, start)
}

type deployStep struct {
	Name            string   `json:"name"`
	Type            string   `json:"type"`
	Cmd             string   `json:"cmd,omitempty"`
	CmdFile         string   `json:"cmdfile,omitempty"`
	LocalPath       string   `json:"local_path,omitempty"`
	RemotePath      string   `json:"remote_path,omitempty"`
	FileName        string   `json:"file_name,omitempty"`
	TargetPath      string   `json:"target_path,omitempty"`
	TempPath        string   `json:"temp_path,omitempty"`
	Mode            string   `json:"mode,omitempty"`
	Timeout         string   `json:"timeout,omitempty"`
	ContinueOnError bool     `json:"continue_on_error,omitempty"`
	Direction       string   `json:"direction,omitempty"`
	Recursive       bool     `json:"recursive,omitempty"`
	Delete          bool     `json:"delete,omitempty"`
	DryRun          bool     `json:"dry_run,omitempty"`
	Conflict        string   `json:"conflict,omitempty"`
	Include         []string `json:"include,omitempty"`
	Exclude         []string `json:"exclude,omitempty"`
}

type deployPlan struct {
	Steps []deployStep `json:"steps"`
}

func (h *SSHHandler) handleDeploy(conn *sshConn, p *SSHPayload, source string, cmdTimeout time.Duration, start time.Time) *output.Response {
	var planData []byte
	var err error

	if p.PlanJSON != "" {
		planData = []byte(p.PlanJSON)
	} else if p.Plan != "" {
		planData, err = os.ReadFile(expandHome(p.Plan))
		if err != nil {
			return output.Fail("ssh", source, "SSH_DEPLOY_PLAN_ERROR",
				fmt.Sprintf("read plan: %v", err), "", start)
		}
	} else {
		return output.Fail("ssh", source, "SSH_DEPLOY_NO_PLAN",
			"plan or plan_json is required", "", start)
	}

	var plan deployPlan
	if err := json.Unmarshal(planData, &plan); err != nil {
		return output.Fail("ssh", source, "SSH_DEPLOY_PARSE_ERROR",
			fmt.Sprintf("parse plan: %v", err), "", start)
	}

	if len(plan.Steps) == 0 {
		return output.Fail("ssh", source, "SSH_DEPLOY_NO_STEPS",
			"deploy plan must contain at least one step", "", start)
	}

	sftpClient, err := conn.sftpClient()
	if err != nil {
		return output.Fail("ssh", source, "SSH_SFTP_ERROR", err.Error(), "", start)
	}

	type stepResult struct {
		Step    int         `json:"step"`
		Name    string      `json:"name"`
		Type    string      `json:"type"`
		Success bool        `json:"success"`
		Data    interface{} `json:"data,omitempty"`
		Error   string      `json:"error,omitempty"`
	}

	results := make([]stepResult, 0, len(plan.Steps))
	for i, step := range plan.Steps {
		stepTimeout := cmdTimeout
		if step.Timeout != "" {
			if d, e := time.ParseDuration(step.Timeout); e == nil {
				stepTimeout = d
			}
		}

		var sr stepResult
		sr.Step = i + 1
		sr.Name = step.Name
		sr.Type = step.Type

		switch strings.ToLower(step.Type) {
		case "cmd":
			cmds, e := readCommands(step.Cmd, step.CmdFile)
			if e != nil {
				sr.Error = e.Error()
				break
			}
			if len(cmds) == 0 {
				sr.Error = "cmd step requires cmd or cmdfile"
				break
			}
			var cmdResults []SSHCommandResult
			allOk := true
			for _, c := range cmds {
				r, e := runRemoteCommand(conn.client, c, stepTimeout)
				if e != nil {
					sr.Error = e.Error()
					allOk = false
					break
				}
				if r.ExitCode != 0 {
					sr.Error = fmt.Sprintf("command %q exited with code %d", c, r.ExitCode)
					allOk = false
					cmdResults = append(cmdResults, *r)
					break
				}
				cmdResults = append(cmdResults, *r)
			}
			if allOk {
				sr.Success = true
				sr.Data = cmdResults
			}

		case "upload":
			lp := expandHome(step.LocalPath)
			rp := resolveRemoteUploadPath(lp, step.RemotePath, step.FileName)
			n, e := uploadFile(sftpClient, lp, rp)
			if e != nil {
				sr.Error = e.Error()
			} else {
				sr.Success = true
				sr.Data = map[string]interface{}{"bytes": n, "remote_path": rp}
			}

		case "download":
			lp := expandHome(step.LocalPath)
			rp := step.RemotePath
			resolvedLocal := resolveLocalDownloadPath(lp, rp, step.FileName)
			n, e := downloadFile(sftpClient, rp, resolvedLocal)
			if e != nil {
				sr.Error = e.Error()
			} else {
				sr.Success = true
				sr.Data = map[string]interface{}{"bytes": n, "local_path": resolvedLocal}
			}

		case "upload_atomic":
			tmp := step.TempPath
			if tmp == "" {
				tmp = step.RemotePath + ".tmp"
			}
			lp := expandHome(step.LocalPath)
			n, e := uploadFile(sftpClient, lp, tmp)
			if e != nil {
				sr.Error = e.Error()
				break
			}
			if e := moveRemotePath(sftpClient, tmp, step.RemotePath); e != nil {
				sr.Error = e.Error()
				break
			}
			sr.Success = true
			sr.Data = map[string]interface{}{"bytes": n}

		case "mkdir":
			if e := sftpClient.MkdirAll(step.RemotePath); e != nil {
				sr.Error = e.Error()
			} else {
				sr.Success = true
			}

		case "remove":
			if e := removeRemotePath(sftpClient, step.RemotePath); e != nil {
				sr.Error = e.Error()
			} else {
				sr.Success = true
			}

		case "chmod":
			mode := step.Mode
			if mode == "" {
				mode = step.FileName
			}
			if step.RemotePath == "" || mode == "" {
				sr.Error = "remote_path and mode required for chmod"
				break
			}
			parsed, e := strconv.ParseUint(mode, 8, 32)
			if e != nil {
				sr.Error = e.Error()
				break
			}
			if e := sftpClient.Chmod(step.RemotePath, os.FileMode(parsed)); e != nil {
				sr.Error = e.Error()
			} else {
				sr.Success = true
			}

		case "move":
			tp := step.TargetPath
			if tp == "" {
				tp = step.LocalPath
			}
			if e := moveRemotePath(sftpClient, step.RemotePath, tp); e != nil {
				sr.Error = e.Error()
			} else {
				sr.Success = true
			}

		case "sync":
			direction := step.Direction
			if direction == "" {
				direction = "push"
			}
			conflict := step.Conflict
			if conflict == "" {
				conflict = "fail_on_conflict"
			}
			syncRes, e := runSync(sftpClient, expandHome(step.LocalPath), step.RemotePath,
				direction, step.Recursive, step.Delete, step.DryRun, conflict, step.Include, step.Exclude)
			if e != nil {
				sr.Error = e.Error()
			} else {
				sr.Success = true
				sr.Data = syncRes
			}

		default:
			sr.Error = fmt.Sprintf("unsupported step type: %s", step.Type)
		}

		results = append(results, sr)

		if !sr.Success && !step.ContinueOnError {
			return output.Success("ssh", source, map[string]interface{}{
				"action":    "deploy",
				"completed": i,
				"total":     len(plan.Steps),
				"failed_at": sr,
				"results":   results,
			}, start)
		}
	}

	return output.Success("ssh", source, map[string]interface{}{
		"action":  "deploy",
		"total":   len(plan.Steps),
		"results": results,
	}, start)
}

func (h *SSHHandler) handleSync(conn *sshConn, p *SSHPayload, source string, start time.Time) *output.Response {
	sftpClient, err := conn.sftpClient()
	if err != nil {
		return output.Fail("ssh", source, "SSH_SFTP_ERROR", err.Error(), "", start)
	}

	localPath := expandHome(p.LocalPath)
	result, err := runSync(sftpClient, localPath, p.RemotePath,
		p.Direction, p.Recursive, p.Delete, p.DryRun, p.Conflict, p.Include, p.Exclude)
	if err != nil {
		return output.Fail("ssh", source, "SSH_SYNC_ERROR", err.Error(), "", start)
	}

	return output.Success("ssh", source, map[string]interface{}{
		"action":    "sync",
		"direction": p.Direction,
		"result":    result,
	}, start)
}

func runSync(sftpClient *sftp.Client, localPath, remotePath, direction string, recursive, deleteExtra, dryRun bool, conflict string, includePatterns, excludePatterns []string) (map[string]interface{}, error) {
	direction = strings.ToLower(strings.TrimSpace(direction))
	conflict = strings.ToLower(strings.TrimSpace(conflict))
	if direction == "" {
		direction = "push"
	}
	if conflict == "" {
		conflict = "fail_on_conflict"
	}

	localInfo, localExists, err := statLocalPath(localPath)
	if err != nil {
		return nil, err
	}
	remoteInfo, remoteExists, err := statRemotePath(sftpClient, remotePath)
	if err != nil {
		return nil, err
	}

	switch direction {
	case "push":
		if !localExists {
			return nil, fmt.Errorf("local path does not exist: %s", localPath)
		}
		if localInfo.IsDir() {
			return syncDirectoryPush(sftpClient, localPath, remotePath, recursive, deleteExtra, dryRun, includePatterns, excludePatterns)
		}
		return syncFilePush(sftpClient, localPath, remotePath, remoteInfo, remoteExists, dryRun)

	case "pull":
		if !remoteExists {
			return nil, fmt.Errorf("remote path does not exist: %s", remotePath)
		}
		if remoteInfo.IsDir() {
			return syncDirectoryPull(sftpClient, localPath, remotePath, recursive, deleteExtra, dryRun, includePatterns, excludePatterns)
		}
		return syncFilePull(sftpClient, localPath, remotePath, localInfo, localExists, dryRun)

	case "bidirectional":
		if deleteExtra {
			return nil, fmt.Errorf("delete is not supported for bidirectional sync")
		}
		kind, err := detectBidirectionalKind(localInfo, localExists, remoteInfo, remoteExists)
		if err != nil {
			return nil, err
		}
		if kind == "dir" {
			return syncDirectoryBidirectional(sftpClient, localPath, remotePath, recursive, dryRun, conflict, includePatterns, excludePatterns)
		}
		return syncFileBidirectional(sftpClient, localPath, remotePath, localInfo, localExists, remoteInfo, remoteExists, dryRun, conflict)

	default:
		return nil, fmt.Errorf("invalid sync direction: %s (use push, pull, bidirectional)", direction)
	}
}

func readCommands(command, commandFile string) ([]string, error) {
	if commandFile != "" {
		data, err := os.ReadFile(expandHome(commandFile))
		if err != nil {
			return nil, fmt.Errorf("read command file: %w", err)
		}
		var cmds []string
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				cmds = append(cmds, line)
			}
		}
		return cmds, nil
	}
	if command == "" {
		return nil, nil
	}
	return []string{command}, nil
}

func runRemoteCommand(client *ssh.Client, cmd string, timeout time.Duration) (*SSHCommandResult, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if timeout <= 0 {
		if err := session.Run(cmd); err != nil {
			output := stdout.String() + stderr.String()
			if exitErr, ok := err.(*ssh.ExitError); ok {
				return &SSHCommandResult{Output: output, ExitCode: exitErr.ExitStatus()}, nil
			}
			return nil, fmt.Errorf("run command: %w", err)
		}
		return &SSHCommandResult{Output: stdout.String() + stderr.String(), ExitCode: 0}, nil
	}

	if err := session.Start(cmd); err != nil {
		return nil, fmt.Errorf("start command: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- session.Wait()
	}()

	select {
	case err := <-done:
		output := stdout.String() + stderr.String()
		if err != nil {
			if exitErr, ok := err.(*ssh.ExitError); ok {
				return &SSHCommandResult{Output: output, ExitCode: exitErr.ExitStatus()}, nil
			}
			return nil, fmt.Errorf("run command: %w", err)
		}
		return &SSHCommandResult{Output: output, ExitCode: 0}, nil
	case <-time.After(timeout):
		output := stdout.String() + stderr.String()
		session.Close()
		return &SSHCommandResult{Output: output, ExitCode: -1, TimedOut: true}, nil
	}
}

func uploadFile(sftpClient *sftp.Client, localPath, remotePath string) (int64, error) {
	if localPath == "" || remotePath == "" {
		return 0, fmt.Errorf("local_path and remote_path are required for upload")
	}

	if err := sftpClient.MkdirAll(path.Dir(remotePath)); err != nil {
		return 0, fmt.Errorf("create remote parent dir: %w", err)
	}

	localFile, err := os.Open(localPath)
	if err != nil {
		return 0, fmt.Errorf("open local file: %w", err)
	}
	defer localFile.Close()

	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return 0, fmt.Errorf("create remote file: %w", err)
	}
	defer remoteFile.Close()

	n, err := io.Copy(remoteFile, localFile)
	if err != nil {
		return n, fmt.Errorf("write remote file: %w", err)
	}
	return n, nil
}

func uploadFilePreserveTime(sftpClient *sftp.Client, localPath, remotePath string) (int64, error) {
	n, err := uploadFile(sftpClient, localPath, remotePath)
	if err != nil {
		return n, err
	}
	if info, statErr := os.Stat(localPath); statErr == nil {
		_ = sftpClient.Chtimes(remotePath, info.ModTime(), info.ModTime())
	}
	return n, nil
}

func downloadFile(sftpClient *sftp.Client, remotePath, localPath string) (int64, error) {
	if localPath == "" || remotePath == "" {
		return 0, fmt.Errorf("local_path and remote_path are required for download")
	}

	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return 0, fmt.Errorf("create local directory: %w", err)
	}

	remoteFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return 0, fmt.Errorf("open remote file: %w", err)
	}
	defer remoteFile.Close()

	localFile, err := os.Create(localPath)
	if err != nil {
		return 0, fmt.Errorf("create local file: %w", err)
	}
	defer localFile.Close()

	n, err := io.Copy(localFile, remoteFile)
	if err != nil {
		return n, fmt.Errorf("read remote file: %w", err)
	}
	return n, nil
}

func downloadFilePreserveTime(sftpClient *sftp.Client, localPath, remotePath string) (int64, error) {
	n, err := downloadFile(sftpClient, remotePath, localPath)
	if err != nil {
		return n, err
	}
	if info, statErr := sftpClient.Stat(remotePath); statErr == nil {
		_ = os.Chtimes(localPath, info.ModTime(), info.ModTime())
	}
	return n, nil
}

func resolveRemoteUploadPath(localPath, remotePath, fileName string) string {
	destFileName := fileName
	if destFileName == "" {
		destFileName = filepath.Base(localPath)
	}
	if strings.HasSuffix(remotePath, "/") {
		return remotePath + destFileName
	}
	return remotePath
}

func resolveLocalDownloadPath(localPath, remotePath, fileName string) string {
	destFileName := fileName
	if destFileName == "" {
		destFileName = path.Base(remotePath)
	}
	info, err := os.Stat(localPath)
	if err == nil && info.IsDir() {
		return filepath.Join(localPath, destFileName)
	}
	if strings.HasSuffix(localPath, string(os.PathSeparator)) || strings.HasSuffix(localPath, "/") || strings.HasSuffix(localPath, "\\") {
		return filepath.Join(localPath, destFileName)
	}
	return localPath
}

func moveRemotePath(sftpClient *sftp.Client, src, dst string) error {
	if err := sftpClient.MkdirAll(path.Dir(dst)); err != nil {
		return fmt.Errorf("create target parent dir: %w", err)
	}
	_ = sftpClient.Remove(dst)
	if err := sftpClient.PosixRename(src, dst); err != nil {
		if err := sftpClient.Rename(src, dst); err != nil {
			return fmt.Errorf("move %s -> %s: %w", src, dst, err)
		}
	}
	return nil
}

func removeRemotePath(sftpClient *sftp.Client, remotePath string) error {
	info, err := sftpClient.Stat(remotePath)
	if err != nil {
		if isNotExistError(err) {
			return nil
		}
		return fmt.Errorf("stat remote path %s: %w", remotePath, err)
	}
	if !info.IsDir() {
		if err := sftpClient.Remove(remotePath); err != nil {
			return fmt.Errorf("remove remote file %s: %w", remotePath, err)
		}
		return nil
	}
	return removeRemoteTree(sftpClient, remotePath)
}

func removeRemoteTree(sftpClient *sftp.Client, dir string) error {
	entries, err := sftpClient.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read remote dir %s: %w", dir, err)
	}
	for _, entry := range entries {
		child := path.Join(dir, entry.Name())
		if entry.IsDir() {
			if err := removeRemoteTree(sftpClient, child); err != nil {
				return err
			}
			continue
		}
		if err := sftpClient.Remove(child); err != nil {
			return fmt.Errorf("remove remote file %s: %w", child, err)
		}
	}
	if err := sftpClient.RemoveDirectory(dir); err != nil {
		return fmt.Errorf("remove remote directory %s: %w", dir, err)
	}
	return nil
}

func isNotExistError(err error) bool {
	if err == nil {
		return false
	}
	if os.IsNotExist(err) {
		return true
	}
	if pathErr, ok := err.(*os.PathError); ok {
		if pathErr.Err != nil && strings.Contains(pathErr.Err.Error(), "not exist") {
			return true
		}
	}
	return strings.Contains(err.Error(), "does not exist") ||
		strings.Contains(err.Error(), "not found") ||
		strings.Contains(err.Error(), "no such file")
}

func statLocalPath(localPath string) (os.FileInfo, bool, error) {
	info, err := os.Stat(localPath)
	if err == nil {
		return info, true, nil
	}
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	return nil, false, fmt.Errorf("stat local path %s: %w", localPath, err)
}

func statRemotePath(sftpClient *sftp.Client, remotePath string) (os.FileInfo, bool, error) {
	info, err := sftpClient.Stat(remotePath)
	if err == nil {
		return info, true, nil
	}
	if isNotExistError(err) {
		return nil, false, nil
	}
	return nil, false, fmt.Errorf("stat remote path %s: %w", remotePath, err)
}

type fileState struct {
	Size    int64
	ModTime time.Time
}

type syncTree struct {
	Exists    bool
	Root      string
	RootIsDir bool
	Files     map[string]fileState
	Dirs      map[string]struct{}
}

func collectLocalTree(root string, recursive bool, includePatterns, excludePatterns []string) (*syncTree, error) {
	tree, err := collectLocalTreeAllowMissing(root, recursive, includePatterns, excludePatterns)
	if err != nil {
		return nil, err
	}
	if !tree.Exists {
		return nil, fmt.Errorf("local path does not exist: %s", root)
	}
	return tree, nil
}

func collectLocalTreeAllowMissing(root string, recursive bool, includePatterns, excludePatterns []string) (*syncTree, error) {
	info, exists, err := statLocalPath(root)
	if err != nil {
		return nil, err
	}
	tree := &syncTree{Exists: exists, Root: root, Files: map[string]fileState{}, Dirs: map[string]struct{}{}}
	if !exists {
		return tree, nil
	}
	tree.RootIsDir = info.IsDir()
	if !info.IsDir() {
		tree.Files[""] = fileState{Size: info.Size(), ModTime: info.ModTime()}
		return tree, nil
	}
	if !recursive {
		return nil, fmt.Errorf("directory sync requires recursive=true: %s", root)
	}
	err = filepath.Walk(root, func(current string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == root {
			return nil
		}
		rel, e := filepath.Rel(root, current)
		if e != nil {
			return e
		}
		rel = filepath.ToSlash(rel)
		if !shouldSyncPath(rel, fi.IsDir(), includePatterns, excludePatterns) {
			if fi.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if fi.IsDir() {
			tree.Dirs[rel] = struct{}{}
			return nil
		}
		tree.Files[rel] = fileState{Size: fi.Size(), ModTime: fi.ModTime()}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan local tree %s: %w", root, err)
	}
	addParentDirs(tree)
	return tree, nil
}

func collectRemoteTree(sftpClient *sftp.Client, root string, recursive, allowMissing bool, includePatterns, excludePatterns []string) (*syncTree, error) {
	info, exists, err := statRemotePath(sftpClient, root)
	if err != nil {
		return nil, err
	}
	tree := &syncTree{Exists: exists, Root: root, Files: map[string]fileState{}, Dirs: map[string]struct{}{}}
	if !exists {
		if allowMissing {
			return tree, nil
		}
		return nil, fmt.Errorf("remote path does not exist: %s", root)
	}
	tree.RootIsDir = info.IsDir()
	if !info.IsDir() {
		tree.Files[""] = fileState{Size: info.Size(), ModTime: info.ModTime()}
		return tree, nil
	}
	if !recursive {
		return nil, fmt.Errorf("directory sync requires recursive=true: %s", root)
	}
	if err := walkRemoteTree(sftpClient, root, "", tree, includePatterns, excludePatterns); err != nil {
		return nil, err
	}
	addParentDirs(tree)
	return tree, nil
}

func walkRemoteTree(sftpClient *sftp.Client, root, rel string, tree *syncTree, includePatterns, excludePatterns []string) error {
	remotePath := root
	if rel != "" {
		remotePath = path.Join(root, rel)
	}
	entries, err := sftpClient.ReadDir(remotePath)
	if err != nil {
		return fmt.Errorf("read remote directory %s: %w", remotePath, err)
	}
	for _, entry := range entries {
		entryRel := entry.Name()
		if rel != "" {
			entryRel = rel + "/" + entry.Name()
		}
		if !shouldSyncPath(entryRel, entry.IsDir(), includePatterns, excludePatterns) {
			continue
		}
		if entry.IsDir() {
			tree.Dirs[entryRel] = struct{}{}
			if err := walkRemoteTree(sftpClient, root, entryRel, tree, includePatterns, excludePatterns); err != nil {
				return err
			}
			continue
		}
		tree.Files[entryRel] = fileState{Size: entry.Size(), ModTime: entry.ModTime()}
	}
	return nil
}

func addParentDirs(tree *syncTree) {
	for rel := range tree.Files {
		parent := path.Dir(rel)
		for parent != "." && parent != "/" && parent != "" {
			tree.Dirs[parent] = struct{}{}
			parent = path.Dir(parent)
		}
	}
}

func sameFileState(a, b fileState) bool {
	return a.Size == b.Size && modTimesClose(a.ModTime, b.ModTime)
}

func sameFileInfo(a, b os.FileInfo) bool {
	if a == nil || b == nil {
		return false
	}
	return sameFileState(fileState{Size: a.Size(), ModTime: a.ModTime()}, fileState{Size: b.Size(), ModTime: b.ModTime()})
}

func modTimesClose(a, b time.Time) bool {
	diff := a.Sub(b)
	if diff < 0 {
		diff = -diff
	}
	return diff <= time.Second
}

func sortedDirs(dirs map[string]struct{}) []string {
	result := make([]string, 0, len(dirs))
	for rel := range dirs {
		result = append(result, rel)
	}
	sort.Strings(result)
	return result
}

func reverseSortedDirs(dirs map[string]struct{}) []string {
	result := sortedDirs(dirs)
	sort.Slice(result, func(i, j int) bool {
		if len(result[i]) == len(result[j]) {
			return result[i] > result[j]
		}
		return len(result[i]) > len(result[j])
	})
	return result
}

func joinLocalRoot(root, rel string) string {
	if rel == "" {
		return root
	}
	return filepath.Join(root, filepath.FromSlash(rel))
}

func joinRemoteRoot(root, rel string) string {
	if rel == "" {
		return root
	}
	return path.Join(root, rel)
}

func ensureRemoteRootDir(sftpClient *sftp.Client, remotePath string, exists, isDir bool) error {
	if exists {
		if !isDir {
			return fmt.Errorf("remote path is not a directory: %s", remotePath)
		}
		return nil
	}
	if err := sftpClient.MkdirAll(remotePath); err != nil {
		return fmt.Errorf("create remote root dir %s: %w", remotePath, err)
	}
	return nil
}

func ensureLocalRootDir(localPath string, exists, isDir bool) error {
	if exists {
		if !isDir {
			return fmt.Errorf("local path is not a directory: %s", localPath)
		}
		return nil
	}
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return fmt.Errorf("create local root dir %s: %w", localPath, err)
	}
	return nil
}

func detectBidirectionalKind(localInfo os.FileInfo, localExists bool, remoteInfo os.FileInfo, remoteExists bool) (string, error) {
	if localExists && remoteExists {
		if localInfo.IsDir() != remoteInfo.IsDir() {
			return "", fmt.Errorf("local and remote path types differ")
		}
		if localInfo.IsDir() {
			return "dir", nil
		}
		return "file", nil
	}
	if localExists {
		if localInfo.IsDir() {
			return "dir", nil
		}
		return "file", nil
	}
	if remoteExists {
		if remoteInfo.IsDir() {
			return "dir", nil
		}
		return "file", nil
	}
	return "", fmt.Errorf("both local and remote paths are missing")
}

func resolveBidirectionalAction(localState, remoteState fileState, conflict string) (string, error) {
	switch conflict {
	case "local_wins":
		return "upload", nil
	case "remote_wins":
		return "download", nil
	case "newer_wins":
		if localState.ModTime.After(remoteState.ModTime) {
			return "upload", nil
		}
		if remoteState.ModTime.After(localState.ModTime) {
			return "download", nil
		}
		return "", fmt.Errorf("same modification time but different content")
	case "fail_on_conflict":
		return "", fmt.Errorf("local and remote versions differ")
	default:
		return "", fmt.Errorf("invalid conflict policy: %s", conflict)
	}
}

func resolveRemoteSyncPath(localPath, remotePath string, remoteInfo os.FileInfo, remoteExists bool) string {
	if strings.HasSuffix(remotePath, "/") || (remoteExists && remoteInfo != nil && remoteInfo.IsDir()) {
		return path.Join(remotePath, filepath.Base(localPath))
	}
	return remotePath
}

func resolveLocalSyncPath(localPath, remotePath string, localInfo os.FileInfo, localExists bool) string {
	if strings.HasSuffix(localPath, string(os.PathSeparator)) || strings.HasSuffix(localPath, "/") || strings.HasSuffix(localPath, "\\") || (localExists && localInfo != nil && localInfo.IsDir()) {
		return filepath.Join(localPath, filepath.Base(remotePath))
	}
	return localPath
}

func syncFilePush(sftpClient *sftp.Client, localPath, remotePath string, remoteInfo os.FileInfo, remoteExists, dryRun bool) (map[string]interface{}, error) {
	resolvedRemotePath := resolveRemoteSyncPath(localPath, remotePath, remoteInfo, remoteExists)
	localInfo, err := os.Stat(localPath)
	if err != nil {
		return nil, fmt.Errorf("stat local file %s: %w", localPath, err)
	}
	resolvedRemoteInfo, resolvedRemoteExists, err := statRemotePath(sftpClient, resolvedRemotePath)
	if err != nil {
		return nil, err
	}
	if resolvedRemoteExists && sameFileInfo(localInfo, resolvedRemoteInfo) {
		return map[string]interface{}{
			"uploaded":  0,
			"skipped":   true,
			"local_path":  localPath,
			"remote_path": resolvedRemotePath,
		}, nil
	}
	if dryRun {
		return map[string]interface{}{
			"uploaded":  1,
			"dry_run":   true,
			"local_path":  localPath,
			"remote_path": resolvedRemotePath,
		}, nil
	}
	n, err := uploadFilePreserveTime(sftpClient, localPath, resolvedRemotePath)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"uploaded":    1,
		"bytes":       n,
		"local_path":  localPath,
		"remote_path": resolvedRemotePath,
	}, nil
}

func syncFilePull(sftpClient *sftp.Client, localPath, remotePath string, localInfo os.FileInfo, localExists, dryRun bool) (map[string]interface{}, error) {
	resolvedLocalPath := resolveLocalSyncPath(localPath, remotePath, localInfo, localExists)
	resolvedLocalInfo, resolvedLocalExists, err := statLocalPath(resolvedLocalPath)
	if err != nil {
		return nil, err
	}
	remoteInfo, remoteExists, err := statRemotePath(sftpClient, remotePath)
	if err != nil {
		return nil, err
	}
	if !remoteExists {
		return nil, fmt.Errorf("remote file does not exist: %s", remotePath)
	}
	if resolvedLocalExists && sameFileInfo(resolvedLocalInfo, remoteInfo) {
		return map[string]interface{}{
			"downloaded": 0,
			"skipped":    true,
			"local_path":  resolvedLocalPath,
			"remote_path": remotePath,
		}, nil
	}
	if dryRun {
		return map[string]interface{}{
			"downloaded": 1,
			"dry_run":    true,
			"local_path":  resolvedLocalPath,
			"remote_path": remotePath,
		}, nil
	}
	n, err := downloadFilePreserveTime(sftpClient, resolvedLocalPath, remotePath)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"downloaded":  1,
		"bytes":       n,
		"local_path":  resolvedLocalPath,
		"remote_path": remotePath,
	}, nil
}

func syncFileBidirectional(sftpClient *sftp.Client, localPath, remotePath string, localInfo os.FileInfo, localExists bool, remoteInfo os.FileInfo, remoteExists, dryRun bool, conflict string) (map[string]interface{}, error) {
	if !localExists && !remoteExists {
		return nil, fmt.Errorf("both files are missing: %s and %s", localPath, remotePath)
	}
	if localExists && !remoteExists {
		return syncFilePush(sftpClient, localPath, remotePath, nil, false, dryRun)
	}
	if !localExists && remoteExists {
		return syncFilePull(sftpClient, localPath, remotePath, nil, false, dryRun)
	}
	if sameFileInfo(localInfo, remoteInfo) {
		return map[string]interface{}{
			"skipped":     true,
			"local_path":  localPath,
			"remote_path": remotePath,
		}, nil
	}
	localState := fileState{Size: localInfo.Size(), ModTime: localInfo.ModTime()}
	remoteState := fileState{Size: remoteInfo.Size(), ModTime: remoteInfo.ModTime()}
	action, err := resolveBidirectionalAction(localState, remoteState, conflict)
	if err != nil {
		return map[string]interface{}{
			"conflict":    true,
			"error":       err.Error(),
			"local_path":  localPath,
			"remote_path": remotePath,
		}, nil
	}
	if action == "upload" {
		return syncFilePush(sftpClient, localPath, remotePath, remoteInfo, true, dryRun)
	}
	return syncFilePull(sftpClient, localPath, remotePath, localInfo, true, dryRun)
}

func syncDirectoryPush(sftpClient *sftp.Client, localPath, remotePath string, recursive, deleteExtra, dryRun bool, includePatterns, excludePatterns []string) (map[string]interface{}, error) {
	source, err := collectLocalTree(localPath, recursive, includePatterns, excludePatterns)
	if err != nil {
		return nil, err
	}
	target, err := collectRemoteTree(sftpClient, remotePath, recursive, true, includePatterns, excludePatterns)
	if err != nil {
		return nil, err
	}

	if !dryRun {
		if err := ensureRemoteRootDir(sftpClient, remotePath, target.Exists, target.RootIsDir); err != nil {
			return nil, err
		}
	}

	return syncTreePush(sftpClient, source, target, remotePath, deleteExtra, dryRun)
}

func syncDirectoryPull(sftpClient *sftp.Client, localPath, remotePath string, recursive, deleteExtra, dryRun bool, includePatterns, excludePatterns []string) (map[string]interface{}, error) {
	source, err := collectRemoteTree(sftpClient, remotePath, recursive, false, includePatterns, excludePatterns)
	if err != nil {
		return nil, err
	}
	target, err := collectLocalTreeAllowMissing(localPath, recursive, includePatterns, excludePatterns)
	if err != nil {
		return nil, err
	}

	if !dryRun {
		if err := ensureLocalRootDir(localPath, target.Exists, target.RootIsDir); err != nil {
			return nil, err
		}
	}

	return syncTreePull(sftpClient, source, target, localPath, deleteExtra, dryRun)
}

func syncDirectoryBidirectional(sftpClient *sftp.Client, localPath, remotePath string, recursive, dryRun bool, conflict string, includePatterns, excludePatterns []string) (map[string]interface{}, error) {
	localTree, err := collectLocalTreeAllowMissing(localPath, recursive, includePatterns, excludePatterns)
	if err != nil {
		return nil, err
	}
	remoteTree, err := collectRemoteTree(sftpClient, remotePath, recursive, true, includePatterns, excludePatterns)
	if err != nil {
		return nil, err
	}

	if !localTree.Exists && !remoteTree.Exists {
		return nil, fmt.Errorf("both paths are missing: %s and %s", localPath, remotePath)
	}
	if !dryRun {
		if localTree.Exists {
			if !localTree.RootIsDir {
				return nil, fmt.Errorf("local path is not a directory: %s", localPath)
			}
		} else if err := os.MkdirAll(localPath, 0755); err != nil {
			return nil, fmt.Errorf("create local root dir: %w", err)
		}
		if remoteTree.Exists {
			if !remoteTree.RootIsDir {
				return nil, fmt.Errorf("remote path is not a directory: %s", remotePath)
			}
		} else if err := sftpClient.MkdirAll(remotePath); err != nil {
			return nil, fmt.Errorf("create remote root dir: %w", err)
		}
	}

	return syncTreeBidirectional(sftpClient, localPath, remotePath, localTree, remoteTree, dryRun, conflict)
}

func syncTreePush(sftpClient *sftp.Client, source, target *syncTree, remoteRoot string, deleteExtra, dryRun bool) (map[string]interface{}, error) {
	var uploaded, createdDirs, deleted []string

	dirs := sortedDirs(source.Dirs)
	for _, rel := range dirs {
		remoteDir := joinRemoteRoot(remoteRoot, rel)
		if dryRun {
			createdDirs = append(createdDirs, "[dry-run] mkdir "+remoteDir)
			continue
		}
		if err := sftpClient.MkdirAll(remoteDir); err != nil {
			return nil, fmt.Errorf("create remote directory %s: %w", remoteDir, err)
		}
		createdDirs = append(createdDirs, remoteDir)
	}

	for rel, state := range source.Files {
		if targetState, ok := target.Files[rel]; ok && sameFileState(state, targetState) {
			continue
		}
		localFile := joinLocalRoot(source.Root, rel)
		remoteFile := joinRemoteRoot(remoteRoot, rel)
		if dryRun {
			uploaded = append(uploaded, fmt.Sprintf("[dry-run] upload %s -> %s", localFile, remoteFile))
			continue
		}
		if _, err := uploadFilePreserveTime(sftpClient, localFile, remoteFile); err != nil {
			return nil, err
		}
		uploaded = append(uploaded, remoteFile)
	}

	if deleteExtra {
		for rel := range target.Files {
			if _, ok := source.Files[rel]; ok {
				continue
			}
			remotePath := joinRemoteRoot(remoteRoot, rel)
			if dryRun {
				deleted = append(deleted, "[dry-run] remove "+remotePath)
				continue
			}
			if err := sftpClient.Remove(remotePath); err != nil && !isNotExistError(err) {
				return nil, fmt.Errorf("remove remote file %s: %w", remotePath, err)
			}
			deleted = append(deleted, remotePath)
		}
		for _, rel := range reverseSortedDirs(target.Dirs) {
			if _, ok := source.Dirs[rel]; ok {
				continue
			}
			remotePath := joinRemoteRoot(remoteRoot, rel)
			if dryRun {
				deleted = append(deleted, "[dry-run] rmdir "+remotePath)
				continue
			}
			if err := sftpClient.RemoveDirectory(remotePath); err != nil && !isNotExistError(err) {
				return nil, fmt.Errorf("remove remote directory %s: %w", remotePath, err)
			}
			deleted = append(deleted, remotePath)
		}
	}

	return map[string]interface{}{
		"uploaded":     len(uploaded),
		"dirs_created": len(createdDirs),
		"deleted":      len(deleted),
		"details": map[string]interface{}{
			"uploads":   uploaded,
			"mkdirs":    createdDirs,
			"deletions": deleted,
		},
	}, nil
}

func syncTreePull(sftpClient *sftp.Client, source, target *syncTree, localRoot string, deleteExtra, dryRun bool) (map[string]interface{}, error) {
	var downloaded, createdDirs, deleted []string

	dirs := sortedDirs(source.Dirs)
	for _, rel := range dirs {
		localDir := joinLocalRoot(localRoot, rel)
		if dryRun {
			createdDirs = append(createdDirs, "[dry-run] mkdir "+localDir)
			continue
		}
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return nil, fmt.Errorf("create local directory %s: %w", localDir, err)
		}
		createdDirs = append(createdDirs, localDir)
	}

	for rel, state := range source.Files {
		if targetState, ok := target.Files[rel]; ok && sameFileState(state, targetState) {
			continue
		}
		remoteFile := joinRemoteRoot(source.Root, rel)
		localFile := joinLocalRoot(localRoot, rel)
		if dryRun {
			downloaded = append(downloaded, fmt.Sprintf("[dry-run] download %s -> %s", remoteFile, localFile))
			continue
		}
		if _, err := downloadFilePreserveTime(sftpClient, localFile, remoteFile); err != nil {
			return nil, err
		}
		downloaded = append(downloaded, localFile)
	}

	if deleteExtra {
		for rel := range target.Files {
			if _, ok := source.Files[rel]; ok {
				continue
			}
			localPath := joinLocalRoot(localRoot, rel)
			if dryRun {
				deleted = append(deleted, "[dry-run] remove "+localPath)
				continue
			}
			if err := os.Remove(localPath); err != nil && !os.IsNotExist(err) {
				return nil, fmt.Errorf("remove local file %s: %w", localPath, err)
			}
			deleted = append(deleted, localPath)
		}
		for _, rel := range reverseSortedDirs(target.Dirs) {
			if _, ok := source.Dirs[rel]; ok {
				continue
			}
			localPath := joinLocalRoot(localRoot, rel)
			if dryRun {
				deleted = append(deleted, "[dry-run] rmdir "+localPath)
				continue
			}
			if err := os.Remove(localPath); err != nil && !os.IsNotExist(err) {
				return nil, fmt.Errorf("remove local directory %s: %w", localPath, err)
			}
			deleted = append(deleted, localPath)
		}
	}

	return map[string]interface{}{
		"downloaded":   len(downloaded),
		"dirs_created": len(createdDirs),
		"deleted":      len(deleted),
		"details": map[string]interface{}{
			"downloads": downloaded,
			"mkdirs":    createdDirs,
			"deletions": deleted,
		},
	}, nil
}

func syncTreeBidirectional(sftpClient *sftp.Client, localRoot, remoteRoot string, localTree, remoteTree *syncTree, dryRun bool, conflict string) (map[string]interface{}, error) {
	var uploaded, downloaded, createdLocalDirs, createdRemoteDirs, conflicts []string

	allDirs := mergeDirKeys(localTree.Dirs, remoteTree.Dirs)
	for _, rel := range allDirs {
		localDir := joinLocalRoot(localRoot, rel)
		remoteDir := joinRemoteRoot(remoteRoot, rel)
		if dryRun {
			if _, ok := localTree.Dirs[rel]; !ok {
				createdLocalDirs = append(createdLocalDirs, "[dry-run] mkdir "+localDir)
			}
			if _, ok := remoteTree.Dirs[rel]; !ok {
				createdRemoteDirs = append(createdRemoteDirs, "[dry-run] mkdir "+remoteDir)
			}
			continue
		}
		if _, ok := localTree.Dirs[rel]; !ok {
			if err := os.MkdirAll(localDir, 0755); err != nil {
				return nil, fmt.Errorf("create local directory %s: %w", localDir, err)
			}
			createdLocalDirs = append(createdLocalDirs, localDir)
		}
		if _, ok := remoteTree.Dirs[rel]; !ok {
			if err := sftpClient.MkdirAll(remoteDir); err != nil {
				return nil, fmt.Errorf("create remote directory %s: %w", remoteDir, err)
			}
			createdRemoteDirs = append(createdRemoteDirs, remoteDir)
		}
	}

	allFiles := mergeFileKeys(localTree.Files, remoteTree.Files)
	for _, rel := range allFiles {
		localState, localOK := localTree.Files[rel]
		remoteState, remoteOK := remoteTree.Files[rel]
		localFile := joinLocalRoot(localRoot, rel)
		remoteFile := joinRemoteRoot(remoteRoot, rel)

		switch {
		case localOK && !remoteOK:
			if dryRun {
				uploaded = append(uploaded, fmt.Sprintf("[dry-run] upload %s -> %s", localFile, remoteFile))
				continue
			}
			if _, err := uploadFilePreserveTime(sftpClient, localFile, remoteFile); err != nil {
				return nil, err
			}
			uploaded = append(uploaded, remoteFile)

		case !localOK && remoteOK:
			if dryRun {
				downloaded = append(downloaded, fmt.Sprintf("[dry-run] download %s -> %s", remoteFile, localFile))
				continue
			}
			if _, err := downloadFilePreserveTime(sftpClient, localFile, remoteFile); err != nil {
				return nil, err
			}
			downloaded = append(downloaded, localFile)

		case sameFileState(localState, remoteState):
			continue

		default:
			action, err := resolveBidirectionalAction(localState, remoteState, conflict)
			if err != nil {
				conflicts = append(conflicts, fmt.Sprintf("%s: %v", rel, err))
				continue
			}
			switch action {
			case "upload":
				if dryRun {
					uploaded = append(uploaded, fmt.Sprintf("[dry-run] upload %s -> %s (%s)", localFile, remoteFile, conflict))
					continue
				}
				if _, err := uploadFilePreserveTime(sftpClient, localFile, remoteFile); err != nil {
					return nil, err
				}
				uploaded = append(uploaded, remoteFile)
			case "download":
				if dryRun {
					downloaded = append(downloaded, fmt.Sprintf("[dry-run] download %s -> %s (%s)", remoteFile, localFile, conflict))
					continue
				}
				if _, err := downloadFilePreserveTime(sftpClient, localFile, remoteFile); err != nil {
					return nil, err
				}
				downloaded = append(downloaded, localFile)
			default:
				return nil, fmt.Errorf("unsupported bidirectional action: %s", action)
			}
		}
	}

	return map[string]interface{}{
		"uploaded":    len(uploaded),
		"downloaded":  len(downloaded),
		"conflicts":   len(conflicts),
		"details": map[string]interface{}{
			"uploads":     uploaded,
			"downloads":   downloaded,
			"conflicts":   conflicts,
			"local_dirs":  createdLocalDirs,
			"remote_dirs": createdRemoteDirs,
		},
	}, nil
}

func mergeDirKeys(a, b map[string]struct{}) []string {
	merged := make(map[string]struct{}, len(a)+len(b))
	for rel := range a {
		merged[rel] = struct{}{}
	}
	for rel := range b {
		merged[rel] = struct{}{}
	}
	return sortedDirs(merged)
}

func mergeFileKeys(a, b map[string]fileState) []string {
	merged := make(map[string]struct{}, len(a)+len(b))
	for rel := range a {
		merged[rel] = struct{}{}
	}
	for rel := range b {
		merged[rel] = struct{}{}
	}
	result := make([]string, 0, len(merged))
	for rel := range merged {
		result = append(result, rel)
	}
	sort.Strings(result)
	return result
}

func shouldSyncPath(rel string, isDir bool, includePatterns, excludePatterns []string) bool {
	rel = filepath.ToSlash(strings.TrimPrefix(rel, "./"))
	if rel == "" {
		return true
	}
	if matchesAnyPattern(rel, excludePatterns) {
		return false
	}
	if len(includePatterns) == 0 {
		return true
	}
	if matchesAnyPattern(rel, includePatterns) {
		return true
	}
	if isDir {
		prefix := rel + "/"
		for _, pattern := range includePatterns {
			if strings.HasPrefix(pattern, prefix) || strings.HasPrefix(pattern, rel+"/**") || strings.Contains(pattern, prefix) {
				return true
			}
		}
	}
	return false
}

func matchesAnyPattern(rel string, patterns []string) bool {
	for _, pattern := range patterns {
		if matchPattern(rel, pattern) {
			return true
		}
	}
	return false
}

func matchPattern(rel, pattern string) bool {
	rel = filepath.ToSlash(rel)
	pattern = filepath.ToSlash(strings.TrimSpace(pattern))
	if pattern == "" {
		return false
	}
	baseOnly := !strings.Contains(pattern, "/")
	if pattern == rel {
		return true
	}
	if !strings.ContainsAny(pattern, "*?[") {
		return path.Base(rel) == pattern || strings.HasPrefix(rel, pattern+"/")
	}
	re, err := regexp.Compile(globToRegex(pattern))
	if err != nil {
		return false
	}
	if re.MatchString(rel) {
		return true
	}
	if baseOnly {
		return re.MatchString(path.Base(rel))
	}
	return false
}

func globToRegex(pattern string) string {
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(pattern); i++ {
		ch := pattern[i]
		switch ch {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				b.WriteString(".*")
				i++
			} else {
				b.WriteString("[^/]*")
			}
		case '?':
			b.WriteString("[^/]")
		case '.', '+', '(', ')', '|', '^', '$', '{', '}', '\\':
			b.WriteByte('\\')
			b.WriteByte(ch)
		default:
			b.WriteByte(ch)
		}
	}
	b.WriteString("$")
	return b.String()
}

func expandHome(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, p[2:])
	}
	return p
}
