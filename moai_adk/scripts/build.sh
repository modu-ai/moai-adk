#!/bin/bash
# MoAI-ADK 빌드 스크립트
# 자동 버전 동기화를 포함한 패키지 빌드

set -e  # 에러 발생 시 스크립트 종료

echo "🗿 MoAI-ADK Build Script"
echo "="*50

# 프로젝트 루트로 이동
cd "$(dirname "$0")/.."

# 빌드 전 버전 동기화
echo "🔄 Step 1: Pre-build version synchronization..."
python3 build_hooks.py --pre-build

# 이전 빌드 아티팩트 정리
echo "🧹 Step 2: Cleaning previous build artifacts..."
rm -rf dist/ build/ *.egg-info/

# 패키지 빌드
echo "📦 Step 3: Building package..."
python3 -m build

# 빌드 결과 확인
echo "✅ Step 4: Verifying build artifacts..."
if [ -d "dist" ] && [ "$(ls -A dist)" ]; then
    echo "Build artifacts created:"
    ls -la dist/
else
    echo "❌ No build artifacts found!"
    exit 1
fi

echo "="*50
echo "🗿 MoAI-ADK build completed successfully!"
echo ""
echo "📦 Next steps (optional):"
echo "  • Test install: pip install dist/*.whl"
echo "  • Upload to PyPI: python -m twine upload dist/*"
echo "  • Create Git tag: git tag v\$(python -c 'from src.moai_adk._version import __version__; print(__version__)')"