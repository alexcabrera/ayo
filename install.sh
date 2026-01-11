#!/bin/bash
set -e

# Parse flags
DEV_MODE=false
while [[ $# -gt 0 ]]; do
    case $1 in
        --dev)
            DEV_MODE=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# If --dev flag is set, always install locally
if [[ "$DEV_MODE" == "true" ]]; then
    echo "Installing to .local/bin/ (--dev mode)..."
    mkdir -p .local/bin
    GOBIN="$(pwd)/.local/bin" go install ./cmd/ayo
    echo ""
    .local/bin/ayo setup --dev
    exit 0
fi

# Determine if we're on an unmodified main branch in sync with origin
branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
dirty=$(git status --porcelain 2>/dev/null)
behind_ahead=$(git rev-list --left-right --count origin/main...HEAD 2>/dev/null || echo "0 0")

if [[ "$branch" == "main" && -z "$dirty" && "$behind_ahead" == "0	0" ]]; then
    # Clean main branch in sync with origin - install to standard location
    echo "Installing to standard GOBIN location..."
    go install ./cmd/ayo
    echo ""
    ayo setup
else
    # Any other state - install to local .local/bin
    echo "Installing to .local/bin/ (branch: $branch, dirty: ${dirty:+yes}${dirty:-no})..."
    mkdir -p .local/bin
    GOBIN="$(pwd)/.local/bin" go install ./cmd/ayo
    echo ""
    .local/bin/ayo setup --dev
fi
