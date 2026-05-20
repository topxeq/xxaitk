package handler

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func initGitRepo(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "git_test_*")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		os.RemoveAll(dir)
		t.Fatalf("git init failed: %s: %v", string(out), err)
	}
	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = dir
	cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = dir
	cmd.Run()
	return dir
}

func commitFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "add", name)
	cmd.Dir = dir
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "add "+name)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("commit failed: %s: %v", string(out), err)
	}
}

func TestGitNoAction(t *testing.T) {
	h := &GitHandler{}
	resp := h.Handle(`{}`, "")
	if resp.Ok {
		t.Fatal("expected failure for empty action")
	}
	if resp.Error.Code != "GIT_NO_ACTION" {
		t.Errorf("expected GIT_NO_ACTION, got: %s", resp.Error.Code)
	}
}

func TestGitUnknownAction(t *testing.T) {
	h := &GitHandler{}
	resp := h.Handle(`{"action":"bad"}`, "")
	if resp.Ok {
		t.Fatal("expected failure for unknown action")
	}
	if resp.Error.Code != "GIT_UNKNOWN_ACTION" {
		t.Errorf("expected GIT_UNKNOWN_ACTION, got: %s", resp.Error.Code)
	}
}

func TestGitInit(t *testing.T) {
	dir, err := os.MkdirTemp("", "git_init_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	h := &GitHandler{}
	repoDir := filepath.Join(dir, "newrepo")
	os.MkdirAll(repoDir, 0755)

	payload := GitPayload{Action: "init", Repo: repoDir}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*GitResult)
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); os.IsNotExist(err) {
		t.Error("expected .git directory to exist")
	}
}

func TestGitStatus(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)

	h := &GitHandler{}
	payload := GitPayload{Action: "status", Repo: dir}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*GitResult)
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
}

func TestGitAddCommit(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)

	os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello"), 0644)

	h := &GitHandler{}
	addPayload := GitPayload{Action: "add", Repo: dir, Files: []string{"hello.txt"}}
	data, _ := json.Marshal(addPayload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("add expected ok, got error: %v", resp.Error)
	}

	commitPayload := GitPayload{Action: "commit", Repo: dir, Message: "initial"}
	data, _ = json.Marshal(commitPayload)
	resp = h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("commit expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*GitResult)
	if !result.Success {
		t.Errorf("commit expected success, got error: %s", result.Error)
	}
}

func TestGitLog(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)

	commitFile(t, dir, "a.txt", "aaa")

	h := &GitHandler{}
	payload := GitPayload{Action: "log", Repo: dir, N: 5}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*GitResult)
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
	if len(result.Output) == 0 {
		t.Error("expected non-empty log output")
	}
}

func TestGitDiff(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)

	commitFile(t, dir, "a.txt", "aaa")
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("bbb"), 0644)

	h := &GitHandler{}
	payload := GitPayload{Action: "diff", Repo: dir}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*GitResult)
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
}

func TestGitBranch(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)

	commitFile(t, dir, "a.txt", "aaa")

	h := &GitHandler{}
	payload := GitPayload{Action: "branch", Repo: dir, Branch: "feature"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*GitResult)
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}

	listPayload := GitPayload{Action: "branch", Repo: dir, All: true}
	data, _ = json.Marshal(listPayload)
	resp = h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result = resp.Data.(*GitResult)
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
}

func TestGitTag(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)

	commitFile(t, dir, "a.txt", "aaa")

	h := &GitHandler{}
	payload := GitPayload{Action: "tag", Repo: dir, Tag: "v1.0"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*GitResult)
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}

	listPayload := GitPayload{Action: "tag", Repo: dir}
	data, _ = json.Marshal(listPayload)
	resp = h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitConfig(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)

	h := &GitHandler{}
	payload := GitPayload{Action: "config", Repo: dir, Args: []string{"--list"}}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*GitResult)
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
}

