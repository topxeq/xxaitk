package handler

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/topxeq/xxaitk/internal/output"
)

type ArchiveHandler struct{}

type ArchivePayload struct {
	Action   string   `json:"action"`
	Format   string   `json:"format"`
	Source   string   `json:"source"`
	Target   string   `json:"target"`
	Files    []string `json:"files"`
	Dir      string   `json:"dir"`
	Overwrite bool    `json:"overwrite,omitempty"`
}

type ArchiveResult struct {
	Action  string   `json:"action"`
	Format  string   `json:"format"`
	Source  string   `json:"source,omitempty"`
	Target  string   `json:"target,omitempty"`
	Files   []string `json:"files,omitempty"`
	Count   int      `json:"count,omitempty"`
	Size    int64    `json:"size,omitempty"`
}

func (h *ArchiveHandler) Handle(data string, source string) *output.Response {
	start := time.Now()

	payload := h.parsePayload(data)
	action := strings.ToLower(payload.Action)
	if action == "" {
		action = "list"
	}

	format := strings.ToLower(payload.Format)
	if format == "" {
		if payload.Source != "" {
			format = detectFormat(payload.Source)
		} else {
			format = "zip"
		}
	}

	switch action {
	case "pack":
		return h.pack(payload, format, source, start)
	case "unpack":
		return h.unpack(payload, format, source, start)
	case "list":
		return h.list(payload, format, source, start)
	default:
		return output.Fail("archive", source, "ARCHIVE_UNKNOWN_ACTION",
			fmt.Sprintf("unknown action: %s (use pack, unpack, list)", action), "", start)
	}
}

func (h *ArchiveHandler) pack(p *ArchivePayload, format string, source string, start time.Time) *output.Response {
	if p.Target == "" {
		return output.Fail("archive", source, "ARCHIVE_NO_TARGET", "no target archive path specified", "", start)
	}
	if len(p.Files) == 0 && p.Dir == "" {
		return output.Fail("archive", source, "ARCHIVE_NO_FILES", "no files or directory specified", "", start)
	}

	files := p.Files
	if p.Dir != "" {
		filepath.Walk(p.Dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			files = append(files, path)
			return nil
		})
	}

	switch format {
	case "zip":
		return h.packZip(files, p, source, start)
	case "tar", "tar.gz", "tgz":
		return h.packTar(files, p, format, source, start)
	default:
		return output.Fail("archive", source, "ARCHIVE_UNSUPPORTED_FORMAT",
			fmt.Sprintf("unsupported format: %s (use zip, tar, tar.gz, tgz)", format), "", start)
	}
}

func (h *ArchiveHandler) packZip(files []string, p *ArchivePayload, source string, start time.Time) *output.Response {
	f, err := os.Create(p.Target)
	if err != nil {
		return output.Fail("archive", source, "ARCHIVE_CREATE_ERROR", err.Error(), "", start)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	var written []string

	for _, path := range files {
		if err := addFileToZip(w, path); err != nil {
			continue
		}
		written = append(written, path)
	}
	w.Close()

	info, _ := f.Stat()

	return output.Success("archive", source, &ArchiveResult{
		Action: "pack",
		Format: "zip",
		Target: p.Target,
		Files:  written,
		Count:  len(written),
		Size:   info.Size(),
	}, start)
}

func (h *ArchiveHandler) packTar(files []string, p *ArchivePayload, format string, source string, start time.Time) *output.Response {
	f, err := os.Create(p.Target)
	if err != nil {
		return output.Fail("archive", source, "ARCHIVE_CREATE_ERROR", err.Error(), "", start)
	}
	defer f.Close()

	var w io.Writer = f
	var gw *gzip.Writer
	if format == "tar.gz" || format == "tgz" {
		gw = gzip.NewWriter(f)
		w = gw
	}

	tw := tar.NewWriter(w)
	var written []string

	for _, path := range files {
		if err := addFileToTar(tw, path); err != nil {
			continue
		}
		written = append(written, path)
	}
	tw.Close()
	if gw != nil {
		gw.Close()
	}

	info, _ := f.Stat()

	return output.Success("archive", source, &ArchiveResult{
		Action: "pack",
		Format: format,
		Target: p.Target,
		Files:  written,
		Count:  len(written),
		Size:   info.Size(),
	}, start)
}

func (h *ArchiveHandler) unpack(p *ArchivePayload, format string, source string, start time.Time) *output.Response {
	if p.Source == "" {
		return output.Fail("archive", source, "ARCHIVE_NO_SOURCE", "no source archive path specified", "", start)
	}
	target := p.Target
	if target == "" {
		target = "."
	}

	switch format {
	case "zip":
		return h.unpackZip(p, target, source, start)
	case "tar", "tar.gz", "tgz":
		return h.unpackTar(p, format, target, source, start)
	default:
		return output.Fail("archive", source, "ARCHIVE_UNSUPPORTED_FORMAT",
			fmt.Sprintf("unsupported format: %s", format), "", start)
	}
}

func (h *ArchiveHandler) unpackZip(p *ArchivePayload, target string, source string, start time.Time) *output.Response {
	r, err := zip.OpenReader(p.Source)
	if err != nil {
		return output.Fail("archive", source, "ARCHIVE_OPEN_ERROR", err.Error(), "", start)
	}
	defer r.Close()

	var extracted []string
	for _, f := range r.File {
		path := filepath.Join(target, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, 0755)
			continue
		}
		os.MkdirAll(filepath.Dir(path), 0755)
		rc, err := f.Open()
		if err != nil {
			continue
		}
		out, err := os.Create(path)
		if err != nil {
			rc.Close()
			continue
		}
		io.Copy(out, rc)
		out.Close()
		rc.Close()
		extracted = append(extracted, f.Name)
	}

	return output.Success("archive", source, &ArchiveResult{
		Action: "unpack",
		Format: "zip",
		Source: p.Source,
		Target: target,
		Files:  extracted,
		Count:  len(extracted),
	}, start)
}

