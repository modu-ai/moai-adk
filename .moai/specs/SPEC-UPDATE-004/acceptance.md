# 수락 기준: SPEC-UPDATE-004

> **Sub-agents를 Skills로 통합**

---

## 1. 기능 수락 기준

### 1.1 TAG 스캔 기능 (Phase 1)

#### Scenario 1: TAG 스캔 기본 동작
```gherkin
Given Alfred가 사용자 요청을 받음
And 요청이 "TAG 스캔"을 포함함
When Alfred가 moai-alfred-tag-scanning Skill을 JIT 로드
Then TAG 스캔이 정상적으로 실행되어야 함
And 결과가 기존 tag-agent와 동일해야 함
```

**검증 방법**:
```bash
# Before (tag-agent 방식)
@agent-tag-agent "AUTH 도메인 TAG 목록 조회"
# 결과: AUTH-001, AUTH-002, ...

# After (Skills 참조 방식)
# Alfred가 자동으로 moai-alfred-tag-scanning Skill 로드
# 결과: AUTH-001, AUTH-002, ... (동일)
```

#### Scenario 2: 고아 TAG 탐지
```gherkin
Given .moai/specs/에 @SPEC:ORPHAN-001이 있음
And src/에 @CODE:ORPHAN-001은 없음
When Alfred가 고아 TAG 탐지를 실행
Then "ORPHAN-001은 고아 TAG입니다" 메시지가 반환되어야 함
```

**검증 방법**:
```bash
# 고아 TAG 생성 (테스트용)
echo "# @SPEC:ORPHAN-001" > .moai/specs/SPEC-ORPHAN-001/spec.md

# 탐지 실행
@agent-tag-agent "고아 TAG 탐지"

# 예상 결과
# ❌ ORPHAN-001: @CODE 없음 (고아 TAG)
```

#### Scenario 3: TAG 체인 검증
```gherkin
Given @SPEC:AUTH-001이 존재함
And @TEST:AUTH-001이 존재함
And @CODE:AUTH-001이 존재함
When Alfred가 TAG 체인 검증을 실행
Then "AUTH-001 TAG 체인이 완전합니다" 메시지가 반환되어야 함
```

**검증 방법**:
```bash
# TAG 체인 검증
rg "@SPEC:AUTH-001" .moai/specs/
rg "@TEST:AUTH-001" tests/
rg "@CODE:AUTH-001" src/

# 결과: 모두 존재하면 통과
```

---

### 1.2 TRUST 검증 기능 (Phase 1)

#### Scenario 4: TRUST 5원칙 검증
```gherkin
Given 프로젝트에 테스트 코드가 있음
And 린터가 설정되어 있음
When Alfred가 TRUST 검증을 실행
Then TRUST 5원칙 모두 검증되어야 함
And 보고서가 생성되어야 함
```

**검증 방법**:
```bash
# TRUST 검증 실행
@agent-trust-checker "현재 프로젝트 TRUST 원칙 준수도 확인"

# 예상 보고서 형식
# ✅ T - Test First: 커버리지 87%
# ✅ R - Readable: 린터 통과
# ✅ U - Unified: 타입 검증 통과
# ✅ S - Secured: 보안 스캔 통과
# ✅ T - Trackable: TAG 체인 검증 통과
```

#### Scenario 5: 테스트 커버리지 부족 경고
```gherkin
Given 테스트 커버리지가 80%임
And 목표 커버리지가 85%임
When Alfred가 TRUST 검증을 실행
Then "⚠️ 테스트 커버리지 부족: 현재 80% (목표 85%)" 메시지가 반환되어야 함
```

**검증 방법**:
```bash
# 커버리지 측정
pytest --cov --cov-report=term

# TRUST 검증
@agent-trust-checker "TRUST 검증"

# 예상 결과
# ⚠️ T - Test First: 커버리지 80% (목표 85%)
# → 추가 테스트 케이스 작성 권장
```

