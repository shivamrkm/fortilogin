#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "usage: $0 <version>" >&2
  exit 1
fi

VERSION="${1#v}"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="$ROOT_DIR/dist/deb"
PKG_DIR="$BUILD_DIR/fortilogin_${VERSION}_amd64"

rm -rf "$PKG_DIR"
mkdir -p "$PKG_DIR/DEBIAN" "$PKG_DIR/usr/bin" "$PKG_DIR/usr/share/fortilogin"

go build -o "$PKG_DIR/usr/bin/fortilogin" "$ROOT_DIR/cmd/fortilogin"

install -m 0644 \
  "$ROOT_DIR/packaging/systemd/fortilogin.service" \
  "$PKG_DIR/usr/share/fortilogin/fortilogin.service.example"

cat > "$PKG_DIR/DEBIAN/control" <<EOF
Package: fortilogin
Version: $VERSION
Section: net
Priority: optional
Architecture: amd64
Maintainer: Shivam Mishra <shivamrkm3010@gmail.com>
Description: NIT Kurukshetra firewall auto login daemon
 FortiLogin automatically logs in to the NIT KKR firewall when
 connectivity is blocked and provides login/logout/update commands.
 The package installs the binary and a sample systemd unit file.
EOF

dpkg-deb --build "$PKG_DIR"
