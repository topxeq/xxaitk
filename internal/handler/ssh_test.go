package handler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
	"time"
)

func TestReadCommands(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "commands.txt")
	if err := os.WriteFile(path, []byte("date\n\nls -la\n"), 0644); err != nil {
		t.Fatalf("write command file: %v", err)
	}

	got, err := readCommands("", path)
	if err != nil {
		t.Fatalf("readCommands returned error: %v", err)
	}
	want := []string{"date", "ls -la"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("readCommands = %#v, want %#v", got, want)
	}

	got, err = readCommands("echo hello", "")
	if err != nil {
		t.Fatalf("readCommands with cmd returned error: %v", err)
	}
	if !reflect.DeepEqual(got, []string{"echo hello"}) {
		t.Fatalf("readCommands with cmd = %#v", got)
	}

	got, err = readCommands("", "")
	if err != nil {
		t.Fatalf("readCommands empty returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("readCommands empty = %#v, want empty", got)
	}
}

func TestResolveRemoteUploadPath(t *testing.T) {
	if got := resolveRemoteUploadPath(`/tmp/app.bin`, "/opt/bin/", ""); got != "/opt/bin/app.bin" {
		t.Fatalf("directory upload path = %q", got)
	}
	if got := resolveRemoteUploadPath(`/tmp/app.bin`, "/opt/bin/custom.bin", "ignored.bin"); got != "/opt/bin/custom.bin" {
		t.Fatalf("file upload path with fileName = %q", got)
	}
	if got := resolveRemoteUploadPath(`/tmp/app.bin`, "/opt/bin/", "renamed.bin"); got != "/opt/bin/renamed.bin" {
		t.Fatalf("file upload path with custom fileName = %q", got)
	}
	if got := resolveRemoteUploadPath(`/tmp/app.bin`, "/opt/bin/app.bin", ""); got != "/opt/bin/app.bin" {
		t.Fatalf("file upload path no fileName = %q", got)
	}
}

func TestResolveLocalDownloadPath(t *testing.T) {
	dir := t.TempDir()
	got := resolveLocalDownloadPath(dir, "/remote/archive.tar.gz", "")
	want := filepath.Join(dir, "archive.tar.gz")
	if got != want {
		t.Fatalf("resolveLocalDownloadPath dir = %q, want %q", got, want)
	}

	filePath := filepath.Join(dir, "output.log")
	got = resolveLocalDownloadPath(filePath, "/remote/archive.tar.gz", "")
	if got != filePath {
		t.Fatalf("resolveLocalDownloadPath file = %q, want %q", got, filePath)
	}

	got = resolveLocalDownloadPath(dir, "/remote/archive.tar.gz", "custom.tar.gz")
	want = filepath.Join(dir, "custom.tar.gz")
	if got != want {
		t.Fatalf("resolveLocalDownloadPath with fileName = %q, want %q", got, want)
	}

	got = resolveLocalDownloadPath(dir+string(os.PathSeparator), "/remote/archive.tar.gz", "")
	want = filepath.Join(dir, "archive.tar.gz")
	if got != want {
		t.Fatalf("resolveLocalDownloadPath trailing separator = %q, want %q", got, want)
	}
}

func TestIsNotExistError(t *testing.T) {
	if isNotExistError(nil) {
		t.Fatal("nil should not be not-exist")
	}
	if !isNotExistError(os.ErrNotExist) {
		t.Fatal("os.ErrNotExist should be not-exist")
	}
	if !isNotExistError(&os.PathError{Err: os.ErrNotExist}) {
		t.Fatal("PathError with ErrNotExist should be not-exist")
	}
}

func TestSameFileState(t *testing.T) {
	now := time.Now()
	a := fileState{Size: 100, ModTime: now}
	b := fileState{Size: 100, ModTime: now}
	if !sameFileState(a, b) {
		t.Fatal("identical states should match")
	}

	c := fileState{Size: 200, ModTime: now}
	if sameFileState(a, c) {
		t.Fatal("different sizes should not match")
	}

	d := fileState{Size: 100, ModTime: now.Add(2 * time.Second)}
	if sameFileState(a, d) {
		t.Fatal("times > 1s apart should not match")
	}

	e := fileState{Size: 100, ModTime: now.Add(500 * time.Millisecond)}
	if !sameFileState(a, e) {
		t.Fatal("times <= 1s apart should match")
	}
}

