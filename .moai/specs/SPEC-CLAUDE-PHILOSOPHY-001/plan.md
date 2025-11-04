# @SPEC:CLAUDE-PHILOSOPHY-001 실행 계획

## 프로젝트 개요

- **SPEC ID**: CLAUDE-PHILOSOPHY-001
- **제목**: CLAUDE.md 철학 재정렬 및 Skill 분리
- **우선순위**: High
- **복잡도**: Medium
- **예상 산출물**: CLAUDE.md 재구조화, 2개 Skill 생성, 패키지 템플릿 동기화

---

## Phase 1: CLAUDE.md 구조 재설계 (Tier 1-4 계층화)

### 목표
CLAUDE.md를 Tier 1-4 계층으로 재구조화하여 핵심 규칙 가독성 향상

### 작업 내용

#### 1.1 현재 구조 분석
- **작업**: 기존 CLAUDE.md 섹션 매핑
- **도구**: Read, Grep
- **산출물**: 섹션 목록 (섹션명, 줄 수, 현재 위치)
- **검증**: 전체 섹션 수 확인 (예상: 15-20개)

#### 1.2 Tier 분류
- **작업**: 각 섹션을 Tier 1-4로 분류
- **기준**:
  - Tier 1: 세션 시작 시 필수 (워크플로우, 언어 경계, Permissions, TRUST)
  - Tier 2: 실행 시 참조 (Sub-agents, 커맨드, Git)
  - Tier 3: 고급 기능 (페르소나, 자동 수정, 보고)
  - Tier 4: 참조 자료 (프로젝트 정보, 설정)
- **산출물**: 섹션별 Tier 분류 테이블
- **검증**: Tier 1이 500줄 이내인지 확인

#### 1.3 섹션 재배치
- **작업**: Tier 순서로 섹션 재배치
- **도구**: Edit
- **변경 사항**:
  - Tier 1 섹션을 문서 상단으로 이동
  - Tier 3-4 섹션을 Skill 링크로 대체
- **산출물**: 재배치된 CLAUDE.md (초안)
- **검증**: 기존 앵커 링크가 깨지지 않는지 확인

#### 1.4 Tier 1 최적화
- **작업**: Tier 1 섹션을 500줄 이내로 압축
- **방법**:
  - 중복 내용 제거
  - 상세 설명은 Skill로 이동
  - 핵심 규칙만 유지
- **산출물**: 최적화된 Tier 1 섹션
- **검증**: 줄 수 카운트 (target: 400-500줄)

### 산출물
- [ ] 섹션 분류 테이블 (Tier 1-4)
- [ ] 재배치된 CLAUDE.md 초안
- [ ] Tier 1 최적화 완료 (500줄 이내)

### 위험 요소
- **위험**: 섹션 재배치로 기존 링크 깨짐
  - **완화**: 섹션 제목 유지, 앵커 ID 보존
  - **검증**: `grep -n '\[.*\](#.*)'` 로 모든 링크 검사

### 검증 방법
```bash
# Tier 1 줄 수 확인
head -500 CLAUDE.md | grep -c '^#'

# 링크 유효성 검증
grep -o '\[.*\](#.*\)' CLAUDE.md | while read link; do
  anchor=$(echo "$link" | sed -E 's/.*\(#(.*)\)/\1/')
  grep -q "^#.*$anchor" CLAUDE.md || echo "Broken: $link"
done
```

---

## Phase 2: 2개 Skill 분리 작업

### 목표
세션 분석과 고급 설정을 독립 Skill로 분리하여 CLAUDE.md 경량화

### 작업 내용

#### 2.1 moai-alfred-session-analytics Skill 생성

##### 2.1.1 디렉토리 구조 생성
- **작업**: `.claude/skills/moai-alfred-session-analytics/` 생성
- **도구**: Bash (mkdir)
- **파일 구조**:
  ```
  .claude/skills/moai-alfred-session-analytics/
  ├── SKILL.md           # Skill 메타데이터 + 개요
  ├── reference.md       # 세션 메트릭 정의
  └── examples.md        # 세션 분석 예시
  ```

