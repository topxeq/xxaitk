package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/topxeq/xxaitk/internal/dispatcher"
	"github.com/topxeq/xxaitk/internal/handler"
	"github.com/topxeq/xxaitk/internal/repl"
	"github.com/topxeq/xxaitk/internal/scriptlib"
)

var version = "0.5.0"

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		r := repl.NewWithVersion(false, version)
		r.Run()
		return
	}

	if args[0] == "lib" {
		handleLibCommand(args[1:])
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
		r := repl.NewWithVersion(debug, version)
		r.Run()
		return
	}

	registerHandlers()

	d := dispatcher.New(debug)
	d.Dispatch(remaining[0])
}

func handleLibCommand(args []string) {
	if len(args) == 0 {
		printLibHelp()
		return
	}

	switch args[0] {
	case "list", "ls":
		installed, err := scriptlib.ListInstalled()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if len(installed) == 0 {
			fmt.Println("No libraries installed.")
			fmt.Println("Use 'aitk lib search' to browse available libraries.")
			return
		}
		fmt.Println("Installed libraries:")
		for _, name := range installed {
			fmt.Printf("  %s\n", name)
		}

	case "search":
		reg, err := scriptlib.ListRemote()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if len(reg.Libraries) == 0 {
			fmt.Println("No libraries available in registry.")
			return
		}
		fmt.Println("Available libraries:")
		data, _ := json.MarshalIndent(reg.Libraries, "", "  ")
		fmt.Println(string(data))

	case "get":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: aitk lib get <name>")
			os.Exit(1)
		}
		name := args[1]
		fmt.Printf("Downloading library '%s'...\n", name)
		if err := scriptlib.Get(name); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Library '%s' installed successfully.\n", name)

	case "remove", "rm":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: aitk lib remove <name>")
			os.Exit(1)
		}
		name := args[1]
		if err := scriptlib.Remove(name); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Library '%s' removed.\n", name)

	default:
		printLibHelp()
	}
}

func printLibHelp() {
	fmt.Println("Usage: aitk lib <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list              List installed libraries")
	fmt.Println("  search            Search remote library registry")
	fmt.Println("  get <name>        Download and install a library")
	fmt.Println("  remove <name>     Remove an installed library")
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
	handler.Register("GIT", &handler.GitHandler{})
	handler.Register("PORT", &handler.PortHandler{})
	handler.Register("NETDOWNLOAD", &handler.NetDownloadHandler{})
	handler.Register("CAPABILITIES", &handler.CapabilitiesHandler{})
}

func printHelp() {
	fmt.Printf("aitk v%s - AI Agent Toolkit\n\n", version)
	fmt.Println("Usage:")
	fmt.Println("  aitk                              Enter REPL mode")
	fmt.Println("  aitk <PREFIX>[_SOURCE]_<HEXDATA>  Execute command")
	fmt.Println("  aitk lib <command>               Manage script libraries")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --version, -v   Print version")
	fmt.Println("  --help, -h      Print this help")
	fmt.Println("  --debug         Enable debug output")
	fmt.Println()
	fmt.Println("Prefixes:")
	fmt.Println("  Execution:    SHELL SCRIPT EVAL")
	fmt.Println("  HTTP:         HTTPGET HTTPPOST HTTPPUT HTTPPATCH HTTPDELETE")
	fmt.Println("  Filesystem:   FILE READFILE WRITEFILE LISTDIR DELETE")
	fmt.Println("  Network:      PING NETDOWNLOAD PORT")
	fmt.Println("  Encoding:     DECODE ENCODE B64ENC B64DEC URLENC URLDEC")
	fmt.Println("  Crypto:       HASH")
	fmt.Println("  VCS:          GIT")
	fmt.Println("  Process:      PROCESS")
	fmt.Println("  Diff:         DIFF")
	fmt.Println("  Archive:      ARCHIVE")
	fmt.Println("  Database:     SQL")
	fmt.Println("  System:       INFO CAPABILITIES")
	fmt.Println()
	fmt.Println("Source modifiers:")
	fmt.Println("  FILE_          Read command data from file path")
	fmt.Println("  URL_           Read command data from URL")
}
