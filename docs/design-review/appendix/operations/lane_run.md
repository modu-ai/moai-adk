# `lane-run`

operation `lane.run` · 코드 `src/ops/lane_run.rs`, 프로필 키 `src/profile.rs`(`lane_run`, `LaneRunSettings`) · SPEC-20261122-nonstop-lane

레인을 멈추지 않고 돌릴지(`nonstop`), 카드 하나씩 돌릴지(`single`)를 고르고 그 선택에 따른 글을 돌려준다. 선택은 프로젝트 프로필의 최상위 키 `lane_run` 하나에 산다: `default` 는 프로젝트의 선택, `lanes` 는 레인 이름별 선택이다. 레인 이름 규칙은 `session-register` 와 같다(`lane_registry::valid_name`). 이 명령은 `/loop` 도 `/goal` 도 시작하지 않는다.

- `lane-run show [--lane NAME] [--project DIR]`: 저장된 선택(`saved`: `nonstop`·`single`·`unset`), 출처(`source`: `lane`·`project`·`none`, 레인 값이 프로젝트 값보다 먼저), 실제로 쓰일 선택(`effective`), 사실(`facts`)과 버전 하한 글(`floors`)을 읽기만 해서 돌려준다. `CLAUDE_CODE_DISABLE_CRON` 이 참 값(`1`·`true`·`yes`·`on`)이면 그 사실이 `blocks: true` 가 되고 `effective` 는 `single` 이다. moai 가 읽을 수 없는 사실(루프를 쓸 수 있는지, Claude Code 버전)은 `read: false` 로만 알리고 막지 않는다. 선택이 없으면(`unset`) `effective` 는 `single` 이다. `lane_run` 은 객체이거나 없어야 한다: `null`·`[]` 은 프로필 읽기가 미선택·빈 값으로 받아도 `show`·`set`·`prompt` 가 같은 메시지로 거부하고, 문자열·숫자는 프로필 읽기가 이미 거부한다.
- `lane-run set nonstop|single [--lane NAME] [--project DIR]`: 묻지 않고 프로젝트 값이나 레인 값을 쓴다. 프로필을 먼저 읽어 검사하고, 원래 바이트의 다른 키와 그 순서를 그대로 둔 채 `lane_run` 만 바꿔 같은 폴더의 임시 파일과 이름 바꾸기로 쓴 뒤 다시 읽어 확인한다. 다른 값, 받지 않는 레인 이름, 프로필이 없는 프로젝트, 정규 파일이 아닌 프로필은 아무것도 쓰기 전에 거부한다. 파일의 권한(mode)은 쓰기 전과 같게 둔다.
- `lane-run prompt --lane NAME [--project DIR]`: 실제 선택이 `nonstop` 이면 Claude Code 의 자기 속도 루프(간격 없는 루프)에 줄 평문 한 개를 `prompt` 로 돌려준다. 그 글은 지금 카드 계속, READY 뒤 `moai card next` 카드 잡기, 기다리는 동안 Monitor 로 붙잡은 카드들의 `moai card show` 를 읽어 새 노트 보기(`watch` 가 들어오는 조각 3에서 그 명령으로 바뀐다), 판정·수리 요청 먼저, 결정은 LANE-BLOCKED 노트, 카드도 감시도 없는 채로 턴을 끝내지 않기, 할 카드가 없으면 idle 한 줄 뒤 `ScheduleWakeup` 의 `stop: true` 로 끝내기를 담는다. 카드 번호, 키, 프로젝트 밖 경로는 담지 않는다. `single` 이면 `prompt` 는 `null` 이고 루프를 시작하지 않는다고 알린다.
- `lane-run goal --card C-N [--turns N] [--project DIR]`: 카드 하나의 `/goal` 조건(4,000자 이하)을 돌려준다. 끝 상태(READY 노트와 git 에서 읽은 head·base, 카드 테스트 통과, review 상태), 확인 방법, 턴 한도(기본 40, 1~200)를 담는다. `/goal` 은 사용자가 입력한다.
- `lane-run watch` 는 다음 조각(SPEC Plan 3)에서 더한다.
