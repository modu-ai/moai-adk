#!/usr/bin/env python3
"""t520 AC-UMH-014 mutation probe.

Each mutant is applied with MUST-REPLACE semantics: if the needle is absent or
ambiguous the run aborts, so a mutant that silently failed to apply cannot
masquerade as "not caught by the guard".
"""
import subprocess
import sys
import pathlib

ROOT = pathlib.Path(__file__).resolve().parents[3]
REPAIR = ROOT / "internal/template/skill_mirror_repair.go"
GATE = ROOT / "internal/cli/update_mirror_heal.go"
DOCTOR = ROOT / "internal/cli/doctor_codex.go"

MUTANTS = [
    ("AC-UMH-003", "ignore the version gate", GATE,
     "\tif !mirrorRepairGateOpen(stamp) {",
     "\tif false { // MUTANT",
     "TestUpdateMirrorHeal_NoCreateBelowStamp"),

    ("AC-UMH-004", 'treat "0.0.0" as satisfying the gate', GATE,
     "func mirrorRepairGateOpen(templateVersion string) bool {\n",
     'func mirrorRepairGateOpen(templateVersion string) bool {\n\tif templateVersion == "0.0.0" {\n\t\treturn true // MUTANT\n\t}\n',
     "TestUpdateMirrorHeal_NoCreateWithoutStamp"),

    ("AC-UMH-006", "rewrite published files unconditionally", REPAIR,
     "\t\tif _, err := os.Stat(dest); err == nil {\n\t\t\t// Occupied — restore-missing-only, never overwrite.\n\t\t\tcontinue\n\t\t}\n",
     "\t\t// MUTANT: restore-missing-only guard removed\n",
     "TestUpdateMirrorHeal_HealthyIsNoop"),

    ("AC-UMH-007", "remove and replace a non-symlink occupant", REPAIR,
     "\t\tpresent = append(present, name)\n",
     "\t\tpresent = append(present, name)\n\t\t_ = os.RemoveAll(filepath.Join(mirrorDir, name)) // MUTANT\n",
     "TestUpdateMirrorHeal_SkipsForeignOccupant"),

    ("AC-UMH-008", "widen the scope set to every dir under .claude/skills (S1)", REPAIR,
     "\tres.repairPathA(projectRoot, canonical)\n",
     "\tif des, derr := os.ReadDir(filepath.Join(projectRoot, \".claude\", \"skills\")); derr == nil { // MUTANT\n\t\tcanonical = nil\n\t\tfor _, de := range des {\n\t\t\tif de.IsDir() {\n\t\t\t\tcanonical = append(canonical, de.Name())\n\t\t\t}\n\t\t}\n\t}\n\tres.repairPathA(projectRoot, canonical)\n",
     "TestUpdateMirrorHeal_ScopeExcludesLocalSkills"),

    ("AC-UMH-011", 'reintroduce the "does not restore it" phrasing', DOCTOR,
     "where it is still absent after an update, %s",
     "does not restore it, so %s",
     "TestUpdateMirrorHeal_DoctorGuidance"),

    ("AC-UMH-012", "add a write call inside the doctor mirror span", DOCTOR,
     "func inspectSkillMirror(root string) skillMirrorState {\n\tvar st skillMirrorState\n",
     "func inspectSkillMirror(root string) skillMirrorState {\n\tvar st skillMirrorState\n\t_ = os.WriteFile(filepath.Join(root, \"mutant-probe\"), nil, 0o644) // MUTANT\n",
     "TestUpdateMirrorHeal_DoctorRemainsReadOnly"),

    ("AC-UMH-015", "remove the REQ-UMH-010 target-existence filter", REPAIR,
     "\t\tinfo, err := os.Stat(filepath.Join(projectRoot, \".claude\", \"skills\", name))\n\t\tif err != nil || !info.IsDir() {\n\t\t\tcontinue\n\t\t}\n",
     "\t\t// MUTANT: target-existence filter removed\n",
     "TestUpdateMirrorHeal_PartialDeployPathA"),

    ("EXTRA-noop", "neuter the whole repair pass (positive-criteria RED probe)", REPAIR,
     "\tcanonical, published, err := mirrorRepairCandidates(fsys)\n",
     "\tif true {\n\t\treturn res // MUTANT\n\t}\n\tcanonical, published, err := mirrorRepairCandidates(fsys)\n",
     "TestUpdateMirrorHeal_RestoresPathA|TestUpdateMirrorHeal_RestoresPathB|TestUpdateMirrorHeal_EarlyReturnPreserved|TestUpdateMirrorHeal_ScopeExcludesLocalSkills|TestUpdateMirrorHeal_PartialDeployPathB"),
]


def run(guard):
    p = subprocess.run(
        ["go", "test", "./internal/cli/", "-run", guard, "-count=1", "-v"],
        cwd=ROOT, capture_output=True, text=True)
    return p.returncode, (p.stdout + p.stderr)


def main():
    results = []
    for ac, desc, path, needle, repl, guard in MUTANTS:
        original = path.read_text()
        n = original.count(needle)
        if n != 1:
            print(f"ABORT {ac}: needle occurs {n} times in {path.name} (must be exactly 1)")
            sys.exit(1)
        try:
            path.write_text(original.replace(needle, repl))
            rc, out = run(guard)
            caught = rc != 0
            keep = [line for line in out.splitlines()
                    if line.startswith("--- ") or line.startswith("FAIL")
                    or line.startswith("ok ") or line.startswith("#")
                    or "_test.go:" in line]
            results.append((ac, desc, guard, caught, rc, "\n".join(keep)[:1500]))
        finally:
            path.write_text(original)

    print("\n================ MUTANT RESULTS ================")
    for ac, desc, guard, caught, rc, tail in results:
        print(f"\n### {ac} — {desc}")
        print(f"guard: {guard}")
        print(f"exit={rc}  caught={'YES' if caught else 'NO'}")
        print(tail)

    p = subprocess.run(["grep", "-rn", "MUTANT", "internal/cli", "internal/template"],
                       cwd=ROOT, capture_output=True, text=True)
    print("\n================ REVERT CHECK ================")
    print(f"grep -rn MUTANT internal/cli internal/template -> exit={p.returncode}")
    print(p.stdout[:2000] if p.stdout else "(no output — clean)")


if __name__ == "__main__":
    main()
