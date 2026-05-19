package repl

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/topxeq/xxaitk/internal/handler"
)

const version = "0.1.0"

type REPL struct {
	debug bool
}

func New(debug bool) *REPL {
	return &REPL{debug: debug}
}

func (r *REPL) Run() {
	fmt.Printf("aitk v%s | type .help for help\n\n", version)

	reader := bufio.NewReader(os.Stdin)
	history := []string{}
	_ = history

	for {
		fmt.Print("aitk> ")

		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		history = append(history, line)

		if strings.HasPrefix(line, ".") {
			if r.handleDotCommand(line) {
				break
			}
			continue
		}

		if strings.HasPrefix(line, ":") {
			r.handleColonCommand(line)
			continue
		}

		r.executeScript(line)
	}
}

func (r *REPL) handleDotCommand(line string) bool {
	cmd := strings.TrimSpace(line)
	switch cmd {
	case ".help":
		fmt.Println("  .help        Show this help")
		fmt.Println("  .prefixes    List all operation prefixes")
		fmt.Println("  .builtins    List all script builtins")
		fmt.Println("  .debug on    Enable debug mode")
		fmt.Println("  .debug off   Disable debug mode")
		fmt.Println("  .quit        Exit REPL")
		fmt.Println()
		fmt.Println("  :<prefix> <args>   Execute prefix command (no hex encoding needed)")
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
	h := &handler.ScriptHandler{}
	resp := h.Handle(line, "repl")
	resp.Print()
}

func readLine(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}
