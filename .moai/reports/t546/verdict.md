# Card Verdict — t546 (GitHub 열린 이슈 스윕)

> Tree: `.claude/worktrees/t546` · branch `WT-issue-sweep` · base `f3517996a` (진입 시점 origin/develop 팁, 2026-09-08)
> Dispatch: lead, 2026-09-08 (운영자 지명) · session `a403835e` · evidence: `.moai/reports/t546/`

---

## Claim

1. 열린 이슈 24건 전수 분류: **해소 확정 8건 · 문서-응답 후보 1건 · 보류 2건 · 카드 없음 13건**.
2. 리드 초기 분류(해소 후보 6건) 전부 재확인 성립. 스윕에서 **#1686(t511 초안 동봉)·#1694(t510 회신 대기) 2건이 추가 해소 후보로 확인**됐다.
3. 회신 초안 7파일 정식화 완료. **게시는 전부 리드 승인 게이트 대기** (외부 공개 행위).
4. #1659 증상을 오늘(2026-09-08, Claude Code 2.1.263) 생관측으로 재확인 — upstream 소관 확정.

## Evidence — 착지 조상 검증

명령: `git merge-base --is-ancestor <sha> origin/develop` (각 1회, 이번 실행). baseline: origin/develop = `f3517996a` (2026-09-08).

