#!/bin/bash
# @CODE:SECURITY-001 | 로컬 보안 스캔 실행 스크립트

echo "🔍 MoAI-ADK Security Scan"
echo "=========================="
echo ""

# 보안 도구 설치 확인
echo "📦 Checking security tools..."
if ! command -v pip-audit &> /dev/null; then
    echo "Installing pip-audit..."
    pip install pip-audit
fi

if ! command -v bandit &> /dev/null; then
    echo "Installing bandit..."
    pip install bandit
fi

echo ""
echo "🔍 Step 1: Running pip-audit (dependency vulnerability scan)..."
echo "-------------------------------------------------------------------"

# pip-audit 실행 (실패 시 경고만 출력)
if pip-audit; then
    echo "✅ No vulnerabilities found"
else
    echo "⚠️ Vulnerabilities detected. Please review above."
    PIP_AUDIT_FAILED=1
fi

echo ""
echo "🔍 Step 2: Running bandit (code security scan)..."
echo "-------------------------------------------------------------------"

# bandit 실행 (Low severity 무시)
if bandit -r src/ -ll; then
    echo "✅ No high/medium security issues found"
else
    echo "❌ Security issues detected. Please review above."
    BANDIT_FAILED=1
fi

echo ""
echo "=========================="
if [ -n "$PIP_AUDIT_FAILED" ] || [ -n "$BANDIT_FAILED" ]; then
    echo "⚠️ Security scan completed with warnings/errors"
    echo "   Please review the issues above and fix them."
    exit 1
else
    echo "✅ All security scans passed!"
fi
