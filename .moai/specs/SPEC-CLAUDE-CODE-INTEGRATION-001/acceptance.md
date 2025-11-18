---
spec_id: SPEC-CLAUDE-CODE-INTEGRATION-001
title: "Claude Code v2.0.43 통합 - 수용 기준 및 테스트"
version: 1.0
created_at: 2025-11-18
---

## 수용 기준 (Acceptance Criteria)

본 문서는 Claude Code v2.0.43 통합이 성공적으로 완료되었는지 검증하기 위한 테스트 시나리오와 품질 기준을 정의합니다.

---

## 테스트 시나리오 (Given-When-Then Format)

### TC-1: SessionStart Hook - Haiku 모델 사용

**Given** (조건)
- MoAI-ADK 프로젝트가 초기화되고 `.claude/settings.json`이 ConfigManager로 로드됨
- SessionStart Hook이 haiku 모델로 설정됨

**When** (행동)
- Claude Code 세션이 시작되고 SessionStart Hook이 트리거됨
- subagent_start__show_project_info.py 가 실행됨

**Then** (결과)
- Hook이 haiku 모델로 실행됨
- 실행 시간이 100ms 이내
- 프로젝트 정보가 표시됨
- systemMessage가 반환됨
- 비용: Sonnet 대비 73% 절감

**검증 방법**
```bash
# .claude/settings.json 에서 model 필드 확인
grep -A 5 "SessionStart" .claude/settings.json | grep "model"
# 예상 출력: "model": "haiku"

# Hook 실행 시간 측정
time uv run .claude/hooks/alfred/session_start__show_project_info.py < /dev/null
# 예상: real 0m0.100s 이내
```

---

### TC-2: PreToolUse Hook - 파일 수정 전 자동 체크포인트

**Given** (조건)
- Edit 또는 Write 도구가 호출될 준비 상태
- `.moai/temp/checkpoint/` 디렉토리가 존재

**When** (행동)
- 파일 수정 작업 (Edit 또는 Write)이 시작됨
- PreToolUse Hook이 트리거됨

**Then** (결과)
- pre_tool__auto_checkpoint.py 가 haiku 모델로 실행됨
- 현재 파일 상태가 `.moai/temp/checkpoint/` 에 저장됨
- Hook이 성공적으로 완료 (continue: true)
- 파일 수정 작업이 계속 진행됨

**검증 방법**
```bash
# 1. Hook 설정 확인
grep -A 5 "PreToolUse" .claude/settings.json | grep "model"
# 예상: "model": "haiku"

# 2. 체크포인트 디렉토리 확인
ls -la .moai/temp/checkpoint/ 2>/dev/null || echo "디렉토리 없음 (처음 실행)"

# 3. Edit 도구 사용 후 체크포인트 확인
# Edit 작업 후:
ls -la .moai/temp/checkpoint/
# 예상: 타임스탬프 파일 생성됨
```

---

### TC-3: UserPromptSubmit Hook - 복잡한 의도 분석 (Sonnet)

**Given** (조건)
- 사용자가 복잡한 자연어 프롬프트를 입력
- UserPromptSubmit Hook이 sonnet 모델로 설정됨

**When** (행동)
- 사용자 프롬프트가 제출됨
- user_prompt__jit_load_docs.py Hook이 실행됨

**Then** (결과)
- Hook이 sonnet 모델로 실행됨 (복잡한 추론 필요)
- 필요한 문서가 JIT로 로드됨
- 사용자 의도가 정확히 분석됨
- 추가 컨텍스트 메시지가 제공됨

**검증 방법**
```bash
# 1. Hook 설정 확인
grep -A 5 "UserPromptSubmit" .claude/settings.json | grep "model"
# 예상: "model": "sonnet"

# 2. 실행 후 시스템 메시지 확인
# 사용자 프롬프트 입력 후:
# → systemMessage가 출력됨
# → 예: "🎯 Loading documentation... (3 files, 50K tokens)"
```

---

### TC-4: SessionEnd Hook - 자동 정리 (Haiku)

**Given** (조건)
- Claude Code 세션이 정상 종료됨
- `.moai/temp/` 디렉토리에 임시 파일들이 존재