| SHA | 카드 | 결과 |
|---|---|---|
| `e265e69a6` | t511 | ancestor — 착지 |
| `9bf08c698` | t512 | ancestor — 착지 |
| `a9d2fb641` | t513 | ancestor — 착지 |
| `d570eb34a` | t514 (tip) | ancestor — 착지 |
| `dd1439502` | t510 | ancestor — 착지 |
| `436e83529` | t515 | **NOT ancestor — 미착지** (#1690 보류 유지 근거) |
| `c82c47f66` | t401 | **NOT ancestor — 미착지** (#1683 보류 유지 근거) |

리드 기준점 `9ce792637`은 본 트리의 조상으로 확인. 그 이후 develop은 두 번 움직였다: `a849d99d2`(t533 착지) → `f3517996a`(t517 착지). 이번 스윕 중에도 develop이 움직였으므로, **회신 게시 시점에 각 이슈의 착지 상태를 다시 확인할 것**.

## Evidence — 증상-착지 대응 (이슈 본문이 말하는 증상 ↔ 착지 트리 읽기)

| 이슈 | 증상 (이슈 본문) | 착지 트리에서 확인한 것 |
|---|---|---|
| #1658 (jjjh7401) | `rm\s+-rf\s+/` 정규식 양방향 결함 — 플래그 순서 우회 + 전 절대경로 오탐 + 따옴표 미접기 | `internal/hook/dangerous_removal.go:9` 헤더가 직접 기술: "Structural guard … **replacing the six `rm -rf <target>` regexes**" — `tokenizeSegment`( 인용 감지 포함, :88-90) 구조적 판정으로 전환 확인. t511 stat: 배포·로컬·픽스처 `security.yaml` extras 잔여 사본 3본 `-1` 씩 |
| #1686 (hansooha) | 내장 패턴 `rm\s+-rf\s+/` 앵커 부재 — 전 절대경로 차단, 표기 변경 통과 | 위 extras 3본 제거 stat과 정확히 일치. #1658과 같은 커밋에서 함께 정리됨 |
| #1659 (jjjh7401) | 따옴표 heredoc 본문 중괄호만 거절 (명령 치환은 통과) | **오늘 생관측으로 재현 (아래 절)**. t512 병합 stat: `.go` 0줄 — 문서(`worktree-integration.md` +29)·재현 원장·회신·업스트림 초안만 착지. 증상 자체는 CC 바이너리 소관(소스 0건 — t512 원장 grep 원거리) |
| #1678 (GoosLab) · #1687 (mg10009) | codex_task 즉시 실패(~300ms)를 `10m0s timeout` 으로 오보고 | t514 커밋 메시지 + 판정문 E1: 두 결함 분리 — (A) 파생 컨텍스트 Done 감시 → 자기 바운드 무조건 명명(호출자 취소 턴 **50.65ms** 만에 끝났는데 `10m0s` 만료로 보고 = 바운드의 약 1/11,800), (B) background job 이 요청 스코프 컨텍스트 상속 → 핸들러 반환 즉시 사망(관측 **0.01s**). 수리: 부모 컨텍스트 판별 + `context.WithoutCancel`. 2방향 회귀 가드 |
| #1692 (michaelleone) | `spec status --sync-git` 가 `--dry-run` 무시하고 spec.md 기록 | t513 판정문: dry-run 인쇄 후 **트리 바이트 동일**(픽스처 해시 3건 pre/post 일치, `git status` 비움), 실실행은 프론트매터 갱신 유지. 신규 바이너리로 orchestrator 독립 재측정까지 완료 |
| #1693 (michaelleone) | status locator 가 본문 `Status`/`Notes` 표 셀(백틱 산문 포함)을 상태로 판독 — 거짓 drift | t513 판정문: 상태 판독을 YAML 프론트매터에 앵커, 본문 셀(헤더행·백틱 산문 변형 모두) 미판독을 **mutation 검증**. `Summary: 2/3 … drift` 는 설계상 참양성만 잔존 |
| #1694 (binsworld) | `.moai/state/` 캐시 디렉터리가 임의 하위 디렉터리에 생성 | `internal/stateanchor/`(단일 시접, 고정 우선순위 체인) develop 착지 확인(`git ls-tree`). 회신 초안 `gh-1694-reply-draft.md` 도 develop 착지 — sync 확정본(session-memo 귀속 정정 포함). **게시 대기(리드 게이트)** |
| #1682 (bjw202) | 세션 간 메시징 두절 가설 — `tengu_harbor_kite` 공유 슬롯 (#1674 후속) | develop 문서에 이미 수용: `cross-session-messaging.md` 1적중 + `cross-session-messaging-detail.md` 2적중("The shared flag slot" — 진단·기전·수동 탈출구). **문서-응답 후보**: 가설 채택 사실과 문서 포인터를 회신으로 전달하는 형태. 문서 내용의 기술적 재검증은 본 스윕 범위 밖 |

## Evidence — #1659 오늘의 생관측 (2026-09-08)

명령 (그대로 발사):

```
cat > /tmp/t546-heredoc-probe.txt <<'EOF'
{"key": "value"}
EOF
```

관측: **거절** — 명령 실행 안 됨. 문구(verbatim, t512 원장과 일치 — 경로만 다름):

```
This session is isolated in the worktree /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t546, but this command is too complex to verify that it stays inside the worktree. Refusing to run it — ...
```

부가 관측: 대상이 워크트리 밖 `/tmp`여도 거절 — 대상 경로와 무관함이 재확인됐다. 판정 환경: Claude Code **2.1.263** (2026-09-08, 이 세션). 이로써 관측축은 3점: 2.1.238(제보자, 2026-08-24 환경 정보) → 2026-09-07(t512 재현 세션, 버전 미기록) → **2.1.263(2026-09-08, 본 세션)**.

## Classification

### A. 해소 확정 — 회신 게시(리드 승인) → 리포터 확인 → 종결

| 이슈 | 카드 | 회신 초안 (`.moai/reports/t546/`) |
|---|---|---|
| #1658 | t511 | `reply-1658.md` (t511 초안 정식화) |
| #1686 | t511 | `reply-1686.md` (t511 초안 정식화) |
| #1659 | t512 | `reply-1659.md` (t512 초안 + 3.2.0 약속 정합화 + 오늘 관측 반영) |
| #1678 | t514 | `reply-1678-1687.md` §1 (한국어 — 자체 feedback 이슈) |
| #1687 | t514 | `reply-1678-1687.md` §2 (영문 — 외부 제보) |
| #1692 | t513 | `reply-1692-1693.md` §1 (영문 — **재구성**, 아래 Gaps 참조) |
| #1693 | t513 | `reply-1692-1693.md` §2 (영문 — 재구성) |
| #1694 | t510 | 기존 확정본 사용: develop `.moai/reports/t510/gh-1694-reply-draft.md` (재작성 없음) |

### B. 업스트림 제보 — 별도 승인 대상

`upstream-heredoc-brace.md` (영문, #1659 발췌 고정 — 리드 지시대로 「이 표는 특정 버전 관측이다」 문면 못박음). 대상: Claude Code upstream.

### C. 보류 — 닫지 않음

- **#1690** (mihaesinbi) — t515 `436e83529` 미착지 재확인. 창 대기 중.
- **#1683** (Seung-zedd) — t401 `c82c47f66` 미착지. run 진행 중.

### D. 카드 없음 — 발행 문안 (운영자 승인 대기, 직접 발행 안 함)

| 이슈 | 제목 (요지) | 발행 문안 (카드 설명 후보) |
|---|---|---|
| #1631 | moai gate, non-eslint JS 프로젝트에서 lint 전면 생략 (biome/oxlint) | gate lint 탐지 확장 — biome·oxlint 지원 (GH #1631) |
| #1632 | Codex audit gate: 구조화 출력 비어 gate 미집행 | codex_audit 빈 출력 시 fail-closed 처리 (GH #1632) |
| #1639 | gate 통과 시 0바이트 출력 + 타임아웃 미집행(60s 설정→915s 생존) + 수동 gate 미직렬화 | gate 가시성·타임아웃·직렬화 3건 묶음 (GH #1639) |
| #1640 | mcp-server `project_root` 폴백이 스테일 MOAI_PROJECT_DIR 고정 — 문서화된 파라미터 미구현 | mcp-server project_root 파라미터 실구현 (GH #1640) |
| #1641 | pre-commit go vet 이 루트 go.mod 기준 — 하위 모듈 모노레포 전면 차단 | pre-commit vet 모듈 해상도 수리 (GH #1641) |
| #1654 | `moai todo` 미등록 동사+id 가 add 로 유출 — 정크 카드 생성 | todo 동사 파싱 강화 (GH #1654) |
| #1660 | goal_arm MCP 래퍼가 산문 조건을 mechanical 통째 저장 — 매 턴엔드 차단 | goal_arm 조건 분류 수리 (GH #1660) |
| #1661 | graph-freshness 가 origin/main HEAD 에서 실패 — codemaps 스탬프 객체 부재 (#1648 착지 결함) | graph 스탬프 객체 부재 수리 (GH #1661) |
| #1675 | 폴백 프로필에서 statusline 이 프로젝트 statusline.yaml 무시 — gh 폴링 부활 | statusline 프로필 폴백 수리 (GH #1675) |
| #1679 | pre-commit fast-subset vet, 모듈이 하위에 있으면 매 커밋 실패 (다중 루트) | #1641과 동일 패밀리 — 묶음 검토 (GH #1679) |
| #1680 | Heavy gate 언어 마커를 프로젝트 최상단에서만 탐지 — 중첩 모듈 0 커버리지 | gate 언어 탐지 재귀화 (GH #1680) |
| #1691 | moai-gate 실패가 core.bare 뒤집음 + gate-child pytest 가 호스트 리포에 fixture 커밋 | gate 격리(작업 트리 보호) 수리 (GH #1691) |
| #1696 | spec lint CoverageIncomplete 오탐 — spec.md 전용 스캔이 acceptance.md 의 AC 누락 | spec lint 스캔 범위 확장 (GH #1696) |

특기: #1679는 #1641과 동일 패밀리(루트 go.mod 기준 vet)로 보임 — 묶음 카드 후보. #1661은 #1648 착지 결함 명시.

## Gaps (명시적으로 관측하지 않은 것)

1. **t513 `issue-reply-draft.md` 유실** — 판정문이 참조하나 develop `.moai/reports/t513/`은 6파일뿐(초안 부재). t513 워크트리는 이미 폐기돼 원본 회수 불가. 본 카드가 판정문 재료로 재구성했다(`reply-1692-1693.md`, 영문). 원본 문면과 톤이 다를 수 있다.
2. **t514 회신 초안 원본 부재** — develop에 verdict.md만 착지. 본 카드가 판정문 E1·커밋 메시지 재료로 재구성했다(`reply-1678-1687.md`).
3. t512 재현 세션(2026-09-07)의 CC 버전 미기록 — 업스트림 초안에서 "not recorded"로 정직 표기했다.
4. 카드 없음 13건은 제목 수준 분류다 — 본문 정독과 원인 규명은 각 카드 발행 후 그 카드의 몫.
5. #1682 문서의 기술적 내용(플래그 슬롯 기전)은 재검증하지 않았다 — 문서 존재와 적중만 측정.

## Residual-risk

- 회신 초안은 본문의 사실 주장을 착지 트리에서 검증했으나, **게시 전 리드의 최종 문면 낭독이 필요하다** (외부 공개 행위 — 리드 지시 [HARD]).
- develop이 스윕 중에도 움직였다(2회). 회신 게시 시점에 착지 재확인 권장 — 특히 #1690(t515)이 창을 받으면 분류가 D→A로 바뀐다.
- #1692/#1693·#1678/#1687 초안은 원본 없는 재구성이다 — 리드 낭독 시 제보자 호칭·톤을 확인할 것.

## 창 요청 (follows)

카드 t546 · `WT-issue-sweep` · 병합 창은 리드 지명 대기. 병합 전 sync는 본 워크트리 안에서 끝낼 것(착지 전 sync 미완료 방지).
