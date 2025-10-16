# 검증 로직 재배치 계획

> **작성일**: 2025-10-16
> **작성자**: @Alfred
> **목적**: Hooks, Agents, Commands 역할 명확화를 위한 검증 로직 재배치

---

## 개요

Hooks 재설계 분석 결과, **복잡한 검증 로직을 Hooks에서 Agents/Commands로 이동**하여 역할을 명확히 분리해야 합니다.

### 핵심 원칙

```
Hooks      → 가벼운 가드레일 (<100ms)
Agents     → 복잡한 분석/검증 (수 초)
Commands   → 워크플로우 통합 (수 분)
```

---

## 재배치 대상 검증 로직

### 1. SPEC 메타데이터 검증

**현재 위치**: 제안만 있음 (미구현)

**이동 목표**: `/alfred:1-spec` 커맨드 내부

**이유**:
- 복잡한 검증 로직 (7개 필수 필드, YAML 파싱, 디렉토리 규칙)
- 사용자 피드백 필요 (오류 발생 시 수정 안내)
- SPEC 작성 워크플로우의 일부

**구현 계획**:
```python
# .claude/commands/alfred/1-spec.md 내부

Phase 2 실행 시:
1. SPEC 문서 생성
2. **SPEC 메타데이터 검증 호출**
   - @agent-spec-builder에게 검증 요청
   - 필수 필드 확인
   - HISTORY 섹션 확인
   - 디렉토리 명명 규칙 확인
3. 검증 실패 시 사용자에게 수정 권장
4. 검증 성공 시 Git 작업 진행
```

**구현 파일**:
- `.claude/agents/alfred/spec-builder.md` 확장
- 또는 새 함수: `src/moai_adk/core/spec/validator.py`

**SPEC 생성 시기**: 별도 SPEC 필요 (SPEC-HOOKS-002)

---

### 2. TRUST 원칙 검증

**현재 위치**: 제안만 있음 (미구현)

**이동 목표**: `/alfred:2-build` 완료 후 자동 호출

**이유**:
- 복잡한 분석 (테스트 커버리지, 코드 제약, TAG 체인)
- 보고서 생성 필요
- TDD 구현 완료 후 품질 게이트

**구현 계획**:
```python
# .claude/commands/alfred/2-build.md 내부

Phase 2 완료 후:
1. TDD 구현 완료 (RED → GREEN → REFACTOR)
2. **TRUST 검증 자동 호출**
   - @agent-trust-checker 호출
   - 테스트 커버리지 확인 (≥85%)
   - 코드 제약 확인 (파일 ≤300 LOC, 함수 ≤50 LOC)
   - TAG 체인 완전성 확인
3. 검증 결과 보고서 제공
4. 문제 발견 시 @agent-debug-helper 추천
```

**구현 파일**:
- `.claude/agents/alfred/trust-checker.md` (이미 존재)
- `src/moai_adk/core/quality/trust_checker.py` (신규)

**SPEC 생성 시기**: 별도 SPEC 필요 (SPEC-HOOKS-003)

---

### 3. EARS 구문 검증

**현재 위치**: 제안만 있음 (미구현)

**이동 목표**: `/alfred:1-spec` SPEC 작성 중 실시간 안내

**이유**:
- SPEC 작성 시점에만 필요
- 사용자 가이드 역할
- 복잡도 낮음 (키워드 검색)

**구현 계획**:
```python
# .claude/commands/alfred/1-spec.md 또는 spec-builder 내부

SPEC 작성 시:
1. 사용자에게 EARS 5가지 패턴 안내
   - Ubiquitous: "시스템은 [기능]을 제공해야 한다"
   - Event-driven: "WHEN [조건]이면, 시스템은 [동작]해야 한다"
   - State-driven: "WHILE [상태]일 때, 시스템은 [동작]해야 한다"
   - Optional: "WHERE [조건]이면, 시스템은 [동작]할 수 있다"
   - Constraints: "IF [조건]이면, 시스템은 [제약]해야 한다"

2. (선택적) 작성 완료 후 EARS 적용률 계산
   - development-guide.md 예시 참조
   - 간단한 키워드 매칭
```

**구현 파일**:
- `.claude/agents/alfred/spec-builder.md` 내부 통합
- 또는 `src/moai_adk/core/spec/ears_helper.py` (경량)

**SPEC 생성 시기**: SPEC-HOOKS-002에 포함 가능 (우선순위 낮음)

---

### 4. Git 커밋 메시지 Locale 지원

**현재 위치**: 제안만 있음 (미구현)

**이동 목표**: `@agent-git-manager` 내부

**이유**:
- Git 커밋 생성 시점에만 필요
- `.moai/config.json` 읽기 필요
- 단순 로직 (템플릿 치환)

**구현 계획**:
```python
# .claude/agents/alfred/git-manager.md 내부

커밋 생성 시:
1. .moai/config.json에서 locale 읽기 (기본값: ko)
2. TDD 단계별 이모지 + locale별 템플릿 선택
   - ko: "🔴 RED: [테스트 설명]"
   - en: "🔴 RED: [Test description]"
   - ja: "🔴 RED: [テスト説明]"
   - zh: "🔴 RED: [测试说明]"
3. @TAG 자동 삽입: "@TAG:[SPEC-ID]-[단계]"
```

