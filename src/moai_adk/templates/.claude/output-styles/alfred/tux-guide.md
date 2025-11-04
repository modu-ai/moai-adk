---
name: Alfred TUX Guide
description: "Claude Code official TUX guidelines for Alfred's output optimization"
# Translations:
# - ko: "Claude Code 공식 TUX 가이드라인 - Alfred 출력 최적화"
# - ja: "Claude Code公式TUXガイドライン - Alfred出力最適化"
# - zh: "Claude Code官方TUX指南 - Alfred输出优化"
---

# Alfred TUX Guide
> Claude Code 공식 TUX(Text User Experience) 가이드라인을 Alfred의 출력에 최적화

**Audience**: Claude Code Skill 개발자 및 Alfred 출력 최적화 담당자

이 가이드는 Claude Code의 공식 TUX 원칙을 Alfred의 4단계 워크플로우에 통합하는 방법을 설명합니다.

## 📋 Claude Code TUX 핵심 원칙

### 1. 스트리밍 최적화 (Streaming Optimization)

Claude Code는 `outputStyle: "streaming"` 환경에서 작동합니다. Alfred의 출력은 이를 최적화해야 합니다.

```bash
# ✅ 좋은 예: 점진적 정보 공개
🔍 Alfred: 분석 시작...
⠋ 프로젝트 구조 스캔 중...
⠙ 의존성 분석 중...
⠹ SPEC 파일 확인 중...
⠸ 계획 생성 중...
✅ 분석 완료!

# ❌ 나쁜 예: 모든 정보를 한 번에 표시
🔍 Alfred: 분석 결과: 프로젝트 구조를 스캔했고, 의존성을 분석했고, SPEC 파일을 확인했으며, 계획을 생성했습니다. 완료!
```

### 2. 비차단 상호작용 (Non-blocking Interaction)

사용자가 Alfred 처리 중에도 계속 작업할 수 있어야 합니다.

```bash
# ✅ 좋은 예: 백그라운드 처리 표시
🔄 Alfred: 백그라운드에서 테스트 실행 중...
💡 팁: 계속해서 다른 작업을 진행할 수 있습니다
📊 진행률: ████████▂▂▂▂▂ 80%
⏱️  예상 완료: 2분 후

# ❌ 나쁜 예: 처리 중단
⏳ Alfred: 테스트 실행 중... (완료될 때까지 기다려주세요)
```

### 3. 터미널 친화적 디자인 (Terminal-friendly Design)

다양한 터미널 환경에서 일관된 경험을 제공해야 합니다.

```bash
# ✅ 좋은 예: 안전한 Unicode 사용
📊 진행률: [████████▂▂▂] 80%
🔄 상태: 처리 중
✅ 성공: 완료됨
❌ 실패: 오류 발생

# ❌ 나쁜 예: 복잡한 이모지 의존
📊 진행률: 🚀🚀🚀🚀🚀🚀🚀🚀⏳⏳⏳
🔄 상태: 💫✨🔮
✅ 성공: 🎉🎊🥳
```

## 🎨 Alfred의 시각적 언어

### 색상 팔레트 (ANSI Color Codes)

```bash
# Alfred의 표준 색상
\033[32m✅\033[0m 성공 (초록)
\033[33m⚠️\033[0m 경고 (노랑)
\033[31m❌\033[0m 오류 (빨강)
\033[34mℹ️\033[0m 정보 (파랑)
\033[36m🔄\033[0m 처리 중 (청록)
\033[35m📊\033[0m 진행률 (자홍)
```

### 진행 표시기 (Progress Indicators)

```bash
# 스피너 애니메이션 (터미널 안전)
spinner_frames = ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"]

# 진행률 바 (문자 기반)
progress_bars = [
    "▂▂▂▂▂▂▂▂▂▂",  # 0%
    "██▂▂▂▂▂▂▂▂",  # 20%
    "████▂▂▂▂▂▂",  # 40%
    "██████▂▂▂▂",  # 60%
    "████████▂▂",  # 80%
    "██████████"   # 100%
]

# 단계 표시기
phase_indicators = ["◐", "◑", "◒", "◓"]
```

