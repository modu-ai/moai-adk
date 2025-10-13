#!/bin/bash

# MkDocs Build Validation Test
# @TEST:SPEC-DOCS-003

set -e

# Check Python version (3.9+)
python_version=$(python3 --version | cut -d' ' -f2)
python_major=$(echo $python_version | cut -d'.' -f1)
python_minor=$(echo $python_version | cut -d'.' -f2)

if [ $python_major -lt 3 ] || ([ $python_major -eq 3 ] && [ $python_minor -lt 9 ]); then
    echo "❌ Python 3.9+ 필요. 현재 버전: $python_version"
    exit 1
fi

# Attempt uv installation if not present
if ! command -v uv &> /dev/null; then
    echo "📦 uv 설치 중..."
    curl -LsSf https://bit.ly/install-uv | sh
fi

# Install dependencies
uv pip install -r requirements.txt || pip install -r requirements.txt

# Build MkDocs
mkdocs build

# Validate build output
if [ ! -d site ]; then
    echo "❌ site/ 디렉토리 생성 실패"
    exit 1
fi

html_count=$(find site -name "*.html" | wc -l)
if [ $html_count -lt 40 ]; then
    echo "❌ HTML 파일 생성 부족: $html_count개 (최소 40개 필요)"
    exit 1
fi

echo "✅ MkDocs 빌드 검증 완료: $html_count HTML 파일 생성"
exit 0