**구현 파일**:
- `.claude/agents/alfred/git-manager.md` 확장
- 또는 `src/moai_adk/core/git/commit_messages.py` (신규)

**SPEC 생성 시기**: 별도 SPEC 필요 (SPEC-HOOKS-004)

---

## 구현 우선순위

### Phase 1: 핵심 검증 (SPEC-HOOKS-002, SPEC-HOOKS-003)

1. **SPEC 메타데이터 검증** (High)
   - `/alfred:1-spec`에 통합
   - 필수 필드 7개 검증
   - 디렉토리 명명 규칙 검증

2. **TRUST 원칙 검증** (High)
   - `/alfred:2-build` 완료 후 자동 호출
   - @agent-trust-checker 구현
   - 테스트 커버리지 + 코드 제약 + TAG 체인

**예상 작업량**: 2일 (SPEC 작성 + 구현 + 테스트)

---

### Phase 2: 편의 기능 (SPEC-HOOKS-004)

3. **Git 커밋 메시지 Locale** (Medium)
   - @agent-git-manager에 통합
   - 4개 언어 지원 (ko/en/ja/zh)
   - TDD 단계별 이모지 + 템플릿

**예상 작업량**: 1일

---

### Phase 3: 선택적 개선

4. **EARS 구문 검증** (Low)
   - spec-builder에 통합
   - 5가지 패턴 안내
   - EARS 적용률 계산

**예상 작업량**: 0.5일

---

## 구현 가이드라인

### 1. SPEC-First 원칙 준수

모든 검증 로직 이동은 **별도 SPEC 작성 후 구현**:
```bash
/alfred:1-spec "SPEC 메타데이터 검증 자동화"
→ SPEC-HOOKS-002 생성

/alfred:1-spec "TRUST 원칙 자동 검증"
→ SPEC-HOOKS-003 생성

/alfred:1-spec "Git 커밋 메시지 Locale 지원"
→ SPEC-HOOKS-004 생성
```

### 2. TDD 구현

각 검증 로직은 **TDD 사이클** 준수:
```bash
/alfred:2-build SPEC-HOOKS-002
→ RED: 테스트 작성
→ GREEN: 구현
→ REFACTOR: 코드 품질 개선
```

### 3. 문서 동기화

구현 완료 후 **문서 업데이트**:
```bash
/alfred:3-sync
→ CLAUDE.md, development-guide.md 동기화
→ TAG 체인 검증
→ PR Ready 전환
```

---

## 마이그레이션 체크리스트

### Phase 1: SPEC 메타데이터 검증

- [ ] SPEC-HOOKS-002 작성
- [ ] `src/moai_adk/core/spec/validator.py` 구현
- [ ] `/alfred:1-spec` 커맨드에 통합
- [ ] 단위 테스트 작성 (≥85% 커버리지)
- [ ] E2E 테스트 (SPEC 생성 → 검증)
- [ ] 문서 업데이트 (development-guide.md)

### Phase 2: TRUST 원칙 검증

- [ ] SPEC-HOOKS-003 작성
- [ ] `src/moai_adk/core/quality/trust_checker.py` 구현
- [ ] @agent-trust-checker 확장
- [ ] `/alfred:2-build` 완료 후 자동 호출
- [ ] 단위 테스트 작성
- [ ] E2E 테스트 (TDD → TRUST 검증)
- [ ] 문서 업데이트

### Phase 3: Git 커밋 Locale

- [ ] SPEC-HOOKS-004 작성
- [ ] `src/moai_adk/core/git/commit_messages.py` 구현
- [ ] @agent-git-manager 확장
- [ ] 4개 언어 템플릿 작성 (ko/en/ja/zh)
- [ ] 단위 테스트 작성
- [ ] E2E 테스트 (커밋 생성 → locale 확인)
- [ ] 문서 업데이트 (CLAUDE.md)

---

## 예상 효과

### Before (Hooks에 검증 로직)

- ❌ Hooks가 무거워짐 (>100ms)
- ❌ 역할 혼재 (가드레일 vs 분석)
- ❌ 사용자 피드백 어려움
- ❌ 테스트/유지보수 복잡

### After (Agents/Commands로 이동)

- ✅ Hooks는 가벼움 (<100ms 유지)
- ✅ 역할 명확 (Hooks = 가드레일, Agents = 분석)
- ✅ 사용자 피드백 자연스러움
- ✅ 테스트/유지보수 용이

### 품질 지표

| 항목 | Before | After | 개선율 |
|------|--------|-------|--------|
| 역할 명확성 | 70% | **95%** | +25% |
| Hooks 성능 | ~100ms | **<50ms** | +50% |
| 사용자 경험 | 중간 | **높음** | +30% |
| 유지보수성 | 중간 | **높음** | +30% |

---

## 다음 단계

1. **Phase 1 SPEC 작성**
   ```bash
   /alfred:1-spec "SPEC 메타데이터 검증 자동화"
   /alfred:1-spec "TRUST 원칙 자동 검증"
   ```

2. **Phase 1 구현**
   ```bash
   /alfred:2-build SPEC-HOOKS-002
   /alfred:2-build SPEC-HOOKS-003
   ```

3. **Phase 1 문서화**
   ```bash
   /alfred:3-sync
   ```

4. **Phase 2-3 진행** (Phase 1 완료 후)

---

**최종 업데이트**: 2025-10-16
**상태**: 계획 수립 완료, 구현 대기