### 아이콘 전략 (Icon Strategy)

```bash
# 1단계: Unicode 기본 (모든 터미널 지원)
✅ ❌ ⚠️ ℹ️ 🔄 📊 📁 🎯 ⚡

# 2단계: 이모지 (현대 터미널)
🚀 🎉 🎊 💡 🤝 🎓 ⚡️

# 3단계: 폴백 (ASCII 전용)
[OK] [ERR] [WARN] [INFO] [PROC] [PROG] [DIR] [TARGET] [FAST]
```

## 📱 반응형 레이아웃 (Responsive Layout)

### 터미널 너비 감지

```python
def format_status_card(terminal_width):
    if terminal_width >= 100:
        return """
┌─ 프로젝트 상태 ───────────────────────────────────────────┐
│ 📁 위치: /path/to/project                                 │
│ 🌍 언어: 한국어                                           │
│ 📊 진행률: 3개 SPEC, 12개 테스트, 85% 커버리지            │
│ 🔄 마지막 동기화: 2시간 전                                │
│ 🚀 브랜치: develop (3개 커밋 앞섬)                       │
└───────────────────────────────────────────────────────────┘
        """
    elif terminal_width >= 80:
        return """
┌─ 프로젝트 상태 ───────────────────┐
│ 📁 /path/to/project               │
│ 🌍 한국어 | 📊 85% 커버리지       │
│ 🔄 2시간 전 | 🚀 develop +3       │
└─────────────────────────────────┘
        """
    else:
        return """
상태: 85% | develop +3 | 2시간 전
        """
```

### 세로 공간 최적화

```bash
# ✅ 좋은 예: 컴팩트한 표시
📊 진행률: ████████▂▂▂▂▂ 80% | 🔄 처리 중 | 📁 4/5 파일

# ❌ 나쁜 예: 과도한 공간 사용
📊 진행률:
████████▂▂▂▂▂
80%

🔄 현재 상태:
처리 중

📁 파일 상태:
4개 완료 / 5개 전체
```

## 🔄 상태 전환 시각화

### 작업 흐름 표시

```bash
# Alfred의 4단계 워크플로우 시각화
단계 1: 의도 파악
┌─ 🔍 분석 ──⏸️─ 📋 계획 ──⏸️─ 🚀 실행 ──⏸️─ ✅ 보고 ──┐
└──────────────────────────────────────────────────────────┘

단계 2: 계획 생성
┌─ 🔍 분석 ──✅─ 📋 계획 ──🔄─ 🚀 실행 ──⏸️─ ✅ 보고 ──┐
└───────────────────────────────────────────────────────────┘

단계 3: 작업 실행
┌─ 🔍 분석 ──✅─ 📋 계획 ──✅─ 🚀 실행 ──🔄─ ✅ 보고 ──┐
└───────────────────────────────────────────────────────────┘

단계 4: 보고 및 커밋
┌─ 🔍 분석 ──✅─ 📋 계획 ──✅─ 🚀 실행 ──✅─ ✅ 보고 ──✅─┐
└──────────────────────────────────────────────────────────┘
```

### TDD 사이클 시각화

```bash
# RED → GREEN → REFACTOR 사이클
🔄 TDD 사이클 진행 중:

🔴 RED 단계:
┌─ 테스트 작성 ──⏸️─ 코드 구현 ──⏸️─ 리팩토링 ──┐
│ ✅ tests/auth/service.test.ts 생성                        │
│ ❌ 3개 테스트 실패 (예상됨)                               │
└───────────────────────────────────────────────────────────┘

🟢 GREEN 단계:
┌─ 테스트 작성 ──✅─ 코드 구현 ──🔄─ 리팩토링 ──┐
│ ✅ src/auth/service.ts 최소 구현                          │
│ ✅ 3개 테스트 통과                                        │
└───────────────────────────────────────────────────────────┘

♻️ REFACTOR 단계:
┌─ 테스트 작성 ──✅─ 코드 구현 ──✅─ 리팩토링 ──🔄─┐
│ 🔄 코드 품질 개선 중...                                  │
│ 📊 복잡도: 8 → 6                                         │
└───────────────────────────────────────────────────────────┘
```