##### 2.1.2 SKILL.md 작성
- **내용**:
  - YAML frontmatter (name, version, status, description, keywords, allowed-tools)
  - What It Does: 세션 분석, 로깅, 메트릭 수집
  - When to Use: "세션 분석", "로그 확인", "메트릭" 키워드 감지
  - Inputs/Outputs: 세션 데이터 → 분석 리포트
- **산출물**: SKILL.md (약 150줄)

##### 2.1.3 reference.md 작성
- **내용**:
  - 세션 메트릭 정의 (시작/종료 시간, 명령 수, 에러율)
  - 로깅 정책 (로그 레벨, 보존 기간)
  - 메트릭 계산 공식
- **산출물**: reference.md (약 200줄)

##### 2.1.4 examples.md 작성
- **내용**:
  - 세션 분석 예시 (성공/실패 시나리오)
  - 로그 리뷰 예시
  - 메트릭 리포트 예시
- **산출물**: examples.md (약 100줄)

#### 2.2 moai-alfred-config-advanced Skill 생성

##### 2.2.1 디렉토리 구조 생성
- **작업**: `.claude/skills/moai-alfred-config-advanced/` 생성
- **파일 구조**:
  ```
  .claude/skills/moai-alfred-config-advanced/
  ├── SKILL.md           # Skill 메타데이터 + 개요
  ├── reference.md       # 고급 설정 필드 설명
  └── examples.md        # 고급 설정 예시
  ```

##### 2.2.2 SKILL.md 작성
- **내용**:
  - YAML frontmatter (name, version, status, description, keywords, allowed-tools)
  - What It Does: Hook 타임아웃, 권한 세분화, 메타데이터 최적화
  - When to Use: "고급 설정", "Hook 조정", "권한 정책" 키워드 감지
  - Inputs/Outputs: 설정 요청 → 커스터마이징 가이드
- **산출물**: SKILL.md (약 150줄)

##### 2.2.3 reference.md 작성
- **내용**:
  - Hook 타임아웃 설정 (SessionStart, PreToolUse, etc.)
  - Permissions 세분화 (deny/ask/allow 패턴)
  - .moai/config.json 고급 필드 (hooks, tags, constitution)
- **산출물**: reference.md (약 250줄)

##### 2.2.4 examples.md 작성
- **내용**:
  - Hook 타임아웃 조정 예시
  - 권한 정책 커스터마이징 예시
  - 고급 메타데이터 설정 예시
- **산출물**: examples.md (약 120줄)

#### 2.3 CLAUDE.md에 Skill 링크 추가
- **작업**: Tier 3-4 섹션에 Skill 링크 추가
- **도구**: Edit
- **변경 사항**:
  ```markdown
  ## 📚 참조 (Tier 4)
  ### 세션 분석
  자세한 내용은 Skill("moai-alfred-session-analytics")를 참조하세요.

  ### 고급 설정
  자세한 내용은 Skill("moai-alfred-config-advanced")를 참조하세요.
  ```
- **산출물**: 업데이트된 CLAUDE.md

### 산출물
- [ ] moai-alfred-session-analytics Skill (SKILL.md, reference.md, examples.md)
- [ ] moai-alfred-config-advanced Skill (SKILL.md, reference.md, examples.md)
- [ ] CLAUDE.md Skill 링크 추가

### 위험 요소
- **위험**: Skill 분리 후 기존 기능성 손실
  - **완화**: CLAUDE.md에 최소 정보 유지, Skill은 상세 정보만
  - **검증**: Skill("name") 호출 시 정상 로드 확인