func TestGitParsePayloadJSON(t *testing.T) {
	h := &GitHandler{}
	payload := h.parsePayload(`{"action":"status","repo":"/tmp/test","branch":"main"}`)
	if payload.Action != "status" {
		t.Errorf("expected action=status, got: %s", payload.Action)
	}
	if payload.Repo != "/tmp/test" {
		t.Errorf("expected repo=/tmp/test, got: %s", payload.Repo)
	}
	if payload.Branch != "main" {
		t.Errorf("expected branch=main, got: %s", payload.Branch)
	}
}

func TestGitParsePayloadNonJSON(t *testing.T) {
	h := &GitHandler{}
	payload := h.parsePayload("not json")
	if payload.Action != "" {
		t.Errorf("expected empty action for non-JSON, got: %s", payload.Action)
	}
}

func TestGitCheckout(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)
	commitFile(t, dir, "a.txt", "aaa")

	h := &GitHandler{}
	payload := GitPayload{Action: "branch", Repo: dir, Branch: "feature"}
	data, _ := json.Marshal(payload)
	h.Handle(string(data), "")

	checkoutPayload := GitPayload{Action: "checkout", Repo: dir, Branch: "feature"}
	data, _ = json.Marshal(checkoutPayload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitCheckoutForce(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)
	commitFile(t, dir, "a.txt", "aaa")

	h := &GitHandler{}
	payload := GitPayload{Action: "branch", Repo: dir, Branch: "feature2"}
	data, _ := json.Marshal(payload)
	h.Handle(string(data), "")

	checkoutPayload := GitPayload{Action: "checkout", Repo: dir, Branch: "feature2", Force: true}
	data, _ = json.Marshal(checkoutPayload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitCheckoutFile(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)
	commitFile(t, dir, "a.txt", "aaa")

	h := &GitHandler{}
	payload := GitPayload{Action: "checkout", Repo: dir, File: "a.txt"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitStash(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)
	commitFile(t, dir, "a.txt", "aaa")
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("modified"), 0644)

	h := &GitHandler{}
	payload := GitPayload{Action: "stash", Repo: dir}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitStashPop(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)
	commitFile(t, dir, "a.txt", "aaa")
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("modified"), 0644)

	h := &GitHandler{}
	payload := GitPayload{Action: "stash", Repo: dir}
	data, _ := json.Marshal(payload)
	h.Handle(string(data), "")

	popPayload := GitPayload{Action: "stash", Repo: dir, Branch: "pop"}
	data, _ = json.Marshal(popPayload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitRemote(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)

	h := &GitHandler{}
	payload := GitPayload{Action: "remote", Repo: dir}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitMergeRequiresBranch(t *testing.T) {
	h := &GitHandler{}
	resp := h.Handle(`{"action":"merge"}`, "")
	if resp.Ok {
		t.Error("expected failure for merge without branch")
	}
	if resp.Error.Code != "GIT_BUILD_ERROR" {
		t.Errorf("expected GIT_BUILD_ERROR, got: %s", resp.Error.Code)
	}
}

func TestGitRebase(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)
	commitFile(t, dir, "a.txt", "aaa")

	h := &GitHandler{}
	payload := GitPayload{Action: "rebase", Repo: dir, Branch: "main"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitCloneRequiresURL(t *testing.T) {
	h := &GitHandler{}
	resp := h.Handle(`{"action":"clone"}`, "")
	if resp.Ok {
		t.Error("expected failure for clone without url")
	}
	if resp.Error.Code != "GIT_BUILD_ERROR" {
		t.Errorf("expected GIT_BUILD_ERROR, got: %s", resp.Error.Code)
	}
}

func TestGitShow(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)
	commitFile(t, dir, "a.txt", "aaa")

	h := &GitHandler{}
	payload := GitPayload{Action: "show", Repo: dir}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*GitResult)
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
}

func TestGitBlameRequiresFile(t *testing.T) {
	h := &GitHandler{}
	resp := h.Handle(`{"action":"blame"}`, "")
	if resp.Ok {
		t.Error("expected failure for blame without file")
	}
	if resp.Error.Code != "GIT_BUILD_ERROR" {
		t.Errorf("expected GIT_BUILD_ERROR, got: %s", resp.Error.Code)
	}
}

func TestGitBlame(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)
	commitFile(t, dir, "a.txt", "aaa")

	h := &GitHandler{}
	payload := GitPayload{Action: "blame", Repo: dir, File: "a.txt"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitReset(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)
	commitFile(t, dir, "a.txt", "aaa")

	h := &GitHandler{}
	payload := GitPayload{Action: "reset", Repo: dir, Force: true}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitClean(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)
	commitFile(t, dir, "a.txt", "aaa")
	os.WriteFile(filepath.Join(dir, "untracked.txt"), []byte("junk"), 0644)

	h := &GitHandler{}
	payload := GitPayload{Action: "clean", Repo: dir}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitCleanForce(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)
	commitFile(t, dir, "a.txt", "aaa")
	os.WriteFile(filepath.Join(dir, "untracked.txt"), []byte("junk"), 0644)

	h := &GitHandler{}
	payload := GitPayload{Action: "clean", Repo: dir, Force: true}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitPush(t *testing.T) {
	h := &GitHandler{}
	resp := h.Handle(`{"action":"push","remote":"origin","branch":"main"}`, "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitPushForce(t *testing.T) {
	h := &GitHandler{}
	resp := h.Handle(`{"action":"push","remote":"origin","branch":"main","force":true}`, "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitPushTag(t *testing.T) {
	h := &GitHandler{}
	resp := h.Handle(`{"action":"push","tag":"v1.0"}`, "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitFetch(t *testing.T) {
	h := &GitHandler{}
	resp := h.Handle(`{"action":"fetch","remote":"origin"}`, "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitFetchAll(t *testing.T) {
	h := &GitHandler{}
	resp := h.Handle(`{"action":"fetch","all":true}`, "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitPull(t *testing.T) {
	h := &GitHandler{}
	resp := h.Handle(`{"action":"pull","remote":"origin","branch":"main"}`, "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitCloneWithDepth(t *testing.T) {
	h := &GitHandler{}
	resp := h.Handle(`{"action":"clone","url":"https://github.com/example/repo.git","depth":1,"repo":"/tmp/clone"}`, "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitLogWithFilters(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)
	commitFile(t, dir, "a.txt", "aaa")

	h := &GitHandler{}
	payload := GitPayload{Action: "log", Repo: dir, N: 5, Author: "Test", File: "a.txt"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitDiffStaged(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)
	commitFile(t, dir, "a.txt", "aaa")
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("bbb"), 0644)

	cmd := exec.Command("git", "add", "a.txt")
	cmd.Dir = dir
	cmd.Run()

	h := &GitHandler{}
	payload := GitPayload{Action: "diff", Repo: dir, Staged: true}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitStatusAll(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)
	commitFile(t, dir, "a.txt", "aaa")

	h := &GitHandler{}
	payload := GitPayload{Action: "status", Repo: dir, All: true}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitCommitWithAuthor(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("aaa"), 0644)

	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	cmd.Run()

	h := &GitHandler{}
	payload := GitPayload{Action: "commit", Repo: dir, Message: "test", Author: "Test <test@test.com>"}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
}

func TestGitParsePayloadEmpty(t *testing.T) {
	h := &GitHandler{}
	payload := h.parsePayload("")
	if payload.Action != "" {
		t.Errorf("expected empty action, got: %s", payload.Action)
	}
}

func TestGitBuildRevParse(t *testing.T) {
	h := &GitHandler{}
	args, err := h.buildRevParse(&GitPayload{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(args) < 2 || args[0] != "rev-parse" || args[1] != "HEAD" {
		t.Errorf("expected rev-parse HEAD, got: %v", args)
	}
}

func TestGitBuildRevParseCustomArgs(t *testing.T) {
	h := &GitHandler{}
	args, err := h.buildRevParse(&GitPayload{Args: []string{"--short", "HEAD"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if args[1] != "--short" || args[2] != "HEAD" {
		t.Errorf("expected custom args, got: %v", args)
	}
}

func TestGitRevParseAction(t *testing.T) {
	dir := initGitRepo(t)
	defer os.RemoveAll(dir)
	commitFile(t, dir, "a.txt", "aaa")

	h := &GitHandler{}
	payload := GitPayload{Action: "rev-parse", Repo: dir, Args: []string{"HEAD"}}
	data, _ := json.Marshal(payload)
	resp := h.Handle(string(data), "")
	if !resp.Ok {
		t.Fatalf("expected ok, got error: %v", resp.Error)
	}
	result := resp.Data.(*GitResult)
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
	if len(result.Output) == 0 {
		t.Error("expected non-empty rev-parse output")
	}
}

func TestGitBuildCleanDefault(t *testing.T) {
	h := &GitHandler{}
	args, err := h.buildClean(&GitPayload{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if args[1] != "-n" {
		t.Errorf("expected -n (dry-run), got: %v", args)
	}
}

func TestGitBuildResetDefault(t *testing.T) {
	h := &GitHandler{}
	args, err := h.buildReset(&GitPayload{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if args[0] != "reset" {
		t.Errorf("expected reset, got: %v", args)
	}
}

func TestGitBuildResetHardWithBranch(t *testing.T) {
	h := &GitHandler{}
	args, err := h.buildReset(&GitPayload{Force: true, Branch: "HEAD~1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if args[1] != "--hard" || args[2] != "HEAD~1" {
		t.Errorf("expected reset --hard HEAD~1, got: %v", args)
	}
}

func TestGitBuildBlameNoFile(t *testing.T) {
	h := &GitHandler{}
	_, err := h.buildBlame(&GitPayload{})
	if err == nil {
		t.Error("expected error for blame without file")
	}
}

func TestGitBuildShowWithBranch(t *testing.T) {
	h := &GitHandler{}
	args, err := h.buildShow(&GitPayload{Branch: "main"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if args[len(args)-1] != "main" {
		t.Errorf("expected main at end, got: %v", args)
	}
}

func TestGitBuildLogDefaults(t *testing.T) {
	h := &GitHandler{}
	args, err := h.buildLog(&GitPayload{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if args[0] != "log" {
		t.Errorf("expected log, got: %v", args)
	}
	if args[2] != "-20" {
		t.Errorf("expected -20 default, got: %v", args)
	}
}

func TestGitBuildLogWithN(t *testing.T) {
	h := &GitHandler{}
	args, _ := h.buildLog(&GitPayload{N: 5})
	if args[2] != "-5" {
		t.Errorf("expected -5, got: %v", args)
	}
}

func TestGitBuildDiffStaged(t *testing.T) {
	h := &GitHandler{}
	args, _ := h.buildDiff(&GitPayload{Staged: true})
	found := false
	for _, a := range args {
		if a == "--staged" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected --staged in diff args: %v", args)
	}
}

func TestGitBuildDiffWithFile(t *testing.T) {
	h := &GitHandler{}
	args, _ := h.buildDiff(&GitPayload{File: "main.go"})
	if args[len(args)-1] != "main.go" {
		t.Errorf("expected main.go at end: %v", args)
	}
}

func TestGitBuildAddDefault(t *testing.T) {
	h := &GitHandler{}
	args, _ := h.buildAdd(&GitPayload{})
	if args[1] != "." {
		t.Errorf("expected add ., got: %v", args)
	}
}

func TestGitBuildAddFiles(t *testing.T) {
	h := &GitHandler{}
	args, _ := h.buildAdd(&GitPayload{Files: []string{"a.go", "b.go"}})
	if len(args) != 3 || args[1] != "a.go" || args[2] != "b.go" {
		t.Errorf("expected add a.go b.go, got: %v", args)
	}
}

func TestGitBuildCommitWithAuthor(t *testing.T) {
	h := &GitHandler{}
	args, _ := h.buildCommit(&GitPayload{Message: "test", Author: "Bob"})
	found := false
	for _, a := range args {
		if strings.HasPrefix(a, "--author=") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected --author in commit args: %v", args)
	}
}