**When** (행동)
- SessionEnd Hook이 트리거됨
- session_end__auto_cleanup.py 가 실행됨

**Then** (결과)
- Hook이 haiku 모델로 실행됨
- `.moai/temp/` 디렉토리 정리
- 캐시 파일 제거
- 로그 압축 (선택사항)
- systemMessage 반환

**검증 방법**
```bash
# 1. Hook 설정 확인
grep -A 5 "SessionEnd" .claude/settings.json | grep "model"
# 예상: "model": "haiku"

# 2. 세션 종료 후 임시 파일 확인
# 세션 종료 후:
ls -la .moai/temp/ 2>/dev/null
# 예상: 최소 파일로 정리됨
```

---

### TC-5: SubagentStart Hook - Context 최적화

**Given** (조건)
- 에이전트가 시작될 준비 상태
- `.claude/hooks/alfred/subagent_start__context_optimizer.py` 존재

**When** (행동)
- 에이전트 (예: spec-builder)가 시작됨
- SubagentStart Hook이 트리거됨

**Then** (결과)
- subagent_start__context_optimizer.py 가 실행됨
- 에이전트별 최적화 전략이 적용됨
  - spec-builder: max_tokens=20000, priority_files=[".moai/specs/", ...]
  - tdd-implementer: max_tokens=30000, priority_files=["src/", "tests/", ...]
  - backend-expert: max_tokens=30000, priority_files=["src/", ...]
- 메타데이터가 `.moai/logs/agent-transcripts/agent-{agent_id}.json` 에 저장됨
- systemMessage가 반환됨

**검증 방법**
```bash
# 1. Hook 파일 존재 확인
ls -l .claude/hooks/alfred/subagent_start__context_optimizer.py
# 예상: 파일 존재

# 2. 에이전트 시작 시 메타데이터 확인
# 에이전트 실행 후:
cat .moai/logs/agent-transcripts/agent-spec-builder.json
# 예상:
# {
#   "agent_id": "spec-builder",
#   "agent_name": "spec-builder",
#   "started_at": "2025-11-18T...",
#   "strategy": "SPEC 작성 - 최소 컨텍스트 로드",
#   "max_tokens": 20000,
#   "priority_files": [...]
# }

# 3. 다양한 에이전트로 테스트
for agent in spec-builder tdd-implementer backend-expert; do
  cat .moai/logs/agent-transcripts/agent-${agent}.json | jq '.strategy, .max_tokens'
done
```

---

### TC-6: SubagentStop Hook - 성능 추적

**Given** (조건)
- 에이전트가 실행 중이고 완료됨
- `.claude/hooks/alfred/subagent_stop__lifecycle_tracker.py` 존재

**When** (행동)
- 에이전트 실행이 완료됨
- SubagentStop Hook이 트리거됨

**Then** (결과)
- subagent_stop__lifecycle_tracker.py 가 실행됨
- 메타데이터가 업데이트됨
  - execution_time_ms: 실행 시간 (ms)
  - execution_time_seconds: 변환된 시간 (s)
  - completed_at: 완료 시각
  - status: "completed" 또는 "failed"
- 성능 통계가 `.moai/logs/agent-performance.jsonl` 에 JSONL 형식으로 저장됨
- systemMessage가 반환됨

**검증 방법**
```bash
# 1. Hook 파일 존재 확인
ls -l .claude/hooks/alfred/subagent_stop__lifecycle_tracker.py
# 예상: 파일 존재

# 2. 메타데이터 확인
cat .moai/logs/agent-transcripts/agent-spec-builder.json | jq '.execution_time_ms, .completed_at, .status'
# 예상:
# 267333
# "2025-11-18T10:35:12.456789"
# "completed"

# 3. 성능 통계 (JSONL) 확인
tail -5 .moai/logs/agent-performance.jsonl
# 예상:
# {"timestamp": "2025-11-18T10:35:12.456789", "agent_id": "spec-builder", "agent_name": "spec-builder", "execution_time_ms": 267333, "success": true}
# ...

# 4. 성능 분석
jq -r '.agent_name' .moai/logs/agent-performance.jsonl | sort | uniq -c
# 예상: 에이전트별 실행 횟수 통계

# 5. 성공률 확인
jq -r 'select(.success == true) | .agent_name' .moai/logs/agent-performance.jsonl | wc -l
# 예상: 성공한 작업 수

# 6. 가장 느린 에이전트
jq -s 'sort_by(-.execution_time_ms) | .[0:3]' .moai/logs/agent-performance.jsonl | jq '.[] | {agent_name, execution_time_ms}'
```