### 검증 방법
```bash
# Skill 파일 존재 확인
ls .claude/skills/moai-alfred-session-analytics/
ls .claude/skills/moai-alfred-config-advanced/

# Skill 링크 확인
grep 'Skill("moai-alfred-session-analytics")' CLAUDE.md
grep 'Skill("moai-alfred-config-advanced")' CLAUDE.md

# YAML frontmatter 검증
head -10 .claude/skills/moai-alfred-session-analytics/SKILL.md
head -10 .claude/skills/moai-alfred-config-advanced/SKILL.md
```

---

## Phase 3: 부정적 제약 → 긍정적 가이드라인 변환

### 목표
최소 20개 부정적 표현을 긍정적 가이드라인으로 변환하여 사용자 친화성 향상

### 작업 내용

#### 3.1 부정적 표현 스캔
- **작업**: CLAUDE.md에서 부정적 표현 검색
- **도구**: Grep
- **검색 패턴**:
  ```bash
  grep -n 'DO NOT\|NEVER\|DON'\''T\|AVOID\|MUST NOT' CLAUDE.md
  ```
- **산출물**: 부정적 표현 목록 (위치, 내용)
- **목표**: 최소 30개 발견 (변환 대상 20개 선정)

#### 3.2 변환 우선순위 선정
- **작업**: 변환 대상 20개 선정
- **기준**:
  - 우선순위 1: 사용자 행동 가이드 (DO NOT create X → CREATE in Y instead)
  - 우선순위 2: 도구 사용 가이드 (NEVER use X → USE Y tool instead)
  - 우선순위 3: Git 워크플로우 (DO NOT commit X → COMMIT when Y)
  - **제외**: 핵심 금지사항 (git push --force는 deny → 유지)
- **산출물**: 변환 대상 20개 목록

#### 3.3 긍정적 가이드라인 작성
- **작업**: 각 부정적 표현을 긍정적으로 재작성
- **도구**: Edit
- **변환 패턴**:
  - 패턴 1: ❌ "DO NOT X" → ✅ "INSTEAD: Y"
  - 패턴 2: ❌ "NEVER X" → ✅ "PREFER: Y"
  - 패턴 3: ❌ "AVOID X" → ✅ "USE: Y"
- **산출물**: 변환된 CLAUDE.md (20개 이상 변환 완료)

#### 3.4 핵심 제약 유지
- **작업**: 필수 금지사항은 부정적 표현 유지
- **유지 대상**:
  - "NEVER run git push --force to main/master"
  - "NEVER amend other developers' commits"
  - "NEVER skip hooks (--no-verify)"
  - "NEVER hardcode secrets"
- **검증**: 핵심 제약이 명확한지 확인

### 산출물
- [ ] 부정적 표현 목록 (30개 이상)
- [ ] 변환 대상 20개 선정
- [ ] 긍정적 가이드라인 적용 완료
- [ ] 핵심 제약 유지 확인

### 위험 요소
- **위험**: 부정적 제약 제거로 금지사항 불명확
  - **완화**: 핵심 제약 유지 (git push --force는 deny)
  - **검증**: 20개 변환 후 필수 금지사항 명확성 확인

### 검증 방법
```bash
# 변환 전후 비교
grep -c 'DO NOT\|NEVER\|DON'\''T' CLAUDE.md  # Before
# 변환 후
grep -c 'DO NOT\|NEVER\|DON'\''T' CLAUDE.md  # After (최소 20개 감소)

# 긍정적 표현 증가 확인
grep -c 'INSTEAD\|PREFER\|USE:' CLAUDE.md  # 최소 20개 증가

# 핵심 제약 유지 확인
grep -n 'NEVER.*git push --force' CLAUDE.md
grep -n 'NEVER.*amend other' CLAUDE.md
```

---

## Phase 4: 패키지 템플릿 동기화

### 목표
로컬 CLAUDE.md 변경사항을 패키지 템플릿에 동기화 (언어만 영어로)

### 작업 내용

