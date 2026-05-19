package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/topxeq/xxaitk/internal/output"
)

type DiffHandler struct{}

type DiffPayload struct {
	SourceA   string `json:"source_a"`
	SourceB   string `json:"source_b"`
	FileA     string `json:"file_a"`
	FileB     string `json:"file_b"`
	ContentA  string `json:"content_a"`
	ContentB  string `json:"content_b"`
	Context   int    `json:"context,omitempty"`
}

type DiffLine struct {
	Type  string `json:"type"`
	LineA int    `json:"line_a,omitempty"`
	LineB int    `json:"line_b,omitempty"`
	Text  string `json:"text"`
}

type DiffResult struct {
	Same    bool       `json:"same"`
	Lines   []DiffLine `json:"diff_lines"`
	Adds    int        `json:"adds"`
	Dels    int        `json:"dels"`
	Changes int        `json:"changes"`
}

func (h *DiffHandler) Handle(data string, source string) *output.Response {
	start := time.Now()

	payload := h.parsePayload(data)

	textA, textB, err := h.resolveInputs(payload)
	if err != nil {
		return output.Fail("diff", source, "DIFF_INPUT_ERROR", err.Error(), "", start)
	}

	linesA := strings.Split(textA, "\n")
	linesB := strings.Split(textB, "\n")

	if textA == textB {
		return output.Success("diff", source, &DiffResult{Same: true}, start)
	}

	diffLines := lcsDiff(linesA, linesB)

	adds, dels, changes := 0, 0, 0
	for _, dl := range diffLines {
		switch dl.Type {
		case "add":
			adds++
		case "del":
			dels++
		case "change":
			changes++
		}
	}

	return output.Success("diff", source, &DiffResult{
		Same:    false,
		Lines:   diffLines,
		Adds:    adds,
		Dels:    dels,
		Changes: changes,
	}, start)
}

func (h *DiffHandler) resolveInputs(p *DiffPayload) (string, string, error) {
	var textA, textB string

	if p.FileA != "" {
		data, err := os.ReadFile(p.FileA)
		if err != nil {
			return "", "", fmt.Errorf("failed to read file_a: %w", err)
		}
		textA = string(data)
	} else if p.ContentA != "" {
		textA = p.ContentA
	}

	if p.FileB != "" {
		data, err := os.ReadFile(p.FileB)
		if err != nil {
			return "", "", fmt.Errorf("failed to read file_b: %w", err)
		}
		textB = string(data)
	} else if p.ContentB != "" {
		textB = p.ContentB
	}

	return textA, textB, nil
}

func lcsDiff(a, b []string) []DiffLine {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	var result []DiffLine
	i, j := m, n
	var temp []DiffLine

	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			temp = append(temp, DiffLine{Type: "same", LineA: i, LineB: j, Text: a[i-1]})
			i--
			j--
		} else if dp[i-1][j] >= dp[i][j-1] {
			temp = append(temp, DiffLine{Type: "del", LineA: i, Text: a[i-1]})
			i--
		} else {
			temp = append(temp, DiffLine{Type: "add", LineB: j, Text: b[j-1]})
			j--
		}
	}
	for i > 0 {
		temp = append(temp, DiffLine{Type: "del", LineA: i, Text: a[i-1]})
		i--
	}
	for j > 0 {
		temp = append(temp, DiffLine{Type: "add", LineB: j, Text: b[j-1]})
		j--
	}

	for k := len(temp) - 1; k >= 0; k-- {
		result = append(result, temp[k])
	}

	return result
}

func (h *DiffHandler) parsePayload(data string) *DiffPayload {
	payload := &DiffPayload{}
	trimmed := strings.TrimSpace(data)
	if len(trimmed) == 0 {
		return payload
	}
	if trimmed[0] == '{' {
		json.Unmarshal([]byte(trimmed), payload)
	}
	return payload
}
