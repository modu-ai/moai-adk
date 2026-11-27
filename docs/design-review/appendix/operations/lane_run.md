# `lane-run`

operation `lane.run` · 코드 `src/ops/lane_run.rs`, 프로필 키 `src/profile.rs`(`lane_run`, `LaneRunSettings`) · SPEC-20261122-nonstop-lane

레인을 멈추지 않고 돌릴지(`nonstop`), 카드 하나씩 돌릴지(`single`)를 고르고 그 선택에 따른 글을 돌려준다. 선택은 프로젝트 프로필의 최상위 키 `lane_run` 하나에 산다: `default` 는 프로젝트의 선택, `lanes` 는 레인 이름별 선택이다. 레인 이름 규칙은 `session-register` 와 같다(`lane_registry::valid_name`). 이 명령은 `/loop` 도 `/goal` 도 시작하지 않는다.

- `lane-run show [--lane NAME] [--project DIR]`: 저장된 선택(`saved`: `nonstop`·`single`·`unset`), 출처(`source`: `lane`·`project`·`none`, 레인 값이 프로젝트 값보다 먼저), 실제로 쓰일 선택(`effective`), 사실(`facts`)과 버전 하한 글(`floors`)을 읽기만 해서 돌려준다. `CLAUDE_CODE_DISABLE_CRON` 이 참 값(`1`·`true`·`yes`·`on`)이면 그 사실이 `blocks: true` 가 되고 `effective` 는 `single` 이다. moai 가 읽을 수 없는 사실(루프를 쓸 수 있는지, Claude Code 버전)은 `read: false` 로만 알리고 막지 않는다. 선택이 없으면(`unset`) `effective` 는 `single` 이다. `lane_run` 은 객체이거나 없어야 한다: `null`·`[]` 은 프로필 읽기가 미선택·빈 값으로 받아도 `show`·`set`·`prompt` 가 같은 메시지로 거부하고, 문자열·숫자는 프로필 읽기가 이미 거부한다.
- `lane-run set nonstop|single [--lane NAME] [--project DIR]`: 묻지 않고 프로젝트 값이나 레인 값을 쓴다. 프로필을 먼저 읽어 검사하고, `lane_run` 값을 바꾼 파일 전체를 pretty JSON 으로 다시 써서 같은 폴더의 임시 파일과 이름 바꾸기로 바꿔 넣은 뒤 다시 읽어 확인한다. 다른 키와 값, 키 순서는 유지하지만 `lane_run` 밖의 바이트 형식(들여쓰기, 줄바꿈 같은 것)은 보장하지 않는다. 다른 값, 받지 않는 레인 이름, 프로필이 없는 프로젝트, 정규 파일이 아닌 프로필은 아무것도 쓰기 전에 거부한다. 파일의 권한(mode)은 Unix 에서만 쓰기 전과 같게 둔다.
- `lane-run prompt --lane NAME [--project DIR]`: 실제 선택이 `nonstop` 이면 Claude Code 의 자기 속도 루프(간격 없는 루프)에 줄 평문 한 개를 `prompt` 로 돌려준다. 카드 번호, 키, 프로젝트 밖 경로는 담지 않는다. `single` 이면 `prompt` 는 `null` 이고 루프를 시작하지 않는다고 알린다. 평문은 다음을 담는다.
  - 깰 때마다 먼저 `moai session-register --show --name leader` 로 리더 등록을 읽는다. `held` 면 리더가 있고 `free` 면 없다. 다만 이 세션이 등록된 리더를 이미 봤거나 배정을 받았으면 `free` 는 만료된 등록이다. 만료, 그 밖의 답, 실패에는 새 카드를 잡지 않고 가진 카드만 이어 가며, 마지막에 일한 카드에 그 답을 적은 노트 하나를 남긴다.
  - 진행 중인 카드를 먼저 잇는다. 리더가 있으면 READY 노트 뒤에 카드를 넘기고 완료 메시지 하나를 보낸 다음 리더의 배정을 기다리며, 스스로 다음 카드를 잡지 않는다. 리더가 없으면 원장에서 앞 카드가 done 이거나 해제된 것을 본 뒤 `moai card next` 가 알려 주는 카드를 잡고 본문과 노트를 다 읽는다.
  - 감시는 `moai card wait` 하나다. 대상은 `moai card next` 의 `mine` 카드와, `moai card list --state review` 카드 가운데 `moai card show` 의 `roles.authors` 에 `moai session-current` 가 찍는 `source_session_id` 가 있는 카드다. 마감 30분의 Monitor 가 `moai card wait --card <card> --timeout 25m --not-author <your session id>` 를 돌리고, `--card` 는 대상 카드마다 하나씩 넣는다. 감시할 카드가 없으면 감시를 시작하지 않는다. 처음에는 시작한 뒤 대상 카드를 `moai card show` 로 읽는다. 끝나면 돌려받은 가장 큰 `note_id`, 시간 초과면 답의 `mark` 를 `--since` 로 넘겨 다시 시작하고, 답 없이 끝났으면 처음 시작을 다시 한다.
  - 판정·수리 요청이 먼저다. 필요한 결정은 그 카드의 `LANE-BLOCKED kind=decision` 노트에 적고 루프에서 사용자에게 묻지 않는다.
  - 진행 중인 카드, 감시, 예약한 다음 깨우기 가운데 아무것도 없는 채로 턴을 끝내지 않는다. `ScheduleWakeup` 간격은 Monitor 가 도는 동안 1200~1800초, 감시할 것이 없으면 3600초다.
  - 일할 수 있는 카드가 없으면(결정을 기다리는 카드는 세지 않는다) idle 한 줄을 한 번 찍는다. 리더를 본 적 없는 세션은 감시를 멈추고 `ScheduleWakeup` 의 `stop: true` 로 루프를 끝낸다. 리더를 본 세션은 3600초 깨우기 하나를 남기고, 리더의 세션 종료 알림이나 사용자의 말이 있을 때만 끝낸다.
  - 원장은 moai 명령으로만 읽고 쓴다. 루프는 시작한 지 30일 뒤에 끝나고(Claude Code 2.1.330 전에는 7일) resume 으로 되살아나지 않는다. 답의 `note` 도 같은 만료를 알린다.
- `lane-run goal --card C-N [--turns N] [--project DIR]`: 카드 하나의 `/goal` 조건(4,000자 이하)을 돌려준다. 끝 상태(READY 노트와 git 에서 읽은 head·base, 카드 테스트 통과, review 상태), 확인 방법, 턴 한도(기본 40, 1~200)를 담는다. `/goal` 은 사용자가 입력한다.

예(`<card>` 자리에는 카드 이름을 넣는다):

```sh
moai lane-run show --lane agent-1
moai lane-run set nonstop --lane agent-1
moai lane-run prompt --lane agent-1
moai lane-run goal --card <card>
```