func TestModTimesClose(t *testing.T) {
	now := time.Now()
	if !modTimesClose(now, now) {
		t.Fatal("same time should be close")
	}
	if !modTimesClose(now, now.Add(999*time.Millisecond)) {
		t.Fatal("999ms apart should be close")
	}
	if modTimesClose(now, now.Add(1001*time.Millisecond)) {
		t.Fatal("1001ms apart should not be close")
	}
}

func TestSameFileInfo(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "test.txt")
	os.WriteFile(fp, []byte("hello world"), 0644)

	info, _ := os.Stat(fp)
	if !sameFileInfo(info, info) {
		t.Fatal("same FileInfo should match")
	}
	if sameFileInfo(nil, info) {
		t.Fatal("nil should not match")
	}
	if sameFileInfo(info, nil) {
		t.Fatal("nil should not match")
	}
}

func TestDetectBidirectionalKind(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "file.txt")
	os.WriteFile(filePath, []byte("x"), 0644)

	fileInfo, _ := os.Stat(filePath)
	dirInfo, _ := os.Stat(dir)

	kind, err := detectBidirectionalKind(fileInfo, true, nil, false)
	if err != nil || kind != "file" {
		t.Fatalf("detect file kind = %q, err=%v", kind, err)
	}

	kind, err = detectBidirectionalKind(nil, false, dirInfo, true)
	if err != nil || kind != "dir" {
		t.Fatalf("detect dir kind = %q, err=%v", kind, err)
	}

	_, err = detectBidirectionalKind(fileInfo, true, dirInfo, true)
	if err == nil {
		t.Fatal("expected mismatch error")
	}

	_, err = detectBidirectionalKind(nil, false, nil, false)
	if err == nil {
		t.Fatal("expected both-missing error")
	}
}

func TestResolveBidirectionalAction(t *testing.T) {
	local := fileState{Size: 10, ModTime: time.Unix(20, 0)}
	remote := fileState{Size: 10, ModTime: time.Unix(10, 0)}

	action, err := resolveBidirectionalAction(local, remote, "newer_wins")
	if err != nil || action != "upload" {
		t.Fatalf("newer_wins (local newer) = %q, err=%v", action, err)
	}

	action, err = resolveBidirectionalAction(remote, local, "newer_wins")
	if err != nil || action != "download" {
		t.Fatalf("newer_wins (remote newer) = %q, err=%v", action, err)
	}

	action, err = resolveBidirectionalAction(local, remote, "local_wins")
	if err != nil || action != "upload" {
		t.Fatalf("local_wins = %q, err=%v", action, err)
	}

	action, err = resolveBidirectionalAction(local, remote, "remote_wins")
	if err != nil || action != "download" {
		t.Fatalf("remote_wins = %q, err=%v", action, err)
	}

	_, err = resolveBidirectionalAction(local, remote, "fail_on_conflict")
	if err == nil {
		t.Fatal("expected conflict error")
	}

	sameTime := fileState{Size: 10, ModTime: time.Unix(5, 0)}
	diffSize := fileState{Size: 20, ModTime: time.Unix(5, 0)}
	_, err = resolveBidirectionalAction(sameTime, diffSize, "newer_wins")
	if err == nil {
		t.Fatal("expected same-time-different-content error for newer_wins")
	}

	_, err = resolveBidirectionalAction(local, remote, "invalid_policy")
	if err == nil {
		t.Fatal("expected invalid policy error")
	}
}

func TestResolveRemoteSyncPath(t *testing.T) {
	got := resolveRemoteSyncPath(`/tmp/bundle.js`, "/opt/app/", nil, false)
	if got != "/opt/app/bundle.js" {
		t.Fatalf("resolveRemoteSyncPath = %q", got)
	}

	got = resolveRemoteSyncPath(`/tmp/bundle.js`, "/opt/app/bundle.js", nil, false)
	if got != "/opt/app/bundle.js" {
		t.Fatalf("resolveRemoteSyncPath exact = %q", got)
	}
}

