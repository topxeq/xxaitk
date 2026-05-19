package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/topxeq/xxaitk/internal/dispatcher"
	"github.com/topxeq/xxaitk/internal/handler"
	"github.com/topxeq/xxaitk/internal/repl"
)

var version = "0.2.0"

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		r := repl.New(false)
		r.Run()
		return
	}

	debug := false

	remaining := []string{}
	for _, arg := range args {
		switch arg {
		case "--version", "-v":
			fmt.Printf("aitk v%s\n", version)
			return
		case "--help", "-h":
			printHelp()
			return
		case "--debug":
			debug = true
		default:
			if !strings.HasPrefix(arg, "-") {
				remaining = append(remaining, arg)
			}
		}
	}

	if len(remaining) == 0 {
		r := repl.New(debug)
		r.Run()
		return
	}

	registerHandlers()

	d := dispatcher.New(debug)
	d.Dispatch(remaining[0])
}

func registerHandlers() {
	handler.Register("SHELL", &handler.ShellHandler{})
	handler.Register("SCRIPT", &handler.ScriptHandler{})
	handler.Register("EVAL", &handler.EvalHandler{})
	handler.Register("HTTPGET", &handler.HTTPGetHandler{})
	handler.Register("HTTPPOST", &handler.HTTPPostHandler{})
	handler.Register("HTTPPUT", &handler.HTTPMethodHandler{Method: "PUT"})
	handler.Register("HTTPPATCH", &handler.HTTPMethodHandler{Method: "PATCH"})
	handler.Register("HTTPDELETE", &handler.HTTPMethodHandler{Method: "DELETE"})
	handler.Register("FILE", &handler.FileHandler{})
	handler.Register("READFILE", &handler.FileHandler{})
	handler.Register("WRITEFILE", &handler.WriteFileHandler{})
	handler.Register("LISTDIR", &handler.ListDirHandler{})
	handler.Register("DELETE", &handler.DeleteHandler{})
	handler.Register("INFO", &handler.InfoHandler{})
	handler.Register("DECODE", &handler.DecodeHandler{})
	handler.Register("ENCODE", &handler.EncodeHandler{})
	handler.Register("B64ENC", &handler.B64EncHandler{})
	handler.Register("B64DEC", &handler.B64DecHandler{})
	handler.Register("URLENC", &handler.URLEncHandler{})
	handler.Register("URLDEC", &handler.URLDecHandler{})
	handler.Register("PING", &handler.PingHandler{})
	handler.Register("HASH", &handler.HashHandler{})
	handler.Register("PROCESS", &handler.ProcessHandler{})
	handler.Register("DIFF", &handler.DiffHandler{})
	handler.Register("ARCHIVE", &handler.ArchiveHandler{})
	handler.Register("SQL", &handler.SQLHandler{})
	handler.Register("PROCESS", &handler.ProcessHandler{})
	handler.Register("DIFF", &handler.DiffHandler{})
	handler.Register("HASH", &handler.HashHandler{})
}

func printHelp() {
	fmt.Printf("aitk v%s - AI Agent Toolkit\n\n", version)
	fmt.Println("Usage:")
	fmt.Println("  aitk                              Enter REPL mode")
	fmt.Println("  aitk <PREFIX>[_SOURCE]_<HEXDATA>  Execute command")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --version, -v   Print version")
	fmt.Println("  --help, -h      Print this help")
	fmt.Println("  --debug         Enable debug output")
	fmt.Println()
	fmt.Println("Prefixes:")
	fmt.Println("  SHELL      Execute shell command")
	fmt.Println("  SCRIPT     Execute built-in script")
	fmt.Println("  EVAL       Evaluate expression (single-line SCRIPT)")
	fmt.Println("  HTTPGET    HTTP GET request")
	fmt.Println("  HTTPPOST   HTTP POST request")
	fmt.Println("  FILE       Read file (alias: READFILE)")
	fmt.Println("  WRITEFILE  Write file")
	fmt.Println("  LISTDIR    List directory")
	fmt.Println("  DELETE     Delete file/directory")
	fmt.Println("  INFO       System information")
	fmt.Println("  DECODE     Hex decode")
	fmt.Println("  ENCODE     Hex encode")
	fmt.Println("  B64ENC     Base64 encode")
	fmt.Println("  B64DEC     Base64 decode")
	fmt.Println("  URLENC     URL encode")
	fmt.Println("  URLDEC     URL decode")
	fmt.Println("  PING       Network connectivity test")
	fmt.Println()
	fmt.Println("Source modifiers:")
	fmt.Println("  FILE_      Read command data from file path")
	fmt.Println("  URL_       Read command data from URL")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  aitk SHELL_6c73202d6c61                    # ls -la")
	fmt.Println("  aitk SHELL_FILE_2f746d702f636d642e7368     # run commands from file")
	fmt.Println("  aitk FILE_2f6574632f686f737473              # read /etc/hosts")
}
