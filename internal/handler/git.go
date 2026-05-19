package handler

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/topxeq/xxaitk/internal/output"
)

type GitHandler struct{}

type GitPayload struct {
	Action    string   `json:"action"`
	Repo      string   `json:"repo"`
	Files     []string `json:"files"`
	Message   string   `json:"message"`
	Branch    string   `json:"branch"`
	Remote    string   `json:"remote"`
	Tag       string   `json:"tag"`
	N         int      `json:"n"`
	File      string   `json:"file"`
	Author    string   `json:"author"`
	Since     string   `json:"since"`
	Until     string   `json:"until"`
	All       bool     `json:"all,omitempty"`
	Force     bool     `json:"force,omitempty"`
	Recursive bool     `json:"recursive,omitempty"`
	Depth     int      `json:"depth,omitempty"`
	Staged    bool     `json:"staged,omitempty"`
	URL       string   `json:"url"`
	Args      []string `json:"args"`
}

type GitResult struct {
	Action  string `json:"action"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
	Success bool   `json:"success"`
}

func (h *GitHandler) Handle(data string, source string) *output.Response {
	start := time.Now()

	payload := h.parsePayload(data)
	action := strings.ToLower(payload.Action)
	if action == "" {
		return output.Fail("git", source, "GIT_NO_ACTION",
			"no action specified", "", start)
	}

	var gitArgs []string
	var err error

	switch action {
	case "status":
		gitArgs, err = h.buildStatus(payload)
	case "log":
		gitArgs, err = h.buildLog(payload)
	case "diff":
		gitArgs, err = h.buildDiff(payload)
	case "add":
		gitArgs, err = h.buildAdd(payload)
	case "commit":
		gitArgs, err = h.buildCommit(payload)
	case "branch":
		gitArgs, err = h.buildBranch(payload)
	case "checkout":
		gitArgs, err = h.buildCheckout(payload)
	case "pull":
		gitArgs, err = h.buildPull(payload)
	case "push":
		gitArgs, err = h.buildPush(payload)
	case "fetch":
		gitArgs, err = h.buildFetch(payload)
	case "tag":
		gitArgs, err = h.buildTag(payload)
	case "stash":
		gitArgs, err = h.buildStash(payload)
	case "merge":
		gitArgs, err = h.buildMerge(payload)
	case "rebase":
		gitArgs, err = h.buildRebase(payload)
	case "remote":
		gitArgs, err = h.buildRemote(payload)
	case "clone":
		gitArgs, err = h.buildClone(payload)
	case "init":
		gitArgs, err = h.buildInit(payload)
	case "revparse", "rev-parse":
		gitArgs, err = h.buildRevParse(payload)
	case "show":
		gitArgs, err = h.buildShow(payload)
	case "blame":
		gitArgs, err = h.buildBlame(payload)
	case "reset":
		gitArgs, err = h.buildReset(payload)
	case "config":
		gitArgs, err = h.buildConfig(payload)
	case "clean":
		gitArgs, err = h.buildClean(payload)
	default:
		return output.Fail("git", source, "GIT_UNKNOWN_ACTION",
			fmt.Sprintf("unknown action: %s", action), "", start)
	}

	if err != nil {
		return output.Fail("git", source, "GIT_BUILD_ERROR", err.Error(), "", start)
	}

	result := h.execGit(payload.Repo, gitArgs...)

	return output.Success("git", source, result, start)
}

func (h *GitHandler) execGit(repo string, args ...string) *GitResult {
	cmd := exec.Command("git", args...)
	if repo != "" {
		cmd.Dir = repo
	}

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := &GitResult{
		Output:  stdout.String(),
		Error:   strings.TrimSpace(stderr.String()),
		Success: err == nil,
	}

	return result
}

func (h *GitHandler) buildStatus(p *GitPayload) ([]string, error) {
	args := []string{"status", "--porcelain=v2", "--branch"}
	if p.All {
		args = append(args, "--", ".")
	}
	return args, nil
}

func (h *GitHandler) buildLog(p *GitPayload) ([]string, error) {
	args := []string{"log", "--format=%H|%an|%ae|%aI|%s"}
	n := p.N
	if n <= 0 {
		n = 20
	}
	args = append(args, fmt.Sprintf("-%d", n))
	if p.Author != "" {
		args = append(args, "--author="+p.Author)
	}
	if p.Since != "" {
		args = append(args, "--since="+p.Since)
	}
	if p.Until != "" {
		args = append(args, "--until="+p.Until)
	}
	if p.File != "" {
		args = append(args, "--", p.File)
	}
	return args, nil
}

func (h *GitHandler) buildDiff(p *GitPayload) ([]string, error) {
	args := []string{"diff"}
	if p.Staged {
		args = append(args, "--staged")
	}
	if p.File != "" {
		args = append(args, "--", p.File)
	}
	return args, nil
}

func (h *GitHandler) buildAdd(p *GitPayload) ([]string, error) {
	args := []string{"add"}
	if len(p.Files) > 0 {
		args = append(args, p.Files...)
	} else {
		args = append(args, ".")
	}
	return args, nil
}

func (h *GitHandler) buildCommit(p *GitPayload) ([]string, error) {
	args := []string{"commit", "-m", p.Message}
	if p.Author != "" {
		args = append(args, "--author="+p.Author)
	}
	return args, nil
}

func (h *GitHandler) buildBranch(p *GitPayload) ([]string, error) {
	args := []string{"branch"}
	if p.All {
		args = append(args, "-a")
	}
	if p.Branch != "" {
		args = []string{"branch", p.Branch}
	}
	return args, nil
}

func (h *GitHandler) buildCheckout(p *GitPayload) ([]string, error) {
	args := []string{"checkout"}
	if p.Branch != "" {
		args = append(args, p.Branch)
	}
	if p.Force {
		args = append(args, "--force")
	}
	if p.File != "" {
		args = append(args, "--", p.File)
	}
	return args, nil
}

func (h *GitHandler) buildPull(p *GitPayload) ([]string, error) {
	args := []string{"pull"}
	if p.Remote != "" {
		args = append(args, p.Remote)
	}
	if p.Branch != "" {
		args = append(args, p.Branch)
	}
	return args, nil
}

func (h *GitHandler) buildPush(p *GitPayload) ([]string, error) {
	args := []string{"push"}
	if p.Remote != "" {
		args = append(args, p.Remote)
	}
	if p.Branch != "" {
		args = append(args, p.Branch)
	}
	if p.Force {
		args = append(args, "--force")
	}
	if p.Tag != "" {
		args = append(args, "origin", p.Tag)
	}
	return args, nil
}

func (h *GitHandler) buildFetch(p *GitPayload) ([]string, error) {
	args := []string{"fetch"}
	if p.Remote != "" {
		args = append(args, p.Remote)
	}
	if p.All {
		args = append(args, "--all")
	}
	return args, nil
}

func (h *GitHandler) buildTag(p *GitPayload) ([]string, error) {
	if p.Tag != "" {
		return []string{"tag", p.Tag}, nil
	}
	return []string{"tag", "-l"}, nil
}

func (h *GitHandler) buildStash(p *GitPayload) ([]string, error) {
	if p.Branch != "" {
		return []string{"stash", "pop"}, nil
	}
	return []string{"stash"}, nil
}

func (h *GitHandler) buildMerge(p *GitPayload) ([]string, error) {
	if p.Branch == "" {
		return nil, fmt.Errorf("merge requires a branch")
	}
	return []string{"merge", p.Branch}, nil
}

func (h *GitHandler) buildRebase(p *GitPayload) ([]string, error) {
	args := []string{"rebase"}
	if p.Branch != "" {
		args = append(args, p.Branch)
	}
	return args, nil
}

func (h *GitHandler) buildRemote(p *GitPayload) ([]string, error) {
	args := []string{"remote", "-v"}
	return args, nil
}

func (h *GitHandler) buildClone(p *GitPayload) ([]string, error) {
	if p.URL == "" {
		return nil, fmt.Errorf("clone requires url")
	}
	args := []string{"clone"}
	if p.Depth > 0 {
		args = append(args, fmt.Sprintf("--depth=%d", p.Depth))
	}
	args = append(args, p.URL)
	if p.Repo != "" {
		args = append(args, p.Repo)
	}
	return args, nil
}

func (h *GitHandler) buildInit(p *GitPayload) ([]string, error) {
	return []string{"init"}, nil
}

func (h *GitHandler) buildRevParse(p *GitPayload) ([]string, error) {
	args := []string{"rev-parse"}
	if len(p.Args) > 0 {
		args = append(args, p.Args...)
	} else {
		args = append(args, "HEAD")
	}
	return args, nil
}

func (h *GitHandler) buildShow(p *GitPayload) ([]string, error) {
	args := []string{"show", "--format=%H|%an|%ae|%aI|%s", "--stat"}
	if p.Branch != "" {
		args = append(args, p.Branch)
	}
	return args, nil
}

func (h *GitHandler) buildBlame(p *GitPayload) ([]string, error) {
	if p.File == "" {
		return nil, fmt.Errorf("blame requires a file")
	}
	return []string{"blame", p.File}, nil
}

func (h *GitHandler) buildReset(p *GitPayload) ([]string, error) {
	args := []string{"reset"}
	if p.Force {
		args = append(args, "--hard")
	}
	if p.Branch != "" {
		args = append(args, p.Branch)
	}
	return args, nil
}

func (h *GitHandler) buildConfig(p *GitPayload) ([]string, error) {
	args := []string{"config"}
	if len(p.Args) > 0 {
		args = append(args, p.Args...)
	} else {
		args = append(args, "--list")
	}
	return args, nil
}

func (h *GitHandler) buildClean(p *GitPayload) ([]string, error) {
	args := []string{"clean"}
	if p.Force {
		args = append(args, "-fd")
	} else {
		args = append(args, "-n")
	}
	return args, nil
}

func (h *GitHandler) parsePayload(data string) *GitPayload {
	payload := &GitPayload{}
	trimmed := strings.TrimSpace(data)
	if len(trimmed) == 0 {
		return payload
	}
	if trimmed[0] == '{' {
		json.Unmarshal([]byte(trimmed), payload)
	}
	return payload
}
