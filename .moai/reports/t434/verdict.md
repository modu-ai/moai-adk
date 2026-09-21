# t434 — §E.5 잔재 5건 처리 방침 판정 (verdict)

- **브랜치**: WT-e5-residue-verdict (develop 2660bcd09 기반)
- **카드**: t434 lane-10 (G2b 순차 3장째), Tier S
- **[HARD] 첫 작업 이행**: 정리가 아니라 판정부터. 이 판정서가 카드의 주 산출물.

## Claim

sanctioned chicken-and-egg 형태(`mx_commit_sha: (this commit)` placeholder)는 **삭제하지 않는다**. 규약(§E.5 은퇴)은 "새 작성 금지"의 시간 방향 규제이지 "기존 기록 삭제" 명령이 아니다. 5건의 처우는 균일하지 않다 — 상태별로 분리하며, 실측상 3건은 이 브랜치에 존재하지 않아( primary 전용 untracked) 이 브랜치에서 실행 가능한 정리는 0건이다.

## Evidence

### 5건 실측 (이번 실행, WT-e5-residue-verdict 트리 + primary 판독)

| SPEC | 트리 존재 | §E.5 잔재 형태 | 귀속 근거 |
|---|---|---|---|
| SPEC-V3R6-GRAPH-FRESHNESS-001 | develop ✅ | progress.md:306 `mx_commit_sha: (this commit)` | **완비** — 002 progress.md:136이 `f32e9a346`(close commit)으로 귀속 확정 |
| SPEC-V3R6-GRAPH-FRESHNESS-002 | develop ✅ | **없음** — research/plan이 "no §E.5 (schema-correct)" 서술만 | — (sanctioned 선언자: :135 "the sanctioned chicken-and-egg form") |
| SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001 | **primary 전용 untracked** | progress.md:106 `(this commit)` | 부분 — SKIP-JUSTIFIED 서술 있으나 SHA 없음 |
| SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001 | **primary 전용 untracked** | :192 §E.5 섹션 + :236 `(this commit)` | 부분 — Step C 판정 상세하나 chore 커밋 SHA 없음 |
| SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 | **primary 전용 untracked** | :644 `§E.5 Mx-phase audit-ready signal \| pending` | — **거짓 대기**: 은퇴 후 채워질 수 없는 pending |

측정 명령: `grep -n 'mx_commit_sha\|## §E\.5\|E\.5 Mx' <progress.md>` (5건 전부), `git log --oneline -- .moai/specs/SPEC-V3R6-LIFECYCLE-SYNC-GATE-001/` → 무출력(이 트리 이력에 경로 없음), `git check-ignore -v .moai/specs/...` → exit=1(ignored 아님 → untracked 확정).

### 판정 근거 (sanctioned 유지)

1. **002 progress.md:135-136**: `(this commit)` L60 placeholder를 "**the sanctioned chicken-and-egg form**"으로 명시 선언 + 귀속 규칙 제공("resolves to f32e9a346 (the close commit that appended it, per the M4 row-3 record)"). 카드가 말한 sanctioned 선언의 정확한 위치.
2. **closer.go:256**: "Backfills §E.5 mx_commit_sha per the L60 atomic-backfill chicken-and-egg (**placeholder acceptable** — the close requires a non-empty mx_commit_sha)" — 코드 층도 placeholder를 L60 해법의 일부로 명명.
3. **closer_test.go:422-437**: "The 5 already-discharged target SPECs"의 §E.5 mx_commit_sha 상태 형태 집합(absent / `null` / `(this commit)` / empty)을 진실표 픽스처(`completedNoOpVariants`)로 명명 — no-op 불변식(AC-LSG-018/022)이 보호하는 프로덕션 현실. (테스트는 자체 문자열 픽스처라 progress.md 직접 참조는 아니나, 명명된 상태 집합의 실체가 이 5건.)
4. **귀속 정보 파괴 방지**: `(this commit)`은 mx 감사 완료 시점을 가리키는 유일 기록일 수 있다(001처럼 귀속 완비 사례). 삭제는 규약 정화가 아니라 이력 소실.