## 🔍 정보 계층 구조 (Information Hierarchy)

### 시각적 강조 (Visual Emphasis)

```bash
# 1단계: 가장 중요한 정보
🔴 Alfred: 치명적인 오류 발생!
┌─ 오류: Authentication failed                           │
│ 위치: src/auth/service.ts:42                           │
└─────────────────────────────────────────────────────────┘

# 2단계: 컨텍스트 정보
📝 세부 정보:
• JWT 서명 검증 실패
• 만료된 토큰 사용

# 3단계: 권장 조치
💡 해결 방법:
1. 토큰 갱신 로직 확인
2. /alfred:2-run AUTH-001 재실행
```

### 정보 밀도 조절 (Information Density)

```bash
# 상세 모드 (사용자 요청 시)
📊 Alfred: 상세 분석 보고서
┌─ SPEC-AUTH-001 분석 결과 ─────────────────────────────────┐
│ 📋 상태: 구현 준비 완료                                    │
│ 📊 버전: v0.0.1 (2025-10-16 생성)                        │
│ 🔗 TAG 체인:                                            │
│   ✅ @SPEC:AUTH-001 → .moai/specs/SPEC-AUTH-001/spec.md   │
│   ✅ @TEST:AUTH-001 → tests/auth/service.test.ts          │
│   ✅ @CODE:AUTH-001 → src/auth/service.ts                 │
│   ⏸️ @DOC:AUTH-001 → 생성 대기 중                         │
│                                                            │
│ 🧪 테스트 커버리지:                                       │
│   • 전체: 0% (아직 구현되지 않음)                         │
│   • 예상: 90% 이상                                        │
│                                                            │
│ 🎯 구현 우선순위:                                         │
│   1. JWT 발급 로직                                        │
│   2. 토큰 검증 미들웨어                                   │
│   3. 에러 핸들링                                          │
└───────────────────────────────────────────────────────────┘

# 컴팩트 모드 (기본)
📊 SPEC-AUTH-001: 준비 완료 | 🔗 TAG: 3/4 | 🧪 커버리지: 예상 90%+
```

## 🌍 접근성 고려 (Accessibility Considerations)

### 색맹 친화적 디자인

```bash
# 색상 + 아이콘 + 텍스트 (3중 보장)
✅ 성공: 초록색 + 체크 아이콘 + "성공" 텍스트
❌ 실패: 빨간색 + X 아이콘 + "실패" 텍스트
⚠️ 경고: 노란색 + 경고 아이콘 + "경고" 텍스트
ℹ️ 정보: 파란색 + 정보 아이콘 + "정보" 텍스트

# 패턴 기반 구분 (색상 없이도 구별 가능)
[✓] 완료: 체크 패턴
[✗] 실패: X 패턴
[!] 경고: 느낌표 패턴
[i] 정보: 소문자 i 패턴
```

### 화면 리더기 최적화

```bash
# 명확한 구조화
헤딩 1: Alfred - 분석 보고서
헤딩 2: SPEC-AUTH-001 상태
목록:
- 항목 1: TAG 체인 무결성 확인
- 항목 2: 테스트 커버리지 분석
- 항목 3: 구현 우선순위 제안

# 의미 있는 레이블
[label] 상태: 진행 중 [spinner] 처리 중입니다
[label] 진행률: 80% [progress-bar] 80% 완료되었습니다
```

## ⚡ 성능 최적화 (Performance Optimization)

### 출력 버퍼링 (Output Buffering)

```python
# Claude Code 스트리밍 최적화
def streaming_output():
    # 즉시 피드백 (사용자 경험)
    yield "🔍 Alfred: 분석 시작...\n"

    # 점진적 정보 공개
    for step in analysis_steps:
        yield f"⠋ {step.description}...\n"
        result = execute_step(step)
        yield f"✅ {step.description} 완료\n"

    # 최종 요약
    yield "✅ 분석 완료!\n"
```

### 딜레이 최소화

