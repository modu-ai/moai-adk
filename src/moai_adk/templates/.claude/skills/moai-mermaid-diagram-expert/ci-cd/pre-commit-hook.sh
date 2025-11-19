#!/bin/sh
# .git/hooks/pre-commit
# Auto-convert modified Mermaid diagrams before commit

MERMAID_DIR="diagrams"
OUTPUT_DIR="docs/images"
SCRIPT_PATH="scripts/mermaid-to-svg-png.py"

# Check if Mermaid directory exists
if [ ! -d "$MERMAID_DIR" ]; then
    exit 0
fi

# Find modified .mmd files
MODIFIED_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep "\.mmd$")

if [ -z "$MODIFIED_FILES" ]; then
    exit 0
fi

echo "Converting modified Mermaid diagrams..."

# Create output directory if needed
mkdir -p "$OUTPUT_DIR"

# Convert diagrams
uv run "$SCRIPT_PATH" "$MERMAID_DIR" \
    --format png \
    --batch \
    --output "$OUTPUT_DIR" \
    --no-overwrite

if [ $? -eq 0 ]; then
    # Stage converted images
    git add "$OUTPUT_DIR"
    echo "✓ Diagram conversion successful"
    exit 0
else
    echo "✗ Error: Diagram conversion failed"
    exit 1
fi
