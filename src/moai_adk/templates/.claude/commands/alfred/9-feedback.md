---
name: alfred:9-feedback
description: "GitHub 이슈 빠르게 생성하기 (자동 정보 수집 + 템플릿)"
allowed-tools:
- Bash(gh:*)
- Bash(uv:*)
- AskUserQuestion
- Skill
skills:
- moai-alfred-issue-labels
- moai-alfred-feedback-templates
---

# 🎯 MoAI-ADK Alfred 9-Feedback: GitHub 이슈 빠른 작성 도구

> **목적**: 버그, 기능 요청, 개선 제안, 질문을 빠르고 정확하게 GitHub에 기록합니다.

## 📋 명령어 목적

개발자가 버그를 발견하거나 아이디어가 생기면 즉시 GitHub 이슈로 기록할 수 있도록 지원합니다.

- ✅ **빠름**: 2-3단계로 이슈 생성 완료
- ✅ **정확함**: 자동으로 버전, 환경 정보 수집
- ✅ **정리됨**: 라벨별 구조화된 템플릿
- ✅ **간단함**: 명령어만 실행하면 끝 (`/alfred:9-feedback`)

**사용 방법**:
```bash
/alfred:9-feedback
```

완료!

---

## 🚀 실행 프로세스 (2단계)

### Step 1: 명령어 실행
```bash
/alfred:9-feedback
```

이렇게만 입력하면, Alfred가 나머지를 처리합니다.

---

### Step 2: 필수 정보 한 번에 수집 (AskUserQuestion - multiSelect)

**한 번의 질문**으로 다음을 모두 선택합니다:

```
┌─ 이슈 타입 (필수, 중복선택 불가)
│  ├─ 🐛 버그 리포트 - 문제 발생
│  ├─ ✨ 기능 요청 - 새로운 기능 제안
│  ├─ ⚡ 개선 사항 - 기존 기능 개선
│  ├─ 📚 문서 - 문서 개선
│  ├─ 🔄 리팩토링 - 코드 구조 개선
│  └─ ❓ 질문 - 팀에 물어보기
│
├─ 우선순위 (기본값: 중간)
│  ├─ 🔴 긴급 - 시스템 중단, 데이터 손실
│  ├─ 🟠 높음 - 주요 기능 장애
│  ├─ 🟡 중간 - 일반 우선순위
│  └─ 🟢 낮음 - 나중에 괜찮음
│
└─ 템플릿 선택 (선택사항)
   ├─ ✅ 자동 템플릿 생성 (권장)
   └─ 📝 직접 작성하기
```

---

### Step 3: 자동 생성된 템플릿 확인 & 입력

Alfred가 선택한 이슈 타입에 맞는 템플릿을 자동으로 생성합니다.

예를 들어, **버그 리포트** 선택 시:

```markdown
## 버그 설명

[사용자가 입력할 공간]

## 재현 단계

1. [사용자가 입력]
2. [사용자가 입력]
3. [사용자가 입력]

## 예상 동작

[사용자가 입력할 공간]

## 실제 동작

[사용자가 입력할 공간]

## 환경 정보

🔍 자동 수집된 정보:
- MoAI-ADK 버전: 0.22.5
- Python 버전: 3.11.5
- OS: macOS 14.2
- 현재 브랜치: feature/SPEC-001
- 커밋되지 않은 변경사항: 3개
```

사용자는 `[사용자가 입력할 공간]` 부분만 채우면 됩니다.

---

Alfred가 자동으로 처리합니다:

1. **환경 정보 수집** (`python3 .moai/scripts/feedback-collect-info.py`):
   - MoAI-ADK 버전
   - Python 버전, OS
   - Git 상태 (현재 브랜치, 커밋되지 않은 변경사항)
   - 작업 중인 SPEC

2. **라벨 매핑** (`Skill("moai-alfred-issue-labels")`):
   - 이슈 타입 → 라벨 (예: 버그 → "bug", "reported")
   - 우선순위 → 라벨 (예: 높음 → "priority-high")

3. **제목 자동 생성**: "🐛 [BUG] 버그 설명..."

