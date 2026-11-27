# `secret-deny-apply`

`operation secret_deny.apply` · 코드 `src/secret_deny.rs`, `src/claude.rs`, `src/ops/secret_deny_apply.rs` · SPEC-20261122-secret-file-deny-rules

`moai secret-deny-apply [프로젝트 폴더] --plan-digest D --permission-reference REF [--pick 규칙]... [--retire-claudeignore]` 는 확인한 `secret-deny-plan` 의 골라낸 규칙만 루트 `.claude/settings.json` 의 `permissions.deny` 에 더해요. `--pick` 과 `--retire-claudeignore` 는 미리보기와 같게 주어야 해요.

- 쓰기 전 거절(아무것도 쓰지 않아요): 권한 참조가 없을 때, 고른 규칙이 미리보기에 없을 때, 설정 파일이 링크이거나 `.claude` 폴더가 링크일 때, JSON 객체가 아닐 때, 사용자 설정 폴더일 때, git 이 `.claude/settings.json` 을 무시할 때(규칙이 워크트리 세션에 닿지 않으니 그 파일을 추적하라고만 안내해요), digest 가 다를 때.
- git 이 무시하지 않으면 써요. 답의 `message` 가 파일을 커밋해야 규칙이 워크트리에 닿는다고 말하고, git 이 답하지 못하면(`unknown`) 모른다고 말해요.
- 쓰기 순서: 바뀐 설정 파일의 원본을 설정 백업 폴더(`<상태>/settings-backups/<UTC>-project/`)에 남기고, 영수증이 있으면 `pending` 으로 적은 뒤, 설정 파일을 원자적으로 교체하고, 영수증의 파일 지문만 새로 적어요. 영수증의 `owned` 목록은 바이트 그대로예요. `permissions` 를 moai 소유로 만들지 않아요. 영수증이 없으면 만들지 않아요. 설정 파일이 없으면 `permissions.deny` 만 가진 파일을 새로 만들고 백업은 없어요.
- 목록 순서: 이미 있는 항목은 그대로 두고, 고른 규칙 가운데 목록에 없는 것을 뒤에 더해요(키 파일 규칙이 아닌 것 먼저). 마지막 `!` 항목 앞에 선 키 파일 규칙은 그 `!` 항목 바로 뒤로 옮겨요. 같은 규칙으로 다시 적용하면 목록은 바뀌지 않고 `written` 은 `false` 예요.
- `--retire-claudeignore`: 설정 쓰기가 끝난 뒤 마지막 단계로 루트 `.claudeignore`(보통 파일만, 링크는 거절)를 같은 백업 폴더로 `rename` 해요. 복사하고 지우지 않아요. 옮기기가 실패하면 규칙은 적용된 채로 두고 `partial apply` 로 말하고, 같은 규칙과 플래그로 미리보기를 다시 하면 이동만 끝내요. 거절되거나 무시된 적용은 아무것도 옮기지 않아요.
- 답: `applied`, `written`, `state`(`applied`·`unchanged`), `final_deny`, `backup_dir`, `retired_claudeignore`, `message`. 쓰기 검사가 거절할 내용(`would_refuse`)은 보고만 하고 쓰기를 막지 않아요.
