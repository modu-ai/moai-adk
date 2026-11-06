# REQ-OWR-005 개정 제안 (SPEC-UPDATE-OWNERSHIP-REPAIR-001)

- 제안 출처: 카드 t1702 (브랜치 `WT-profile-skill-reach`)
- 대상: SPEC-UPDATE-OWNERSHIP-REPAIR-001 `spec.md` REQ-OWR-005 및 AC-OWR-08
- 성격: 제안서다. 해당 SPEC은 다른 카드의 보류 브랜치에 있어 이 카드는 SPEC 파일을 고치지 않는다. 채택 여부는 운영자와 그 카드의 담당자가 정한다.
- 근거 결정: d-20261106T055331Z-c28f (2026-11-06, 운영자 판정 — 프로필 도달은 읽기 전용 검증으로 판단)

## 현행 문구와 충돌하는 지점

1. **읽기·stat 전면 금지와 판정의 정면 충돌.** 현행 문구는 프로필 루트를 open, read, enumerate, stat, walk 하는 것을 모두 금한다. 그런데 운영자 판정은 프로젝트 사본을 지우기 전에 실제 프로필 루트마다 같은 사본(내용 해시 동일)이 있는지 읽어서 확인하라고 요구한다. 판정을 따르려면 바로 그 금지 행위를 해야 한다. 이 브랜치의 `internal/userassets/profile_reach.go`(`CheckProfileReach`, `SafeToRemoveProjectCopy`)가 그 읽기를 구현했다.
2. **동결된 호출 지점 기준선과 AC-OWR-08.** 현행 문구는 기존 접점을 넘는 새 호출 지점을 금하고, AC-OWR-08은 새 호출 지점이나 늘어난 호출 수에서 빨간불이 되는 소스 스캔이다. 위 함수들은 `config.UserProfileDirNames` 열거, 루트 Lstat, `os.OpenRoot`, 상대 파일 Lstat, 제한 읽기를 새로 더하므로 현행 AC-OWR-08을 그대로 두면 반드시 실패한다.
3. **변경 금지 목록과 이 브랜치의 변경.** 현행 문구는 `internal/config/userroots.go`와 `internal/config/protected_zone.go`의 변경을 금한다. 이 브랜치는 `protected_zone.go`에 `foldProfileSlug`를 더해, 배포된 `user-root:claude-skills/`·`user-root:claude-agents/` 보호 구역이 `claude-skills@<profile>` 같은 프로필 파생 루트까지 덮게 했다(프로필 루트 쓰기를 더 어렵게 하는 보호 강화). 또 `userroots.go`를 단일 루트 출처로 삼는다. 현행 문구 아래서는 둘 다 위반이 된다.

## 개정안 — REQ-OWR-005

> **REQ-OWR-005** (Unwanted behavior): The update shall not create, write, rename, remove, link, chmod, or otherwise modify anything inside any profile root (the `skills` and `agents` directories under a profile configuration directory, and the `claude-skills@<profile>` and `claude-agents@<profile>` directories), and shall not supply a profile root as a destination to any write-capable operation; this write ban has no exception.
> The update shall not open, read, enumerate, stat, or walk any profile root for any purpose other than the profile-reach judgment defined below, and shall not add any other call site of a symbol that produces or opens a path in a profile configuration directory beyond the frozen existing contacts.
> The single permitted read path is the profile-reach judgment in `internal/userassets` — `CheckProfileReach` and `SafeToRemoveProjectCopy` — which may enumerate real profiles through `config.UserProfileDirNames`, and, per call, Lstat a profile `skills` or `agents` root, open it under an `os.Root`, Lstat and read with a fixed upper bound one named relative file inside it, solely to decide whether a project copy of a user skill or agent may be removed; a relative path that leaves the root shall be rejected, and any result other than an equal content hash for every real profile — including zero profiles, a link, a non-regular file, or any read error — shall keep the project copy.
> The frozen existing contacts, including the `os.Stat` of the profile configuration directory entry inside `launchCandidateIsUsable` (`internal/profile/profile.go:439`), remain tolerated and unchanged; this SPEC neither calls them anew nor alters them.
> This SPEC shall not change `internal/profile`, `internal/config/userroots.go`, or `internal/config/protected_zone.go`; a change to `userroots.go` or `protected_zone.go` that only adds protection or a shared root derivation is permitted solely when made by the card that owns profile reach, and is not part of this SPEC's change set.

## 개정안 — AC-OWR-08

> **AC-OWR-08**: The source scan stays red on any write-capable call whose target is or derives from a profile root, on any open, read, stat, or Lstat of a profile root outside the allowlist, and on any enumeration or walk of profile roots outside the allowlist, where the allowlist is exactly `internal/userassets.CheckProfileReach` and `internal/userassets.SafeToRemoveProjectCopy`, plus the frozen existing contacts at their frozen baseline count.

## 채택하지 않은 안

- **설치 단계에서 모든 프로필 루트에 대응 사본을 넣는 안.** 프로젝트 사본을 지우기 전에 각 프로필 루트로 사본을 써 넣어 도달을 보장하는 방식이다. 운영자 판정(d-20261106T055331Z-c28f)이 읽기 전용 검증을 택했기 때문에 채택하지 않았다.

## 미확인

- 이 제안은 SPEC의 REQ-OWR-005 문구(spec.md 67행)와 AC-OWR-08 요지만 근거로 썼고, SPEC의 나머지 부분은 읽지 않았다.
- 따라서 프로필 루트를 언급하는 다른 요구사항·인수 기준(예: AC-OWR-08을 참조하는 항목, L-12, OQ-5 등)도 같은 개정이 필요할 수 있으나 이 제안은 다루지 않는다.
- `internal/profile/profile.go:439` 행 번호는 현행 문구에서 옮긴 것이며, 이 카드에서 다시 확인하지 않았다.
- AC-OWR-08 소스 스캔의 실제 구현(검사 대상 심볼 목록·기준선 수치)은 보지 못했으므로, 허용 목록을 스캔에 반영하는 방법은 SPEC 담당 카드가 정해야 한다.
