# Figma 디자인 추출 및 분석 워크플로우

**프로젝트**: 모듈의사주-프로젝트-와디즈
**파일 키**: m2odCIWVPWv84ygT5w43Ur
**노드 ID**: 689:1242
**작성 날짜**: 2025-11-19

---

## 🚀 빠른 시작 (5분)

### 방법 1: Figma UI에서 JSON 복사 (가장 간단)

```bash
# 1단계: Figma 파일 열기
# https://www.figma.com/file/m2odCIWVPWv84ygT5w43Ur

# 2단계: 노드 689:1242 선택

# 3단계: 우클릭 → Copy as JSON

# 4단계: 파일에 저장
cat > .moai/research/figma-metadata.json << 'EOF'
# 여기에 복사된 JSON 붙여넣기
EOF

# 5단계: 분석 스크립트 실행
uv run .moai/research/figma_analyzer.py \
  --json .moai/research/figma-metadata.json \
  --analyze colors,typography,components,images \
  --css \
  --report
```

**출력물**:
- `.moai/research/design-tokens.css` - CSS 변수
- `.moai/research/analysis-report.md` - 마크다운 리포트
- `.moai/research/analysis-metadata.json` - 메타데이터

---

### 방법 2: Figma REST API 사용 (권장)

```bash
# 1단계: 개인 액세스 토큰 발급
# https://www.figma.com/settings/account → Personal access tokens
# 토큰 생성 → 복사

# 2단계: 환경 변수로 저장 (⚠️ .env에 저장, .gitignore에 추가)
echo "FIGMA_TOKEN=your_token_here" > .env

# 3단계: API로 파일 메타데이터 다운로드
FIGMA_TOKEN=$(grep FIGMA_TOKEN .env | cut -d= -f2)

curl -H "X-Figma-Token: $FIGMA_TOKEN" \
  "https://api.figma.com/v1/files/m2odCIWVPWv84ygT5w43Ur/nodes?ids=689:1242" \
  > .moai/research/figma-metadata.json

# 4단계: 분석 실행
uv run .moai/research/figma_analyzer.py \
  --json .moai/research/figma-metadata.json \
  --css \
  --report
```

---

## 📊 상세 워크플로우

### Phase 1: 메타데이터 수집 (10분)

#### Step 1a: Figma 파일 접근

```bash
# 파일 확인 링크
https://www.figma.com/file/m2odCIWVPWv84ygT5w43Ur/모듈의사주-프로젝트-와디즈
```

#### Step 1b: 노드 689:1242 메타데이터 추출

**Option A: Figma UI (UI 클릭)**

1. Figma 파일 열기
2. "모듈의사주-프로젝트-와디즈" 탭 클릭
3. 왼쪽 레이어 패널에서 ID `689:1242` 찾기 또는 검색
4. 해당 노드 클릭
5. 우클릭 → **"Copy as JSON"** (또는 **"Export as JSON"**)
6. 텍스트 에디터 열기
7. Ctrl+V (또는 Cmd+V) 붙여넣기
8. `.moai/research/figma-metadata.json` 저장

**Option B: Figma API (자동화)**

```bash
#!/bin/bash
# .moai/scripts/fetch-figma-metadata.sh

set -e

FILE_KEY="m2odCIWVPWv84ygT5w43Ur"
NODE_ID="689:1242"
OUTPUT_DIR=".moai/research"

# 토큰 확인
if [ -z "$FIGMA_TOKEN" ]; then
    echo "Error: FIGMA_TOKEN not set"
    echo "Set it: export FIGMA_TOKEN=your_token_here"
    exit 1
fi

# 디렉토리 생성
mkdir -p "$OUTPUT_DIR"

# 메타데이터 다운로드
echo "Fetching Figma metadata..."
curl -H "X-Figma-Token: $FIGMA_TOKEN" \
  "https://api.figma.com/v1/files/$FILE_KEY/nodes?ids=$NODE_ID" \
  -o "$OUTPUT_DIR/figma-metadata.json"

echo "✅ Metadata saved to $OUTPUT_DIR/figma-metadata.json"

# 이미지 URL 생성
echo "Generating image URLs..."
curl -H "X-Figma-Token: $FIGMA_TOKEN" \
  "https://api.figma.com/v1/images/$FILE_KEY?ids=$NODE_ID&scale=2&format=png" \
  -o "$OUTPUT_DIR/figma-images.json"

echo "✅ Image URLs saved to $OUTPUT_DIR/figma-images.json"
```

