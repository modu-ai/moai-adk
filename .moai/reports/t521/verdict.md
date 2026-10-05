# t521 — lane verdict (doctor mirror layout literals → producer exports)

card: t521 (Class B, Tier S)
worktree: .claude/worktrees/t521
branch: WT-doctor-mirror-path
base: e4ecbf854 (local develop at dispatch)
measured by: lane-8 orchestrator, directly

## Claim

1. Doctor's Codex-wiring check no longer restates the skill mirror layout. The private var
   block (`mirrorSkillsRelDir` / `canonicalSkillsRelDir`, `internal/cli/doctor_codex.go`) is
   deleted; the check reads `template.MirrorSkillsRelDir` / `template.CanonicalSkillsRelDir`.
   The producer exports `MirrorSkillsRelDir` (kept a var — the regression's seam),
   `MirrorLinkTarget(skill)`, and `CanonicalSkillsRelDir`, with `canonicalSkillsPrefix`
   derived from the canonical dir so the walker's slash-terminated form keeps one source too.
2. The drift-following guard is discriminating, not a literal-count assertion. Two tests
   (`internal/cli/doctor_codex_mirror_drift_test.go`) repoint `template.MirrorSkillsRelDir`
   to `.agents/moved` and require the reader to follow — the mirror at the repointed
   location reads clean, and the absent finding names the producer's CURRENT path. Both
   were observed FAILING against an injected hardcoded-literal mutant and PASSING on the
   clean tree, same commands, same session.
3. Default-path behavior is unchanged: the exported values render the same strings the
   deleted literals did (`filepath.Join(".agents","skills")` → ".agents/skills";
   ".claude/skills"), so doctor messages and golden fixtures stay byte-identical on
   default trees.

## Evidence

Two-run table (selector `-run 'TestInspectSkillMirrorFollowsProducerMirrorRelDir|TestMirrorAbsentFindingNamesProducerRelDir' -count=1 -v`):

| tree | follows-guard | names-producer-guard | run exit |
|---|---|---|---|
| mutant (`.agents/skills` literal re-hardcoded in doctor) | **FAIL** — "doctor did not follow the producer's mirror dir \".agents/moved\": mirror read as absent" | **FAIL** — "absent finding does not name the producer's mirror dir \".agents/moved\": .agents/skills mirror absent — …" | `FAIL` |
| clean (mutant reverted) | PASS | PASS | `ok` |

Supporting runs, this session, this tree:

- `gofmt -l` on the four changed Go files: empty list.
- `go vet ./internal/template/... ./internal/cli/...`: clean (`VET_OK`).
- `go build ./internal/template/... ./internal/cli/...`: `BUILD_OK`.
- `go test ./internal/template/... -count=1`: `ok … 36.636s`.
- `go test ./internal/cli/ -run 'Mirror|DoctorCodex' -count=1`: `ok … 4.593s`.
- `go test ./internal/cli/ -run 'Doctor|Mirror|Skill' -count=1`: `ok … 334.882s`.
- `GOOS=windows GOARCH=amd64 go build ./internal/template/... ./internal/cli/...`: `WIN_BUILD_OK`.
- Load condition at the heavy run: load averages 7.07 / 6.17 / 6.37 (measured before start).

Mutant provenance: the two mutant lines were injected by edit into
`internal/cli/doctor_codex.go` (the `inspectSkillMirror` mirror-dir join and the absent
finding's summary argument), run, then reverted by edit; the clean-tree re-run confirmed
the revert (`git diff` contains no mutant line — the committed tree carries the
`template.MirrorSkillsRelDir` reads only).

## Baseline-attribution

All figures above were measured in this run, in this worktree
(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t521`), on branch `WT-doctor-mirror-path`
based at local develop `e4ecbf854`. No figure is carried from another tree or session.

## Gaps

- The full `internal/cli` package suite was NOT run locally (lane-local verification scope;
  the package runs ~26 min under load). The `Doctor|Mirror|Skill` selector, `go vet`, both
  scoped full-package runs where cheap, and the Windows build stand in; the full-suite
  verdict belongs to CI on the develop push.
- The sibling literal `mirrorSkillsRel` (`internal/cli/codex_skills_disable.go:54`), which
  restates ".agents/skills" for the disable flow, was discovered but NOT fixed — outside
  this card's doctor scope. Flagged to the lead for a follow-up card.
- `repairPathA`'s canonical literal (`filepath.Join(projectRoot, ".claude", "skills", name)`,
  `skill_mirror_repair.go`) is producer-internal (same package as the SSOT) and was left
  as-is for the same scope reason.

## Residual-risk

- A future producer-side layout rename now compiles cleanly through doctor (compile-time
  binding), but the golden fixtures pin the message text; a rename will fail goldens until
  they are regenerated — that is the drift guard surfacing, not a defect.
- Until the disable-path card lands, `moai codex-skills disable` reads a restated path and
  would show the same false-absent shape if the producer moved the layout.