---

### TC-7: permissionMode - Auto Mode (안전한 작업)

**Given** (조건)
- spec-builder 또는 docs-manager 에이전트가 auto 모드로 설정됨
- 읽기 또는 문서 생성 작업 예정

**When** (행동)
- 에이전트가 안전한 작업 (Read, Write 문서)을 수행
- permissionMode: auto 적용됨

**Then** (결과)
- 사용자 승인 없이 작업 계속 진행
- 작업이 성공적으로 완료됨

**검증 방법**
```bash
# 1. Auto Mode 에이전트 확인
grep "permissionMode: auto" .claude/agents/alfred/*.md | wc -l
# 예상: 11

# 2. spec-builder 확인
grep -A 1 "permissionMode:" .claude/agents/alfred/spec-builder.md
# 예상: permissionMode: auto

# 3. 실제 실행
# spec-builder 에이전트 호출 시:
# → 사용자 승인 없이 SPEC 생성 진행
# → 작업 완료
```

---

### TC-8: permissionMode - Ask Mode (코드 수정)

**Given** (조건)
- tdd-implementer 또는 backend-expert 에이전트가 ask 모드로 설정됨
- 코드 수정 작업 예정

**When** (행동)
- 에이전트가 코드를 수정 (Edit, Write)하려고 함
- permissionMode: ask 적용됨

**Then** (결과)
- Claude Code가 사용자 승인 요청
- 사용자가 승인하면 작업 계속 진행
- 사용자가 거부하면 작업 중단
- 거부 사유 로그 기록

**검증 방법**
```bash
# 1. Ask Mode 에이전트 확인
grep "permissionMode: ask" .claude/agents/alfred/*.md | wc -l
# 예상: 21

# 2. tdd-implementer 확인
grep -A 1 "permissionMode:" .claude/agents/alfred/tdd-implementer.md
# 예상: permissionMode: ask

# 3. 실제 실행
# tdd-implementer 에이전트 호출 시:
# → Claude Code가 승인 요청 표시
# → 사용자 선택 후 진행 또는 중단
```

---

### TC-9: Skills Frontmatter - 도메인 전문 지식 로드

**Given** (조건)
- 에이전트가 skills 필드를 frontmatter에 가지고 있음
- moai-foundation-ears, moai-lang-python 등 Skills 정의됨

**When** (행동)
- 에이전트가 시작되면서 auto_load_skills: true 적용
- SubagentStart Hook이 실행됨

**Then** (결과)
- 에이전트의 frontmatter에서 skills 자동 추출
- 도메인별 Skills 자동 로드
- 최신 기술 정보 및 패턴 제공
- 성능 향상

**검증 방법**
```bash
# 1. Skills 필드 확인
grep -h "^skills:" .claude/agents/alfred/*.md | head -5
# 예상:
# skills:
# - moai-foundation-ears
# - moai-foundation-specs

# 2. 에이전트별 Skills 확인
grep -A 5 "^skills:" .claude/agents/alfred/backend-expert.md
# 예상:
# skills:
# - moai-domain-backend
# - moai-lang-python
# - moai-context7-lang-integration

# 3. Skills 필드가 있는 에이전트 수
grep -l "^skills:" .claude/agents/alfred/*.md | wc -l
# 예상: 32 (모두)
```

---

### TC-10: Graceful Degradation - Hook 실패 처리

**Given** (조건)
- Hook 파일에 버그나 예외 상황 발생
- Hook 실행 중 IOError, JSONDecodeError 등 발생

**When** (행동)
- Hook 실행 중 예외 발생
- try-except 블록에서 캐치됨

**Then** (결과)
- Hook이 graceful degradation 처리
- `continue: true` 로 응답하여 에이전트 계속 실행
- systemMessage에 경고 메시지 포함 (⚠️ 접두사)
- 예: "⚠️ Context optimization skipped: [error]"
- 에러 로그 기록 (선택사항)

