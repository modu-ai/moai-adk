#!/bin/bash

# testPyPI 업로드 스크립트
# 사용법: ./upload_testpypi.sh [your-api-token]

if [ -z "$1" ]; then
    echo "❌ 사용법: ./upload_testpypi.sh [testpypi-api-token]"
    echo ""
    echo "🔑 API 토큰 생성 방법:"
    echo "1. https://test.pypi.org/account/register/ - 계정 생성"
    echo "2. https://test.pypi.org/manage/account/token/ - 토큰 생성"
    echo "3. 생성된 토큰을 이 스크립트 인수로 전달"
    echo ""
    echo "예시: ./upload_testpypi.sh pypi-AgEIAHN..."
    exit 1
fi

API_TOKEN=$1

echo "🚀 testPyPI 업로드 시작..."
echo "📦 패키지: MoAI-ADK v0.1.9"

# 환경변수 설정
export TWINE_USERNAME=__token__
export TWINE_PASSWORD=$API_TOKEN
export TWINE_REPOSITORY=testpypi

# 업로드 실행
echo "📤 업로드 중..."
twine upload dist/moai_adk-0.1.9*

if [ $? -eq 0 ]; then
    echo "✅ 업로드 성공!"
    echo ""
    echo "🧪 설치 테스트:"
    echo "pip install -i https://test.pypi.org/simple/ moai-adk==0.1.9"
    echo ""
    echo "🔍 기능 검증:"
    echo "moai --version"
else
    echo "❌ 업로드 실패"
    echo ""
    echo "🔧 문제 해결:"
    echo "1. API 토큰 확인 (pypi-로 시작하는지)"
    echo "2. testPyPI 계정 및 토큰 권한 확인"
    echo "3. 패키지명 충돌 확인"
fi