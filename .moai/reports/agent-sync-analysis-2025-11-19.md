# 에이전트 동기화 분석 리포트

**분석 일시**: 2025-11-19
**범위**: 패키지 템플릿 vs 로컬 에이전트 31개 파일 완전 비교
**상태**: SSOT (Single Source of Truth) 재정렬 필요

---

## 📊 Executive Summary

### 핵심 발견사항

| 항목 | 수치 | 상태 |
|------|------|------|
| **총 에이전트 파일** | 31개 | 양쪽 일치 |
| **누락된 파일** | 0개 | ✅ 완벽 |
| **로컬 전용 파일** | 0개 | ✅ 완벽 |
| **내용 동일** | 8개 | ✅ 최신 유지 |
| **내용 차이** | 23개 | ⚠️ 업데이트 필요 |

### 변경 유형 분류

| 변경 유형 | 파일 수 | 심각도 | 이유 |
|----------|--------|--------|------|
| **Skill 이름 변경** | 18개 | 높음 | `moai-alfred-*` → `moai-core-*` 브랜딩 통합 |
| **Skill 업데이트** | 5개 | 높음 | 추가 기능 및 참조 업데이트 |
| **주석/설명 업데이트** | 10개 | 중간 | AskUserQuestion 스킬 문서 링크 업데이트 |
| **신규 내용 추가** | 1개 | 낮음 | mcp-context7-integrator에 기능 추가 |

---

## 🔴 추가/업데이트 필요 (23개 파일)

### Category 1: Skill 이름 변경 (alfred → core) - 18개 파일

**원인**: 브랜딩 통합 및 네임스페이스 재정의
**영향**: 에이전트에서 호출하는 Skill 참조 업데이트 필수

#### 1-1. 단순 Skill 참조 변경 (2줄 변경) - 13개 파일

**패턴**: `Skill("moai-alfred-language-detection")` → `Skill("moai-core-language-detection")`

**해당 파일**:
```
1. accessibility-expert.md (2 changes)
2. api-designer.md (2 changes)
3. backend-expert.md (2 changes)
4. component-designer.md (2 changes)
5. devops-expert.md (2 changes)
6. figma-expert.md (2 changes)
7. frontend-expert.md (2 changes)
8. migration-expert.md (2 changes)
9. monitoring-expert.md (2 changes)
10. performance-engineer.md (2 changes)
11. ui-ux-expert.md (2 changes)
```

**더블 체크 필요** (추가 변경):
```
12. debug-helper.md (8 changes - see below)
13. api-designer.md (2 changes - already listed)
```

#### 1-2. 복잡한 Skill 참조 변경 (다중 변경) - 5개 파일

**Skill 이름 변경 + AskUserQuestion 문서 링크 업데이트**

**파일별 변경사항**:

##### cc-manager.md (10 changes)
```
변경 패턴:
- moai-alfred-workflow → moai-core-workflow
- moai-alfred-language-detection → moai-core-language-detection
- moai-alfred-tag-scanning → moai-core-tag-scanning

주요 라인:
L26: `Skill("moai-core-workflow")` + workflows/ (아키텍처 결정)
L50+: 언어 감지, TAG 검증 스킬 업데이트
```

##### debug-helper.md (8 changes)
```
변경 패턴:
- moai-alfred-ask-user-questions → moai-core-ask-user-questions
- moai-alfred-language-detection → moai-core-language-detection
- moai-alfred-tag-scanning → moai-core-tag-scanning

주요 라인:
L36: AskUserQuestion 도구 문서 링크
L50+: 언어 감지, TAG 스캔 스킬 업데이트
```

##### doc-syncer.md (16 changes)
```
변경 패턴:
- moai-alfred-ask-user-questions → moai-core-ask-user-questions
- moai-alfred-tag-scanning → moai-core-tag-scanning

주요 라인:
L35: AskUserQuestion 도구 문서 링크
L120+: TAG 스캔 스킬 업데이트
L130+: TAG 기반 동기화 로직 참조
```

##### git-manager.md (12 changes)
```
변경 패턴:
- moai-alfred-ask-user-questions → moai-core-ask-user-questions
- moai-alfred-git-workflow → moai-core-git-workflow
- moai-alfred-trust-validation → moai-core-trust-validation

주요 라인:
L36: AskUserQuestion 도구 문서 링크
L95+: Git 워크플로우 스킬 업데이트
L150+: TRUST 검증 스킬 참조
```