#### 4.1 로컬 변경사항 확정
- **작업**: Phase 1-3 변경사항 Git 커밋
- **도구**: Bash (git)
- **커밋 메시지**:
  ```
  refactor(docs): Phase 6 CLAUDE.md 재구조화 (Tier 1-4, Skill 분리, 긍정적 가이드라인)

  - Tier 1-4 계층 구조 도입 (핵심 규칙 500줄 이내)
  - 2개 Skill 분리 (session-analytics, config-advanced)
  - 20개 이상 부정적 제약 → 긍정적 가이드라인 변환
  - 패키지 템플릿 동기화 준비

  🤖 Generated with Claude Code

  Co-Authored-By: 🎩 Alfred@MoAI
  ```
- **산출물**: Git 커밋 완료

#### 4.2 패키지 템플릿 구조 복사
- **작업**: 로컬 CLAUDE.md 구조를 패키지로 복사
- **도구**: Read, Edit
- **복사 항목**:
  - 섹션 순서 (Tier 1-4)
  - Skill 링크 (Skill("name"))
  - YAML frontmatter (필드명은 동일)
- **산출물**: `src/moai_adk/templates/CLAUDE.md` 업데이트 (구조만)

#### 4.3 언어 변환 (한국어 → 영어)
- **작업**: 내용을 영어로 번역
- **도구**: Edit
- **변환 대상**:
  - 섹션 제목
  - 본문 설명
  - 예시 코드 주석
- **유지 항목**:
  - Skill("name") (영어 유지)
  - YAML frontmatter 필드명 (영어 유지)
  - 기술 용어 (Git, SPEC, TDD 등)
- **산출물**: `src/moai_adk/templates/CLAUDE.md` (영어 버전)

#### 4.4 동기화 검증
- **작업**: 로컬과 패키지 구조 일치 확인
- **도구**: Bash (diff)
- **검증 항목**:
  ```bash
  # 섹션 수 비교
  grep -c '^##' CLAUDE.md
  grep -c '^##' src/moai_adk/templates/CLAUDE.md

  # Skill 링크 비교
  grep 'Skill("' CLAUDE.md | wc -l
  grep 'Skill("' src/moai_adk/templates/CLAUDE.md | wc -l

  # Tier 구조 비교
  grep '^## .*Tier' CLAUDE.md
  grep '^## .*Tier' src/moai_adk/templates/CLAUDE.md
  ```
- **산출물**: 검증 리포트 (일치 여부)

#### 4.5 패키지 커밋
- **작업**: 패키지 템플릿 변경사항 커밋
- **도구**: Bash (git)
- **커밋 메시지**:
  ```
  refactor(templates): Sync CLAUDE.md template with Phase 6 restructure

  - Apply Tier 1-4 structure to package template
  - Add Skill links (session-analytics, config-advanced)
  - Convert 20+ negative constraints to positive guidelines
  - Language: English (structure identical to local Korean version)

  🤖 Generated with Claude Code

  Co-Authored-By: 🎩 Alfred@MoAI
  ```
- **산출물**: Git 커밋 완료

### 산출물
- [ ] 로컬 CLAUDE.md Git 커밋
- [ ] 패키지 템플릿 구조 복사
- [ ] 언어 변환 (한국어 → 영어)
- [ ] 동기화 검증 완료
- [ ] 패키지 템플릿 Git 커밋

### 위험 요소
- **위험**: 패키지 동기화 수동 누락
  - **완화**: 단계별 체크리스트 사용
  - **검증**: diff 비교로 구조 일치 확인

### 검증 방법
```bash
# 구조 일치 검증 스크립트
#!/bin/bash
LOCAL="CLAUDE.md"
PKG="src/moai_adk/templates/CLAUDE.md"

echo "Comparing structure..."
echo "Sections: Local=$(grep -c '^##' $LOCAL), Package=$(grep -c '^##' $PKG)"
echo "Skills: Local=$(grep -c 'Skill("' $LOCAL), Package=$(grep -c 'Skill("' $PKG)"
echo "Tiers: Local=$(grep -c 'Tier [1-4]' $LOCAL), Package=$(grep -c 'Tier [1-4]' $PKG)"

# 차이점 확인 (구조만, 언어 제외)
diff <(grep '^##' $LOCAL) <(grep '^##' $PKG) || echo "Structure differs!"
```

