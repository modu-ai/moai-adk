# t525 M3.4 1부 — 의존 간선 확인

트리: `.claude/worktrees/t525`, 브랜치 `WT-speclint-red`, HEAD `02063f715`(흡수 트리 `3b9e8016` 위 증거 커밋) + 미커밋 spec.md 편집.
판정 바이너리: 흡수 트리 `42c17b5b6` 에서 빌드한 스크래치 바이너리(`moai-t525-absorb2`, `cmd/moai` 는 이후 커밋에서 바뀌지 않음), 경로로 호출.

## 편집 (manager-spec 위임)

`.moai/specs/SPEC-SPECLINT-GATE-SIGNAL-001/spec.md` 프런트매터 두 줄:

- `updated: 2026-09-08` → `updated: 2026-09-11`
- `tier: M` 다음 줄에 `dependencies: [SPEC-SPEC-LINT-BLIND-AXES-001]` 추가

`git diff --stat` → `1 file changed, 2 insertions(+), 1 deletion(-)`. 편집 뒤 blob `7b686256b`.
필드 이름은 린터 바인딩 `Dependencies []string \`yaml:"dependencies"\``(`internal/spec/lint.go:513`)에 맞췄다.
의존 대상 `.moai/specs/SPEC-SPEC-LINT-BLIND-AXES-001/spec.md`: `id: SPEC-SPEC-LINT-BLIND-AXES-001`, `status: completed`.
t518 착지: `git merge-base --is-ancestor a4fbaeb82 HEAD` → exit 0. 흡수한 develop tip: `git merge-base --is-ancestor 81c1d58f9 HEAD` → exit 0.

## 편집 뒤 lint

`spec lint --json` → exit 0, 발견 3336, error 0, non-advisory warning 0, 코드 이름에 `Dependenc` 가 들어간 발견 0, 이 SPEC 파일의 발견 8. 편집 전 흡수 트리 측정(`../absorb/census.txt`)과 발견 수가 같다.

## 양성 대조 — 린터가 이 줄을 읽는가

발견 0만으로는 "간선이 유효하다"와 "린터가 줄을 안 읽는다"를 가를 수 없다. 그래서 값을 존재하지 않는 id 로 잠시 바꿨다.

1. 원본 백업(세션 스크래치 `t525-spec-edge.bak`).
2. 프런트매터 줄만 `dependencies: [SPEC-NOSUCH-PROBE-999]` 로 바꿈(본문 REQ-SLGS-011 의 같은 문자열은 건드리지 않음).
3. `spec lint` → **exit 1**. 출력:
   - `ERROR     MissingDependency              …/.moai/specs/SPEC-SPECLINT-GATE-SIGNAL-001/spec.md                      1     Dependency SPEC "SPEC-NOSUCH-PROBE-999" not found`
   - `1 error(s), 3133 warning(s)`
4. 백업으로 복원 → `cmp` exit 0, blob `7b686256b`, `NOSUCH-PROBE` 0회.

판독: 린터는 이 프런트매터 줄을 읽고 대상 id 를 코퍼스에서 찾는다. 실제 id 로는 발견이 없으므로 간선은 유효하다. 의존 대상은 프런트매터 `id` 와 디렉터리 이름이 같아(`SPEC-SPEC-LINT-BLIND-AXES-001`) 어느 쪽으로 찾는지는 이 대조로 가를 수 없다.