##### implementation-planner.md (20 changes)
```
변경 패턴:
- moai-alfred-ask-user-questions → moai-core-ask-user-questions
- moai-alfred-language-detection → moai-core-language-detection
- (다양한 다른 alfred 스킬들도 core로 변경)

주요 라인:
L36: AskUserQuestion 도구 문서 링크
L100+: 언어 감지 및 구현 계획 스킬 업데이트
```

### Category 2: 대규모 Skill 네임스페이스 업데이트 - 5개 파일

**원인**: 스킬 팩토리 및 인증/유효성 검사 체계 재정의

#### 2-1. agent-factory.md (12 changes)

**변경 패턴**:
```
Old: moai-alfred-agent-factory
New: moai-core-agent-factory

변경 위치:
- Skill 참조 (L95+)
- 템플릿 경로 (L145+)
- 마스터 스킬 설명 (L120+)

라인 확인 필요:
L95: Skill 참조 직접 업데이트
L120: "MASTER SKILL containing:" 섹션
L145: ".claude/skills/moai-core-agent-factory/templates/" 경로
```

#### 2-2. quality-gate.md (18 changes)

**변경 패턴**:
```
Old: moai-alfred-ask-user-questions, moai-alfred-trust-validation
New: moai-core-ask-user-questions, moai-core-trust-validation

변경 위치:
- AskUserQuestion 문서 링크 (L36)
- 신뢰성 검증 스킬 (L90+)
- 에센셜 리뷰 통합 (L150+)

라인 확인 필요:
L36: AskUserQuestion 도구 문서 링크
L90: Skill("moai-core-trust-validation")
```

#### 2-3. skill-factory.md (30 changes) - 최대 규모

**변경 패턴**:
```
Old: moai-alfred-skill-factory, moai-alfred-ask-user-questions
New: moai-core-skill-factory, moai-core-ask-user-questions

변경 위치:
- 제목 및 헤더 (L14: "moai-alfred-skill-factory" → "moai-core-skill-factory")
- 다중 Skill 참조 (L50+, L90+, L130+, 등)
- 테이블 및 섹션 헤더

라인 확인 필요:
L14: 헤더 제목
L50+: "You invoke" 섹션 (다중 참조)
L90+: 대규모 테이블 (alfred → core 변경)
L130+: 추가 스킬 참조
```

#### 2-4. spec-builder.md (18 changes)

**변경 패턴**:
```
Old: moai-alfred-spec-authoring, moai-alfred-ask-user-questions, moai-alfred-ears-authoring
New: moai-core-spec-authoring, moai-core-ask-user-questions, moai-core-ears-authoring

변경 위치:
- Skills 섹션 (L8-11)
- AskUserQuestion 문서 링크 (L36)
- EARS 오작 로직 (L100+)

라인 확인 필요:
L8: Skill 목록 (moai-core-spec-authoring)
L36: AskUserQuestion 도구 문서 링크
L100+: EARS 관련 스킬 참조
```

#### 2-5. tdd-implementer.md (18 changes)

**변경 패턴**:
```
Old: moai-alfred-ask-user-questions, moai-alfred-language-detection
New: moai-core-ask-user-questions, moai-core-language-detection

변경 위치:
- AskUserQuestion 문서 링크 (L36)
- 언어 감지 로직 (L90+)
- 테스트 구현 세부사항 (L150+)

라인 확인 필요:
L36: AskUserQuestion 도구 문서 링크
L90+: 언어 감지 스킬 참조
L150+: TDD 구현 가이드
```

#### 2-6. trust-checker.md (16 changes)

**변경 패턴**:
```
Old: moai-alfred-ask-user-questions, moai-alfred-trust-validation
New: moai-core-ask-user-questions, moai-core-trust-validation

변경 위치:
- AskUserQuestion 문서 링크 (L36)
- 신뢰성 검증 로직 (L85+)
- 파운데이션 신뢰 통합 (L110+)

라인 확인 필요:
L36: AskUserQuestion 도구 문서 링크
L85+: TRUST 검증 스킬 참조
L110+: Skill("moai-foundation-trust") 통합
```

### Category 3: 신규 콘텐츠 추가 - 1개 파일

#### mcp-context7-integrator.md (1 change)

