#!/usr/bin/env bash
# Builds the signed update manifest and SHA256SUMS for a release.
#
#   scripts/release-manifest.sh <tag> <assets-dir> <private-key> <public-key>
#
# Run from the release job against the flat directory of staged assets. The
# manifest lists exactly one in-app artifact per platform/arch, because the
# updater installs the first entry matching the running platform:
#   macOS    adex-darwin-<arch>.app.zip
#   Windows  adex-windows-<arch>.exe   (never the NSIS installer)
#   Linux    the AppImage; for an arch without one, the raw binary, so those
#            users still hear about the release (they get a link, since
#            only AppImages update in place).
# Distro packages and installers are left out and stay manual downloads.
set -euo pipefail

if [ $# -ne 4 ]; then
  echo "usage: $0 <tag> <assets-dir> <private-key> <public-key>" >&2
  exit 2
fi
tag=$1
dir=$2
key=$(realpath "$3")
pub=$(realpath "$4")
version=${tag#v}
repo=${GITHUB_REPOSITORY:-AnoRebel/aDex-UI}

cd "$dir"
shopt -s nullglob

picks=(*.app.zip *.AppImage)
for f in adex-windows-*.exe; do
  [[ $f == *-installer-* ]] || picks+=("$f")
done
for f in adex-linux-*; do
  [[ $f == *.* ]] && continue # .deb, .rpm, .pkg.tar.zst
  arch=${f#adex-linux-}
  appimages=(*-linux-"$arch".AppImage)
  [ ${#appimages[@]} -eq 0 ] && picks+=("$f")
done
if [ ${#picks[@]} -eq 0 ]; then
  echo "no in-app artifacts found in $dir" >&2
  exit 1
fi

wails3 updater manifest \
  -version "$version" \
  -name "aDex $version" \
  -key "$key" \
  -url-prefix "https://github.com/$repo/releases/download/$tag/" \
  -output update-manifest.json \
  "${picks[@]}"

wails3 updater verify -manifest update-manifest.json -publickey "$pub"

# Every platform/arch that was built (each build ships a raw adex-<os>-<arch>
# binary) must have exactly one manifest entry, or its users would never be
# told about the release.
built=$(for f in adex-*-*; do
  if [[ $f =~ ^adex-(linux|darwin|windows)-(amd64|arm64)(\.exe)?$ ]]; then
    echo "${BASH_REMATCH[1]}/${BASH_REMATCH[2]}"
  fi
done | sort -u)
listed=$(jq -r '.artifacts[] | "\(.platform)/\(.arch)"' update-manifest.json | sort)
if [ "$built" != "$(sort -u <<<"$listed")" ] || [ -n "$(uniq -d <<<"$listed")" ]; then
  echo "manifest entries do not match the built platforms" >&2
  echo "built:" >&2; echo "$built" >&2
  echo "listed:" >&2; echo "$listed" >&2
  exit 1
fi

files=()
for f in *; do
  [ -f "$f" ] && [ "$f" != SHA256SUMS ] && files+=("$f")
done
sha256sum -- "${files[@]}" > SHA256SUMS
echo "manifest and SHA256SUMS written for $tag"
