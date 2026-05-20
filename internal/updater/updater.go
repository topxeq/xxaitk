package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const DefaultBaseURL = "https://github.com/topxeq/xxaitk/releases"

type Updater struct {
	CurrentVersion string
	BaseURL        string
	Client         *http.Client
	ExePath        string
}

func New(currentVersion string) *Updater {
	return &Updater{
		CurrentVersion: currentVersion,
		BaseURL:        DefaultBaseURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (u *Updater) getExePath() (string, error) {
	if u.ExePath != "" {
		return u.ExePath, nil
	}
	return os.Executable()
}

func (u *Updater) CheckUpdate() (latestVersion string, hasUpdate bool, err error) {
	checkClient := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	url := u.BaseURL + "/latest"
	resp, err := checkClient.Get(url)
	if err != nil {
		return "", false, fmt.Errorf("check update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusMovedPermanently {
		return "", false, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return "", false, fmt.Errorf("no Location header in redirect")
	}

	parts := strings.Split(location, "/")
	tag := parts[len(parts)-1]
	latestVersion = strings.TrimPrefix(tag, "v")

	if latestVersion == u.CurrentVersion {
		return latestVersion, false, nil
	}

	return latestVersion, true, nil
}

func (u *Updater) downloadArchive(url string) (string, error) {
	resp, err := u.Client.Get(url)
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: status %d", resp.StatusCode)
	}

	tmpDir, err := os.MkdirTemp("", "aitk-update-*")
	if err != nil {
		return "", fmt.Errorf("temp dir: %w", err)
	}

	archiveName, err := PlatformArchiveName()
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}
	archivePath := filepath.Join(tmpDir, archiveName+ArchiveExt())

	f, err := os.Create(archivePath)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("create archive file: %w", err)
	}

	_, err = io.Copy(f, resp.Body)
	f.Close()
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("write archive: %w", err)
	}

	return archivePath, nil
}

func (u *Updater) extractBinary(archivePath string) (string, error) {
	binaryName := BinaryName()

	if IsWindows() {
		return u.extractFromZip(archivePath, binaryName)
	}
	return u.extractFromTarGz(archivePath, binaryName)
}

func (u *Updater) extractFromTarGz(archivePath, binaryName string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("gzip: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("tar: %w", err)
		}

		base := filepath.Base(hdr.Name)
		if base == binaryName {
			tmpDir := filepath.Dir(archivePath)
			outPath := filepath.Join(tmpDir, binaryName+".new")
			out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return "", err
			}
			out.Close()
			return outPath, nil
		}
	}

	return "", fmt.Errorf("binary '%s' not found in archive", binaryName)
}

func (u *Updater) extractFromZip(archivePath, binaryName string) (string, error) {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer zr.Close()

	for _, f := range zr.File {
		base := filepath.Base(f.Name)
		if base == binaryName {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()

			tmpDir := filepath.Dir(archivePath)
			outPath := filepath.Join(tmpDir, binaryName+".new")
			out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(out, rc); err != nil {
				out.Close()
				return "", err
			}
			out.Close()
			return outPath, nil
		}
	}

	return "", fmt.Errorf("binary '%s' not found in archive", binaryName)
}

func (u *Updater) replaceBinary(newBinaryPath string) error {
	currentExe, err := u.getExePath()
	if err != nil {
		return fmt.Errorf("get current exe: %w", err)
	}

	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		return fmt.Errorf("resolve symlink: %w", err)
	}

	oldInfo, err := os.Stat(currentExe)
	if err != nil {
		return fmt.Errorf("stat current exe: %w", err)
	}

	if err := os.Chmod(newBinaryPath, oldInfo.Mode()); err != nil {
		return fmt.Errorf("chmod new binary: %w", err)
	}

	if IsWindows() {
		oldBackup := currentExe + ".old"
		os.Remove(oldBackup)
		if err := os.Rename(currentExe, oldBackup); err != nil {
			return fmt.Errorf("rename old binary: %w", err)
		}
		if err := os.Rename(newBinaryPath, currentExe); err != nil {
			os.Rename(oldBackup, currentExe)
			return fmt.Errorf("rename new binary: %w", err)
		}
		os.Remove(oldBackup)
	} else {
		if err := os.Rename(newBinaryPath, currentExe); err != nil {
			return fmt.Errorf("replace binary: %w", err)
		}
	}

	return nil
}

func (u *Updater) Run() error {
	fmt.Printf("Current version: v%s\n", u.CurrentVersion)
	fmt.Println("Checking for updates...")

	latestVersion, hasUpdate, err := u.CheckUpdate()
	if err != nil {
		return fmt.Errorf("check update: %w", err)
	}

	if !hasUpdate {
		fmt.Printf("Already up to date (v%s)\n", latestVersion)
		return nil
	}

	fmt.Printf("Latest version:  v%s\n", latestVersion)

	archiveName, err := PlatformArchiveName()
	if err != nil {
		return err
	}
	downloadURL := fmt.Sprintf("%s/download/v%s/%s%s", u.BaseURL, latestVersion, archiveName, ArchiveExt())

	fmt.Printf("Downloading %s...\n", filepath.Base(downloadURL))

	archivePath, err := u.downloadArchive(downloadURL)
	if err != nil {
		return err
	}
	defer os.RemoveAll(filepath.Dir(archivePath))

	fmt.Println("Extracting...")
	newBinaryPath, err := u.extractBinary(archivePath)
	if err != nil {
		return err
	}

	fmt.Println("Updating aitk binary...")
	if err := u.replaceBinary(newBinaryPath); err != nil {
		return err
	}

	fmt.Printf("Updated aitk v%s → v%s\n", u.CurrentVersion, latestVersion)
	return nil
}

func CleanupOldBinary() {
	if runtime.GOOS != "windows" {
		return
	}
	exe, err := os.Executable()
	if err != nil {
		return
	}
	oldBackup := exe + ".old"
	os.Remove(oldBackup)
}
