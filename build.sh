#!/usr/bin/env bash
set -euo pipefail

# Пакет, который собираем (можно указать ./cmd/lib и т.п.)
PKG="./"

DIST_DIR="./dist"
mkdir -p "${DIST_DIR}"

echo "==> Listing Go platforms..."
PLATFORMS=$(go tool dist list)

for PLATFORM in ${PLATFORMS}; do
  GOOS="${PLATFORM%/*}"
  GOARCH="${PLATFORM#*/}"

  # Некоторые платформы не поддерживают c-shared (js/wasm, ios, план9 и т.п.)
  case "${GOOS}" in
    js|ios|wasip1|plan9)
      echo "Skipping unsupported GOOS=${GOOS} GOARCH=${GOARCH}"
      continue
      ;;
  esac

  case "${GOOS}" in
    windows) EXT="dll" ;;
    darwin)  EXT="dylib" ;;
    *)       EXT="so" ;;
  esac

  OUT="${DIST_DIR}/xray_cshare-${GOOS}-${GOARCH}.${EXT}"

  echo "==> Building ${OUT} (GOOS=${GOOS}, GOARCH=${GOARCH})"
  GOOS="${GOOS}" GOARCH="${GOARCH}" CGO_ENABLED=1 \
    go build -buildmode=c-shared -o "${OUT}" "${PKG}" || {
      echo "!! Failed for ${GOOS}/${GOARCH}, continuing..."
      continue
    }
done

echo "==> Done. Artifacts in ${DIST_DIR}"
