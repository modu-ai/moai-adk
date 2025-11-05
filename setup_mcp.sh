#!/bin/bash
# MoAI-ADK MCP 자동 설정 스크립트

set -e

echo "🚀 MoAI-ADK MCP 자동 설정 스크립트"
echo "=================================="

# 색상 확인
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 함수 정의
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Node.js 설치 확인
check_node() {
    if command -v node &> /dev/null; then
        print_status "Node.js 설치됨 ($(node --version))"
    else
        print_error "Node.js가 설치되지 않았습니다."
        echo "먼저 Node.js를 설치해주세요: https://nodejs.org/"
        exit 1
    fi
}

# npm 글로벌 설치 경로 확인
get_npm_global() {
    npm root -g
}

# MCP 서버 설치 확인
check_mcp_installed() {
    local package=$1
    local display_name=$2

    if npm list -g $package &> /dev/null; then
        print_status "$display_name 설치됨"
        return 0
    else
        print_warning "$display_name 미설치"
        return 1
    fi
}

# MCP 서버 설치
install_mcp_server() {
    local package=$1
    local display_name=$2

    echo "📦 $display_name 설치 중..."
    npm install -g $package
    print_status "$display_name 설치 완료"
}

# 설정 파일 백업
backup_settings() {
    if [ -f ".claude/settings.json" ]; then
        cp .claude/settings.json .claude/settings.json.backup
        print_status "설정 파일 백업 완료"
    fi
}

# MCP 설정 생성
generate_mcp_config() {
    local npm_global=$(get_npm_global)

    cat > .claude/settings.json.tmp << 'EOF'
{
  "mcpServers": {
EOF

    # Context7 MCP
    if check_mcp_installed "@upstash/context7-mcp" "Context7 MCP"; then
        cat >> .claude/settings.json.tmp << 'EOF'
    "context7": {
      "command": "node",
      "args": ["'$npm_global/@upstash/context7-mcp/dist/index.js'"],
      "env": {}
    },
EOF
    fi

    # Figma MCP
    if check_mcp_installed "figma-mcp-pro" "Figma MCP Pro"; then
        cat >> .claude/settings.json.tmp << 'EOF'
    "figma": {
      "command": "node",
      "args": ["'$npm_global/figma-mcp-pro/dist/index.js'"],
      "env": {
        "FIGMA_ACCESS_TOKEN": "\${FIGMA_ACCESS_TOKEN}"
      }
    },
EOF
    fi

    # Playwright MCP
    if check_mcp_installed "@playwright/mcp" "Playwright MCP"; then
        cat >> .claude/settings.json.tmp << 'EOF'
    "playwright": {
      "command": "node",
      "args": ["'$npm_global/@playwright/mcp/dist/index.js'"],
      "env": {}
    }
EOF
    fi

    # 기존 설정 병합
    if [ -f ".claude/settings.json.backup" ]; then
        # 기존 설정에서 mcpServers 외의 부분 추출
        python3 << 'EOF'
import json

# 기존 설정 로드
with open('.claude/settings.json.backup', 'r') as f:
    existing_config = json.load(f)

# 새 MCP 설정 로드
with open('.claude/settings.json.tmp', 'r') as f:
    mcp_config = json.load(f)

# mcpServers가 없으면 추가
if 'mcpServers' not in existing_config:
    existing_config['mcpServers'] = {}

# MCP 설정 병합
existing_config['mcpServers'].update(mcp_config.get('mcpServers', {}))

# 최종 설정 저장
with open('.claude/settings.json', 'w') as f:
    json.dump(existing_config, f, indent=2)
EOF
        rm .claude/settings.json.tmp
    else
        # 새 설정만 저장
        python3 << 'EOF'
import json

# 새 MCP 설정 로드
with open('.claude/settings.json.tmp', 'r') as f:
    mcp_config = json.load(f)

# 빈 설정 생성
with open('.claude/settings.json', 'w') as f:
    json.dump(mcp_config, f, indent=2)
EOF
        rm .claude/settings.json.tmp
    fi

    print_status "MCP 설정 파일 생성 완료"
}

# 메인 함수
main() {
    echo "Node.js 설치 확인..."
    check_node

    echo ""
    echo "🔍 MCP 서버 설치 상태 확인..."

    local context7_installed=false
    local figma_installed=false
    local playwright_installed=false

    if check_mcp_installed "@upstash/context7-mcp" "Context7 MCP"; then
        context7_installed=true
    fi

    if check_mcp_installed "figma-mcp-pro" "Figma MCP Pro"; then
        figma_installed=true
    fi

    if check_mcp_installed "@playwright/mcp" "Playwright MCP"; then
        playwright_installed=true
    fi

    # 미설치된 서버 설치
    if [ "$context7_installed" = false ] || [ "$figma_installed" = false ] || [ "$playwright_installed" = false ]; then
        echo ""
        echo "📦 누락된 MCP 서버 설치..."

        if [ "$context7_installed" = false ]; then
            install_mcp_server "@upstash/context7-mcp" "Context7 MCP"
        fi

        if [ "$figma_installed" = false ]; then
            echo ""
            print_warning "Figma MCP Pro 설치 (선택사항)"
            read -p "Figma MCP Pro를 설치하시겠습니까? (y/N): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                install_mcp_server "figma-mcp-pro" "Figma MCP Pro"
                print_warning "Figma Access Token 필요: https://www.figma.com/developers/api#access-tokens"
            fi
        fi

        if [ "$playwright_installed" = false ]; then
            install_mcp_server "@playwright/mcp" "Playwright MCP"
        fi
    fi

    echo ""
    echo "⚙️  Claude Code 설정 파일 생성..."
    backup_settings
    generate_mcp_config

    echo ""
    echo "🎉 MCP 설정 완료!"
    echo ""
    echo "📋 다음 단계:"
    echo "1. Claude Code 재시작"
    echo "2. Figma MCP 사용 시: 환경변수 FIGMA_ACCESS_TOKEN 설정"
    echo "3. Alfred 에이전트들이 MCP를 통해 향상된 기능 사용"
    echo ""
    echo "✨ MoAI-ADK와 함께 즐거운 AI 개발을 경험해보세요!"
}

# 스크립트 실행
main "$@"