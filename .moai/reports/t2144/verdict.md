auditor-model: claude-opus-5-5

# t2144 sync 최종 판정 (수리 1회차 델타, Codex 필수 게이트 충족)

verdict: PASS-WITH-DEBT
audited_sha: b73989497adbfeb6c8bfdcc6929ea2212349b088
score: 0.91
audit_receipt: rcpt-185288dcaa28642f8c9fd53b

- 카드: t2144 (Class B, SPEC 없음), 브랜치 `WT-merge-readiness-resolves`, 작업 트리 HEAD `ee633e411d` (감사 코드 끝 `b73989497a` 위에 `.moai/reports/t2144/` 증거 파일만 더한 커밋), 로컬 main `f9d8866951` 이 조상
- 감사 범위: 수리 델타 `git diff c89eac461c89e9bbb017a7d1bb44fdccd146196e..b73989497adbfeb6c8bfdcc6929ea2212349b088 -- internal/` (merge.go 해석기 교체 + 새 시험 파일 1개)
- 감사일: 2026-11-07, 감사자: sync-auditor (claude-opus-5-5)
- 판정 요지: 차단 결함 없음. 2회차의 유일한 FAIL 사유였던 필수 Codex 게이트가 이번 실행에서 충족됐다(Codex `pass`, 결함 0건, 영수증 `rcpt-185288dcaa28642f8c9fd53b`). N1(선택)은 부채로 넘긴다.

## 반복 이력

| 회차 | 감사 SHA | 판정 | 점수 | 차단 사유 |
|---|---|---|---|---|
| 1 | c89eac461c | FAIL | 0.79 | F1 — 태그 우선 대체 경로가 `origin/main` 태그로 원격 추적 ref 를 가림 (`sync-audit-iter1.md`) |
| 2 | b73989497a | FAIL | 0.90 | 코드 차단 결함 없음. 필수 Codex 게이트 미충족(`stale_server`, 영수증 없음) (`sync-audit-iter2.md`) |
| 3 | b73989497a | PASS-WITH-DEBT | 0.91 | 없음. Codex 게이트 충족, N1 부채 |

## 차원 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 92/100 | PASS | 이번 실행 `go test ./internal/factorylane/ -run T2144 -count=1` → `ok  	github.com/modu-ai/moai-adk/internal/factorylane	61.661s`, exit 0. 2회차가 같은 코드 끝에서 패키지 전체·cli 4패턴 통과와 P3(태그 `origin/main` 이 원격 추적 ref 를 못 가림) 재현을 기록 |
| Security (25%) | 88/100 | PASS | 해석기가 `show-ref --verify --hash` 정확 경로만 받아 태그·리비전 식 경로가 사라짐(델타 직독). 2회차 P1 옵션 주입 5종 거절·부작용 없음. Codex 정적 검토 결함 0건. N1 은 fail-closed 로 끝나 Low |
| Craft (20%) | 92/100 | PASS | 이번 실행 `go vet ./internal/factorylane/` → `vet_exit=0`. 2회차 변이 7/7 사망(같은 코드 끝) |
| Consistency (15%) | 90/100 | PASS | `branchTipSHA` 와 같은 `show-ref --verify --hash` 모양, 거절 문구가 후보 형식을 `strings.Join(candidates, " or ")` 로 모두 댐, 주석이 F2 동작을 의도로 명시 |

가중 점수 0.368 + 0.220 + 0.184 + 0.135 = 0.907 → 0.91. must-pass(Functionality, Security) 둘 다 PASS.

## Codex 결과 (요약 원문)

`mcp__moai__codex_audit` (mode native, target baseBranch, delta_base `c89eac461c89e9bbb017a7d1bb44fdccd146196e`, project_root `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2144`) 1회 호출:

```
codex_result: pass
summary: Verdict: pass
  구스 행님, 지정한 기준 커밋 대비 변경에서 수정할 결함은 찾지 못했네. 정적 검토만 했으며, 지침대로 테스트·빌드·재현 명령은 실행하지 않았네.
findings: []
next_steps: []
provenance.backend: codex / transport: codex app-server / resolved_model: gpt-6.1-sol / requested_effort: high
review_base: delta_base (merge base c89eac461c89e9bbb017a7d1bb44fdccd146196e)
codex_receipt: rcpt-185288dcaa28642f8c9fd53b
build_commit: f9d8866951
```

Codex 가 블로킹 결함을 보고하지 않았으므로 FAIL 사유가 아니다.

## 결함 처리 상태

| ID | 상태 |
|---|---|
| F1 [High, blocking] 태그 우선 대체 경로 | 닫힘 — 델타에서 `rev-parse --verify --quiet name^{commit}` 대체 경로 삭제 확인, 후보는 `refs/heads/<name>`, `refs/remotes/<name>` (또는 `refs/` 이름 그대로)뿐 |
| F2 [Low] 로컬 브랜치 `origin/main` 우선 | 의도된 동작, 시험 `f2_local_branch_named_origin_main_wins_like_git` 이 고정 |
| F3 [Low] 접미사 이름 | 닫힘 — 시험 `c_suffixed_names_fail_closed` (`main~0`, `main^`, `main@{0}`, `--all`) |
| F4 [Low, outside] `Develop=HEAD` | 부수적으로 닫힘(2회차 P6) |
| F5 [Low, outside] 안내 문구 맨이름 | 범위 밖, 기록만 |

