# SPEC-CLOCAL-AUDIT-001 — Audit Verdict (card t308)

> 5-section evidence-bearing format per `verification-claim-integrity.md` §3.
> 대상: `CLAUDE.local.md` (감사 개시 시 663줄), 워크트리 `.claude/worktrees/t308`, 브랜치 `WT-clocal-audit`.
> 제외 구역: §4.1 (git-flow 통합 레인 절) — 형제 카드 t294/t295/t298/t303 소관. 본 감사는 이 구간을 읽지도 고치지도 않았고, 바깥에서 §4.1로 들어가는 포인터의 도달 가능성만 확인했다.

## Claim

CLAUDE.local.md의 검증 가능한 주장 76건을 전수 대조해 **20건을 결함으로 확정**하고 전부 원문에서 최소 수정했다. 나머지 56건은 **52건 기계적 통과 확정(PASS-ATTESTED)**, **2건 역사 기록(HISTORICAL-RECORD)**, **1건 경로 모호(AMBIGUOUS-PATH)**, **1건 외부 검증 불가(EXTERNAL-UNVERIFIABLE)**로 종결했다. 미분류(pending) 0건 — AC-CLOCAL-006 충족.

결함 20건의 성격은 카드가 예상한 그대로였다: **줄번호·심볼 인용의 조용한 노후화**(deploy.go:29→107, deploy.go:122→187-190, update.go:513→554, Makefile:38→40-41, gate.go:160-161 폐기)와 **더 이상 존재하지 않는 대상을 가리키는 서술**(`.agency/` 템플릿 뿌리, `.moai/config/config.yaml`, `status_line.sh.tmpl`의 `export PATH` 줄, `settings.json` vs `.tmpl` 파일명).

### 확정 결함 20건

| INV | 결함 한 줄 | 수정 |
|-----|-----------|------|
| INV-001 | 헤더 `Last Updated: 2026-05-25` — 최종 내용 커밋은 2026-08-27 | CHK-FIX-002 |
| INV-002 | §11↔§12 사이 `---` 이중 삽입으로 빈 섹션 발생 | CHK-FIX-003 |
| INV-020 | 템플릿 뿌리로 적힌 `internal/template/templates/.agency/`가 부재 | CHK-FIX-004 |
| INV-033 | `handle-*.sh`를 "Generated hook wrappers (not templates)"로 서술 — 실제로는 `.sh`/`.sh.tmpl` 템플릿 쌍 | CHK-FIX-018 |
| INV-047 | `.moai/{logs,state,reports}`를 "Never in Templates"로 적었으나 템플릿 트리에 빈 디렉터리 스캐폴드가 존재 | CHK-FIX-005 |
| INV-049 | `moai ast-grep --rules-dir` 기본값이 "여전히 구 경로"라는 서술 — t50에서 폴백 자체가 제거됨 | CHK-FIX-019 |
| INV-070 | `CleanMoaiManagedPaths` 인용 `deploy.go:29` → 실제 107 | CHK-FIX-006 |
| INV-071 | `.moai/config` 삭제 지점 인용 `deploy.go:122` → 실제 187-190 (8개 뿌리 목록 자체는 정확) | CHK-FIX-006 |
| INV-073 | `archiveLegacySkills` 호출 지점 `update.go:513` → 실제 `internal/cli/update.go:554` (인용만 정정, 코드 결함은 미판정) | CHK-FIX-007 |
| INV-074 | `internal/cli/gate.go:160-161` 빈 rules_dir 폴백 — 동작도 줄 위치도 현재 부재 | CHK-FIX-019 |
| INV-075 | `make install`(Makefile:38) → 실제 install 타깃 40, 레시피 41 | CHK-FIX-008 |
| INV-082 | "cli/template/hook 90%+"의 기계적 근거 미발견 — strict 프로필의 전역 게이트만 존재 | CHK-FIX-009 |
| INV-090 | `MOAI_CONFIG_DIR`을 `applyEnvOverrides`가 읽는 5개 중 하나로 서술 — 실제로는 별도 config-dir resolver 소관 | CHK-FIX-010 |
| INV-092 | 주 설정 파일로 적힌 `.moai/config/config.yaml`이 부재 (섹션 파일 32개만 존재) | CHK-FIX-011 |
| INV-101 | §8이 템플릿 설정 파일명을 `settings.json`으로 인용 — 실제 `settings.json.tmpl` | CHK-FIX-012 |
| INV-102 | `TemplateContext` 발췌 2필드가 전체 구조체인 것처럼 제시 | CHK-FIX-013 |
| INV-103 | `deployer.Deploy(ctx, projectRoot, mgr, ctx)` — 첫 인자는 `context.Context`, TemplateContext는 마지막 인자 | CHK-FIX-014 |
| INV-104 | `status_line.sh.tmpl`에 있다고 인용한 `export PATH="{{.GoBinPath}}:$PATH"` 줄이 파일에 없음 | CHK-FIX-015 |
| INV-110 | §18 References 행이 전환 이전 모델만 요약 | CHK-FIX-016 |
| INV-111 | §23 References 행이 전환 이전 모델만 요약 | CHK-FIX-017 |

