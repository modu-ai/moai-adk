# `audit-done`

operation `audit.done` · 코드 `src/ops/audit_done.rs`, `src/audit_prep.rs` · SPEC-20261124-audit-prep

카드의 감사를 끝내요. `audit-prep` 이 카드에 잡아 준 자리를 찾아 깨끗할 때만 풀고, 자리의 `target/` 은 다음 감사가 이어 쓰도록 남겨요.

`audit-done CARD [--project DIR]`

- `CARD` 는 `C-` 와 숫자예요. 카드가 쥔 감사 자리를 holder 기록에서 찾아요. 임대가 끝난 기록도 같은 이름이면 그 카드의 자리예요. 세션은 비교하지 않아요(`audit-prep` 과 같아요).
- 깨끗한 자리(추적 파일 변경도, `target/` 밖의 추적 안 된 파일도 없음)는 holder 기록을 지우고 풀어요. 답은 `audit-slot done` 과 같고(`slot`, `path`, `head`, `released`, `target_kept`) `card` 가 붙어요.
- 변경이 남은 자리(뮤턴트, 스테이지한 새 파일, 떨어진 파일)는 파일 이름을 대며 거부하고 아무것도 바꾸지 않아요. 버리는 옵션도 강제하는 옵션도 없어요(`--discard`, `--force` 는 받지 않아요). 자리에서 변경을 되돌린 뒤(`git -C <자리> checkout -- <경로>`) 다시 돌려요. 막힌 자리를 리더가 고칠 때는 `audit-slot done <자리> --holder <카드> --discard` 를 써요.
- 카드가 쥔 자리가 없으면 거부해요.
- 무시된 파일(`target/` 말고)은 지우지 않아요. 자세한 것은 `audit_slot.md`.
