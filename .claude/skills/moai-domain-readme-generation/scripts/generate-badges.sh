#!/bin/bash
set -euo pipefail

# generate-badges.sh - Generate badge markdown for README
# Usage: ./generate-badges.sh [LANGUAGE] [PACKAGE_NAME] [VERSION] [LICENSE]

LANGUAGE="${1:-}"
PACKAGE_NAME="${2:-}"
VERSION="${3:-}"
LICENSE="${4:-MIT}"

if [ -z "$LANGUAGE" ] || [ -z "$PACKAGE_NAME" ]; then
    echo "Usage: $0 <language> <package-name> [version] [license]" >&2
    echo "Example: $0 python my-package 1.0.0 MIT" >&2
    exit 1
fi

echo "# Generated Badges for $PACKAGE_NAME"
echo ""

# Version badge (language-specific)
case "$LANGUAGE" in
    python)
        echo "[![PyPI version](https://img.shields.io/pypi/v/$PACKAGE_NAME)](https://pypi.org/project/$PACKAGE_NAME/)"
        ;;
    javascript|typescript|node)
        echo "[![npm version](https://img.shields.io/npm/v/$PACKAGE_NAME)](https://www.npmjs.com/package/$PACKAGE_NAME)"
        ;;
    rust)
        echo "[![Crates.io](https://img.shields.io/crates/v/$PACKAGE_NAME)](https://crates.io/crates/$PACKAGE_NAME)"
        ;;
    go)
        echo "[![Go Reference](https://pkg.go.dev/badge/$PACKAGE_NAME.svg)](https://pkg.go.dev/$PACKAGE_NAME)"
        ;;
    *)
        echo "[![Version](https://img.shields.io/badge/version-$VERSION-blue)]"
        ;;
esac

# License badge
case "$LICENSE" in
    MIT)
        echo "[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)"
        ;;
    Apache-2.0|Apache)
        echo "[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)"
        ;;
    GPL-3.0|GPL)
        echo "[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)"
        ;;
    BSD-3-Clause|BSD)
        echo "[![License: BSD 3-Clause](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause)"
        ;;
    *)
        echo "[![License: $LICENSE](https://img.shields.io/badge/License-$LICENSE-blue.svg)]"
        ;;
esac

# Language badge
case "$LANGUAGE" in
    python)
        echo "[![Python](https://img.shields.io/badge/Python-3.12+-blue)](https://www.python.org/)"
        ;;
    javascript)
        echo "[![JavaScript](https://img.shields.io/badge/JavaScript-ES2022+-yellow)](https://developer.mozilla.org/en-US/docs/Web/JavaScript)"
        ;;
    typescript)
        echo "[![TypeScript](https://img.shields.io/badge/TypeScript-5.3+-blue)](https://www.typescriptlang.org/)"
        ;;
    go)
        echo "[![Go](https://img.shields.io/badge/Go-1.22+-blue)](https://golang.org/)"
        ;;
    rust)
        echo "[![Rust](https://img.shields.io/badge/Rust-2021-orange)](https://www.rust-lang.org/)"
        ;;
    java)
        echo "[![Java](https://img.shields.io/badge/Java-17+-red)](https://www.oracle.com/java/)"
        ;;
    *)
        echo "[![Language: $LANGUAGE](https://img.shields.io/badge/Language-$LANGUAGE-blue)]"
        ;;
esac

exit 0