func TestResolveLocalSyncPath(t *testing.T) {
	dir := t.TempDir()
	dirInfo, _ := os.Stat(dir)

	got := resolveLocalSyncPath(dir, "/opt/app/bundle.js", dirInfo, true)
	want := filepath.Join(dir, "bundle.js")
	if got != want {
		t.Fatalf("resolveLocalSyncPath = %q, want %q", got, want)
	}

	got = resolveLocalSyncPath("/tmp/output.txt", "/opt/app/bundle.js", nil, false)
	if got != "/tmp/output.txt" {
		t.Fatalf("resolveLocalSyncPath exact = %q", got)
	}

	got = resolveLocalSyncPath(dir+string(os.PathSeparator), "/opt/app/bundle.js", dirInfo, true)
	if got != filepath.Join(dir, "bundle.js") {
		t.Fatalf("resolveLocalSyncPath trailing sep = %q", got)
	}
}

func TestShouldSyncPath(t *testing.T) {
	tests := []struct {
		name    string
		rel     string
		isDir   bool
		include []string
		exclude []string
		want    bool
	}{
		{name: "include nested file", rel: "dist/app.js", include: []string{"dist/**"}, want: true},
		{name: "exclude nested file", rel: "dist/app.js", include: []string{"dist/**"}, exclude: []string{"dist/*.js"}, want: false},
		{name: "keep parent dir for include", rel: "dist", isDir: true, include: []string{"dist/**"}, want: true},
		{name: "skip unrelated file", rel: "docs/readme.md", include: []string{"dist/**"}, want: false},
		{name: "basename match", rel: "logs/app.log", include: []string{"*.log"}, want: true},
		{name: "no patterns", rel: "anything.txt", want: true},
		{name: "exclude only", rel: "secret.key", exclude: []string{"*.key"}, want: false},
		{name: "exact match", rel: "config.json", include: []string{"config.json"}, want: true},
		{name: "prefix match dir", rel: "src", isDir: true, include: []string{"src/main.go"}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSyncPath(tt.rel, tt.isDir, tt.include, tt.exclude)
			if got != tt.want {
				t.Fatalf("shouldSyncPath(%q) = %v, want %v", tt.rel, got, tt.want)
			}
		})
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		rel     string
		pattern string
		want    bool
	}{
		{"dist/app.js", "dist/**", true},
		{"dist/sub/deep.js", "dist/**", true},
		{"dist/app.js", "dist/*.js", true},
		{"dist/sub/deep.js", "dist/*.js", false},
		{"app.js", "*.js", true},
		{"app.js", "app.?", false},
		{"app.x", "app.?", true},
		{"app.jsx", "app.?", false},
		{"config.json", "config.json", true},
		{"other.txt", "config.json", false},
	}

	for _, tt := range tests {
		got := matchPattern(tt.rel, tt.pattern)
		if got != tt.want {
			t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.rel, tt.pattern, got, tt.want)
		}
	}
}

func TestGlobToRegex(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		want    bool
	}{
		{"dist/**", "dist/app.js", true},
		{"dist/**", "dist/sub/deep.js", true},
		{"dist/**", "other/app.js", false},
		{"*.js", "app.js", true},
		{"*.js", "app.ts", false},
		{"dist/*.js", "dist/app.js", true},
		{"dist/*.js", "dist/sub/app.js", false},
		{"app.?", "app.x", true},
		{"app.?", "app.jsx", false},
	}

	for _, tt := range tests {
		re := globToRegex(tt.pattern)
		regex := mustCompileRegex(re)
		got := regex.MatchString(tt.input)
		if got != tt.want {
			t.Errorf("globToRegex(%q) match(%q) = %v, want %v (regex=%s)", tt.pattern, tt.input, got, tt.want, re)
		}
	}
}

func mustCompileRegex(s string) *regexp.Regexp {
	re, err := regexp.Compile(s)
	if err != nil {
		panic(err)
	}
	return re
}

func TestSortedDirs(t *testing.T) {
	dirs := map[string]struct{}{
		"z":    {},
		"a":    {},
		"a/b":  {},
		"a/b/c": {},
	}
	got := sortedDirs(dirs)
	want := []string{"a", "a/b", "a/b/c", "z"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sortedDirs = %v, want %v", got, want)
	}
}

