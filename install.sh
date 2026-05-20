#!/usr/bin/env bash
set -euo pipefail

REPO="topxeq/xxaitk"
BINARY="aitk"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

info()  { printf "${CYAN}[INFO]${RESET} %s\n" "$*"; }
warn()  { printf "${YELLOW}[WARN]${RESET} %s\n" "$*"; }
error() { printf "${RED}[ERROR]${RESET} %s\n" "$*"; }
ok()    { printf "${GREEN}[OK]${RESET} %s\n" "$*"; }

detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "Linux" ;;
        Darwin*) echo "Darwin" ;;
        MINGW*|MSYS*|CYGWIN*) echo "Windows" ;;
        FreeBSD*) echo "FreeBSD" ;;
        *) error "Unsupported OS: $(uname -s)"; exit 1 ;;
    esac
}

detect_arch() {
    local arch="$(uname -m)"
    case "$arch" in
        x86_64|amd64) echo "x86_64" ;;
        aarch64|arm64) echo "aarch64" ;;
        armv7l|armv7) echo "armv7" ;;
        armv6l|armv6) echo "armv6" ;;
        i386|i686) echo "i386" ;;
        *) error "Unsupported architecture: $arch"; exit 1 ;;
    esac
}

get_latest_version() {
    local url="https://github.com/${REPO}/releases/latest"
    local tmp
    tmp=$(curl -fsSL -o /dev/null -w '%{url_effective}' "$url" 2>/dev/null) || {
        error "Failed to check latest version. Network issue?"
        exit 1
    }
    local tag="${tmp##*/}"
    echo "${tag#v}"
}

check_existing() {
    if command -v "$BINARY" &>/dev/null; then
        local current
        current=$("$BINARY" --version 2>/dev/null | grep -oP 'v\K[0-9]+\.[0-9]+\.[0-9]+' || echo "unknown")
        info "Existing installation found: aitk v${current}"
        return 0
    fi
    return 1
}

tmpdir=""
cleanup() {
    if [ -n "$tmpdir" ] && [ -d "$tmpdir" ]; then
        rm -rf "$tmpdir"
    fi
}
trap cleanup EXIT

main() {
    printf "\n${BOLD}${CYAN}  aitk - AI Agent Toolkit Installer${RESET}\n\n"

    local os arch version archive_name archive_ext download_url
    os=$(detect_os)
    arch=$(detect_arch)

    info "OS:   ${os}"
    info "Arch: ${arch}"

    version=$(get_latest_version)
    info "Latest version: v${version}"

    check_existing

    archive_name="xxaitk_${os}_${arch}"
    if [ "$os" = "Windows" ]; then
        archive_ext=".zip"
    else
        archive_ext=".tar.gz"
    fi

    download_url="https://github.com/${REPO}/releases/download/v${version}/${archive_name}${archive_ext}"

    tmpdir=$(mktemp -d)
    local archive_path="${tmpdir}/${archive_name}${archive_ext}"

    info "Downloading ${archive_name}${archive_ext} ..."
    curl -fsSL --progress-bar -o "$archive_path" "$download_url" || {
        error "Download failed!"
        error "URL: ${download_url}"
        exit 1
    }

    info "Extracting ..."
    if [ "$os" = "Windows" ]; then
        if ! command -v unzip &>/dev/null; then
            error "unzip is required but not installed"
            exit 1
        fi
        (cd "$tmpdir" && unzip -o "$archive_path" "$BINARY.exe")
    else
        tar xzf "$archive_path" -C "$tmpdir" "$BINARY"
    fi

    local bin_name="$BINARY"
    [ "$os" = "Windows" ] && bin_name="${BINARY}.exe"
    local extracted="${tmpdir}/${bin_name}"

    if [ ! -f "$extracted" ]; then
        error "Binary not found in archive"
        exit 1
    fi
    chmod +x "$extracted"

    local target="${INSTALL_DIR}/${bin_name}"

    mkdir -p "$(dirname "$target")"
    if [ -w "$(dirname "$target")" ] 2>/dev/null; then
        mv "$extracted" "$target"
    else
        info "Installing to ${target} (requires sudo) ..."
        sudo mv "$extracted" "$target"
    fi

    ok "Installed aitk v${version} to ${target}"

    if command -v "$BINARY" &>/dev/null; then
        printf "\n${GREEN}${BOLD}  aitk v${version} installed successfully!${RESET}\n\n"
        printf "  Quick start:\n"
        printf "    aitk --help        Show help\n"
        printf "    aitk               Enter REPL\n"
        printf "    aitk update        Self-update\n\n"
    else
        warn "aitk is not in your PATH."
        warn "Add ${INSTALL_DIR} to your PATH or re-open your terminal."
    fi
}

main "$@"