func (h *ArchiveHandler) unpackTar(p *ArchivePayload, format string, target string, source string, start time.Time) *output.Response {
	f, err := os.Open(p.Source)
	if err != nil {
		return output.Fail("archive", source, "ARCHIVE_OPEN_ERROR", err.Error(), "", start)
	}
	defer f.Close()

	var tr *tar.Reader
	if format == "tar.gz" || format == "tgz" {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return output.Fail("archive", source, "ARCHIVE_GZIP_ERROR", err.Error(), "", start)
		}
		tr = tar.NewReader(gz)
	} else {
		tr = tar.NewReader(f)
	}

	var extracted []string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		path := filepath.Join(target, hdr.Name)
		if hdr.Typeflag == tar.TypeDir {
			os.MkdirAll(path, 0755)
			continue
		}
		os.MkdirAll(filepath.Dir(path), 0755)
		out, err := os.Create(path)
		if err != nil {
			continue
		}
		io.Copy(out, tr)
		out.Close()
		extracted = append(extracted, hdr.Name)
	}

	return output.Success("archive", source, &ArchiveResult{
		Action: "unpack",
		Format: format,
		Source: p.Source,
		Target: target,
		Files:  extracted,
		Count:  len(extracted),
	}, start)
}

func (h *ArchiveHandler) list(p *ArchivePayload, format string, source string, start time.Time) *output.Response {
	if p.Source == "" {
		return output.Fail("archive", source, "ARCHIVE_NO_SOURCE", "no source archive path specified", "", start)
	}

	switch format {
	case "zip":
		return h.listZip(p, source, start)
	case "tar", "tar.gz", "tgz":
		return h.listTar(p, format, source, start)
	default:
		return output.Fail("archive", source, "ARCHIVE_UNSUPPORTED_FORMAT",
			fmt.Sprintf("unsupported format: %s", format), "", start)
	}
}

func (h *ArchiveHandler) listZip(p *ArchivePayload, source string, start time.Time) *output.Response {
	r, err := zip.OpenReader(p.Source)
	if err != nil {
		return output.Fail("archive", source, "ARCHIVE_OPEN_ERROR", err.Error(), "", start)
	}
	defer r.Close()

	var files []string
	for _, f := range r.File {
		files = append(files, f.Name)
	}

	return output.Success("archive", source, &ArchiveResult{
		Action: "list",
		Format: "zip",
		Source: p.Source,
		Files:  files,
		Count:  len(files),
	}, start)
}

func (h *ArchiveHandler) listTar(p *ArchivePayload, format string, source string, start time.Time) *output.Response {
	f, err := os.Open(p.Source)
	if err != nil {
		return output.Fail("archive", source, "ARCHIVE_OPEN_ERROR", err.Error(), "", start)
	}
	defer f.Close()

	var tr *tar.Reader
	if format == "tar.gz" || format == "tgz" {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return output.Fail("archive", source, "ARCHIVE_GZIP_ERROR", err.Error(), "", start)
		}
		tr = tar.NewReader(gz)
	} else {
		tr = tar.NewReader(f)
	}

	var files []string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		files = append(files, hdr.Name)
	}

	return output.Success("archive", source, &ArchiveResult{
		Action: "list",
		Format: format,
		Source: p.Source,
		Files:  files,
		Count:  len(files),
	}, start)
}

func detectFormat(path string) string {
	if strings.HasSuffix(path, ".zip") {
		return "zip"
	}
	if strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".tgz") {
		return "tar.gz"
	}
	if strings.HasSuffix(path, ".tar") {
		return "tar"
	}
	return "zip"
}

func addFileToZip(w *zip.Writer, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = path
	header.Method = zip.Deflate

	writer, err := w.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, f)
	return err
}

func addFileToTar(tw *tar.Writer, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = path

	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	_, err = io.Copy(tw, f)
	return err
}

func (h *ArchiveHandler) parsePayload(data string) *ArchivePayload {
	payload := &ArchivePayload{}
	trimmed := strings.TrimSpace(data)
	if len(trimmed) == 0 {
		return payload
	}
	if trimmed[0] == '{' {
		json.Unmarshal([]byte(trimmed), payload)
	}
	return payload
}