**검증 방법**
```bash
# 1. Hook 코드에서 graceful degradation 확인
grep -A 5 "except Exception" .claude/hooks/alfred/subagent_start__context_optimizer.py
# 예상:
# except Exception as e:
#     print(json.dumps({
#         "continue": True,
#         "systemMessage": f"⚠️ Context optimization skipped: {str(e)}"
#     }))

# 2. 실제 테스트 (Hook에 버그 주입)
# .claude/hooks/alfred/subagent_start__context_optimizer.py 수정:
# raise ValueError("Test error")
# 에이전트 실행 후:
# → systemMessage에 ⚠️ Context optimization skipped 출력
# → 에이전트는 계속 실행됨
```

---

### TC-11: Cost Savings - 비용 절감 검증

**Given** (조건)
- 100시간 규모의 개발 작업 로그 데이터 확보
- 에이전트 성능 통계 데이터 수집됨

**When** (행동)
- 성능 통계 JSONL 파일 분석
- Haiku vs Sonnet 모델 사용량 비교

**Then** (결과)
- Haiku 사용 비율: ~70-80%
- Sonnet 사용 비율: ~20-30%
- 평균 비용: 약 70% 절감 달성
- 월간 비용 절감: $20-$500 (조직 규모별)

**검증 방법**
```bash
# 1. 현재 Hook별 모델 집계
grep '"model"' .claude/settings.json | grep -o '"model": "[^"]*"' | sort | uniq -c
# 예상:
# 5 "model": "haiku"  (SessionStart, PreToolUse, SessionEnd, SubagentStart, SubagentStop)
# 1 "model": "sonnet" (UserPromptSubmit)

# 2. 에이전트별 실행 횟수 집계 (100회 이상)
jq -r '.agent_name' .moai/logs/agent-performance.jsonl | sort | uniq -c | sort -rn
# 예상: 에이전트별 실행 횟수

# 3. 비용 계산 (Haiku 기반 Hook들)
# Hook 비용 = token_count × model_cost
# Haiku: 2-5K tokens × $0.0008 = $0.0016-$0.004
# 1,000회 Hook 실행 시: $1.6-$4 (기존: $6-$15)
```

---

### TC-12: 에지 케이스 - 파일 손상 및 복구

**Given** (조건)
- `.moai/logs/agent-transcripts/agent-{id}.json` 파일이 손상됨 (invalid JSON)
- 또는 파일이 누락됨

**When** (행동)
- Hook이 메타데이터를 읽으려고 함
- JSONDecodeError 또는 FileNotFoundError 발생

**Then** (결과)
- Hook이 예외를 처리
- 기본 메타데이터 또는 빈 dict로 초기화
- 새로운 메타데이터로 덮어쓰기
- 에이전트는 계속 실행됨

**검증 방법**
```bash
# 1. 손상된 JSON 파일 생성
echo "invalid json {" > .moai/logs/agent-transcripts/agent-test.json

# 2. 에이전트 실행
# → Hook이 JSONDecodeError 처리
# → 새로운 메타데이터로 재작성
# → 파일이 정상화됨

# 3. 결과 확인
cat .moai/logs/agent-transcripts/agent-test.json | jq .
# 예상: 정상적인 JSON 구조
```

---

## 품질 기준 (Quality Gates)

### 1. Hook 실행 성공률

**기준**: 99% 이상

**검증**:
```bash
# 성공률 계산
success=$(jq -r 'select(.success == true)' .moai/logs/agent-performance.jsonl | wc -l)
total=$(wc -l < .moai/logs/agent-performance.jsonl)
echo "Success Rate: $((success * 100 / total))%"
# 예상: >= 99%
```

### 2. Hook 실행 시간

**기준**: 평균 < 500ms, 95 percentile < 1s

**검증**:
```bash
# 실행 시간 통계
jq '.execution_time_ms' .moai/logs/agent-performance.jsonl | \
  jq -s '{
    mean: (add / length),
    min: min,
    max: max,
    p95: sort[length * 0.95 | floor]
  }'
# 예상:
# {
#   "mean": 250,
#   "min": 50,
#   "max": 800,
#   "p95": 450
# }
```

