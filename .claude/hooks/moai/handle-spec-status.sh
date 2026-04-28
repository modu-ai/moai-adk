#!/bin/bash
# Hook wrapper for SPEC status auto-update
# Triggered by PostToolUse event after git commit

INPUT=$(cat)
moai hook spec-status <<< "$INPUT"
