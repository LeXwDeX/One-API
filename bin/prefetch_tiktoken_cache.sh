#!/usr/bin/env bash

set -euo pipefail

CACHE_DIR="${1:-./data/tiktoken-cache}"
PYTHON_BIN="${PYTHON_BIN:-python3}"
ENCODINGS="${ENCODINGS:-cl100k_base o200k_base p50k_base r50k_base}"

if ! command -v "${PYTHON_BIN}" >/dev/null 2>&1; then
  echo "Error: ${PYTHON_BIN} is not available. Please install Python 3 first."
  exit 1
fi

mkdir -p "${CACHE_DIR}"
CACHE_DIR="$(cd "${CACHE_DIR}" && pwd)"

if ! "${PYTHON_BIN}" - <<'PY' >/dev/null 2>&1; then
import tiktoken  # noqa: F401
PY
then
  echo "tiktoken not found, attempting to install via pip..."
  if ! "${PYTHON_BIN}" -m pip install --user --upgrade tiktoken >/dev/null 2>&1; then
    echo "pip install failed, retrying with --break-system-packages..."
    if ! "${PYTHON_BIN}" -m pip install --user --upgrade --break-system-packages tiktoken >/dev/null 2>&1; then
      echo "Error: failed to install tiktoken. Please install it manually and rerun this script."
      exit 1
    fi
  fi
fi

echo "Caching tokenizers into ${CACHE_DIR}"
TIKTOKEN_CACHE_DIR="${CACHE_DIR}" ENCODINGS="${ENCODINGS}" "${PYTHON_BIN}" - <<'PY'
import os
from pathlib import Path

import tiktoken

cache_dir = Path(os.environ["TIKTOKEN_CACHE_DIR"])
encodings = os.environ.get("ENCODINGS", "cl100k_base").split()

cache_dir.mkdir(parents=True, exist_ok=True)

for name in encodings:
    try:
        tiktoken.get_encoding(name)
        print(f"[OK] cached {name}")
    except Exception as exc:
        raise SystemExit(f"[ERR] failed to cache {name}: {exc}")

print(f"Cache populated under {cache_dir}")
PY
