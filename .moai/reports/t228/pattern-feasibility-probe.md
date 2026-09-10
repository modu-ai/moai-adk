# t228 — 새 언어 패턴 실현 가능성 탐침

측정 트리: 워크트리 `.claude/worktrees/t228`, 브랜치 `WT-astgrep-16-langs`,
HEAD `294b4b6ab` (= `origin/main`). 측정 위치: 워크트리 루트. 도구: `ast-grep 0.40.5`.

목적: 마일스톤 M4~M7 이 전제하는 것 — 새 언어에서 **의미 있는 패턴이 실제로 발화하고,
무해한 코드에는 발화하지 않는다** — 을 착수 전에 확인. 발화만 확인하면 "무엇이든 잡는 룰"이
통과하므로 두 방향을 짝으로 쟀습니다.

## 양성 탐침 — 발화해야 하는 입력

| 언어 | 패턴 | 입력 | 결과 |
|---|---|---|---|
| java | `Runtime.getRuntime().exec($CMD)` | `Runtime.getRuntime().exec(cmd)` | 매치 |
| rust | `Command::new("sh")` | `Command::new("sh").arg("-c").arg(cmd)` | 매치 |
| php | `shell_exec($CMD)` | `<?php shell_exec($cmd); ?>` | 매치 |
| ruby | `system($CMD)` | `system(cmd)` | 매치 |
| csharp | `MD5.Create()` | `MD5.Create()` | 매치 |
| elixir | `System.cmd("sh", $ARGS)` | `System.cmd("sh", ["-c", c])` | 매치 |

명령: `printf '<입력>' \| sg run -p '<패턴>' -l <lang> --stdin`

## 음성 탐침 — 발화하면 안 되는 무해한 입력

| 언어 | 패턴 | 무해 입력 | 결과 |
|---|---|---|---|
| ruby | `system($CMD)` | `Process.spawn("ls", "-l")` | 무매치 |
| java | `Runtime.getRuntime().exec($CMD)` | `new ProcessBuilder("ls", "-l").start()` | 무매치 |
| rust | `Command::new("sh")` | `Command::new("ls").arg("-l")` | 무매치 |

## 판정

6개 언어에서 양성 발화 확인, 그중 3개에서 음성 무발화까지 확인. 패턴 작성 자체는
실현 가능하며, 남은 위험은 문법 가용성이 아니라 **패턴 정밀도**(무해한 동형 코드에
걸리지 않는지)로 좁혀집니다 — 이는 SPEC 의 `error` 승격 전제 조건이 이미 다루는 축이고,
각 룰의 `sg test` `valid` 케이스가 룰마다 반복 측정합니다.

이 탐침은 6개 언어 × 1패턴 표본이며, 10개 언어 × 8패밀리 전수의 근거가 아닙니다.
전수 판정은 M4~M7 에서 `sg test` 로 룰마다 측정합니다.