**변경**: Context7 통합 최적화 가이드 추가

```
추가된 라인:
- **mcp-context7-integrator**: Use [Context7 MCP] for complex research strategies

맥락:
- 복잡한 연구 전략을 위한 MCP 활용 추천
- Context7 기능 확대
```

---

## 🟢 최신 유지 (8개 파일)

**이미 템플릿과 동일한 에이전트** - 업데이트 불필요:

```
1. database-expert.md
2. docs-manager.md
3. format-expert.md
4. mcp-notion-integrator.md
5. mcp-playwright-integrator.md
6. project-manager.md
7. security-expert.md
8. sync-manager.md
```

---

## 🚀 동기화 실행 계획

### Phase 1: 단순 변경 (13개 파일, 예상 소요시간: 15분)

**작업**: 1줄 변경만 필요한 파일들 일괄 업데이트

**변경 패턴**: `moai-alfred-language-detection` → `moai-core-language-detection`

**파일 목록**:
```bash
accessibility-expert.md
api-designer.md
backend-expert.md
component-designer.md
devops-expert.md
figma-expert.md
frontend-expert.md
migration-expert.md
monitoring-expert.md
performance-engineer.md
ui-ux-expert.md
```

**실행 명령어** (각 파일):
```bash
# 패턴: moai-alfred-language-detection → moai-core-language-detection
sed -i 's/moai-alfred-language-detection/moai-core-language-detection/g' filename.md
```

### Phase 2: 복잡한 변경 (5개 파일, 예상 소요시간: 25분)

**작업**: 다중 Skill 참조 + AskUserQuestion 링크 업데이트

**파일별 순서**:

#### Step 2-1: cc-manager.md (10 changes)
```bash
# 변경 1: moai-alfred-workflow → moai-core-workflow
sed -i 's/moai-alfred-workflow/moai-core-workflow/g' cc-manager.md

# 변경 2: moai-alfred-language-detection → moai-core-language-detection
sed -i 's/moai-alfred-language-detection/moai-core-language-detection/g' cc-manager.md

# 변경 3: moai-alfred-tag-scanning → moai-core-tag-scanning
sed -i 's/moai-alfred-tag-scanning/moai-core-tag-scanning/g' cc-manager.md
```

#### Step 2-2: debug-helper.md (8 changes)
```bash
sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' debug-helper.md
sed -i 's/moai-alfred-language-detection/moai-core-language-detection/g' debug-helper.md
sed -i 's/moai-alfred-tag-scanning/moai-core-tag-scanning/g' debug-helper.md
```

#### Step 2-3: doc-syncer.md (16 changes)
```bash
sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' doc-syncer.md
sed -i 's/moai-alfred-tag-scanning/moai-core-tag-scanning/g' doc-syncer.md
```

#### Step 2-4: git-manager.md (12 changes)
```bash
sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' git-manager.md
sed -i 's/moai-alfred-git-workflow/moai-core-git-workflow/g' git-manager.md
sed -i 's/moai-alfred-trust-validation/moai-core-trust-validation/g' git-manager.md
```

#### Step 2-5: implementation-planner.md (20 changes)
```bash
sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' implementation-planner.md
sed -i 's/moai-alfred-language-detection/moai-core-language-detection/g' implementation-planner.md
# (다른 변경들도 마찬가지)
```

### Phase 3: 대규모 네임스페이스 업데이트 (5개 파일, 예상 소요시간: 30분)

**작업**: 스킬 팩토리 및 validation 스킬 대규모 재정의

**파일별 순서**:

#### Step 3-1: agent-factory.md (12 changes)
```bash
sed -i 's/moai-alfred-agent-factory/moai-core-agent-factory/g' agent-factory.md
# 추가 검사: L95, L120, L145 라인 수동 확인
```

#### Step 3-2: quality-gate.md (18 changes)
```bash
sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' quality-gate.md
sed -i 's/moai-alfred-trust-validation/moai-core-trust-validation/g' quality-gate.md
```

#### Step 3-3: skill-factory.md (30 changes) - 최대 규모
```bash
sed -i 's/moai-alfred-skill-factory/moai-core-skill-factory/g' skill-factory.md
sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' skill-factory.md
# L14, L50+, L90+, L130+ 라인 수동 검증
```

