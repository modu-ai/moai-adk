# t214 — PreToolUse 동기 차단 비용 (Class B)

카드: t214 · 브랜치 `WT-pretool-blocking` · base `origin/main cd0cee1b8`
근거 감사: `.moai/reports/hook-audit/tool-path.md` (per-tool-call 축)

## 1. Claim

F-1과 F-3을 닫았습니다. **F-2b·F-4는 이 PR에 넣지 않았습니다** — 이유는 §4에 명시.

| 항목 | 상태 | 실체 |
|---|---|---|
| **F-1** 커밋 게이트 30초 동기 차단 | 닫음 | PreToolUse에서 게이트 제거; git `pre-commit` 훅이 원래 자리 |
| **F-3** `status-transition-ownership.sh` 동기 | 닫음(전제 정정 후) | 템플릿 3개 엔트리에 `async: true` |
| **F-2b** security-scan → post-tool 병합 | 미착수 | 보안 스캔 표면 변경 — 별도 카드 |
| **F-4** PreToolUse ast-grep 스캔 제거 | 미착수 | 위와 같은 이유 |

## 2. Evidence

### 2-A. F-1 — 게이트는 훅의 예산 밖으로 나갔습니다

배차문의 실측(30,033 / 30,016 / 30,020 ms → deny)을 이 트리에서 재현했습니다. 다만
**30초짜리 fixture를 쓰지 않았습니다** — 게이트가 도는지 여부가 쟁점이지, 얼마나 오래 도는지가
쟁점이 아니기 때문입니다. `go vet`이 반드시 잡는 printf 불일치 파일 하나만 있는 tiny fixture로
같은 분기를 때립니다.

```
# RED (수정 전)
--- FAIL: TestPreTool_GitCommitDoesNotRunTheQualityGate (0.15s)
    PreToolUse denied a plain `git commit` against a vet-failing fixture …
    reason: quality gate failed: go vet / bad.go:6:29: fmt.Printf format %d has arg …

# GREEN (수정 후)
--- PASS: TestPreTool_GitCommitDoesNotRunTheQualityGate (0.00s)
```

`0.15s → 0.00s`. 게이트가 아예 돌지 않습니다.

**조치**: 배차문의 1안(게이트를 PreToolUse에서 빼기)을 택했습니다. 2안(8초 예산 + allow+warning)을
버린 이유는 실측에 있습니다 — 감사가 tiny 모듈 1.1s / 이 저장소 30s를 쟀으므로, 8초 예산은
**실제 저장소에서 100% 타임아웃**합니다. 즉 2안은 매 커밋마다 8초를 내고 아무 신호도 못 받는
구조입니다. 차단 30초를 차단 8초로 줄이는 게 아니라, 쓸모없는 8초를 영구화합니다.

**게이트를 잃지 않았습니다.** `moai init`/`moai update`가 까는 git `pre-commit` 훅이
`moai gate`를 셸아웃합니다(`internal/cli/hook_install_precommit.go:89-96`). 같은 16-언어 게이트,
10초 천장 없음, 실패 시 tool-permission deny가 아니라 커밋 clean abort,
`SKIP_MOAI_PRECOMMIT=1` 우회 문서화. 코드 주석이 이미 이것을 "relocated heavy quality gate"라
부르고 있었고 — PreToolUse 사본이 **이전을 안 마친 절반**이었습니다.

**낙진 1건.** `TestPreTool_AstGrepSkipReasonSurfaces`는 ast-grep 스킵 사유가
`RunAstGrepGateV2 → QualityGate.Run → preToolHandler.Handle` 세 프레임을 건너오는지 보던
테스트입니다. 세 번째 프레임이 사라졌습니다. **삭제하지 않고 살아있는 두 프레임으로 재조준**했습니다
(`TestGateRun_AstGrepSkipReasonSurfaces`) — 전파 자체는 그대로 살아 있고, 소비자가
`moai gate`(`runGate`가 pass 경로 output을 stderr로 흘림)로 바뀌었을 뿐입니다. 가드를 잃는 삭제는
하지 않았습니다.

**의도적으로 남긴 것.** `loadGateConfig`와 `firstLine`은 이제 프로덕션 호출자가 없고 테스트만
부릅니다. 지우려다 멈췄습니다 — `pre_tool_gate_disabled_steps_test.go`가 이슈 #1265
(`disabled_steps`를 FALSE로 읽는 매핑 결함)의 **유일한 회귀 가드**이고, CLI 쪽 미러
`mapConfigGateToQuality`에는 대응 테스트가 없습니다(`grep DisabledSteps internal/cli/*_test.go` → 0건).
지우면 실재했던 결함의 가드가 사라집니다. 가드를 CLI 미러로 옮긴 뒤 삭제하는 것은 후속 카드 감.

### 2-B. F-3 — 전제 정정 후 조치

배차문은 "47.5ms × 359회 ≈ 17초 차단"이라 했고, 그 숫자는 참입니다 — 다만 **로컬
`.claude/settings.json`에 대해서만** 참입니다. 배포되는 템플릿은 이미 다릅니다:

| | 로컬 `.claude/settings.json` | 템플릿 `settings.json.tmpl` |
|---|---|---|
| 엔트리 수 | 1 | 3 (Write/Edit/MultiEdit 분리) |
| 경로 스코프 | **없음** — 모든 Write/Edit에 발화 | `"if": "<Tool>(**/.moai/specs/**)"` |
| `async` | 없음 | 없음 |

