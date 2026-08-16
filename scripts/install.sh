#!/usr/bin/env sh
set -eu

repository="d3cie/dottie"
install_dir="${DOTTIE_INSTALL_DIR:-/usr/local/bin}"
version="${DOTTIE_VERSION:-}"

if [ -z "$version" ]; then
  version=$(curl -fsSL "https://api.github.com/repos/$repository/releases/latest" | sed -n 's/.*"tag_name": "v\([^"]*\)".*/\1/p' | head -n 1)
fi
if [ -z "$version" ]; then
  echo "Could not determine the latest Dottie version." >&2
  exit 1
fi
version=${version#v}

case "$(uname -s)" in
  Linux) platform=linux ;;
  Darwin) platform=darwin ;;
  *) echo "Unsupported operating system. Download a release archive manually." >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64) architecture=x86_64 ;;
  arm64|aarch64) architecture=arm64 ;;
  *) echo "Unsupported architecture. Download a release archive manually." >&2; exit 1 ;;
esac

archive="dottie_${version}_${platform}_${architecture}.tar.gz"
release_url="https://github.com/$repository/releases/download/v${version}"
temporary=$(mktemp -d "${TMPDIR:-/tmp}/dottie-install.XXXXXX")
trap 'rm -rf "$temporary"' EXIT INT TERM

curl -fsSL "$release_url/$archive" -o "$temporary/$archive"
curl -fsSL "$release_url/checksums.txt" -o "$temporary/checksums.txt"
expected=$(sed -n "s/^\([[:xdigit:]]*\)[[:space:]]*${archive}$/\1/p" "$temporary/checksums.txt")
if [ -z "$expected" ]; then
  echo "The release checksum does not list $archive." >&2
  exit 1
fi
if command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "$temporary/$archive" | awk '{print $1}')
else
  actual=$(shasum -a 256 "$temporary/$archive" | awk '{print $1}')
fi
if [ "$actual" != "$expected" ]; then
  echo "Checksum verification failed for $archive." >&2
  exit 1
fi
tar -xzf "$temporary/$archive" -C "$temporary"
mkdir -p "$install_dir"
install -m 0755 "$temporary/dottie" "$install_dir/dottie"
echo "Installed Dottie $version to $install_dir/dottie"