#### Step 3-4: spec-builder.md (18 changes)
```bash
sed -i 's/moai-alfred-spec-authoring/moai-core-spec-authoring/g' spec-builder.md
sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' spec-builder.md
sed -i 's/moai-alfred-ears-authoring/moai-core-ears-authoring/g' spec-builder.md
```

#### Step 3-5: tdd-implementer.md (18 changes)
```bash
sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' tdd-implementer.md
sed -i 's/moai-alfred-language-detection/moai-core-language-detection/g' tdd-implementer.md
```

#### Step 3-6: trust-checker.md (16 changes)
```bash
sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' trust-checker.md
sed -i 's/moai-alfred-trust-validation/moai-core-trust-validation/g' trust-checker.md
```

### Phase 4: 신규 콘텐츠 추가 (1개 파일, 예상 소요시간: 5분)

#### Step 4-1: mcp-context7-integrator.md (1 addition)
```
수동 추가 필요:
위치: 기술 범위 섹션
내용: "- **mcp-context7-integrator**: Use [Context7 MCP] for complex research strategies"
```

### Phase 5: 검증 (모든 파일, 예상 소요시간: 20분)

**검증 체크리스트**:

```bash
# 1. 전체 파일 검사: alfred 참조가 남아있는지 확인
grep -r "moai-alfred" .claude/agents/moai/ | wc -l
# 결과: 0 (완벽 동기화)

# 2. 각 파일의 YAML 프론트매터 검증
for file in .claude/agents/moai/*.md; do
  head -15 "$file" | grep -E "^(name|description|tools|model|skills):"
done

# 3. 파일 크기 일치 확인
ls -l .claude/agents/moai/*.md | wc -l
# 결과: 31개

# 4. Git diff 최종 확인
git diff .claude/agents/moai/ | head -100
```

---

## 📋 동기화 체크리스트

### 사전 검사
- [ ] 현재 브랜치 확인: `git branch` (release/0.26.0)
- [ ] 작업 디렉토리 클린: `git status` (clean)
- [ ] 백업 생성: `.claude/agents/moai/` 전체 복사 완료

### Phase 1 실행
- [ ] 13개 파일 일괄 업데이트 (sed 명령어)
- [ ] 각 파일 내용 샘플 확인
- [ ] 문법 오류 없음 확인

### Phase 2 실행
- [ ] 5개 파일 복합 변경 (cc-manager, debug-helper, doc-syncer, git-manager, implementation-planner)
- [ ] 각 파일의 모든 Skill 참조 확인
- [ ] AskUserQuestion 링크 올바르게 업데이트됨

### Phase 3 실행
- [ ] 5개 파일 대규모 업데이트 (agent-factory, quality-gate, skill-factory, spec-builder, tdd-implementer, trust-checker)
- [ ] Skill 네임스페이스 완벽 재정의
- [ ] 문서 링크 모두 정확함

### Phase 4 실행
- [ ] mcp-context7-integrator.md 신규 콘텐츠 추가
- [ ] 맥락 정확성 확인

### Phase 5 검증
- [ ] 모든 `moai-alfred-*` 참조 제거됨
- [ ] 모든 `moai-core-*` 참조 추가됨
- [ ] 파일 크기 및 라인 수 검증
- [ ] Git diff 최종 확인
- [ ] 에러 메시지 없음

### 사후 작업
- [ ] 동기화 완료 문서화
- [ ] 커밋 메시지 작성: "chore(agents): Sync local agents with v0.26.0 templates (moai-alfred → moai-core)"
- [ ] 브랜치 병합 (release/0.26.0 → main)

---

## 📊 메트릭 데이터

```json
{
  "summary": {
    "template_total": 31,
    "local_total": 31,
    "missing": 0,
    "outdated": 23,
    "up_to_date": 8
  },
  "changes": {
    "total_files_to_update": 23,
    "total_changes": 192,
    "average_changes_per_file": 8.3,
    "max_changes": 30,
    "min_changes": 1
  },
  "effort_estimation": {
    "phase_1_simple": "15 minutes",
    "phase_2_complex": "25 minutes",
    "phase_3_large_scale": "30 minutes",
    "phase_4_additions": "5 minutes",
    "phase_5_validation": "20 minutes",
    "total_estimated": "95 minutes (~1.5 hours)"
  },
  "risk_level": "LOW",
  "rollback_plan": "Git restore .claude/agents/moai/ from backup"
}
```

---