즉 **갓 업데이트한 프로젝트에서는 359회가 아니라 SPEC 아티팩트 쓰기 횟수만큼만** 발화합니다.
17초는 드리프트된 로컬 파일을 잰 값입니다. 로컬 파일은 `moai update`가 재생성하는 산출물이라
직접 고칠 대상이 아닙니다.

남은 실체는 `async` 부재 하나이고, 그건 진짜입니다. 템플릿 3개 엔트리 전부에 `async: true`를
붙였습니다. 근거는 스크립트 자신의 주석 — advisory이고 `exit 2` 차단은 "reserved for future
enforcement"이며 stdout을 읽는 곳이 없습니다.

```
$ make build                                    → catalog.yaml updated, go build OK
$ go test ./internal/template/... -count=1      → ok 33.430s / ok 1.161s
```

### 2-C. 회귀·빌드

```
go test ./internal/hook/ -count=1        → ok 57.168s   (수트 전체)
go test ./internal/template/... -count=1 → ok 33.430s, ok 1.161s
go build ./internal/...                  → 무출력
make build                               → 성공 (템플릿 임베드 재컴파일)
```

## 3. Baseline-attribution

- 트리: `WT-pretool-blocking`, base `origin/main cd0cee1b8`.
- **t213 브랜치를 이 워크트리에 병합했습니다** (`WT-gotest-tmp-leak`, PR #1624). 선택이 아니라
  전제였습니다: t213 이전의 `runStep`은 `cmd.Dir`를 안 걸어서, 이 카드의 재현 테스트가
  게이트를 돌리는 순간 **`internal/hook` 수트를 재귀 실행**합니다. 재현을 안전하게 돌리려면
  그 수정이 트리에 있어야 합니다. → **PR #1624가 이 PR의 선행 조건입니다.**
- 감사 절대값은 인용하되 근거로 쓰지 않았습니다(load 8~10에서 측정돼 부풀려짐, 감사 자신이 명시).
  판정 근거는 이 트리에서 다시 잰 RED/GREEN과 수트 결과입니다.

## 4. Gaps (미검증)

- **F-2b·F-4 미착수.** 둘 다 "어떤 보안 스캔이 언제 도는가"를 바꿉니다. F-4는 쓰기 **전** 차단
  능력을 없애는 것이고, F-2b는 PostToolUse 스캔 3종 중 하나를 접는 것입니다. 지연시간 PR에
  묶으면 리뷰어가 보안 표면 변경을 성능 개선으로 읽고 넘길 위험이 있습니다 — 별도 카드에서
  단독 리뷰받아야 할 성질의 변경이라 판단했습니다. **범위 축소는 리드 판정 사항이므로 보고합니다.**
- **`internal/cli` 전체 수트 로컬 미실행** (load 29). CI 몫.
- **훅 지연시간 재측정 안 함.** F-1의 효과는 "게이트가 돌지 않음"으로 단언했지, "훅이 N ms
  빨라짐"으로 단언하지 않았습니다. 감사 조건(swap 고갈, load 8~10)을 재현할 수 없으므로
  절대값 비교는 무의미합니다.
- **`async: true`의 런타임 효과 미관측.** Claude Code가 그 키를 읽어 비동기로 돌리는 것은
  런타임 동작이고, 이 트리에서 훅을 실제로 발화시켜 확인하지 않았습니다. 근거는 같은 파일의
  다른 엔트리들이 이미 같은 키를 쓰고 있다는 것뿐입니다.

## 5. Residual-risk

- **SPEC-STOPCHAIN-TRIM-001 REQ-005의 기계적 실체가 사라졌습니다.** 그 REQ는 커밋 게이트를
  tier-aware로 만든 것이고(`automatic`/`fully-autonomous`에서 OFF), 이제 그 게이트가 없습니다.
  REQ의 **의도**("semi-auto에서 커밋은 사전 검증된다")는 pre-commit 훅이 무조건 수행하므로
  살아 있지만, **tier로 그것을 끄는 능력은 없어졌습니다.** 무인 실행에서 커밋마다 pre-commit
  게이트를 무는 것이 문제라면 `SKIP_MOAI_PRECOMMIT=1`이 우회이고, 이건 tier와 무관한 별개 축입니다.
  SPEC 문서 갱신은 하지 않았습니다 — 이 판단이 맞는지는 리드/운영자 확인이 필요합니다.
- **pre-commit 훅이 안 깔린 사용자.** 설치자는 마커 없는 기존 훅이 있으면 덮어쓰지 않고
  `ErrUserHookExists`를 냅니다(`hook_install_precommit.go`). 그런 사용자는 이제 커밋 게이트가
  전혀 없습니다. 다만 **종전에도 실질 보호는 없었습니다** — 30초 게이트는 10초에 죽었고 그
  deny는 근거 없는 deny였습니다. 없던 보호가 사라진 것이지 있던 보호가 사라진 게 아닙니다.
- **`sourceExts` 없는 언어 스텝.** t213에서 Go 스텝에만 붙였습니다. 다른 15개 툴체인에서
  "마커는 있는데 소스는 없는" 스캐폴드가 같은 형태로 실패할 수 있습니다(미측정).
- **F-3의 로컬/템플릿 드리프트.** 로컬 `.claude/settings.json`은 이 PR로 바뀌지 않습니다.
  `moai update`가 돌아 템플릿판이 배포되기 전까지, 이 저장소 자신은 여전히 스코프 없는
  동기 엔트리를 씁니다.