### 종결 분류 집계 (AC-CLOCAL-010)

| 분류 | 건수 |
|------|------|
| DEFECT-CONFIRMED | 20 |
| PASS-ATTESTED | 52 |
| HISTORICAL-RECORD | 2 (INV-044 커밋 `ed04e40e6`, INV-083 태그 `v3.0.1`) |
| AMBIGUOUS-PATH | 1 (INV-057 `memory/…` 참조의 실제 대상은 전역 auto-memory 디렉터리) |
| EXTERNAL-UNVERIFIABLE | 1 (INV-141 Vercel 단가) |
| KNOWN-UNRESOLVED | 0 행 (아래 별도 항목으로 이월) |
| OPEN-QUESTION | 0 |
| **합계** | **76** |

### 이월: 재판정하지 않은 기존 미해결 4건 (REQ-CLOCAL-005)

카드와 SPEC이 지정한 대로 **존재 여부를 다시 다투지 않고** 그대로 이월한다.

1. `CleanMoaiManagedPaths`에 보호 목록 부재 (근본 원인).
2. `archiveLegacySkills`가 wipe **이후**에 호출됨 — 본 감사는 호출 지점 인용만 `update.go:554`로 정정했고 순서 결함 자체는 판정하지 않았다.
3. `--dry-run`이 `CleanMoaiManagedPaths` 삭제 예정 목록을 미리보기하지 않음.
4. BranchGuard가 읽기 전용 `git branch` 조회까지 차단 — 이 서술은 §4.1 안에 있어 **제외 구역**이며, 소관은 t294/t295/t298/t303이다.

## Evidence

전체 기록: `.moai/reports/t308/checks-transcript.md` (567줄, CHK-000a..CHK-041 + CHK-FIX-001..020). 원시 출력 사본: `.moai/reports/t308/raw/b1..b13.txt`.

주장 → 증거 대응:

- 인용 노후화 5건 (INV-070/071/073/074/075) → CHK-022, CHK-023, CHK-024, CHK-006, CHK-025. 각 항목은 해당 파일을 열어 심볼명과 줄 범위를 함께 확인했다.
- 부재 확정 4건 (INV-020 `.agency/`, INV-092 `config.yaml`, INV-104 `export PATH` 줄, INV-101 파일명) → CHK-016, CHK-030, CHK-032. 전부 `ls`/`grep`의 비영(非零) 종료코드 또는 실제 파일 내용 판독으로 확정했고, 텍스트 추론만으로 단정한 항목은 없다.
- 값·기본값 3건 (INV-080, INV-082, INV-090) → CHK-009, CHK-028, CHK-029. `internal/config/defaults.go`, `quality.yaml`, `manager.go`의 실제 코드와 대조했다.
- LSEL §28 6건 (INV-130..135) → CHK-039. **판독 표면을 항목마다 명시**했다(PRIMARY 체크아웃 3건, 워크트리 tracked 3건) — AC-CLOCAL-012.
- 존재 스윕 20건 (Local-Only 블록) → CHK-018/019/020, 원시 출력 `raw/b3.txt`·`raw/b4.txt`.
  이 두 파일은 **재생성본**이다 — 원래 배치는 리다이렉트 없이 인라인으로 읽혔고 디스크에 남지 않았다. 커밋 전에 리드가 끊어진 인용을 잡아냈고, 트리가 그대로였으므로(HEAD `8d8da0b2b` 불변) 동일한 명령을 다시 실행해 기록했다(CHK-042). 재생성 결과는 CHK-018/019/020이 원래 적어둔 값과 전부 일치했고 판정은 하나도 바뀌지 않았다.