**실행**:
```bash
chmod +x .moai/scripts/fetch-figma-metadata.sh
export FIGMA_TOKEN=your_token_here
uv run bash .moai/scripts/fetch-figma-metadata.sh
```

#### Step 1c: 메타데이터 검증

```bash
# JSON 구조 확인
cat .moai/research/figma-metadata.json | python3 -m json.tool | head -50

# 파일 크기 확인
wc -l .moai/research/figma-metadata.json

# 노드 ID 확인
grep -o '"id":"[^"]*"' .moai/research/figma-metadata.json | head -5
```

---

### Phase 2: 색상 및 타이포그래피 분석 (15분)

#### Step 2a: 분석 스크립트 실행

```bash
uv run .moai/research/figma_analyzer.py \
  --json .moai/research/figma-metadata.json \
  --analyze colors,typography,components,images \
  --output .moai/research \
  --css \
  --report
```

**출력물**:

```
Extracted 24 colors
Extracted 12 typography styles
Found 8 components
Found 3 images
CSS tokens saved to .moai/research/design-tokens.css
Report saved to .moai/research/analysis-report.md
Metadata saved to .moai/research/analysis-metadata.json
```

#### Step 2b: 생성된 파일 검토

```bash
# 디자인 토큰 (CSS)
cat .moai/research/design-tokens.css

# 분석 리포트 (마크다운)
cat .moai/research/analysis-report.md

# 메타데이터 (JSON)
cat .moai/research/analysis-metadata.json | python3 -m json.tool
```

---

### Phase 3: 이미지 자산 다운로드 (10분)

#### Step 3a: 이미지 URL에서 다운로드

```bash
#!/bin/bash
# .moai/scripts/download-figma-images.sh

FILE_KEY="m2odCIWVPWv84ygT5w43Ur"
NODE_ID="689:1242"
OUTPUT_DIR=".moai/research/figma-assets"

mkdir -p "$OUTPUT_DIR"

# 이미지 URL 조회
curl -H "X-Figma-Token: $FIGMA_TOKEN" \
  "https://api.figma.com/v1/images/$FILE_KEY?ids=$NODE_ID&scale=2&format=png" \
  | python3 << 'EOF'
import json
import sys
import urllib.request
import os

data = json.load(sys.stdin)
output_dir = "$OUTPUT_DIR"

if "images" in data:
    for node_id, url in data["images"].items():
        filename = os.path.join(output_dir, f"{node_id}.png")
        print(f"Downloading {node_id}...")
        urllib.request.urlretrieve(url, filename)
        print(f"✅ Saved to {filename}")
EOF
```

**실행**:
```bash
export FIGMA_TOKEN=your_token_here
uv run bash .moai/scripts/download-figma-images.sh
```

#### Step 3b: Figma UI에서 수동 다운로드

1. Figma 파일에서 노드 `689:1242` 선택
2. 오른쪽 패널 → **"Export"** 섹션 열기
3. **"+"** 버튼 클릭 → PNG 추가
4. Scale: **2x** 선택
5. **"Export"** 버튼 클릭
6. 파일 저장: `.moai/research/figma-assets/node-689-1242@2x.png`

---

### Phase 4: 접근성 검증 (15분)

#### Step 4a: WCAG AA 색상 대조 검증

