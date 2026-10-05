# Plan — SPEC-DOCS-LOCALE-PARITY-REPAIR-001

## §A Context

- 카드 t538 · 브랜치 `WT-docs-v313-locales` · 베이스 `bce6d7e08` (= origin/develop 팁).
- 측정 SSOT: `.moai/reports/t538/plan-phase.md` (2026-09-08 실측, `/usr/bin/grep`). **본 plan 이 측정 파일과 어긋나면 측정 파일이 이긴다.**
- 성격: 신규 기능이 아니라 착지 현실에 대한 문서 정렬(e2e en/zh 가 기능 부재를 주장하는 부정확 문서 수리) + t535 가 기록만 남긴 기존 점위반 수리 + v3.1.3 SVG 축 보강(포함 결정).
- 개발 모드: 문서 SPEC — DDD/TDD 사이클 아님. 검증은 grep 카운트 + hugo 빌드 + 4-로케일 패리티(Skill("hns-oss-docs-verify") 레시피).

## §B Known Issues

1. **배차 전제 절반 stale**: v3.1.3 캐치업 축은 이미 반영(t274/t496). 실갭은 G1-G3+G4 뿐 — spec.md §B.1.
2. **G1 시대 귀속**: v3.1.0대 부채(`SPEC-DESKTOP-NATIVE-E2E-001` docs 연기), v3.1.3 아님 — spec.md §B.2. 감사 오귀속 방지용.
3. **t535 측정 정정**: sync-audit F2 가 zh 만 지목했으나 ja 도 누락 → G2 는 두 로케일.
4. **셸 grep 함정**: ugrep 래퍼가 조용히 건너뜀 → 모든 검증 grep 은 `/usr/bin/grep` (REQ-012).
5. **`hns-oss-docs-run` 러너 결함**: 저장-스크립트 문법 오류 — run 페이즈는 스페셜리스트 직접 스폰으로 우회 (REQ-013).

## §C Pre-flight

- [x] 워크트리 진입 확인: `git rev-parse --show-toplevel` → `.claude/worktrees/t538`
- [x] 베이스 확인: `git rev-parse --short HEAD` → `bce6d7e08` = origin/develop 팁
- [x] baseline 재측정 완료: en:170 "not yet provided" 1건, zh:170 "尚未提供" 1건, doctor ja/zh 예시 4행(`hook` 종단), ko:41/ja:39/zh:39 간격위반, skill-guide SVG0 언급 0 — 모두 본 트리에서 직접 관측
- [x] SPEC ID regex PASS + 카탈로그 중복 없음
- [ ] run 페이즈 진입 전: Implementation Kickoff Approval (HUMAN GATE — 본 plan 승인으로 개시)

## §D Constraints

1. ko 정본 → en/zh 파생; ja 는 G1 표 형태 참조 렌더링(열 구조)으로 열람만.
2. [HARD] 4-로케일 동일 착지: 변경 집합 전체가 하나의 커밋 체인 (REQ-011).
3. new-badge 금지, ko·ja e2e 무변경 (REQ-006).
4. G3 수리는 기존 절 3곳 한정, Codex Wiring 절 불가침 (REQ-008).
5. 원어 토큰 검증: 데스크탑-네이티브 / デスクトップネイティブ / 桌面原生 (REQ-012).
6. i18n 숙칙: Mermaid TD-only, 본문 이모지 금지, 강조 간격, adk.mo.ai.kr URL 한정.
7. 시간 예측 금지 — 우선순위·위상 정렬로만 마일스톤 서술.

## §E Self-Verification

| 항목 | 방법 |
|---|---|
| plan-측정 일치 | 본 plan 의 모든 수치는 `.moai/reports/t538/plan-phase.md` 인용 — 재유도 없음 |
| RED-now AC 존재 | AC-001·002(deferral en/zh), AC-003(플래그 행), AC-006(예시행), AC-007(간격 스캔), AC-008(SVG0) — acceptance.md §D 실번호와 정렬 (plan-audit iter1 D3 재정렬) |
| 4-로케일 커버리지 | 변경 파일 ×4 로케일 매핑 표 (§F M0) |
| 빌드 게이트 | hugo exit 0 + WARN/ERROR 0 (AC-009) |

## §F Milestones

> 변경 파일 매핑 (M0 기준 — 전 마일스톤 공통 대상):

| 파일 | ko | en | ja | zh |
|---|---|---|---|---|
| `utility-commands/moai-e2e.md` | 무변경 | M1 | 무변경 | M1 |
| `cli-reference/doctor.md` | M2 (간격 1행) | 무변경 | M2 (예시+간격) | M2 (예시+간격) |
| `advanced/skill-guide.md` | M3 (정본) | M3 (파생) | M3 (파생) | M3 (파생) |
| `CHANGELOG.md` | M3 (단일 파일) | — | — | — |