func TestReverseSortedDirs(t *testing.T) {
	dirs := map[string]struct{}{
		"a":    {},
		"a/b":  {},
		"a/b/c": {},
	}
	got := reverseSortedDirs(dirs)
	if len(got) == 0 {
		t.Fatal("reverseSortedDirs returned empty")
	}
	if got[0] != "a/b/c" {
		t.Fatalf("reverseSortedDirs first = %q, want %q", got[0], "a/b/c")
	}
}

func TestAddParentDirs(t *testing.T) {
	tree := &syncTree{
		Files: map[string]fileState{
			"a/b/c.txt": {Size: 10, ModTime: time.Now()},
		},
		Dirs: map[string]struct{}{},
	}
	addParentDirs(tree)
	if _, ok := tree.Dirs["a"]; !ok {
		t.Fatal("expected parent dir 'a'")
	}
	if _, ok := tree.Dirs["a/b"]; !ok {
		t.Fatal("expected parent dir 'a/b'")
	}
}

func TestCollectLocalTree(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "sub"), 0755)
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(dir, "sub", "b.txt"), []byte("world"), 0644)

	tree, err := collectLocalTree(dir, true, nil, nil)
	if err != nil {
		t.Fatalf("collectLocalTree error: %v", err)
	}
	if !tree.Exists {
		t.Fatal("tree should exist")
	}
	if !tree.RootIsDir {
		t.Fatal("root should be dir")
	}
	if _, ok := tree.Files["a.txt"]; !ok {
		t.Fatal("expected file a.txt")
	}
	if _, ok := tree.Files["sub/b.txt"]; !ok {
		t.Fatal("expected file sub/b.txt")
	}
	if _, ok := tree.Dirs["sub"]; !ok {
		t.Fatal("expected dir sub")
	}
}

func TestCollectLocalTreeFile(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "single.txt")
	os.WriteFile(fp, []byte("data"), 0644)

	tree, err := collectLocalTree(fp, false, nil, nil)
	if err != nil {
		t.Fatalf("collectLocalTree file error: %v", err)
	}
	if !tree.Exists {
		t.Fatal("tree should exist")
	}
	if tree.RootIsDir {
		t.Fatal("root should not be dir")
	}
	if _, ok := tree.Files[""]; !ok {
		t.Fatal("expected file with empty rel path")
	}
}

func TestCollectLocalTreeNotRecursive(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0644)

	_, err := collectLocalTree(dir, false, nil, nil)
	if err == nil {
		t.Fatal("expected error for directory sync without recursive")
	}
}

func TestCollectLocalTreeMissing(t *testing.T) {
	tree, err := collectLocalTreeAllowMissing("/nonexistent/path/xyz", true, nil, nil)
	if err != nil {
		t.Fatalf("collectLocalTreeAllowMissing error: %v", err)
	}
	if tree.Exists {
		t.Fatal("tree should not exist")
	}
}

func TestEnsureRemoteRootDir(t *testing.T) {
	err := ensureRemoteRootDir(nil, "/opt/app", true, true)
	if err != nil {
		t.Fatalf("existing dir should succeed: %v", err)
	}

	err = ensureRemoteRootDir(nil, "/opt/app", true, false)
	if err == nil {
		t.Fatal("existing non-dir should fail")
	}
}

func TestEnsureLocalRootDir(t *testing.T) {
	dir := t.TempDir()
	err := ensureLocalRootDir(dir, true, true)
	if err != nil {
		t.Fatalf("existing dir should succeed: %v", err)
	}

	fp := filepath.Join(dir, "file.txt")
	os.WriteFile(fp, []byte("x"), 0644)
	err = ensureLocalRootDir(fp, true, false)
	if err == nil {
		t.Fatal("existing non-dir should fail")
	}

	newDir := filepath.Join(dir, "newdir")
	err = ensureLocalRootDir(newDir, false, false)
	if err != nil {
		t.Fatalf("create new dir should succeed: %v", err)
	}
	if _, e := os.Stat(newDir); e != nil {
		t.Fatalf("new dir should exist: %v", e)
	}
}

