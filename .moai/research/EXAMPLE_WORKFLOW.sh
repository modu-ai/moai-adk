#!/bin/bash
# Figma 디자인 분석 - 실제 실행 예제
# 사용법: bash .moai/research/EXAMPLE_WORKFLOW.sh

set -e  # 에러 시 중단

echo "=================================================="
echo "Figma 디자인 분석 워크플로우 시작"
echo "=================================================="
echo ""

# 상수 정의
FILE_KEY="m2odCIWVPWv84ygT5w43Ur"
NODE_ID="689:1242"
OUTPUT_DIR=".moai/research"
FIGMA_TOKEN="${FIGMA_TOKEN:-}"

echo "[Step 1] 환경 확인"
echo "=========================================="

# 1. Figma 토큰 확인
if [ -z "$FIGMA_TOKEN" ]; then
    echo "⚠️  FIGMA_TOKEN이 설정되지 않았습니다"
    echo "   설정 방법:"
    echo "     export FIGMA_TOKEN=your_token_here"
    echo ""
    echo "   또는 .env 파일에 저장:"
    echo "     echo 'FIGMA_TOKEN=your_token' >> .env"
    echo "     export FIGMA_TOKEN=$(grep FIGMA_TOKEN .env | cut -d= -f2)"
    echo ""
    echo "   → UI에서 'Copy as JSON'을 사용하여 진행할 수 있습니다"
    echo ""
else
    echo "✅ FIGMA_TOKEN: ${FIGMA_TOKEN:0:10}...***"
fi

# 2. 디렉토리 생성
mkdir -p "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR/figma-assets"
mkdir -p "$OUTPUT_DIR/scripts"
echo "✅ 디렉토리 생성 완료"

echo ""
echo "[Step 2] 메타데이터 추출"
echo "=========================================="

if [ -z "$FIGMA_TOKEN" ]; then
    echo "⚠️  수동으로 메타데이터를 추출하세요:"
    echo "   1. Figma 파일 열기: https://www.figma.com/file/$FILE_KEY"
    echo "   2. 노드 선택: ID $NODE_ID"
    echo "   3. 우클릭 → Copy as JSON"
    echo "   4. 다음 명령어 실행:"
    echo "      cat > $OUTPUT_DIR/figma-metadata.json << 'JSON'"
    echo "      # 복사된 JSON 붙여넣기"
    echo "      JSON"
    echo ""
    echo "   또는 API 사용 (토큰 필요):"
    echo "      export FIGMA_TOKEN=your_token"
    echo "      bash $0"
else
    echo "📥 Figma API에서 메타데이터 다운로드 중..."
    curl -H "X-Figma-Token: $FIGMA_TOKEN" \
        "https://api.figma.com/v1/files/$FILE_KEY/nodes?ids=$NODE_ID" \
        -o "$OUTPUT_DIR/figma-metadata.json"
    
    if [ -f "$OUTPUT_DIR/figma-metadata.json" ]; then
        FILE_SIZE=$(du -h "$OUTPUT_DIR/figma-metadata.json" | cut -f1)
        echo "✅ 메타데이터 저장 완료: $FILE_SIZE"
        echo "   파일: $OUTPUT_DIR/figma-metadata.json"
    else
        echo "❌ 메타데이터 다운로드 실패"
        exit 1
    fi
fi

echo ""
echo "[Step 3] 분석 스크립트 실행"
echo "=========================================="

if [ -f "$OUTPUT_DIR/figma-metadata.json" ]; then
    echo "📊 색상, 타이포그래피, 컴포넌트 분석 중..."
    
    if command -v uv &> /dev/null; then
        uv run "$OUTPUT_DIR/figma_analyzer.py" \
            --json "$OUTPUT_DIR/figma-metadata.json" \
            --analyze colors,typography,components,images \
            --output "$OUTPUT_DIR" \
            --css \
            --report
        echo "✅ 분석 완료"
    else
        echo "❌ uv가 설치되지 않았습니다"
        echo "   설치: curl -LsSf https://astral.sh/uv/install.sh | sh"
        exit 1
    fi
else
    echo "⚠️  figma-metadata.json이 없습니다"
    echo "   Step 2를 먼저 완료하세요"
    exit 1
fi

echo ""
echo "[Step 4] 생성된 파일 확인"
echo "=========================================="

files=(
    "design-tokens.css"
    "analysis-report.md"
    "analysis-metadata.json"
)

for file in "${files[@]}"; do
    if [ -f "$OUTPUT_DIR/$file" ]; then
        size=$(du -h "$OUTPUT_DIR/$file" | cut -f1)
        echo "✅ $file ($size)"
    else
        echo "⚠️  $file (아직 생성되지 않음)"
    fi
done

echo ""
echo "[Step 5] CSS 토큰 미리보기"
echo "=========================================="

if [ -f "$OUTPUT_DIR/design-tokens.css" ]; then
    echo "생성된 CSS 변수 (처음 10줄):"
    echo ""
    head -10 "$OUTPUT_DIR/design-tokens.css"
    echo ""
    echo "... (전체 보기: cat $OUTPUT_DIR/design-tokens.css)"
fi

echo ""
echo "[Step 6] 다음 단계"
echo "=========================================="

echo ""
echo "✅ 분석 완료!"
echo ""
echo "다음 단계:"
echo "  1. 분석 리포트 확인:"
echo "     cat $OUTPUT_DIR/analysis-report.md"
echo ""
echo "  2. CSS 토큰 확인:"
echo "     cat $OUTPUT_DIR/design-tokens.css"
echo ""
echo "  3. 이미지 다운로드 (토큰 필요):"
echo "     bash $OUTPUT_DIR/scripts/download-figma-images.sh"
echo ""
echo "  4. React 컴포넌트 생성 시작:"
echo "     /moai:1-plan \"노드 689:1242 React 컴포넌트 생성\""
echo "     /clear"
echo "     /moai:2-run SPEC-001"
echo ""
echo "=================================================="
echo "워크플로우 완료 ✅"
echo "=================================================="