### M1 — G1: e2e desktop-native en/zh 수리 (Priority High — 내용 결정 대부분이 여기 집중)

1. content-author 스폰: ko:58(플래그 행)·78-80(3-OS 매트릭스)·98(자동감지 행)·176(라우팅 절)·매트릭스 직후(호스트 OS 문단)를 정본으로 열람.
2. en `moai-e2e.md` 수리: 5축 전부 반영, :170 deferral 문장 삭제 → ko:176 라우팅 의미론 영어 파생.
3. zh `moai-e2e.md` 수리: 동일 (ko 정본 → zh 파생, en 재활용 가능하나 zh 자연어 검증).
4. ja:78-80 표 형태 참조로 열 구조 일치 확인 (내용 복사 아님 — en/zh 자연어는 ko→각 로케일 파생).
5. 검증: acceptance.md AC-001~AC-005 (RED-now 해소 + 패리티 + 호스트 OS 문단 토큰), AC-009 (ko·ja 무변경·배지 금지).

### M2 — G2+G3: doctor.md ja·zh 예시 + 기존 절 간격 (Priority High — M1 정본 열람 후 파생 가능하나 파일 독립적이라 병렬 가능)

1. ja `doctor.md`: 예시 블록(:106-111)에 `moai doctor permission`·`moai doctor sandbox` 행 추가 (ko:120-121 구성 동등, 일본어 주석).
2. zh `doctor.md`: 동일 (중국어 주석).
3. 간격 수리 3곳 (i18n 숙칙 §5 — 괄호문은 마커 밖): ko:41 `**권고(advisory)**`→`**권고** (advisory)`, ja:39 `**勧告 (advisory)**`→`**勧告** (advisory)`, zh:39 `**建议 (advisory)**`→`**建议** (advisory)`. Codex Wiring 절(ko:69·ja:67·zh:67) 불가침 — bold+괄호 패턴이 없어 이미 rule §5 준수(2026-09-08 전체-파일 스캔 실측: 위반은 3곳뿐, en 0건). 판정은 전체-파일 스캔 0건(AC-004).
4. 검증: AC-006 (예시행 — 블록 한정 앵커 0→2, 전체 파일 2→4), AC-007 (간격위반 전체-파일 스캔 0).

### M3 — G4 + CHANGELOG (Priority Medium — 기계적 파생 + 장부)

> **G4 동일-카드 근거 (1행)**: G4 는 v3.1.3 Added 중 docs 표면 언급이 없던 마지막 축으로, 이 카드의 v3.1.3 정합 취지를 4-로케일 1-2행 패리티 추가로 완결한다 — 새 문서 주제를 열지 않는다.

1. content-author: ko `skill-guide.md` 기존 svg-infographic 문맥(ko:158, :163 인접)에 SVG060-064(접근성 이름: role="img", aria-labelledby, title 최전단)·SVG070-074(커넥터 기하) 사실문 1-2행 추가.
2. locale-translator: en·ja·zh 파생.
3. `CHANGELOG.md` `[Unreleased]` → `### Added` **최상단** 삽입(t535 선례 = 직하 배치 — `### Docs` 섹션은 현행 [Unreleased] 부재): 수리 집합 요약 + `t538` 카드 id + G2 두-로케일 정정 명시.
4. 최종 게이트: hugo 빌드(WARN/ERROR 0) + 4-로케일 패리티 + 커밋 체인 단일성 확인 (AC-009~AC-011).

## §G Anti-Patterns

- ❌ ko/ja e2e 를 "맞춤 정리"로 건드리기 — 무변경이 요구사항이다 (REQ-006).
- ❌ en 문장을 zh 에 그대로 복사 — zh 는 zh 자연어 파생.
- ❌ new-badge 추가 — 현실 정렬이지 신규 기능이 아니다.
- ❌ Codex Wiring 절 재수리 시도 — 이미 준수.
- ❌ 셸 `grep` 검증 — ugrep 래퍼가 조용히 건너뛴다 (REQ-012).
- ❌ `hns-oss-docs-run` 러너 사용 — 문법 오류 결함 (REQ-013).
- ❌ 부분-로케일 착지 — 4-로케일 동일 착지 HARD (REQ-011).
- ❌ G4 파생 시 ko skill-guide :167-177 구간의 기존 내부-경로 서술 스타일(`.moai/reports/...` 참조) 모방 — 공개 트리에 없는 경로다 (plan-audit iter1 F5 경고, 수리는 후속 카드 소관).

## §H Cross-References

- `.moai/reports/t538/plan-phase.md` — 측정 SSOT
- spec.md §D (REQ-001~013) / acceptance.md §D (AC-001~011)
- Skill("hns-oss-docs-i18n-rules") · Skill("hns-oss-docs-verify")