### 3. Context 로드 효율성

**기준**: 평균 70% 토큰 절감 달성

**검증**:
```bash
# Context 최적화 전후 비교
# Before: 전체 코드베이스 ~100K tokens
# After: 우선순위 파일만 ~20-50K tokens (에이전트별)
# 평균: (100 - 35) / 100 = 65% 절감 ✅

# 에이전트별 max_tokens 확인
jq '.max_tokens' .moai/logs/agent-transcripts/agent-*.json | \
  jq -s '{
    mean: (add / length),
    min: min,
    max: max
  }'
# 예상:
# {
#   "mean": 25000,
#   "min": 15000,
#   "max": 50000
# }
```

### 4. 메타데이터 정합성

**기준**: 모든 completed_at 과 execution_time_ms 가 일관성 있게 기록됨

**검증**:
```bash
# 메타데이터와 JSONL 동기화 확인
# agent-transcripts/ 파일과 agent-performance.jsonl 비교
diff <(jq -r '.agent_id' .moai/logs/agent-transcripts/agent-*.json | sort) \
     <(jq -r '.agent_id' .moai/logs/agent-performance.jsonl | sort | uniq)
# 예상: 동일한 agent_id 목록
```

### 5. 권한 모드 설정 완성도

**기준**: 모든 32개 에이전트가 auto 또는 ask 모드 설정

**검증**:
```bash
# 권한 모드 설정 확인
missing=$(grep -L "^permissionMode:" .claude/agents/alfred/*.md)
if [ -z "$missing" ]; then
  echo "✅ All 32 agents have permissionMode"
else
  echo "❌ Missing permissionMode: $missing"
fi

# auto vs ask 비율
echo "Auto: $(grep -l 'permissionMode: auto' .claude/agents/alfred/*.md | wc -l)"
echo "Ask: $(grep -l 'permissionMode: ask' .claude/agents/alfred/*.md | wc -l)"
# 예상: Auto: 11, Ask: 21
```

### 6. Skills Frontmatter 설정 완성도

**기준**: 모든 32개 에이전트가 skills 필드 보유

**검증**:
```bash
# Skills 필드 설정 확인
missing=$(grep -L "^skills:" .claude/agents/alfred/*.md)
if [ -z "$missing" ]; then
  echo "✅ All 32 agents have skills field"
else
  echo "❌ Missing skills: $missing"
fi

# Skills 개수 확인
jq -r '.agent_name' .claude/agents/alfred/*.md | sort | uniq -c
# 예상: 32개 모두
```

---

## 정의된 완료 기준 (Definition of Done)

### Phase 1 완료 조건

- [ ] `.claude/settings.json` 에 6개 Hook 모두 model 필드 추가
  - [ ] SessionStart: "model": "haiku"
  - [ ] PreToolUse: "model": "haiku"
  - [ ] UserPromptSubmit: "model": "sonnet"
  - [ ] SessionEnd: "model": "haiku"
  - [ ] SubagentStart: "model": "haiku"
  - [ ] SubagentStop: "model": "haiku"

- [ ] 각 Hook 동작 검증 (수동 테스트)
  - [ ] SessionStart Hook 실행 시간 < 100ms
  - [ ] PreToolUse Hook 실행 시간 < 50ms
  - [ ] UserPromptSubmit Hook 실행 완료
  - [ ] SessionEnd Hook 실행 시간 < 100ms

### Phase 2 완료 조건

- [ ] `.claude/hooks/alfred/subagent_start__context_optimizer.py` 완성
  - [ ] 8개 이상 에이전트의 context_strategies 정의
  - [ ] 메타데이터 저장 기능 동작
  - [ ] Graceful Degradation 구현

- [ ] `.claude/hooks/alfred/subagent_stop__lifecycle_tracker.py` 완성
  - [ ] 실행 시간 측정 기능 동작
  - [ ] 메타데이터 업데이트 기능 동작
  - [ ] JSONL 파일에 성능 통계 기록

- [ ] 32개 에이전트에 permissionMode 설정
  - [ ] auto 모드: 11개
  - [ ] ask 모드: 21개

