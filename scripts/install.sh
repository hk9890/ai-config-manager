#!/usr/bin/env sh

set -eu

REPO="${AIMGR_GITHUB_REPO:-dynatrace-oss/ai-config-manager}"
REQUESTED_VERSION="${AIMGR_VERSION:-}"
INSTALL_DIR="${AIMGR_INSTALL_DIR:-}"
API_URL="https://api.github.com/repos/${REPO}/releases/latest"

say() {
    printf '%s\n' "$*"
}

fail() {
    printf 'aimgr install: %s\n' "$*" >&2
    exit 1
}

have_cmd() {
    command -v "$1" >/dev/null 2>&1
}

download() {
    url=$1
    dest=$2

    if have_cmd curl; then
        curl -fsSL "$url" -o "$dest"
        return
    fi

    if have_cmd wget; then
        wget -qO "$dest" "$url"
        return
    fi

    fail "curl or wget is required"
}

hash_file() {
    file=$1

    if have_cmd sha256sum; then
        sha256sum "$file" | cut -d ' ' -f 1
        return
    fi

    if have_cmd shasum; then
        shasum -a 256 "$file" | cut -d ' ' -f 1
        return
    fi

    if have_cmd openssl; then
        openssl dgst -sha256 "$file" | sed 's/^.*= //'
        return
    fi

    fail "sha256sum, shasum, or openssl is required"
}

path_contains() {
    case ":${PATH:-}:" in
        *":$1:"*) return 0 ;;
        *) return 1 ;;
    esac
}

detect_os() {
    os_name=$(uname -s 2>/dev/null || true)

    case "$os_name" in
        Linux) printf 'linux' ;;
        Darwin) printf 'darwin' ;;
        *) fail "unsupported operating system: ${os_name:-unknown}" ;;
    esac
}

detect_arch() {
    arch_name=$(uname -m 2>/dev/null || true)

    case "$arch_name" in
        x86_64 | amd64) printf 'amd64' ;;
        arm64 | aarch64) printf 'arm64' ;;
        *) fail "unsupported architecture: ${arch_name:-unknown}" ;;
    esac
}

normalize_version() {
    version=$1

    case "$version" in
        [vV]*) printf '%s' "${version#?}" ;;
        *) printf '%s' "$version" ;;
    esac
}

normalize_release_tag() {
    version=$(normalize_version "$1")
    printf 'v%s' "$version"
}

resolve_release_tag() {
    if [ -n "$REQUESTED_VERSION" ]; then
        normalize_release_tag "$REQUESTED_VERSION"
        return
    fi

    metadata_file=$(mktemp)
    download "$API_URL" "$metadata_file"

    resolved_tag=$(sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$metadata_file" | head -n 1)
    rm -f "$metadata_file"

    [ -n "$resolved_tag" ] || fail "failed to determine the latest release version"
    normalize_release_tag "$resolved_tag"
}

resolve_install_dir() {
    if [ -n "$INSTALL_DIR" ]; then
        printf '%s' "$INSTALL_DIR"
        return
    fi

    if [ "$OS" = "darwin" ] && [ -d "/usr/local/bin" ] && [ -w "/usr/local/bin" ]; then
        printf '%s' "/usr/local/bin"
        return
    fi

    if path_contains "$HOME/.local/bin" || [ "$OS" = "linux" ]; then
        printf '%s' "$HOME/.local/bin"
        return
    fi

    if path_contains "$HOME/bin"; then
        printf '%s' "$HOME/bin"
        return
    fi

    printf '%s' "$HOME/.local/bin"
}

extract_expected_checksum() {
    checksums_file=$1
    asset_name=$2

    awk -v asset="$asset_name" '
        {
            line = $0
            sub(/^[^[:space:]]+[[:space:]]+\*?/, "", line)
            if (line == asset) {
                print $1
                exit
            }
        }
    ' "$checksums_file"
}

OS=$(detect_os)
ARCH=$(detect_arch)
RELEASE_TAG=$(resolve_release_tag)
VERSION=$(normalize_version "$RELEASE_TAG")
INSTALL_DIR=$(resolve_install_dir)
ASSET="aimgr_${VERSION}_${OS}_${ARCH}.tar.gz"
CHECKSUMS_ASSET="checksums.txt"
RELEASE_BASE_URL="https://github.com/${REPO}/releases/download/${RELEASE_TAG}"

have_cmd tar || fail "tar is required"
have_cmd mktemp || fail "mktemp is required"

tmp_dir=$(mktemp -d)
cleanup() {
    rm -rf "$tmp_dir"
}
trap cleanup EXIT INT HUP TERM

archive_path="$tmp_dir/$ASSET"
checksums_path="$tmp_dir/$CHECKSUMS_ASSET"
extract_dir="$tmp_dir/extract"
binary_path="$extract_dir/aimgr"

say "Downloading aimgr ${VERSION} for ${OS}/${ARCH}..."
download "$RELEASE_BASE_URL/$ASSET" "$archive_path"
download "$RELEASE_BASE_URL/$CHECKSUMS_ASSET" "$checksums_path"

expected_checksum=$(extract_expected_checksum "$checksums_path" "$ASSET")
[ -n "$expected_checksum" ] || fail "checksum not found for $ASSET"

actual_checksum=$(hash_file "$archive_path")
[ "$actual_checksum" = "$expected_checksum" ] || fail "checksum verification failed for $ASSET"

mkdir -p "$extract_dir"
tar -xzf "$archive_path" -C "$extract_dir"
[ -f "$binary_path" ] || fail "archive did not contain aimgr binary"

mkdir -p "$INSTALL_DIR"
cp "$binary_path" "$INSTALL_DIR/aimgr"
chmod 0755 "$INSTALL_DIR/aimgr"

say "Installed aimgr to $INSTALL_DIR/aimgr"

if path_contains "$INSTALL_DIR"; then
    say "Run: aimgr --version"
else
    say "Add this to your shell profile:"
    say "  export PATH=\"$INSTALL_DIR:\$PATH\""
    say "Then run: $INSTALL_DIR/aimgr --version"
fi
