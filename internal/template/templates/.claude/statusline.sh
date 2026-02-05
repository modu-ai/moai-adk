#!/bin/bash
# MoAI Statusline Wrapper
# Cross-platform PATH resolution for moai binary
# This wrapper ensures moai statusline works regardless of PATH configuration

# Read JSON input from stdin (required by Claude Code statusline protocol)
read -r input

# Function to find and execute moai
exec_moai() {
    # Try system PATH first (fastest)
    if command -v moai &> /dev/null; then
        exec moai statusline
    fi

    # Standard Go workspace location
    if [ -f "$HOME/go/bin/moai" ]; then
        exec "$HOME/go/bin/moai" statusline
    fi

    # Alternative Go workspace paths
    if [ -d "$HOME/go" ]; then
        for subdir in bin .bin; do
            if [ -f "$HOME/go/$subdir/moai" ]; then
                exec "$HOME/go/$subdir/moai" statusline
            fi
        done
    fi

    # User local bin
    if [ -f "$HOME/.local/bin/moai" ]; then
        exec "$HOME/.local/bin/moai" statusline
    fi

    # Cargo bin (for Rust-based installations)
    if [ -f "$HOME/.cargo/bin/moai" ]; then
        exec "$HOME/.cargo/bin/moai" statusline
    fi

    # Silent fail - statusline errors are noisy
    exit 0
}

# Execute moai statusline
exec_moai
