package datasource

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/topxeq/xxaitk/internal/hexcodec"
)

type ResolvedData struct {
	Content string
	Source  string
	Err     error
}

func Resolve(source string, hexData string) *ResolvedData {
	switch source {
	case "INLINE":
		decoded, err := hexcodec.DecodeString(hexData)
		if err != nil {
			return &ResolvedData{Err: fmt.Errorf("hex decode error: %w", err)}
		}
		return &ResolvedData{Content: decoded, Source: "inline"}
	case "FILE":
		path, err := hexcodec.DecodeString(hexData)
		if err != nil {
			return &ResolvedData{Err: fmt.Errorf("hex decode error for file path: %w", err)}
		}
		return resolveFile(path)
	case "URL":
		urlStr, err := hexcodec.DecodeString(hexData)
		if err != nil {
			return &ResolvedData{Err: fmt.Errorf("hex decode error for URL: %w", err)}
		}
		return resolveURL(urlStr)
	default:
		return &ResolvedData{Err: fmt.Errorf("unknown source type: %s", source)}
	}
}

func resolveFile(path string) *ResolvedData {
	data, err := os.ReadFile(path)
	if err != nil {
		return &ResolvedData{Source: "file", Err: fmt.Errorf("failed to read file %s: %w", path, err)}
	}
	return &ResolvedData{Content: string(data), Source: "file"}
}

func resolveURL(urlStr string) *ResolvedData {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(urlStr)
	if err != nil {
		return &ResolvedData{Source: "url", Err: fmt.Errorf("failed to fetch URL %s: %w", urlStr, err)}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ResolvedData{Source: "url", Err: fmt.Errorf("failed to read response from %s: %w", urlStr, err)}
	}

	if resp.StatusCode >= 400 {
		return &ResolvedData{
			Source: "url",
			Err:    fmt.Errorf("HTTP %d from %s: %s", resp.StatusCode, urlStr, string(body[:min(len(body), 512)])),
		}
	}

	return &ResolvedData{Content: string(body), Source: "url"}
}

func ResolveJSONSource(sourceType string, sourcePath string, sourceURL string) (string, error) {
	switch sourceType {
	case "file":
		rd := resolveFile(sourcePath)
		if rd.Err != nil {
			return "", rd.Err
		}
		return rd.Content, nil
	case "url":
		rd := resolveURL(sourceURL)
		if rd.Err != nil {
			return "", rd.Err
		}
		return rd.Content, nil
	default:
		return "", fmt.Errorf("unknown source type: %s", sourceType)
	}
}