4. **GitHub Issue 생성**:
   ```bash
   gh issue create \
     --title "🐛 [BUG] 버그 설명" \
     --body "## 버그 설명\n...[템플릿 + 환경 정보]..." \
     --label "bug" \
     --label "reported" \
     --label "priority-high"
   ```

5. **결과 표시**:
   ```
   ✅ GitHub Issue #234 생성 완료!

   📋 제목: 🐛 [BUG] 버그 설명
   🔴 우선순위: 높음
   🏷️ 라벨: bug, reported, priority-high
   🔗 URL: https://github.com/owner/repo/issues/234

   💡 다음: 커밋 메시지에서 이 Issue를 참조하거나 SPEC과 연결하세요
   ```

---

## 📊 라벨 매핑 (via `Skill("moai-alfred-issue-labels")`)

| 타입 | 주요 라벨 | 우선순위 | 최종 라벨 |
|------|---------|---------|---------|
| 🐛 버그 | bug, reported | 높음 | bug, reported, priority-high |
| ✨ 기능 | feature-request, enhancement | 중간 | feature-request, enhancement, priority-medium |
| ⚡ 개선 | improvement, enhancement | 중간 | improvement, enhancement, priority-medium |
| 📚 문서 | documentation | 중간 | documentation, priority-medium |
| 🔄 리팩토링 | refactor | 중간 | refactor, priority-medium |
| ❓ 질문 | question, help-wanted | 중간 | question, help-wanted, priority-medium |

---

## ⚠️ 규칙

### ✅ 해야 할 것

- ✅ multiSelect로 필수 정보 한 번에 수집 (이슈 타입, 우선순위)
- ✅ 사용자 입력을 정확하게 보존하기
- ✅ 자동 정보 수집 스크립트 실행 (`python3 .moai/scripts/feedback-collect-info.py`)
- ✅ `Skill("moai-alfred-issue-labels")`로 라벨 매핑하기
- ✅ `Skill("moai-alfred-feedback-templates")`로 템플릿 제공하기
- ✅ 생성 후 Issue URL 표시하기

### ❌ 하지 말아야 할 것

- ❌ 명령어 인자 사용 (`/alfred:9-feedback --bug` 잘못됨 → 그냥 `/alfred:9-feedback` 사용)
- ❌ 4단계 이상 질문하기
- ❌ 사용자 입력 수정하기
- ❌ 라벨 없이 Issue 생성하기
- ❌ 라벨 하드코딩하기 (스킬 기반 매핑 사용)

---

## 💡 주요 장점

1. **⚡ 빠름**: 2-3단계로 30초 이내 완료
2. **🤖 자동**: 버전, 환경 정보 자동 수집
3. **📋 정확함**: 라벨별 구조화된 템플릿
4. **🏷️ 의미있음**: `moai-alfred-issue-labels` 스킬 기반 분류
5. **🔄 재사용 가능**: `/alfred:1-plan`, `/alfred:3-sync`과 라벨 공유
6. **한국어**: 모든 텍스트가 한국어로 작성됨

---

## 📝 사용 예시

**1단계**: 명령어 실행
```bash
/alfred:9-feedback
```

**2단계**: 필수 정보 선택
```
이슈 타입: [🐛 버그 리포트] 선택
우선순위: [🟠 높음] 선택
템플릿: [✅ 자동 생성] 선택
```

**3단계**: 템플릿 작성
```markdown
## 버그 설명
로그인 버튼을 클릭해도 반응이 없습니다.

## 재현 단계
1. 홈페이지 접속
2. 오른쪽 상단의 로그인 버튼 클릭
3. 아무 반응 없음

## 예상 동작
로그인 모달이 나타나야 함

## 실제 동작
아무 일도 일어나지 않음

## 환경 정보
🔍 자동 수집된 정보:
- MoAI-ADK 버전: 0.22.5
- Python 버전: 3.11.5
- OS: macOS 14.2
```

**결과**: Issue #234 자동 생성 + URL 표시 ✅

---

**지원 버전**: MoAI-ADK v0.22.5+
