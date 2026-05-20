package repl

import (
	"fmt"
	"strings"

	"golang.org/x/term"
	"os"

	"github.com/topxeq/xxaitk/internal/handler"
	"github.com/topxeq/xxaitk/internal/script"
)

type REPL struct {
	debug   bool
	history []string
	version string
}

func New(debug bool) *REPL {
	return &REPL{debug: debug, version: "0.4.0"}
}

func NewWithVersion(debug bool, ver string) *REPL {
	return &REPL{debug: debug, version: ver}
}

func (r *REPL) Run() {
	fmt.Printf("aitk v%s | type .help for help\n\n", r.version)

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		r.runSimple()
		return
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	t := term.NewTerminal(os.Stdin, "aitk> ")

	for {
		line, err := t.ReadLine()
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		r.history = append(r.history, line)

		term.Restore(int(os.Stdin.Fd()), oldState)

		if strings.HasPrefix(line, ".") {
			if r.handleDotCommand(line) {
				return
			}
		} else if strings.HasPrefix(line, ":") {
			r.handleColonCommand(line)
		} else {
			r.executeScript(line)
		}

		oldState, _ = term.MakeRaw(int(os.Stdin.Fd()))
		t = term.NewTerminal(os.Stdin, "aitk> ")
	}
}

func (r *REPL) runSimple() {
	var buf [4096]byte
	for {
		fmt.Print("aitk> ")
		n, _ := os.Stdin.Read(buf[:])
		line := strings.TrimSpace(string(buf[:n]))
		if line == "" {
			continue
		}
		r.history = append(r.history, line)
		if strings.HasPrefix(line, ".") {
			if r.handleDotCommand(line) {
				return
			}
		} else if strings.HasPrefix(line, ":") {
			r.handleColonCommand(line)
		} else {
			r.executeScript(line)
		}
	}
}

func (r *REPL) handleDotCommand(cmd string) bool {
	cmd = strings.TrimSpace(cmd)
	switch cmd {
	case ".help":
		fmt.Println("  .help        Show this help")
		fmt.Println("  .prefixes    List all operation prefixes")
		fmt.Println("  .builtins    List all script builtins")
		fmt.Println("  .debug on    Enable debug mode")
		fmt.Println("  .debug off   Disable debug mode")
		fmt.Println("  .history     Show command history")
		fmt.Println("  .quit        Exit REPL")
		fmt.Println()
		fmt.Println("  :<prefix> <args>   Execute prefix command (no hex encoding)")
		fmt.Println("  <script>           Execute script statement")
	case ".prefixes":
		for _, p := range handler.KnownPrefixes() {
			fmt.Printf("  %s\n", p)
		}
	case ".builtins":
		fmt.Println("  Script builtins: str_*, math_*, list_*, map_*, json_*,")
		fmt.Println("  io_*, net_*, os_*, time_*, log_*, type_*, conv_*")
	case ".debug on":
		r.debug = true
		fmt.Println("  Debug mode enabled")
	case ".debug off":
		r.debug = false
		fmt.Println("  Debug mode disabled")
	case ".history":
		for i, h := range r.history {
			fmt.Printf("  %4d  %s\n", i+1, h)
		}
	case ".quit":
		return true
	default:
		fmt.Printf("  Unknown command: %s\n", cmd)
	}
	return false
}

func (r *REPL) handleColonCommand(line string) {
	parts := strings.SplitN(line[1:], " ", 2)
	if len(parts) == 0 {
		fmt.Println("  Error: empty prefix command")
		return
	}

	prefix := strings.ToUpper(parts[0])
	arg := ""
	if len(parts) > 1 {
		arg = parts[1]
	}

	h, ok := handler.Get(prefix)
	if !ok {
		fmt.Printf("  Unknown prefix: %s\n", prefix)
		return
	}

	resp := h.Handle(arg, "repl")
	resp.Print()
}

func (r *REPL) executeScript(line string) {
	script.PrintCallback = func(s string) {
		fmt.Println(s)
	}
	h := &handler.ScriptHandler{}
	resp := h.Handle(line, "repl")
	if resp.Ok {
		if data, ok := resp.Data.(*handler.ScriptResult); ok {
			if data.RawResult != nil {
				switch v := data.RawResult.(type) {
				case script.NilObject:
				case script.BoolObject:
					fmt.Printf("  => %s\n", v.Inspect())
				case script.IntObject:
					fmt.Printf("  => %s\n", v.Inspect())
				case script.FloatObject:
					fmt.Printf("  => %s\n", v.Inspect())
				case script.StringObject:
					fmt.Printf("  => %s\n", v.Inspect())
				case script.ListObject:
					fmt.Printf("  => %s\n", v.Inspect())
				case script.MapObject:
					fmt.Printf("  => %s\n", v.Inspect())
				default:
					fmt.Printf("  => %s\n", v.Inspect())
				}
			}
		}
	} else {
		resp.Print()
	}
}
