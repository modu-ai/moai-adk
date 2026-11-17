# Language: C and C++

When to read: `internal/quality` sent you here because the project's detected languages include `c` or `cpp`. The detector finds them by file extension only, so read the build files before running anything. The project's own declarations (CI steps, `AGENTS.md`) win over every default below.

## Which build system

| Evidence | Configure and build | Test |
| --- | --- | --- |
| `CMakeLists.txt` | `cmake -S . -B build` then `cmake --build build` | `ctest --test-dir build --output-on-failure` |
| `meson.build` | `meson setup build` then `meson compile -C build` | `meson test -C build` |
| `Makefile` | `make` | `make test` or `make check`, only when that target exists |
| none of these | nothing to run: record build and test as `not_run` and ask how the project builds | |

Use a build folder inside the worktree, and the one the project's docs name when there is one.

## Evidence and defaults

| Kind | Default command | Use it only when |
| --- | --- | --- |
| format | `clang-format --dry-run --Werror <changed files>` | `.clang-format` exists |
| lint | `clang-tidy -p build <changed files>` | `.clang-tidy` exists and the build wrote `compile_commands.json` |

## Dependency existence

There is no single registry. Check in the manager the project uses:

- `vcpkg.json`: `vcpkg search <name>`;
- `conanfile.txt` or `conanfile.py`: `conan search <name>`;
- a system package: the distribution's package search, and say which distribution you checked.

## Traps

- Warnings can be errors in CI (`-Werror`) and not locally; build with the CI's flags.
- A stale build folder hides changes; when a result looks stale, configure a fresh folder and say so.
- `ctest` with a filter (`-R`) that matches nothing reports `No tests were found`: that is `not_run`.
- `clang-format -i` rewrites files; a VERIFY run never rewrites.