```python
#!/usr/bin/env python3
# .moai/scripts/check-contrast.py

import json
import sys
from pathlib import Path

def rgb_to_luminance(rgb_tuple):
    """RGB 튜플을 WCAG 휘도로 변환"""
    r, g, b = [x / 255.0 for x in rgb_tuple]
    r = r / 12.92 if r <= 0.03928 else ((r + 0.055) / 1.055) ** 2.4
    g = g / 12.92 if g <= 0.03928 else ((g + 0.055) / 1.055) ** 2.4
    b = b / 12.92 if b <= 0.03928 else ((b + 0.055) / 1.055) ** 2.4
    return 0.2126 * r + 0.7152 * g + 0.0722 * b

def contrast_ratio(color1, color2):
    """두 색상 간 WCAG 대조 비율"""
    l1 = rgb_to_luminance(color1)
    l2 = rgb_to_luminance(color2)
    lmax, lmin = max(l1, l2), min(l1, l2)
    return round((lmax + 0.05) / (lmin + 0.05), 2)

# 예시: 텍스트 색상과 배경 색상
text_color = (15, 23, 42)       # #0F172A (다크 네이비)
background = (248, 250, 252)    # #F8FAFC (라이트 그레이)

ratio = contrast_ratio(text_color, background)
print(f"색상 대조 비율: {ratio}:1")

# WCAG 규정 확인
wcag_aa = ratio >= 4.5  # 본문 텍스트
wcag_aaa = ratio >= 7.0  # 향상된 대조
print(f"WCAG AA (4.5:1): {'✅ PASS' if wcag_aa else '❌ FAIL'}")
print(f"WCAG AAA (7.0:1): {'✅ PASS' if wcag_aaa else '❌ FAIL'}")

# 모든 색상 조합 검증
with open(".moai/research/analysis-metadata.json") as f:
    data = json.load(f)

print("\n색상 조합 검증:")
print("=" * 60)

# 주요 색상 조합 정의
checks = [
    ("Primary Text", (15, 23, 42), "Background", (248, 250, 252)),
    ("Primary CTA", (14, 165, 233), "White", (255, 255, 255)),
    ("Error Text", (239, 68, 68), "White", (255, 255, 255)),
    ("Success Text", (16, 185, 129), "White", (255, 255, 255)),
]

for name1, color1, name2, color2 in checks:
    ratio = contrast_ratio(color1, color2)
    aa_pass = "✅" if ratio >= 4.5 else "❌"
    aaa_pass = "✅" if ratio >= 7.0 else "❌"
    print(f"{aa_pass} AA | {aaa_pass} AAA | {ratio}:1 | {name1} on {name2}")
```

**실행**:
```bash
python .moai/scripts/check-contrast.py
```

#### Step 4b: 검사 항목 확인

```markdown
## WCAG 2.1 AA 체크리스트

### 색상 대조 (Color Contrast)
- [ ] 본문 텍스트 4.5:1 이상
- [ ] 대형 텍스트 3:1 이상
- [ ] UI 컴포넌트 3:1 이상
- [ ] 포커스 지시자 명확함

### 구조 및 의미론 (Structure & Semantics)
- [ ] 올바른 제목 계층 (h1, h2, h3...)
- [ ] 의미론적 마크업 사용
- [ ] 랜드마크 영역 정의

### 키보드 접근성 (Keyboard Navigation)
- [ ] 모든 기능 키보드로 접근 가능
- [ ] 탭 순서 논리적
- [ ] 포커스 트랩 없음

### 터치 대상 (Touch Targets)
- [ ] 최소 44x44px
- [ ] 요소 간 최소 8px 간격

### 스크린 리더 (Screen Reader Support)
- [ ] 모든 이미지에 alt 텍스트
- [ ] ARIA 레이블 적절함
- [ ] Live region 사용 (동적 콘텐츠)
```

---

### Phase 5: 설계 문서화 (20분)

#### Step 5a: 최종 분석 리포트 생성

```bash
cat > .moai/research/figma-node-689-1242-analysis.md << 'EOF'
# Figma 디자인 분석 리포트: 노드 689:1242

## 기본 정보
- **파일**: 모듈의사주-프로젝트-와디즈
- **파일 키**: m2odCIWVPWv84ygT5w43Ur
- **노드 ID**: 689:1242
- **작성 날짜**: 2025-11-19
- **분석자**: [이름]

## 설계 요소

### 색상 팔레트

$(cat .moai/research/design-tokens.css | grep "color")

### 타이포그래피

$(cat .moai/research/analysis-report.md | grep -A 20 "Typography Summary")

## 접근성 검증

$(python .moai/scripts/check-contrast.py)

## 다음 단계

1. ✅ 메타데이터 추출 완료
2. ✅ 색상 및 타이포그래피 분석 완료
3. ✅ 접근성 검증 완료
4. → React 컴포넌트 생성 시작
5. → Storybook 통합
6. → Playwright 테스트

## 리소스

- 메타데이터: `.moai/research/figma-metadata.json`
- 분석 리포트: `.moai/research/analysis-report.md`
- 디자인 토큰: `.moai/research/design-tokens.css`
- 이미지 자산: `.moai/research/figma-assets/`

---

**다음 단계**: `/moai:1-plan "노드 689:1242 기반 React 컴포넌트 생성"`
EOF

cat .moai/research/figma-node-689-1242-analysis.md
```