- [ ] 32개 에이전트에 skills 필드 추가
  - [ ] 모든 에이전트가 skills 배열 포함

### 검증 완료 조건

- [ ] 비용 절감 검증 (70% 이상)
- [ ] Hook 성공률 99% 이상
- [ ] Hook 평균 실행 시간 < 500ms
- [ ] 모든 메타데이터 파일 생성 확인
- [ ] 모든 권한 모드 설정 완료
- [ ] 모든 Skills 필드 추가 완료

---

## 테스트 실행 및 리포팅

### 테스트 실행 절차

1. **수동 검증 (초기)**
   ```bash
   # Phase 1 검증
   grep -c '"model"' .claude/settings.json
   # 예상: 6 이상

   # Phase 2 검증
   ls -l .claude/hooks/alfred/subagent_*.py
   # 예상: 2개 파일 존재
   ```

2. **Hook 실행 테스트**
   ```bash
   # SessionStart Hook
   echo '{}' | uv run .claude/hooks/alfred/session_start__show_project_info.py

   # SubagentStart Hook
   echo '{"agentId":"spec-builder","agentName":"spec-builder","prompt":"test"}' | \
     uv run .claude/hooks/alfred/subagent_start__context_optimizer.py
   ```

3. **통합 테스트**
   ```bash
   # 에이전트 실행 후 메타데이터 확인
   ls .moai/logs/agent-transcripts/agent-*.json
   cat .moai/logs/agent-performance.jsonl
   ```

### 테스트 리포트 템플릿

```markdown
## Test Report - SPEC-CLAUDE-CODE-INTEGRATION-001

**실행 날짜**: 2025-11-18
**테스터**: [이름]
**환경**: [OS, Python 버전, uv 버전]

### 테스트 결과 요약

| 테스트 케이스 | 상태 | 비고 |
|-------------|------|------|
| TC-1: SessionStart Hook | ✅ | 실행 시간: 95ms |
| TC-2: PreToolUse Hook | ✅ | 체크포인트 생성 확인 |
| TC-3: UserPromptSubmit Hook | ✅ | Sonnet 모델 사용 |
| TC-4: SessionEnd Hook | ✅ | 임시 파일 정리 |
| TC-5: SubagentStart Hook | ✅ | Context 최적화 적용 |
| TC-6: SubagentStop Hook | ✅ | 성능 통계 기록 |
| TC-7: permissionMode Auto | ✅ | 사용자 승인 없이 진행 |
| TC-8: permissionMode Ask | ✅ | 사용자 승인 요청 |
| TC-9: Skills Frontmatter | ✅ | 도메인 지식 로드 |
| TC-10: Graceful Degradation | ✅ | 예외 처리 정상 |
| TC-11: Cost Savings | ✅ | 70% 절감 달성 |
| TC-12: Edge Case 처리 | ✅ | 파일 손상 복구 |

### 품질 기준 검증

| 기준 | 목표 | 결과 | 상태 |
|------|------|------|------|
| Hook 성공률 | 99% | 99.5% | ✅ |
| Hook 실행 시간 | < 500ms | 평균 245ms | ✅ |
| Context 절감 | 70% | 71% | ✅ |
| 메타데이터 정합성 | 100% | 100% | ✅ |
| 권한 모드 설정 | 32/32 | 32/32 | ✅ |
| Skills 필드 | 32/32 | 32/32 | ✅ |

### 최종 결론

**PASS** ✅ - 모든 요구사항 충족

**승인자**: [서명]
**승인 날짜**: YYYY-MM-DD
```

---

## 다음 단계

1. **본 SPEC 제출 전**:
   - [ ] 모든 테스트 케이스 실행
   - [ ] 품질 기준 검증
   - [ ] 테스트 리포트 작성

2. **배포 전**:
   - [ ] 문서화 (`.moai/docs/hook-integration.md`)
   - [ ] 템플릿 동기화
   - [ ] 신규 프로젝트에 설정 적용

3. **배포 후 모니터링**:
   - [ ] 1주일: Hook 동작 모니터링
   - [ ] 1개월: 비용 절감 검증
   - [ ] 3개월: 성능 분석 및 최적화