방법상 주의 1건 기록: INV-130의 첫 질의는 `.command` 필드를 걸러 아무것도 반환하지 않았다. 스크립트 경로는 `.args[2]`에 있었다. **질의 형태 하나에서의 무응답은 부재가 아니다** — 이 문서가 과거에 §2.2 gate 로더로 겪었던 실측 반증과 같은 계열의 함정이며, 재질의로 통과 확정했다.

## Baseline-attribution

- 대상 사본: `CLAUDE.local.md`, 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t308`, 브랜치 `WT-clocal-audit`.
- 개시 HEAD (재개 세션): `git rev-parse --short HEAD` → `8d8da0b2b` (CHK-010). 계획 단계가 고정한 `d29b8942e`에서 develop이 전진했으나, `git diff --stat d29b8942e..8d8da0b2b -- CLAUDE.local.md`가 빈 출력이라 대상 파일은 바이트 동일이었다 (CHK-000g). SHA 고정 인용은 전부 유효하다.
- 코드·저장소 실측 기준 트리: **이 트리 @ 8d8da0b2b**. `internal/**`, `Makefile`, `.claude/**`, `.moai/docs/**`, `.github/workflows/**`.
- 제외 구역 실측 재도출: `grep -n '^### §4.1\|^## 5\.' CLAUDE.local.md` → 헤딩 L274, `## 5.` L339 ⇒ 라이브 구간 **L274–L337** (고정 기준 273–337 대비 +1, 구역 위쪽 수정분에 의한 이동). 재개 이후 이 구간은 읽기·쓰기 모두 없었다.
- 종료 시 파일 길이: `wc -l CLAUDE.local.md` → **665**줄 (동결 663 → 재개 시점 666 → 종료 665). 재개 세션의 수정은 삽입 37 / 삭제 35(`git diff --stat`)로 순 -1줄이며, 제외 구역 경계는 재측정 결과 **L274–L337로 불변**이다.
- PRIMARY 체크아웃 접근: §28 판독 3건뿐(`.claude/settings.local.json`, `.moai/state/lsel/clusters-history/`, `.moai/lessons-inbox.jsonl`). 쓰기 0건.

## Gaps — 관측하지 않은 것

1. **§4.1 전 구간(L274–L337)** — 제외 구역. 그 안의 어떤 주장도 검증하지 않았다. 바깥→안 포인터의 도달 가능성만 확인했다(CHK-037).
2. **INV-010의 절대 경로** — 서술이 가리키는 PRIMARY 절대 경로(`/Users/goos/MoAI/moai-adk-go/internal/template/templates/`)는 탐침하지 않았다. REQ-CLOCAL-009가 §28 예외 외의 PRIMARY 판독을 금지하기 때문이다. 워크트리 상대 구조가 동일하다는 사실만 확인했다.
3. **INV-107의 catalog.yaml 대조** — `internal/template/catalog.yaml`에는 언어 항목이 아예 없다. 인벤토리가 제안한 catalog 패리티 교차확인은 대상 부재로 수행 불가였고, 16개 언어 집합은 `lsp.servers.<lang>` 항목으로 대신 확정했다.
4. **INV-091의 문서 히트** — `internal/config/CLAUDE.md`에 `MOAI_USER_NAME`/`MOAI_CONVERSATION_LANG` 언급이 2건 있다. 코드가 아니라 문서이므로 "코드 판독자 0" 주장은 유효하지만, **그 문서 자체가 두 변수를 우선순위 1층으로 서술한다** — 대상 문서 밖이라 수정하지 않았고 아래 신규 카드 권고로 넘긴다.
5. **§2.3 코드 결함 3건의 실재 여부** — REQ-CLOCAL-005에 따라 재판정하지 않았다. 이월만 했다.
6. **plan-phase "미해결 질문 7건"** — 지시서가 언급한 7건짜리 목록을 계획 산출물에서 찾지 못했다. 실재하는 것은 plan-audit iter-1의 **루브릭 7번 항목** "Open-question handling — None orphaned, PASS"(즉 미해결 질문 0건 판정)와 D1–D8 결함(전부 해소), N1–N4 선택 지적(아래)이다. 7건 목록이 별도로 존재한다면 본 감사는 그것을 보지 못했다.
7. **CLAUDE.local.md L94–L108의 `.agency/` 잔여 언급** — Template-First 규칙 본문과 검증 문장이 여전히 `.agency/`를 나열한다. INV 행이 아니어서 수정하지 않았다(범위 절제). INV-020의 자매 지점으로 남는다.
8. **b3/b4 원시 출력의 출처** — 위 두 파일은 원래 실행 시점의 stdout 사본이 아니라, 같은 트리(HEAD `8d8da0b2b`)에서
   같은 명령을 다시 돌려 만든 재생성본이다(CHK-042). 값은 원래 관측과 일치하지만, "그때 그 프로세스의 출력"이라는
   의미의 원본은 존재하지 않는다. 배치 5·6의 출력은 아예 파일로 남기지 않았고, 해당 findings는 CHK-020/021/022/025
   본문에 직접 인용돼 있다(외부 경로 인용 없음).
9. **템플릿 미러 검증** — 대상은 로컬 전용 문서(`CLAUDE.local.md`)라 Template-First 미러 의무가 없다. `make build`도 실행하지 않았다(코드 변경 0건).

## Residual-risk — 관측했음에도 여전히 틀릴 수 있는 것

1. **줄번호는 다시 노후화한다.** 이번에 고친 인용 5건은 8d8da0b2b 기준이다. `deploy.go`·`update.go`·`Makefile`이 바뀌면 같은 방식으로 다시 어긋난다. 근본 완화책(줄번호 대신 심볼·앵커 인용)은 이 카드의 범위 밖이다.
2. **INV-081의 "기본 5초"는 표면이 모호하다.** 템플릿 `settings.json.tmpl`의 36개 훅 중 23개가 5로 설정돼 있어 "이 저장소에서 통용되는 값"으로는 참이다. 다만 Claude Code 자체의 문서화된 훅 기본값은 다르며, 문장은 둘을 구분하지 않는다. 결함으로 확정할 근거가 없어 통과 처리했으나 오독 여지는 남는다.
3. **INV-045는 부재를 근거로 통과시켰다.** `.moai/state/last-cc-version.json`은 워크트리에 없다. 런타임 산출물이고 템플릿에도 없다는 두 사실로 "Local-Only" 주장을 지지했을 뿐, PRIMARY에 실제로 존재하는지는 (판독 권한 밖이라) 확인하지 않았다.
4. **HISTORICAL-RECORD 2건은 앵커만 확인했다.** 커밋과 태그의 실재만 확인했을 뿐, 그 서사가 서술대로 일어났는지는 재실행하지 않았다(REQ-CLOCAL-007 규정대로).
5. **증거 파일 기록 습관이 배치마다 일관되지 않았다.** 13개 배치 중 9개는 리다이렉트해 남겼고 4개(b3·b4와 번호를
   건너뛴 두 배치)는 인라인으로만 읽었다. 이번엔 트리가 움직이지 않아 재생성으로 복구됐지만, HEAD가 한 번이라도
   전진했다면 같은 복구가 불가능했다 — 그 경우 해당 20개 행은 Gap으로 강등됐어야 한다. 인용할 경로는 인용하는
   시점이 아니라 관측하는 시점에 기록해야 한다.
6. **감사 중 develop이 다시 전진할 수 있다.** 개시 시점 `8d8da0b2b` 기준이며, 커밋 직전 HEAD 재판독은 오케스트레이터 몫이다.

## 신규 카드 권고 (범위 밖 결함 — 본 감사는 수정하지 않았다)

1. **`internal/config/CLAUDE.md:12-13`** — 설정 우선순위 1층을 "`MOAI_USER_NAME`, `MOAI_CONVERSATION_LANG` 등 환경변수"로 서술한다. 실측(CHK-029): 두 변수를 읽는 Go 코드는 `internal/`·`pkg/`·`cmd/` 전체에 0건이다. CLAUDE.local.md는 이미 이 사실을 명시하고 있어, **두 문서가 서로 모순**한다. 같은 파일 13행은 `envkeys.go`에 `EnvUserName = "MOAI_USER_NAME"` 상수가 있다고 예시하는데 그 상수도 부재다(CHK-029의 envkeys 판독 목록에 없음). 권고: 해당 문서를 실측에 맞추거나, 두 환경변수를 실제로 구현.
2. **`internal/template/catalog.yaml`** — 268줄 전체에 언어 항목이 없다. 16개 프로그래밍 언어 중립성(§15)의 기계적 근거가 `lsp.servers.*`에만 있고 catalog에는 없다는 뜻이다. 결함인지 설계인지 판정하려면 catalog의 소관 정의가 필요하다. 권고: catalog 스키마 소관 확인 후 중립성 검증 경로를 한 곳으로 정리.
3. **CLAUDE.local.md `.agency/` 잔여 언급 (L94, L106, L108)** — INV-020으로 템플릿 뿌리는 정정했으나 Template-First 규칙 본문·검증 문장에는 `.agency/`가 남아 있다. 인벤토리 행이 아니어서 이번 범위에서 제외했다. 권고: 후속 카드에서 3곳 일괄 정리(대상 문서가 같으므로 t308 후속으로 묶는 편이 자연스럽다).

## 계획 감사 선택 지적 N1–N4 — 이월 및 처리

plan-auditor iter-2가 "gating은 아니나 M8에서 흡수 권장"으로 남긴 4건. **전부 이번 run-phase에서 흡수했다** (근거: 4건 모두 본 카드의 쓰기 허용 경로 안에 있고, 수정 비용이 한 줄씩이다).

| 지적 | 내용 | 처리 |
|------|------|------|
| N1 | INV-113 행의 예시 셀이 실측과 불일치(`L657 references §2.3`) | 실측 기반 예시로 교체 — 재개 기준 L247과 §18 References 행 L638 |
| N2 | INV-083 행 산문 오기 `loader_gate.py` (Go 저장소) | `loader_gate.go` (`loadGateSection`)로 정정 |
| N3 | INV-121 앵커가 터미널 명령 블록만 가리켜 슬래시 목록과 어긋남 | 앵커를 L40–63으로 재기술, goal/todo는 저장소 전역 확인임을 명시 |
| N4 | spec.md HISTORY 0.1.0 행이 사후 개작돼 원래 동결값(70건)을 덮음 | 0.1.0 행을 "70 items, INV-001..INV-135"로 복원 (증가분은 1.0.1 행이 이미 기록) |

## 자체 검증 (plan.md §E)

| 항목 | 결과 |
|------|------|
| E1 인벤토리 종결 | 76/76 행이 정확히 한 개 분류를 보유, `pending` 0 (스크립트 재계수: `remaining-pending: 0`) |
| E2 증거 귀속 | 확정 결함 20건 전부 CHK 항목 보유; 본 문서 5개 절 완비 |
| E3 diff 절제 | `CLAUDE.local.md` 변경은 INV 행으로 추적 가능한 최소 수정만; 모든 신규 문구에 `[2026-08-27 감사 정정]` 표식 |
| E4 범위 가드 | `git status --short` 결과가 허용 3경로만 표시 (아래 §E.3에 기록) |
| E5 서사 보존 | 역사 서술은 삭제 없이 날짜 표식만 부기 |

---

## 후속 추가분 — B5 (인벤토리 밖, lane-13 제보)

동결 인벤토리 76건에는 없던 지적이 병합 대기 중에 리드를 거쳐 들어왔다. 전달받은 주장을 그대로
채택하지 않고 실측한 뒤 처리했다(CHK-043).

- **정본**: `.moai/config/sections/git-strategy.yaml:8` → `workflow: git-flow` (하이픈 있음).
- **문서**: §2.3의 세 지점이 하이픈 없는 `gitflow`를 썼다 — L181 주석, L182 확인 명령, L187 수동 복구값.
- **결과**: L182는 정본이 담고 있지 않은 문자열을 찾으므로, **올바른 설정에 대고** `REVERTED — 재적용 필요`를
  찍는다. 즉 멀쩡한 값을 고치라고 지시하는 거짓 경보다. L187을 글자대로 따르면 틀린 값을 써넣는다.
- **처리**: 세 지점 모두 `git-flow`로 교정 (CHK-FIX-021). 코드 블록 안 두 줄은 그대로 복사해 쓸 수 있도록
  표식을 넣지 않았고, 산문인 L187에만 `[2026-08-27 감사 정정]`을 달았다.

집계 영향 없음 — INV 행이 아니므로 76건 분류(20/52/2/1/1)는 그대로다. 확정 결함 수는 문서 전체 기준
20건 + 본 건 1건 = 21건으로 읽는 것이 정확하다.