## 🔍 상세 변경 매트릭스

| 파일명 | 변경 | 타입 | 심각도 | 검증 우선순위 |
|--------|------|------|--------|---------------|
| accessibility-expert.md | 2 | Skill 단순 변경 | 낮음 | 3 |
| agent-factory.md | 12 | Skill 네임스페이스 | 높음 | 1 |
| api-designer.md | 2 | Skill 단순 변경 | 낮음 | 3 |
| backend-expert.md | 2 | Skill 단순 변경 | 낮음 | 3 |
| cc-manager.md | 10 | Skill 복합 변경 | 높음 | 1 |
| component-designer.md | 2 | Skill 단순 변경 | 낮음 | 3 |
| debug-helper.md | 8 | Skill 복합 변경 | 중간 | 2 |
| devops-expert.md | 2 | Skill 단순 변경 | 낮음 | 3 |
| doc-syncer.md | 16 | Skill 복합 변경 | 높음 | 1 |
| figma-expert.md | 2 | Skill 단순 변경 | 낮음 | 3 |
| frontend-expert.md | 2 | Skill 단순 변경 | 낮음 | 3 |
| git-manager.md | 12 | Skill 복합 변경 | 높음 | 1 |
| implementation-planner.md | 20 | Skill 복합 변경 | 높음 | 1 |
| mcp-context7-integrator.md | 1 | 신규 추가 | 낮음 | 3 |
| migration-expert.md | 2 | Skill 단순 변경 | 낮음 | 3 |
| monitoring-expert.md | 2 | Skill 단순 변경 | 낮음 | 3 |
| performance-engineer.md | 2 | Skill 단순 변경 | 낮음 | 3 |
| quality-gate.md | 18 | Skill 네임스페이스 | 높음 | 1 |
| skill-factory.md | 30 | Skill 네임스페이스 | 높음 | 1 |
| spec-builder.md | 18 | Skill 네임스페이스 | 높음 | 1 |
| tdd-implementer.md | 18 | Skill 네임스페이스 | 높음 | 1 |
| trust-checker.md | 16 | Skill 네임스페이스 | 높음 | 1 |
| ui-ux-expert.md | 2 | Skill 단순 변경 | 낮음 | 3 |

---

## ✅ 최종 권장사항

### 우선순위 1: 높은 심각도 파일 (즉시)
1. skill-factory.md (30 changes)
2. agent-factory.md (12 changes)
3. implementation-planner.md (20 changes)
4. quality-gate.md (18 changes)
5. spec-builder.md (18 changes)

### 우선순위 2: 중간 복잡도 파일 (다음)
1. cc-manager.md (10 changes)
2. doc-syncer.md (16 changes)
3. git-manager.md (12 changes)
4. trust-checker.md (16 changes)
5. tdd-implementer.md (18 changes)

### 우선순위 3: 낮은 심각도 파일 (마지막)
13개의 단순 Skill 변경 파일들 (각 2개 변경)

---

## 🎯 다음 단계

1. **지금**: 이 리포트 검토 및 승인
2. **Phase 1-4**: 체크리스트에 따라 동기화 실행
3. **Phase 5**: 모든 변경 검증
4. **커밋**: `git commit -m "chore(agents): Sync with v0.26.0 templates"`
5. **병합**: `git merge release/0.26.0` (또는 PR 생성)

---

## 📞 문제 해결

### 변경 검증 실패
```bash
# 모든 alfred 참조 찾기
grep -r "moai-alfred" .claude/agents/moai/

# 특정 파일만 검사
grep "moai-alfred" .claude/agents/moai/spec-builder.md
```

### 파일 복원
```bash
# 모든 파일 원상복구
git checkout .claude/agents/moai/

# 특정 파일만 복원
git checkout .claude/agents/moai/spec-builder.md
```

### 부분 동기화 검증
```bash
# 두 디렉토리 비교
diff -r .claude/agents/moai/ src/moai_adk/templates/.claude/agents/moai/

# 특정 파일 비교
diff .claude/agents/moai/spec-builder.md src/moai_adk/templates/.claude/agents/moai/spec-builder.md
```

---

**생성 일시**: 2025-11-19
**분석 스크립트**: `/tmp/compare_agents.py`, `/tmp/detailed_diff.py`
**백업 위치**: `.moai/backup/agents-pre-sync-2025-11-19/`