### 카드 질문 — "두 경우의 처우가 같아야 하는가" → **같지 않다**

| 경우 | 성격 | 처우 |
|---|---|---|
| t252가 막은 3건 — 지금 새로 은퇴 신호를 기입하려던 시도 | 은퇴 **이후** 기입 시도 | 차단이 옳음(이미 막음) — 규약 준수 |
| 남아 있는 잔재 — 이전 closer가 남긴 기록 | 은퇴 **이전** 기록 | 보존이 옳음 — 이력 데이터 |

구분 기준은 **기록 시점**: 은퇴 규약(lifecycle-sync-gate.md의 §E.5 은퇴)은 새로운 작성을 금지하는 시간 방향 규제이고, 기존 기록에 소급 삭제를 명령하지 않는다. 순서 어긋난 착지가 되돌림이 아니라 기록으로 남는 것과 같은 원리.

## 처우 (판정에 따른)

| SPEC | 처우 |
|---|---|
| GRAPH-FRESHNESS-001 | **보존** (귀속 완비, sanctioned) — 변경 불필요 |
| GRAPH-FRESHNESS-002 | **대상 아님** (잔재 없음) |
| TEMPLATE-MIRROR-CASCADE-001 | **보존 + 권고**: 반입 시 `(this commit)` 옆에 귀속 SHA 확정 부기(서술은 있으나 SHA 부재) |
| ANTHROPIC-AUDIT-TIER3-001 | **보존 + 권고**: 동일 — Step C 판정 기록은 상세하므로 SHA만 보완 |
| LIFECYCLE-SYNC-GATE-001 | **권고: 반입 시 pending → retired 표기로 갱신** — 은퇴 후 영원히 미충족인 거짓 대기는 잔재 보존 대상이 아닌 오표기 |

**이 브랜치에서 실행한 정리: 0건** — 이는 미완이 아니라 판정의 결과다. 실행 가능 대상(develop 내)은 001(보존 확정)과 002(대상 아님)뿐이고, 3건은 이 브랜치에 파일이 없어 primary의 untracked 문서를 이 카드가 수정·반입하는 것은 범위 밖(카드는 "이미 남은 5건의 처우"만 다룸).

## 부수 발견 (리드 보고 축)

**closer.go의 AC-LSG 진실표 원천 SPEC(SPEC-V3R6-LIFECYCLE-SYNC-GATE-001)이 develop에 부재** — primary 체크아웃에 untracked로만 존재. t386이 다룬 "판정 근거가 세션과 함께 사라진다"의 SPEC 문서판: develop 이력에 없으므로 다른 클론에서는 그 SPEC이 존재하지 않는다. TEMPLATE-MIRROR-CASCADE-001·ANTHROPIC-AUDIT-TIER3-001도 같은 상태(2026-05~07의 SPEC 3건). **SPEC 문서 반입(커밋) 판단이 필요한 별도 카드** 후보 — 이 카드에서 실행하지 않음.

## Gaps (미검증)

- 3건 untracked SPEC의 progress.md `pending`/`(this commit)` 주변 정밀 판독은 primary에서만 가능 — worktree 격리 가드상 primary에서 git 조작 없이 파일 Read만으로 부분 확인했고, 전체 파일 대조는 하지 않음.
- 2건(untracked 아님) 중 GRAPH-FRESHNESS-001의 `(this commit)`→f32e9a346 귀속 논리 자체는 002:136 기록 의존 — 다만 `git cat-file -t f32e9a346` → commit 존재, `git show -s` → "chore(SPEC-V3R6-GRAPH-FRESHNESS-001): Mx-phase audit-ready signal + 3-phase close" 제목 직접 관측(이 브랜치 이력 내)으로 대응 커밋의 실재는 확인.

## Residual-risk

- untracked 3건이 primary의 다른 정리(clean/restore)에 노출되어 있음 — SPEC 반입 카드가 나오기 전까지 유일본.
- closer.go 수리 카드(카드 후보 1번)가 진행되면 closer.go:193-194·256 주석의 "5 already-discharged" 서술과 실측 현실(3건은 트리 밖)의 간극이 커질 수 있음 — 그 카드에 이 판정서 참조 권고.