## 부채 (넘김)

- **N1 [Low] [optional] [in-diff]** `internal/factorylane/merge.go` `integrationRevSHA` — `refs/` 이름은 `show-ref --verify --hash` 로 존재만 확인하고 커밋인지는 확인하지 않는다. 주석·거절 문구는 "resolved to a commit" 이라 말한다. 주석 태그는 태그 객체 SHA 를 돌려주고(merge-tree 가 벗겨 써서 결과는 정상), 트리를 가리키는 태그는 `ok=true` 뒤 "unresolved conflicts" 로 잘못된 사유를 낸다(fail-closed 는 유지). 도달 경로는 운영자의 `--develop refs/tags/...` 뿐.
  - 후속 조치: 해석된 SHA 에 `rev-parse --verify --quiet <sha>^{commit}` 을 한 번 더 적용하거나 주석을 "ref 가 가리키는 객체" 로 고칠 것.
- N2 [Info] 맨 SHA·`HEAD`·접미사 이름이 더는 해석되지 않는 동작 변화 — 실제 호출자 영향 없음(2회차 호출자 표).

## 실행한 명령과 관측 출력

### Claim / Evidence

1. `df -h /private/tmp` (시작) → `/dev/disk3s5   3.6Ti   3.6Ti    24Gi   100% …`; (종료) → `… 23Gi   100% …`
2. `moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2144` → exit 0, `"config_status": "ok"`, `"model": "codex"`, 게이트 `off` / `required` / `off`, `"cross_model_active": true`, `"cross_model_required": true`, `"enforced_required": [` (2회차 기록상 codex 만 required)
3. `mcp__moai__codex_audit` 1회 → 위 Codex 결과. 영수증 `rcpt-185288dcaa28642f8c9fd53b`
4. `go test ./internal/factorylane/ -run T2144 -count=1` → `ok  	github.com/modu-ai/moai-adk/internal/factorylane	61.661s`, `exit=0`
5. `go vet ./internal/factorylane/` → 출력 없음, `vet_exit=0`
6. `git diff c89eac461c…..b73989497a… -- internal/` → `merge.go` 해석기 본문 교체(+16/-12 남짓)와 `merge_tag_shadow_repair_t2144_test.go` 신규 121줄. 직독 내용은 위 결함 처리 표에 반영
7. `git status --short` (종료) → 출력 없음 — 추적 파일 무변경(이 파일은 미추적)

2회차(같은 코드 끝 `b73989497a`)에서 이미 실행한 것: factorylane 전체, cli `MergeReady`/`SpecLess`/`FactoryMerge`/`MergeReadinessBeforePR`, `gofmt -l`, 적대 프로브 P1~P9, 변이 7종 — 모두 `sync-audit-iter2.md` 에 출력 원문이 있다. 이번에는 반복하지 않았다.

### Baseline-attribution

이번 실행의 `go test`/`go vet` 은 작업 트리 HEAD `ee633e411d` 에서 돌았다. `ee633e411d` 와 `b73989497a` 의 차이는 `.moai/reports/t2144/` 아래 증거 파일뿐이라 `internal/` 코드는 감사 SHA 와 같다. `go test` 가 트리에서 직접 빌드했으므로 도구 출처(§2.2)는 트리 자체다. Codex 검토는 diff 를 읽는 모델 검토이고, MCP 서버 빌드(`f9d8866951`, HEAD 의 조상 — 결과의 `build_lag` 경고)는 전달 경로일 뿐 코드를 판정하지 않는다.

### Gaps

- CI 미검증: 원격 CI 결과를 읽지 않았다.
- Windows 미검증: 측정은 모두 darwin 에서만 했다.
- `internal/cli` 전체 스위트와 `internal/integration` 은 돌리지 않았다(2회차의 cli 4패턴만 있음).
- 감사 계획은 `cross_model_active: true` 로 `audit_multi` 경로를 가리키지만, 리더 지시에 따라 `audit_multi` 를 부르지 않고 유일한 required 게이트인 Codex 만 `codex_audit` 로 실행했다. claude/glm 게이트는 `off`. 오케스트레이터의 `--result-file` 대조 검사는 `audit_multi` 결과가 없어 해당 없음.
- MCP 서버 빌드 `f9d8866951` 이 트리 HEAD 의 조상이라는 `build_lag` 경고가 있었다. 영수증 발급·전달 경로에 그 사이 수정이 빠져 있을 가능성은 관측하지 않았다.
- Codex 는 정적 검토만 했다고 밝혔다(실행 없음).

### Residual-risk

- N1: `refs/...` 정규형으로 비커밋 객체를 가리키면 거절은 되지만 사유가 "conflict" 로 잘못 표시될 수 있다.
- 로컬 브랜치 `origin/main`(F2)이 생기면 PR 경로가 그 브랜치를 잰다. git 과 같은 규칙이고 시험으로 고정됐다.
- 2회차의 프로브·변이 원본은 머신 로컬 스크래치이며, 그 판정 파일에 옮긴 줄만 증거로 남는다.