---

### 1.3 SPEC 작성 기능 (Phase 2)

#### Scenario 6: EARS 구문 자동 적용
```gherkin
Given 사용자가 "/alfred:1-spec '새 기능'" 실행
When spec-builder가 moai-alfred-ears-authoring Skill을 JIT 로드
Then SPEC 문서에 EARS 5가지 구문이 모두 적용되어야 함
```

**검증 방법**:
```bash
# SPEC 작성
/alfred:1-spec "사용자 인증 기능"

# 생성된 SPEC 확인
cat .moai/specs/SPEC-AUTH-001/spec.md

# EARS 구문 확인
rg "Ubiquitous Requirements" .moai/specs/SPEC-AUTH-001/spec.md
rg "Event-driven Requirements" .moai/specs/SPEC-AUTH-001/spec.md
rg "State-driven Requirements" .moai/specs/SPEC-AUTH-001/spec.md
rg "Optional Features" .moai/specs/SPEC-AUTH-001/spec.md
rg "Constraints" .moai/specs/SPEC-AUTH-001/spec.md

# 결과: 5개 모두 존재하면 통과
```

#### Scenario 7: SPEC 메타데이터 생성
```gherkin
Given 사용자가 SPEC 작성을 요청
When spec-builder가 SPEC 문서를 생성
Then YAML Front Matter에 필수 필드 7개가 모두 포함되어야 함
```

**검증 방법**:
```bash
# 필수 필드 확인
rg "^id:" .moai/specs/SPEC-AUTH-001/spec.md
rg "^version:" .moai/specs/SPEC-AUTH-001/spec.md
rg "^status:" .moai/specs/SPEC-AUTH-001/spec.md
rg "^created:" .moai/specs/SPEC-AUTH-001/spec.md
rg "^updated:" .moai/specs/SPEC-AUTH-001/spec.md
rg "^author:" .moai/specs/SPEC-AUTH-001/spec.md
rg "^priority:" .moai/specs/SPEC-AUTH-001/spec.md

# 결과: 7개 모두 존재하면 통과
```

---

## 2. 성능 수락 기준

### 2.1 LOC 감소율

#### Criteria 1: 전체 LOC 30% 이상 감소
```gherkin
Given 마이그레이션 전 Agent 프롬프트 총 LOC가 측정됨
When 마이그레이션 후 Agent 프롬프트 총 LOC를 측정
Then 감소율이 30% 이상이어야 함
```

**측정 방법**:
```bash
# Before
echo "Before:"
before=$(cat .claude/agents/alfred/{tag-agent,trust-checker,spec-builder}.md | wc -l)
echo "Total: $before LOC"

# After
echo "After:"
after=$(cat .claude/agents/alfred/{tag-agent,trust-checker,spec-builder}.md | wc -l)
echo "Total: $after LOC"

# 감소율 계산
reduction=$(echo "scale=2; 100 - ($after * 100 / $before)" | bc)
echo "Reduction: $reduction%"

# 검증
if (( $(echo "$reduction >= 30" | bc -l) )); then
    echo "✅ PASS: LOC 감소율 $reduction% (목표 ≥30%)"
else
    echo "❌ FAIL: LOC 감소율 $reduction% (목표 ≥30%)"
fi
```

#### Criteria 2: Agent 프롬프트 300 LOC 이하
```gherkin
Given 마이그레이션이 완료됨
When 각 Agent 프롬프트의 LOC를 측정
Then 모든 Agent 프롬프트가 300 LOC 이하여야 함
```

**측정 방법**:
```bash
# 각 Agent LOC 측정
for agent in tag-agent trust-checker spec-builder; do
    loc=$(wc -l < .claude/agents/alfred/$agent.md)
    if [ $loc -le 300 ]; then
        echo "✅ $agent: $loc LOC (목표 ≤300)"
    else
        echo "❌ $agent: $loc LOC (목표 ≤300)"
    fi
done
```

