# t699 판정서 — codex_skills_disable 미러 경로의 생산자 결속

- 카드: t699 (Tier S, Class B — plan 생략, run → sync)
- 브랜치: `WT-codex-skills-bind` (워크트리 `.claude/worktrees/t699`)
- 기점: develop 선두 `0c32a15b2` (배차 기점 `6d5943052` 의 자손 — t835 병합 4커밋 포함. 배차 후 리드가 develop 을 전진시켜 워크트리가 최신 develop 에서 파진 케이스)
- 변경: `internal/cli/codex_skills_disable.go` (4+/8−), 신규 `internal/cli/codex_skills_disable_mirror_drift_test.go`
- 미푸시 커밋 수와 병합 SHA는 완료 보고(리드 전달)가 정본

## Claim (주장)

1. `codex_skills_disable.go` 의 `mirrorSkillsRel` 리터럴(`filepath.Join(".agents","skills")`)은 생산자 `internal/template.MirrorSkillsRelDir` 와 분리된 재진술이며, disable 검문 경로 3개 사용 지점(미정지 사유 메시지 · 프로젝트 미러 경로 · 홈 미러 비교)을 모두 오염시킨다.
2. 수리는 세 사용 지점을 `template.MirrorSkillsRelDir` 직접 참조로 결속하고 로컬 var 를 삭제하는 것 — doctor(`doctor_codex.go`)가 t521 때 택한 같은 결속 선례를 따른다.
3. 변이(리터럴 재주입)를 잡는 two-run 가드가 새 테스트 파일에 있고, 가드는 결손 상태에서 FAIL, 결속 상태에서 PASS, 결속을 깬 변이 재주입에서 다시 FAIL 한다.

## Evidence (증거 — 이번 실행, 이 트리에서 관측한 출력)

**RED (수리 전, 가드만 있는 상태)** — `go test ./internal/cli/ -run 'TestResolveCodexSkillMirrorPathFollowsProducerMirrorRelDir|TestMirrorAbsentReasonNamesProducerRelDir' -count=1`:

```
--- FAIL: TestResolveCodexSkillMirrorPathFollowsProducerMirrorRelDir (0.00s)
    codex_skills_disable_mirror_drift_test.go:66: run 2 (producer repointed to ".agents/moved"): outcome = 1, want resolved — the gate did not follow the producer (reason: the project skill mirror /var/folders/.../002/.agents/skills does not exist, ...)
--- FAIL: TestMirrorAbsentReasonNamesProducerRelDir (0.00s)
    codex_skills_disable_mirror_drift_test.go:89: absent reason does not name the producer's mirror dir ".agents/moved": ... /.agents/skills does not exist, ...
FAIL
```

재지정된 생산자(`.agents/moved`)가 아니라 옛 리터럴(`.agents/skills`)을 계속 보는 결함이 그대로 재현됨.

**GREEN (결속 적용 후)** — 동일 선택자: `ok github.com/modu-ai/moai-adk/internal/cli 0.991s` (2/2 PASS; `-v` 로 `--- PASS` 2행 확인).

**변이 probe (결속 상태에서 리터럴 재주입)** — `projMirror` 한 줄을 리터럴로 되돌린 뒤 동일 가드 실행:

```
--- FAIL: TestResolveCodexSkillMirrorPathFollowsProducerMirrorRelDir (0.27s)
    ... run 2 (producer repointed to ".agents/moved"): outcome = 1, want resolved — the gate did not follow the producer ...
FAIL
```

가드가 변이를 잡음(공허통과 아님 — 측정으로 확인). 결속 복원 후 `ok ... 1.135s`, probe 흔적 grep 0.

**인접 회귀** — 선택자 `Codex|InspectSkillMirror|MirrorAbsent`, 함수 수 사전 대조:

```
선택자 일치 함수 수: 422   (grep 카운트)
422                        (-v 결과 행 수 — 누락 0)
ok  github.com/modu-ai/moai-adk/internal/cli  218.085s   (exit 0)
```

**정적 검사**: `gofmt -l` 두 파일 0행 · `go vet ./internal/cli/` 통과 · `grep -rn "mirrorSkillsRel" internal/cli/` 잔여 0 (`MirrorSkillsRelDir` 제외).

## Baseline-attribution (baseline 귀속)

모든 명령은 이번 턴에 워크트리 `.claude/worktrees/t699` (브랜치 `WT-codex-skills-bind`, HEAD `0c32a15b2` 기점)에서 실행했고 출력은 위에 그대로 옮겼다. 다른 트리·다른 시점 수치를 재사용하지 않았다.

## Gaps (미검증)

- `internal/cli` 패키지 전체 스위트(422 선택 밖 포함)와 전-저장소 스위트는 실행하지 않았다 — 레인 로컬 검증은 변경이 닿을 수 있는 면(Codex·skill-mirror 422건)으로 한정하며, 전체 판정은 develop push 이후 CI 몫이다(§4 레인 검증 규율).
- `GOOS=windows` 크로스 빌드는 돌리지 않았다. 변경은 순수 참조 치환(새 분기·플랫폼 분기 없음)이라 위험도 낮다고 판단했으나, 측정은 아니다. CI 매트릭스가 대신 본다.
- golangci-lint 전체 패키지 실행은 생략했다(시간). `gofmt`+`go vet` 만 통과시켰다.

## Residual-risk (잔여 위험)

- `template.MirrorSkillsRelDir` 는 var(의도된 테스트 봉합, `@MX:WARN` 부착)이므로 결속 후에도 disable 동사는 프로세스 안에서 생산자 재지정의 영향을 받는다. 이는 카드의 목적 자체(드리프트 추종)이며, 생산자 측 `@MX:REASON` 이 수복 규약(must restore before returning)을 이미 규정한다.
- 결속으로 disable 동사와 doctor 가 이제 같은 var 를 바라본다 — 한쪽 테스트의 재지정이 다른 쪽 테스트에 새는 가능성은 패키지의 기존 non-parallel 규약(`repointProducerMirrorRelDir` 호출자 금지)이 계속 지배한다. 신규 테스트 2건은 그 규약을 지켰다(`t.Parallel()` 없음, `t.Cleanup` 복원).

## 동기화 메모 (Class B sync)

- SPEC 부재(plan 생략) — 문서 동기 대상 없음. 코드 주석의 측정 사실 문단은 그대로 유효(수정 없음).
- 가드 테스트는 t521 선례(`doctor_codex_mirror_drift_test.go`)의 헬퍼 `repointProducerMirrorRelDir` 를 재사용하고, 미러 SKILL.md 를 정규 파일로 심는 전용 헬퍼 `writeMirrorCopySkill` 만 새로 뒀다.