---

## 전체 타임라인 (우선순위 기반)

### Primary Goals (필수 완료)
1. **Phase 1**: CLAUDE.md 구조 재설계
2. **Phase 2**: 2개 Skill 분리
3. **Phase 4**: 패키지 템플릿 동기화

### Secondary Goals (가능 시 완료)
4. **Phase 3**: 부정적 제약 → 긍정적 가이드라인 변환 (최소 20개)

### Final Goals (선택 사항)
5. **추가 변환**: 20개 이상 긍정적 가이드라인 변환
6. **문서화**: Phase 6 요약 리포트 작성

---

## 의존성 관계

```
Phase 1 (구조 재설계)
    ↓ (Tier 분류 완료 후)
Phase 2 (Skill 분리)
    ↓ (Skill 생성 완료 후)
Phase 3 (긍정적 가이드라인 변환)
    ↓ (모든 변경 완료 후)
Phase 4 (패키지 동기화)
```

**병렬 가능**:
- Phase 2.1 (session-analytics Skill) ∥ Phase 2.2 (config-advanced Skill)

**순차 필수**:
- Phase 1 → Phase 2 → Phase 4
- Phase 3은 Phase 1-2와 병렬 가능하나 Phase 4 전에 완료 필요

---

## 최종 산출물 체크리스트

### 파일 생성
- [ ] `/Users/goos/MoAI/MoAI-ADK/CLAUDE.md` (재구조화, 한국어)
- [ ] `.claude/skills/moai-alfred-session-analytics/SKILL.md`
- [ ] `.claude/skills/moai-alfred-session-analytics/reference.md`
- [ ] `.claude/skills/moai-alfred-session-analytics/examples.md`
- [ ] `.claude/skills/moai-alfred-config-advanced/SKILL.md`
- [ ] `.claude/skills/moai-alfred-config-advanced/reference.md`
- [ ] `.claude/skills/moai-alfred-config-advanced/examples.md`
- [ ] `src/moai_adk/templates/CLAUDE.md` (동기화, 영어)

### 검증 완료
- [ ] Tier 1이 500줄 이내
- [ ] 2개 Skill 정상 로드
- [ ] 최소 20개 부정적 → 긍정적 변환
- [ ] 패키지 템플릿 구조 일치
- [ ] 모든 Skill 링크 유효성 확인

### Git 커밋
- [ ] 로컬 CLAUDE.md 변경 커밋
- [ ] 2개 Skill 생성 커밋
- [ ] 패키지 템플릿 동기화 커밋

---

## 성공 기준

### 정량적 기준
- Tier 1 줄 수: 400-500줄
- Skill 개수: 2개
- 부정적 → 긍정적 변환: 최소 20개
- 패키지 동기화 일치율: 100% (구조 기준)

### 정성적 기준
- CLAUDE.md 가독성 향상 (Tier 1 집중)
- Skill 분리로 경량화 달성
- 긍정적 가이드라인으로 사용자 친화성 향상
- 패키지 템플릿 동기화로 일관성 유지

---

## 커뮤니케이션 계획

### 단계별 확인
- Phase 1 완료 후: Tier 구조 리뷰 (사용자 확인)
- Phase 2 완료 후: Skill 로드 테스트 (사용자 확인)
- Phase 3 완료 후: 긍정적 가이드라인 리뷰 (사용자 확인)
- Phase 4 완료 후: 패키지 동기화 검증 (자동)

### 최종 리포트
- **포함 항목**:
  - Phase별 완료 상태
  - 변환된 부정적 표현 목록 (20개)
  - 생성된 Skill 목록 (2개)
  - 패키지 동기화 검증 결과

---

_이 실행 계획은 `/alfred:2-run SPEC-CLAUDE-PHILOSOPHY-001`으로 실행됩니다._
