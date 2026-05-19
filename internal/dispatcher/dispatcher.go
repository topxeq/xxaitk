package dispatcher

import (
	"fmt"
	"strings"
	"time"

	"github.com/topxeq/xxaitk/internal/argparser"
	"github.com/topxeq/xxaitk/internal/datasource"
	"github.com/topxeq/xxaitk/internal/handler"
	"github.com/topxeq/xxaitk/internal/output"
)

type Dispatcher struct {
	Debug bool
}

func New(debug bool) *Dispatcher {
	return &Dispatcher{Debug: debug}
}

func (d *Dispatcher) Dispatch(arg string) {
	start := time.Now()

	parsed, err := argparser.Parse(arg)
	if err != nil {
		resp := output.Fail("unknown", "", "PARSE_ERROR", err.Error(), "", start)
		resp.Print()
		return
	}

	if d.Debug {
		fmt.Printf("[DEBUG] operation=%s source=%s hexLen=%d\n",
			parsed.Operation, parsed.Source, len(parsed.HexData))
	}

	if !handler.IsKnownPrefix(parsed.Operation) {
		resp := output.Fail(parsed.Operation, "", "UNKNOWN_PREFIX",
			fmt.Sprintf("unknown operation prefix: %s", parsed.Operation),
			fmt.Sprintf("known prefixes: %v", handler.KnownPrefixes()), start)
		resp.Print()
		return
	}

	resolved := datasource.Resolve(parsed.Source, parsed.HexData)
	if resolved.Err != nil {
		resp := output.Fail(strings.ToLower(parsed.Operation), strings.ToLower(parsed.Source),
			"SOURCE_RESOLVE_ERROR", resolved.Err.Error(), "", start)
		resp.Print()
		return
	}

	h, ok := handler.Get(parsed.Operation)
	if !ok {
		resp := output.Fail(parsed.Operation, parsed.Source, "HANDLER_NOT_FOUND",
			fmt.Sprintf("no handler registered for: %s", parsed.Operation), "", start)
		resp.Print()
		return
	}

	resp := h.Handle(resolved.Content, resolved.Source)
	resp.Print()
}