---

### 2.2 JIT 로딩 시간

#### Criteria 3: JIT 로딩 시간 100ms 이하
```gherkin
Given Alfred가 Skill을 JIT 로드
When 로딩 시간을 측정
Then 로딩 시간이 100ms 이하여야 함
```

**측정 방법**:
```bash
# Skill 로딩 시간 측정
echo "moai-alfred-tag-scanning:"
time rg "@moai-alfred-tag-scanning" .claude/skills/moai-alfred-tag-scanning/skill.md > /dev/null

echo "moai-alfred-trust-validation:"
time rg "@moai-alfred-trust-validation" .claude/skills/moai-alfred-trust-validation/skill.md > /dev/null

echo "moai-alfred-ears-authoring:"
time rg "@moai-alfred-ears-authoring" .claude/skills/moai-alfred-ears-authoring/skill.md > /dev/null

# 목표: real < 0.1s (100ms)
```

---

### 2.3 Skill 파일 크기

#### Criteria 4: Skill 파일 크기 500KB 이하
```gherkin
Given Skill 파일이 생성됨
When 파일 크기를 측정
Then 모든 Skill 파일이 500KB 이하여야 함
```

**측정 방법**:
```bash
# Skill 파일 크기 측정
for skill in moai-alfred-tag-scanning moai-alfred-trust-validation moai-alfred-ears-authoring; do
    size=$(du -k .claude/skills/$skill/skill.md | cut -f1)
    if [ $size -le 500 ]; then
        echo "✅ $skill: ${size}KB (목표 ≤500KB)"
    else
        echo "❌ $skill: ${size}KB (목표 ≤500KB)"
    fi
done
```

---

## 3. 품질 수락 기준

### 3.1 기능 손실 없음

#### Criteria 5: 모든 기능이 정상 동작
```gherkin
Given 마이그레이션이 완료됨
When 기존 기능 테스트를 실행
Then 모든 테스트가 통과해야 함
```

**테스트 체크리스트**:
- [ ] TAG 스캔 기능 (Scenario 1)
- [ ] 고아 TAG 탐지 (Scenario 2)
- [ ] TAG 체인 검증 (Scenario 3)
- [ ] TRUST 검증 (Scenario 4)
- [ ] 테스트 커버리지 경고 (Scenario 5)
- [ ] EARS 구문 적용 (Scenario 6)
- [ ] SPEC 메타데이터 생성 (Scenario 7)

---

### 3.2 호환성 유지

#### Criteria 6: 기존 호출 방식 호환
```gherkin
Given 사용자가 기존 호출 방식을 사용
When Agent 또는 Skills를 호출
Then 기존 방식과 동일한 결과를 반환해야 함
```

**호환성 테스트**:
```bash
# Before (tag-agent 방식)
@agent-tag-agent "AUTH 도메인 TAG 목록 조회"
# 결과 저장: before.txt

# After (Skills 참조 방식)
# Alfred가 자동으로 moai-alfred-tag-scanning Skill 로드
# 결과 저장: after.txt

# 결과 비교
diff before.txt after.txt
# 예상: 차이 없음
```

---

### 3.3 문서 일관성

#### Criteria 7: CLAUDE.md 업데이트 완료
```gherkin
Given 마이그레이션이 완료됨
When CLAUDE.md를 확인
Then Agent 목록이 업데이트되어 있어야 함
And Skills 참조 가이드가 추가되어 있어야 함
```

**검증 방법**:
```bash
# Agent 목록 확인 (tag-agent, trust-checker 제거됨)
rg "tag-agent|trust-checker" .moai/CLAUDE.md
# 예상: 결과 없음 (제거됨)

# Skills 참조 가이드 확인
rg "Skills 참조 가이드|moai-alfred-tag-scanning" .moai/CLAUDE.md
# 예상: 결과 있음 (추가됨)
```

