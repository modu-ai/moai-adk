# API Error 400 "no low surrogate" 예방 가이드

## ⚠️ 에러 개요

**에러 메시지:**
```
API Error: 400 {"type":"error","error":{"type":"invalid_request_error","message":"The request body is not valid JSON: no low surrogate in string: line 1 column 203558 (char 203557)"}}
```

**원인:** Claude Code의 JSON 직렬화 버그 (UTF-16 surrogate pair 불완전 인코딩)
**상태:** 알려진 버그, GitHub Issues #1832, #5440, #1709 등에서 추적 중

---

## 🎯 안전한 작업 패턴

### 1. 세션 관리

**⏱️ 세션 시간 제한**
- 권장: 20-30분 단위로 작업
- 경고: 5-10분 경과 시 context 누적 모니터링
- 한계: 15분 이상 지속되는 복잡한 작업은 세션 분리

**중간 체크포인트 (모든 작업)**
```bash
# 20분마다 실행
git add . && git commit -m "checkpoint: progress update"
```

### 2. 작업 단위 축소

**✅ Recommended (각 SPEC별 독립 세션)**
```bash
/alfred:1-plan "기능명"
/alfred:2-run SPEC-001
/alfred:3-sync auto SPEC-001
# → 세션 종료 후 새 세션 시작
```

**❌ Anti-Pattern (멀티 SPEC 동시 진행)**
```bash
/alfred:2-run SPEC-001 SPEC-002 SPEC-003  # 위험!
```

### 3. 이모지 사용 규칙 (CRITICAL)

**❌ 완전 금지 (JSON 인코딩 에러)**
```python
# AskUserQuestion 필드에서 절대 사용 금지
questions = [{
    "question": "어느 기능을 선택할까요? 🚀",  # ❌ NO
    "header": "기능 선택 ✨",                  # ❌ NO
    "options": [{
        "label": "Option 1 🎯",                # ❌ NO
        "description": "설명 입니다 💡"        # ❌ NO
    }]
}]
```

**✅ 허용 범위**
```python
# 일반 대화 (최소화)
"이 기능은 정말 중요합니다 ⭐"  # 최대 1-2개만

# 마크다운 문서 (제한적)
"## 주요 기능 ✨"  # 문서 가독성을 위해 사용 가능
```

### 4. Context 정리

**자동 정리 (config.json 설정됨)**
- cleanup_days: 3일
- max_reports: 5개
- 타겟: `.moai/reports/`, `.moai/cache/`, `.moai/temp/`, `.moai/memory/`

**수동 정리 (필요 시)**
```bash
# 현재 세션 메모리 제거
rm -rf ~/.claude/memory/*.json
rm -rf .moai/memory/*.json

# 캐시 초기화
rm -rf .moai/cache/*
rm -rf .moai/temp/*
```

---

## 🚨 경고 신호 (Warning Signs)

다음 중 하나라도 해당되면 세션 재시작 권장:

- [ ] 세션 시간 > 10분
- [ ] 연속된 긴 문서 편집 (README, SPEC 등)
- [ ] AskUserQuestion 반복 호출 (5회 이상)
- [ ] 복잡한 마크다운 생성 중
- [ ] 대용량 파일 편집 예정

---

## 🔧 설정 확인

**현재 설정 (최적화됨):**
```bash
cat .moai/config.json | jq '.auto_cleanup'
```

**출력 예상:**
```json
{
  "enabled": true,
  "cleanup_days": 3,
  "max_reports": 5,
  "cleanup_targets": [
    ".moai/reports/*.json",
    ".moai/reports/*.md",
    ".moai/cache/*",
    ".moai/temp/*",
    ".moai/memory/*.json"
  ]
}
```

---

## 🆘 에러 발생 시 대응

### Step 1: 즉시 복구
```bash
# 현재 세션 종료
exit

# 새 세션 시작
claude-code
```

### Step 2: 작업 복구
```bash
# 마지막 git 상태 확인
git status
git log --oneline -5

# 필요시 중간 커밋으로 돌아가기
git reset --hard <commit-hash>
```

### Step 3: 원인 분석
```bash
# context 크기 확인
du -sh .moai/memory/ .moai/cache/

# 최근 에러 로그 확인
cat .moai/logs/sessions/*.json | tail -20
```

---

## 📊 모니터링 지표

**세션별 추적 사항:**
- 세션 시간 (권장: 20-30분)
- Commit 빈도 (권장: 20분마다)
- AskUserQuestion 호출 횟수 (경고: 5회 이상)
- Memory 파일 크기 (경고: 100KB 이상)

**설정 최적화 현황:**
- ✅ cleanup_days: 3일 (이전: 7일)
- ✅ max_reports: 5개 (이전: 10개)
- ✅ 메모리 파일 정리 추가

---

## 📌 GitHub Issues (추적 중)

| Issue | Title | Status |
|-------|-------|--------|
| #1832 | JSON Parsing Error: Invalid Low Surrogate | Open |
| #5440 | JSON Serialization Failure: Unicode Surrogate Pair Error | Open |
| #1709 | no low surrogate in string (대용량 payload) | Open |

**구독 방법:**
```bash
gh issue view 1832 --web  # GitHub에서 열기
```

---

## ✅ 체크리스트

**새 세션 시작 전:**
- [ ] 이전 세션 완료 및 커밋
- [ ] git status 확인 (변경사항 없음)
- [ ] 20-30분 작업 계획 수립

**작업 중:**
- [ ] 매 20분마다 commit
- [ ] AskUserQuestion에 이모지 사용 금지
- [ ] 세션 시간 모니터링

**세션 종료 시:**
- [ ] 모든 변경사항 커밋
- [ ] .moai/memory/ 정리 여부 확인
- [ ] 다음 세션을 위한 메모 작성

---

## 🔗 Related Documentation

- MoAI-ADK CLAUDE.md: `./CLAUDE.md`
- Config Schema: `.moai/config.json`
- GitHub Issues: https://github.com/anthropics/claude-code/issues

**Last Updated:** 2025-11-09
**Author:** Alfred (debug-helper analysis)