func TestExpandHome(t *testing.T) {
	if expanded := expandHome("/absolute/path"); expanded != "/absolute/path" {
		t.Fatalf("absolute path should stay same: %q", expanded)
	}
	home, _ := os.UserHomeDir()
	if home != "" {
		expanded := expandHome("~/test")
		expected := filepath.Join(home, "test")
		if expanded != expected {
			t.Fatalf("expandHome ~/test = %q, want %q", expanded, expected)
		}
	}
}

func TestJoinLocalRoot(t *testing.T) {
	got := joinLocalRoot("/root", "sub/file.txt")
	expected := filepath.Join("/root", "sub", "file.txt")
	if got != expected {
		t.Fatalf("joinLocalRoot = %q, want %q", got, expected)
	}
	got = joinLocalRoot("/root", "")
	if got != "/root" {
		t.Fatalf("joinLocalRoot empty rel = %q", got)
	}
}

func TestJoinRemoteRoot(t *testing.T) {
	got := joinRemoteRoot("/opt", "sub/file.txt")
	if got != "/opt/sub/file.txt" {
		t.Fatalf("joinRemoteRoot = %q", got)
	}
	got = joinRemoteRoot("/opt", "")
	if got != "/opt" {
		t.Fatalf("joinRemoteRoot empty rel = %q", got)
	}
}

func TestMergeDirKeys(t *testing.T) {
	a := map[string]struct{}{"x": {}, "y": {}}
	b := map[string]struct{}{"y": {}, "z": {}}
	got := mergeDirKeys(a, b)
	expected := []string{"x", "y", "z"}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("mergeDirKeys = %v, want %v", got, expected)
	}
}

func TestMergeFileKeys(t *testing.T) {
	a := map[string]fileState{"a.txt": {}, "c.txt": {}}
	b := map[string]fileState{"b.txt": {}, "c.txt": {}}
	got := mergeFileKeys(a, b)
	expected := []string{"a.txt", "b.txt", "c.txt"}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("mergeFileKeys = %v, want %v", got, expected)
	}
}

func TestSSHPayloadParsing(t *testing.T) {
	jsonData := `{
		"host": "192.168.1.1",
		"port": 22,
		"user": "root",
		"password": "secret",
		"action": "cmd",
		"cmd": "date",
		"timeout": "30s",
		"cmd_timeout": "60s"
	}`
	var p SSHPayload
	if err := json.Unmarshal([]byte(jsonData), &p); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if p.Host != "192.168.1.1" {
		t.Fatalf("host = %q", p.Host)
	}
	if p.Port != 22 {
		t.Fatalf("port = %d", p.Port)
	}
	if p.User != "root" {
		t.Fatalf("user = %q", p.User)
	}
	if p.Action != "cmd" {
		t.Fatalf("action = %q", p.Action)
	}
}

func TestSSHPayloadWithArrays(t *testing.T) {
	jsonData := `{
		"host": "1.2.3.4",
		"user": "root",
		"password": "x",
		"action": "sync",
		"direction": "push",
		"recursive": true,
		"include": ["dist/**", "configs/*.json"],
		"exclude": ["**/*.map"]
	}`
	var p SSHPayload
	if err := json.Unmarshal([]byte(jsonData), &p); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(p.Include) != 2 || p.Include[0] != "dist/**" {
		t.Fatalf("include = %v", p.Include)
	}
	if len(p.Exclude) != 1 || p.Exclude[0] != "**/*.map" {
		t.Fatalf("exclude = %v", p.Exclude)
	}
}

func TestDeployPlanParsing(t *testing.T) {
	dir := t.TempDir()
	planPath := filepath.Join(dir, "deploy.json")
	content := `{"steps":[{"name":"mkdir logs","type":"mkdir","remote_path":"/opt/app/logs"},{"type":"cmd","cmd":"echo ok","timeout":"10s"}]}`
	os.WriteFile(planPath, []byte(content), 0644)

	var plan deployPlan
	data, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("read plan: %v", err)
	}
	if err := json.Unmarshal(data, &plan); err != nil {
		t.Fatalf("parse plan: %v", err)
	}
	if len(plan.Steps) != 2 {
		t.Fatalf("plan steps = %d, want 2", len(plan.Steps))
	}
	if plan.Steps[0].Name != "mkdir logs" || plan.Steps[1].Timeout != "10s" {
		t.Fatalf("unexpected plan contents: %+v", plan.Steps)
	}
}