#### Criteria 8: development-guide.md 업데이트 완료
```gherkin
Given 마이그레이션이 완료됨
When development-guide.md를 확인
Then Skills 참조 가이드가 추가되어 있어야 함
```

**검증 방법**:
```bash
# Skills 참조 가이드 확인
rg "JIT 참조 방법|Skills 목록" .moai/memory/development-guide.md
# 예상: 결과 있음 (추가됨)
```

---

## 4. 보안 수락 기준

### 4.1 민감한 정보 누출 방지

#### Criteria 9: Skills에 민감한 정보 없음
```gherkin
Given Skills 파일이 생성됨
When Skills 파일 내용을 검사
Then 민감한 정보(API 키, 비밀번호, 토큰)가 포함되지 않아야 함
```

**검증 방법**:
```bash
# 민감한 정보 검색
rg -i "api_key|password|token|secret" .claude/skills/

# 예상: 결과 없음
```

---

## 5. 롤백 수락 기준

### 5.1 롤백 절차 검증

#### Criteria 10: 롤백 시 기능 복원
```gherkin
Given 마이그레이션 후 문제가 발생함
When 이전 Agent 프롬프트로 롤백
Then 모든 기능이 마이그레이션 전 상태로 복원되어야 함
```

**롤백 테스트**:
```bash
# 1. 백업 확인
ls -la .backup/agents/

# 2. 롤백 실행
cp .backup/agents/tag-agent.md .claude/agents/alfred/
cp .backup/agents/trust-checker.md .claude/agents/alfred/
cp .backup/agents/spec-builder.md .claude/agents/alfred/

# 3. 기능 테스트
@agent-tag-agent "AUTH 도메인 TAG 목록 조회"
@agent-trust-checker "현재 프로젝트 TRUST 원칙 준수도 확인"
/alfred:1-spec "테스트 기능"

# 4. 결과 확인
# 예상: 모든 기능이 마이그레이션 전과 동일하게 동작
```

---

## 6. 완료 조건 (Definition of Done)

### 6.1 Phase 1 완료 조건
- [ ] tag-agent 마이그레이션 완료 (LOC ≤200)
- [ ] trust-checker 마이그레이션 완료 (LOC ≤200)
- [ ] moai-alfred-tag-scanning Skill 생성 완료
- [ ] moai-alfred-trust-validation Skill 생성 완료
- [ ] Scenario 1~5 테스트 통과
- [ ] Criteria 1, 2 충족

### 6.2 Phase 2 완료 조건
- [ ] spec-builder EARS 부분 분리 완료 (LOC ≤500)
- [ ] spec-builder.md 업데이트 완료
- [ ] Scenario 6~7 테스트 통과
- [ ] Criteria 1, 2 충족

### 6.3 Phase 3 완료 조건
- [ ] 모든 기능 테스트 통과 (Scenario 1~7)
- [ ] 모든 성능 기준 충족 (Criteria 1~4)
- [ ] 모든 품질 기준 충족 (Criteria 5~8)
- [ ] 보안 기준 충족 (Criteria 9)
- [ ] 롤백 테스트 통과 (Criteria 10)
- [ ] CLAUDE.md 업데이트 완료
- [ ] development-guide.md 업데이트 완료

---

## 7. 검증 자동화 스크립트

