# 카드 t293 판정 — SPEC-STATUSLINE-PROFILE-RESPECT-001

> Evidence-bearing report (AGENTS.md §1 5-섹션 형식). 작성: factory lane-2, 2026-08-27.
> 브랜치 `WT-statusline-profile` · HEAD `0a92b22f9` · 기점 `3abde7053`(=origin/main 팁) · 병합 보류(리드 지시).

## Claim (주장)

이슈 modu-ai/moai-adk#1675의 수용기준(AC-001..008, AC-010..011)이 본 워크트리에서 구현·검증되었고,
AC-004의 양방향 회귀(opty-out→억제 / none→github 복귀)가 테스트로 고정되었다. AC-009는 킥오프 게이트
결정(D1)에 따라 DEFERRED다. M7 운영 편집(`forge: none`)이 본 저장소 로컬 설정에 적용되었다.

> **[정정 — sync 감사 F2, 2026-08-27]** 이 문단은 원래 "코드 병합 전이라도 이 레포 세션의 gh
> 폴링이 멈춘다"로 끝났다. 실측으로 반증됐다. 워크트리 세션의 `boardRoot`는
> `worktree.original_cwd`, 즉 **primary 체크아웃**이고, primary의
> `.moai/config/sections/statusline.yaml`에는 `forge` 키가 아직 없다(33줄, 병합 전 원본 —
> `grep -n 'forge' <primary>/.moai/config/sections/statusline.yaml` → rc 1). 이 워크트리의 같은
> 파일 12행만 `forge: "none"`이다. 따라서 옵트아웃의 운영 효과는 **병합분이 primary 워킹트리에
> 도달한 뒤에** 발효된다. 코드 결함이 아니라 배포 상태다. 근거:
> `.moai/reports/t293/sync-audit.md` §2.6.

## Evidence (증거 — 명령과 실측 출력)

manager-develop 산출(전부 `.moai/state/verify/t293-dev/`):
- RED: `red-statusline.txt` exit 1 — `spawn attempts = 1, want 0`; `red-profile.txt` exit 1 — subtree resolved "" want alpha
- GREEN: `ac-matrix.txt` — 12 top-level PASS / 0 FAIL (`go test … -count=1 -v -run '…'`)
- 커밋 4건: `62485c918` M1 / `5a193fa4c` M2+M3 / `d615bf374` M4 / `0a92b22f9` M7 — 전부 메시지에 `(t293)` 각인

lane-2 독립 재검증 배치(본 파일 작성 직전, 동일 HEAD에서 관측):
| # | 항목 | 명령 | 관측 |
|---|---|---|---|
| V1 | git 상태 | `git log --oneline -5` 등 | 4커밋 존재, 트리 청결(??는 본 리포트 디렉터리만) |
| V2 | 대상 테스트 | `go test ./internal/statusline/ ./internal/profile/ -count=1` | `ok … 13.722s` / `ok … 0.451s`, exit 0 |
| V3 | 억제 게이트 | `grep -n forgeOverride github.go` | :91 config 읽기로 Suppressed=true |
| V4 | 세그먼트 게이트 | `grep builder.go` | :266 `isSegmentEnabled(SegmentGitHub)` |
| V5 | 조상-walk 배선 | `grep profile.go` | :473 miss-path else-branch에서 `lookupSubtreeProjectKey` 호출(:387 정의) |
| V6 | seam 배치(D3) | `grep githubSpawnProbe` | :126 정의 — TTL 뒤 probe·isSelfInvocable 앞 계약 준수 |
| V7 | M7 read-back | `grep '^  forge:' …statusline.yaml` | `12:  forge: "none"` |
| V8 | 템플릿 무결 | `git diff --stat 3abde7053..HEAD -- internal/template/` | 빈 출력 |
| V9 | vet · 윈도 빌드 | `go vet …` / `GOOS=windows go build …` | 둘 다 OK |

## Baseline-attribution (baseline 귀속)

위 모든 관측은 **본 실행(bare turn)에서, 본 워크트리(`.claude/worktrees/t293`), HEAD `0a92b22f9`** 에 대해
수행한 결과다. 이전 SPEC이나 다른 트리에서 옮겨온 수치는 하나도 없다. 감사 체계: plan-auditor iter-2
**PASS 0.94**(반복 한도 2 도달, R1~R4 MINOR 잔여는 판정 비연계).

## Gaps (미검증 — 명시적으로 관측하지 않은 것)

- 실제 바이너리 통합 렌더(`moai statusline` 실행 후 상태줄 육안 관측)는 수행하지 않았다 — 동일 픽스처가
  유닛 경로를 커버하므로 코드 신뢰도에는 차등 없으나 "화면에서 안 보인다"는 1차 증거는 아니다.
- `internal/profile` 패키지 커버리지 84.3%(<85%) — SPEC 범위 밖 기존 미커버 함수(`Delete`, `GetCurrentName` 경로) 소관.
- 리터럴 `__no_such_profile__` 명명의 발원지 — 레포·깃 이력·바이너리·로컬 설정·tmux 전역환경 전무 확인.
  상위 프로세스 체인 상속 추정, 본 수정으로 의존 자체가 제거됨(SPEC spec.md §F accepted Gap).
- mo.ai.kr 쪽 statusline.yaml의 이중 블록(43~57행 중복 헤더)은 다른 프로젝트라 손대지 않았다.

## Residual-risk (잔여 위험)

- **M7 키 휘발**: `moai update` 시 `.moai/config` 와이프로 `forge: none`이 사라질 수 있다(§2.3) — 인파일
  경고 주석으로 방어했고 근본 해결은 wipe-family 카드(CleanMoaiManagedPaths 보호 목록) 소관이다.
- **원장 성장**: 워크트리마다 projects[] 행이 계속 쌓인다 — REQ-009 후속 카드("launch-ledger 쓰기 정규화")
  등록이 sync 닫기 전 필요(plan-auditor R3 권고, 큐 생산은 리드 권한).
- t215·t211(#1621) statusline 계열과의 충돌 여부는 develop 통합 시점 판정 대기(변경면은 github.go·builder.go
  호출부·profile.go miss-path로 좁혀져 있음).