func TestSSHHandlerMissingParams(t *testing.T) {
	h := &SSHHandler{}
	resp := h.Handle(`{}`, "")
	if resp.Ok {
		t.Fatal("should fail with missing params")
	}

	resp = h.Handle(`{"host":"1.2.3.4","user":"root"}`, "")
	if resp.Ok {
		t.Fatal("should fail without auth")
	}

	resp = h.Handle(`{"host":"1.2.3.4","user":"root","password":"x"}`, "")
	if resp.Ok {
		t.Fatal("should fail connecting to non-existent host")
	}
}

func TestSyncTreePushDryRun(t *testing.T) {
	now := time.Now()
	source := &syncTree{
		Root:      "/local",
		RootIsDir: true,
		Files: map[string]fileState{
			"app.js":    {Size: 100, ModTime: now},
			"config.js": {Size: 50, ModTime: now},
		},
		Dirs: map[string]struct{}{
			"assets": {},
		},
	}
	target := &syncTree{
		Root:      "/remote",
		RootIsDir: true,
		Files: map[string]fileState{
			"config.js": {Size: 50, ModTime: now},
			"old.js":    {Size: 200, ModTime: now},
		},
		Dirs: map[string]struct{}{
			"old_dir": {},
		},
	}

	result, err := syncTreePush(nil, source, target, "/remote", true, true)
	if err != nil {
		t.Fatalf("syncTreePush dry-run error: %v", err)
	}
	if m, ok := result["uploaded"].(int); !ok || m != 1 {
		t.Fatalf("uploaded = %v, want 1 (only app.js is new)", result["uploaded"])
	}
	if m, ok := result["deleted"].(int); !ok || m != 2 {
		t.Fatalf("deleted = %v, want 2 (old.js + old_dir)", result["deleted"])
	}
}

func TestSyncTreeBidirectionalDryRun(t *testing.T) {
	now := time.Now()
	localTree := &syncTree{
		Root:      "/local",
		RootIsDir: true,
		Files: map[string]fileState{
			"only_local.txt":  {Size: 10, ModTime: now},
			"conflict.txt":    {Size: 10, ModTime: now.Add(-time.Hour)},
		},
		Dirs: map[string]struct{}{},
	}
	remoteTree := &syncTree{
		Root:      "/remote",
		RootIsDir: true,
		Files: map[string]fileState{
			"only_remote.txt": {Size: 20, ModTime: now},
			"conflict.txt":    {Size: 30, ModTime: now},
		},
		Dirs: map[string]struct{}{},
	}

	result, err := syncTreeBidirectional(nil, "/local", "/remote", localTree, remoteTree, true, "newer_wins")
	if err != nil {
		t.Fatalf("syncTreeBidirectional dry-run error: %v", err)
	}
	if m, ok := result["uploaded"].(int); !ok || m != 1 {
		t.Fatalf("uploaded = %v, want 1 (only_local + conflict with newer_wins=remote)", result["uploaded"])
	}
	if m, ok := result["downloaded"].(int); !ok || m != 2 {
		t.Fatalf("downloaded = %v, want 2 (only_remote + conflict resolved to download)", result["downloaded"])
	}
}

func TestSyncTreeBidirectionalFailOnConflict(t *testing.T) {
	now := time.Now()
	localTree := &syncTree{
		Root:      "/local",
		RootIsDir: true,
		Files: map[string]fileState{
			"conflict.txt": {Size: 10, ModTime: now},
		},
		Dirs: map[string]struct{}{},
	}
	remoteTree := &syncTree{
		Root:      "/remote",
		RootIsDir: true,
		Files: map[string]fileState{
			"conflict.txt": {Size: 30, ModTime: now},
		},
		Dirs: map[string]struct{}{},
	}

	result, err := syncTreeBidirectional(nil, "/local", "/remote", localTree, remoteTree, true, "fail_on_conflict")
	if err != nil {
		t.Fatalf("syncTreeBidirectional error: %v", err)
	}
	if m, ok := result["conflicts"].(int); !ok || m != 1 {
		t.Fatalf("conflicts = %v, want 1", result["conflicts"])
	}
}