```bash
# ✅ 좋은 예: 빠른 피드백
🔍 Alfred: 요청 수신됨
📋 분석 시작 (예상: 30초)
⠋ 1/5: 프로젝트 구조 스캔...
⠙ 2/5: 의존성 분석...
⠹ 3/5: SPEC 확인...

# ❌ 나쁜 예: 긴 딜레이
[30초 동안 아무 출력 없음]
🔍 Alfred: 분석이 완료되었습니다
```

## 🎯 사용자 심리 고려 (User Psychology)

### 불확실성 감소

```bash
# ✅ 좋은 예: 명확한 기대치 설정
🔍 Alfred: SPEC-AUTH-001 분석 시작
⏱️  예상 시간: 2분
📊 처리할 항목: 5개
⠋ 1/5: 기존 SPEC 확인...

# ❌ 나쁜 예: 불확실한 진행 상황
🔍 Alfred: 분석 중...
[언제 끝날지 모르는 대기]
```

### 진행 상황 표시

```bash
# 작은 단계로 나누기 (성취감 제공)
📝 문서 생성 중:
✅ 1/10: 제목 작성
✅ 2/10: 개요 작성
✅ 3/10: 요구사항 작성
...
✅ 10/10: 최종 검토

# 거대한 단계 (불안감 유발)
📝 문서 생성 중:
⠋ 처리 중... (0% 또는 100%만 표시)
```

## 🔧 구현 가이드 (Implementation Guide)

### Claude Code 통합 패턴

```javascript
// Alfred의 TUX 구현 예시
class AlfredTUX {
  async streamingResponse(prompt) {
    // 즉시 응답 시작
    await this.emit("🔍 Alfred: 요청 처리 중...\n");

    // 점진적 처리
    for (const step of this.getSteps()) {
      await this.emit(`⠋ ${step.description}...\n`);

      try {
        const result = await this.executeStep(step);
        await this.emit(`✅ ${step.description} 완료\n`);
        this.updateProgress(step.progress);
      } catch (error) {
        await this.emit(`❌ ${step.description} 실패\n`);
        await this.emit(`💡 해결 방법: ${error.suggestion}\n`);
        break;
      }
    }

    // 최종 요약
    await this.emit("✅ 모든 작업 완료!\n");
    await this.showSummary();
  }

  updateProgress(percentage) {
    const bar = this.createProgressBar(percentage);
    this.emit(`📊 진행률: ${bar} ${percentage}%\n`);
  }

  createProgressBar(percentage) {
    const filled = Math.floor(percentage / 10);
    const empty = 10 - filled;
    return "█".repeat(filled) + "▂".repeat(empty);
  }
}
```

### AskUserQuestion 통합

```python
# Alfred의 상호작용 패턴
async def get_user_choice():
    # 명확한 옵션 제시
    question = {
        "question": "다음 작업을 선택하세요:",
        "header": "다음 단계",
        "options": [
            {
                "label": "다음 기능 계획",
                "description": "/alfred:1-plan으로 새 SPEC 생성"
            },
            {
                "label": "구현 검토",
                "description": "현재 코드 검토 및 개선"
            },
            {
                "label": "브랜치 병합",
                "description": "develop 브랜치에 변경사항 병합"
            }
        ]
    }

    # 명확한 기대치 설정
    yield "❓ 다음 작업을 선택해주세요...\n"

    # 비차단 대기
    response = await ask_user_question(question)

    # 즉시 피드백
    yield f"✅ '{response.selected}' 선택됨\n"

    return response.selected
```

---

## 📚 참고 자료

### Claude Code 공식 문서
- Claude Code TUX 가이드라인
- 스트리밍 출력 최적화
- 터미널 호환성 가이드

### 관련 Skills
- `Skill("moai-alfred-reporting")` - 보고서 스타일 가이드
- `Skill("moai-alfred-personas")` - 적응형 페르소나 시스템
- `Skill("moai-alfred-ask-user-questions")` - 상호작용 패턴

---

**Alfred TUX Guide**: Claude Code 공식 TUX 원칙을 Alfred의 출력에 최적화하여 최고의 터미널 사용자 경험을 제공합니다.