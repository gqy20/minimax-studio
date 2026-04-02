#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_HOST="${BACKEND_HOST:-127.0.0.1}"
BACKEND_PORT="${BACKEND_PORT:-8037}"
FRONTEND_HOST="${FRONTEND_HOST:-0.0.0.0}"
FRONTEND_PORT="${FRONTEND_PORT:-3023}"
OUTPUT_DIR="${OUTPUT_DIR:-./output}"
PROXY_TARGET="${VITE_PROXY_TARGET:-http://${BACKEND_HOST}:${BACKEND_PORT}}"

if [[ -z "${MINIMAX_API_KEY:-}" ]]; then
  echo "MINIMAX_API_KEY is not set."
  echo "Example:"
  echo "  MINIMAX_API_KEY=your_key make dev"
  exit 1
fi

cleanup() {
  if [[ -n "${SERVER_PID:-}" ]] && kill -0 "${SERVER_PID}" >/dev/null 2>&1; then
    kill "${SERVER_PID}" >/dev/null 2>&1 || true
    wait "${SERVER_PID}" >/dev/null 2>&1 || true
  fi
}

trap cleanup EXIT INT TERM

mkdir -p "${ROOT_DIR}/bin"

if [[ ! -f "${ROOT_DIR}/bin/ms" ]]; then
  echo "Building backend binary..."
  (cd "${ROOT_DIR}" && make build)
fi

if [[ ! -d "${ROOT_DIR}/frontend/node_modules" ]]; then
  echo "Installing frontend dependencies..."
  (cd "${ROOT_DIR}/frontend" && npm install)
fi

echo "Starting backend on ${BACKEND_HOST}:${BACKEND_PORT}..."
(
  cd "${ROOT_DIR}"
  ./bin/ms server --port "${BACKEND_PORT}" --output-dir "${OUTPUT_DIR}"
) &
SERVER_PID=$!

sleep 1
if ! kill -0 "${SERVER_PID}" >/dev/null 2>&1; then
  echo "Backend failed to start."
  exit 1
fi

echo "Starting frontend on ${FRONTEND_HOST}:${FRONTEND_PORT}..."
echo "Open: http://localhost:${FRONTEND_PORT}"

exec env \
  FRONTEND_HOST="${FRONTEND_HOST}" \
  FRONTEND_PORT="${FRONTEND_PORT}" \
  VITE_PROXY_TARGET="${PROXY_TARGET}" \
  npm run dev --prefix "${ROOT_DIR}/frontend"