### 7.1 전체 검증 스크립트
```bash
#!/bin/bash
# verify-spec-update-004.sh

echo "=========================================="
echo "SPEC-UPDATE-004 검증 시작"
echo "=========================================="

# 1. LOC 측정
echo "1. LOC 측정..."
before=1900  # 마이그레이션 전 총 LOC (예상)
after=$(cat .claude/agents/alfred/{tag-agent,trust-checker,spec-builder}.md 2>/dev/null | wc -l)
reduction=$(echo "scale=2; 100 - ($after * 100 / $before)" | bc)

if (( $(echo "$reduction >= 30" | bc -l) )); then
    echo "✅ LOC 감소율: $reduction% (목표 ≥30%)"
else
    echo "❌ LOC 감소율: $reduction% (목표 ≥30%)"
    exit 1
fi

# 2. Agent LOC 검증
echo "2. Agent LOC 검증..."
for agent in tag-agent trust-checker spec-builder; do
    if [ -f .claude/agents/alfred/$agent.md ]; then
        loc=$(wc -l < .claude/agents/alfred/$agent.md)
        if [ $loc -le 300 ]; then
            echo "✅ $agent: $loc LOC (목표 ≤300)"
        else
            echo "❌ $agent: $loc LOC (목표 ≤300)"
            exit 1
        fi
    fi
done

# 3. Skills 파일 존재 확인
echo "3. Skills 파일 존재 확인..."
for skill in moai-alfred-tag-scanning moai-alfred-trust-validation moai-alfred-ears-authoring; do
    if [ -f .claude/skills/$skill/skill.md ]; then
        echo "✅ $skill 존재"
    else
        echo "❌ $skill 누락"
        exit 1
    fi
done

# 4. Skills 파일 크기 검증
echo "4. Skills 파일 크기 검증..."
for skill in moai-alfred-tag-scanning moai-alfred-trust-validation moai-alfred-ears-authoring; do
    size=$(du -k .claude/skills/$skill/skill.md | cut -f1)
    if [ $size -le 500 ]; then
        echo "✅ $skill: ${size}KB (목표 ≤500KB)"
    else
        echo "❌ $skill: ${size}KB (목표 ≤500KB)"
        exit 1
    fi
done

# 5. 민감한 정보 검사
echo "5. 민감한 정보 검사..."
sensitive=$(rg -i "api_key|password|token|secret" .claude/skills/ 2>/dev/null | wc -l)
if [ $sensitive -eq 0 ]; then
    echo "✅ 민감한 정보 없음"
else
    echo "❌ 민감한 정보 발견 ($sensitive건)"
    exit 1
fi

# 6. 문서 업데이트 확인
echo "6. 문서 업데이트 확인..."
if rg -q "Skills 참조 가이드" .moai/CLAUDE.md; then
    echo "✅ CLAUDE.md 업데이트 완료"
else
    echo "❌ CLAUDE.md 업데이트 필요"
    exit 1
fi

if rg -q "JIT 참조 방법" .moai/memory/development-guide.md; then
    echo "✅ development-guide.md 업데이트 완료"
else
    echo "❌ development-guide.md 업데이트 필요"
    exit 1
fi

echo "=========================================="
echo "✅ 모든 검증 통과!"
echo "=========================================="
```

### 7.2 사용 방법
```bash
# 검증 스크립트 실행
bash .moai/specs/SPEC-UPDATE-004/verify-spec-update-004.sh

# 예상 출력
# ==========================================
# SPEC-UPDATE-004 검증 시작
# ==========================================
# 1. LOC 측정...
# ✅ LOC 감소율: 35.79% (목표 ≥30%)
# 2. Agent LOC 검증...
# ✅ tag-agent: 180 LOC (목표 ≤300)
# ✅ trust-checker: 190 LOC (목표 ≤300)
# ✅ spec-builder: 450 LOC (목표 ≤300)
# ...
# ==========================================
# ✅ 모든 검증 통과!
# ==========================================
```

---

## 8. 수락 승인

### 8.1 승인자
- **프로젝트 매니저**: @Goos
- **기술 리드**: @Alfred

### 8.2 승인 기준
- [ ] 모든 Phase 완료 조건 충족
- [ ] 검증 자동화 스크립트 통과
- [ ] 사용자 승인 (기능 테스트 결과 확인)

### 8.3 최종 승인
```yaml
approval:
  date: {YYYY-MM-DD}
  approver: @{GitHub ID}
  status: approved|rejected
  notes: {승인/거부 사유}
```

---

**작성자**: @Goos
**최초 작성일**: 2025-10-19