#### Step 5b: 디자인 시스템 문서화

```bash
cat > .moai/research/design-system.md << 'EOF'
# 디자인 시스템: 노드 689:1242

## 원자 요소 (Atoms)

### 색상 변수
```css
--color-primary-500: #0EA5E9
--color-text: #0F172A
--color-background: #F8FAFC
```

### 타이포그래피
- **제목 L**: 32px / 700 / 1.25
- **본문**: 16px / 400 / 1.5
- **캡션**: 12px / 500 / 1.25

### 간격
- xs: 4px, sm: 8px, md: 16px, lg: 24px, xl: 32px

## 분자 요소 (Molecules)

### 버튼
- 크기: Small, Medium, Large
- 상태: Default, Hover, Active, Disabled
- 종류: Primary, Secondary, Ghost

### 입력 필드
- 텍스트, 이메일, 비밀번호
- 상태: 정상, 포커스, 에러, 비활성화

## 유기체 요소 (Organisms)

### 헤더
- 로고 + 네비게이션
- 반응형 (모바일/태블릿/데스크탑)

### 카드
- 이미지 + 제목 + 설명 + CTA

### 폼
- 입력 필드 + 레이블 + 검증 메시지

---

**스토리북**: `/src/stories/`에서 모든 컴포넌트 확인
EOF

cat .moai/research/design-system.md
```

---

## 📁 최종 파일 구조

```
.moai/research/
├── FIGMA_EXTRACTION_GUIDE.md        # 추출 가이드 (이 문서)
├── FIGMA_WORKFLOW.md                # 워크플로우 (이 문서)
├── figma_analyzer.py                # 분석 스크립트
├── figma-metadata.json              # Figma JSON 메타데이터
├── figma-images.json                # 이미지 URL 목록
├── design-tokens.css                # CSS 변수
├── analysis-report.md               # 마크다운 리포트
├── analysis-metadata.json           # 추출 결과 메타데이터
├── figma-node-689-1242-analysis.md # 최종 분석 리포트
├── design-system.md                 # 디자인 시스템 문서
├── figma-assets/
│   ├── node-689-1242.png            # 1배 해상도
│   ├── node-689-1242@2x.png         # 2배 해상도
│   └── images/                      # 삽입된 이미지들
└── scripts/
    ├── fetch-figma-metadata.sh      # API 메타데이터 다운로드
    ├── download-figma-images.sh     # API 이미지 다운로드
    └── check-contrast.py            # WCAG 대조 검증
```

---

## 🎯 다음 단계

### 즉시 (오늘)
1. ✅ Figma 메타데이터 추출 (위 방법 1 또는 2 사용)
2. ✅ 분석 스크립트 실행
3. ✅ 이미지 자산 다운로드
4. ✅ 접근성 검증

### 내일
1. React 컴포넌트 생성 시작
   ```bash
   /moai:1-plan "노드 689:1242 기반 React 컴포넌트 구현 (TypeScript, Tailwind CSS)"
   /clear
   /moai:2-run SPEC-XXX
   ```

2. Storybook 설정
   ```bash
   npm install --save-dev storybook
   npx storybook init
   ```

3. Playwright 테스트 작성
   ```bash
   npx playwright install
   npx playwright codegen http://localhost:6006
   ```

### 다음주
1. 디자인 시스템 통합
2. CI/CD 파이프라인 설정
3. 성능 최적화

---

## 🆘 문제 해결

### Figma 토큰 오류
```
Error: 401 Unauthorized
```
**해결**: 토큰 재발급 (Settings → Personal access tokens)

### JSON 파싱 오류
```
JSONDecodeError: Expecting value
```
**해결**: JSON 파일이 유효한지 확인
```bash
python3 -m json.tool .moai/research/figma-metadata.json > /dev/null
```

### 이미지 다운로드 실패
```
Error: 404 Not Found
```
**해결**: 이미지 URL 확인, Figma 파일에 이미지 포함 여부 확인

---

**질문이나 막힘이 있으면 언제든 물어보세요!**
