# `home-routing-plan`

operation `home.routing_plan` · 코드 `src/ops/home_routing_plan.rs`, `src/home_routing.rs`, `src/home_routing/rules.rs` (#2638, SPEC-20261122-claude-md-home-routing, SPEC-20261122-claude-md-flow-rule)

`home-routing-plan HOME PACKAGE [--config-dir ABS]... [--remove-old-block]` 는 설정 폴더마다 두 파일을 함께 계획한다: 꾸러미의 `references/home/routing.md` 를 그린 `rules/moai.md`(`{invoke}` → Claude `/moai`, 12줄 이내, `paths` 키 없음)와 `references/home/flow.md` 의 `rules/moai/flow.md`(NeXT-Flow 글, 머리말 없음, 80줄·5 KiB 이내). 꾸러미에 흐름 글이 없으면 그 파일은 계획하지 않고, 한도를 넘거나 `paths` 키가 있는 원본은 한도를 이름으로 대며 계획 전체를 거절한다. 파일은 마지막에 쓴 바이트와 같을 때만 moai 것이다(홈 스타일 파일과 같은 소유 해시): `create`, `merge`, `unchanged`, 그 밖은 `conflict`(고친 파일, 링크인 파일·`rules`·`rules/moai`·설정 폴더, UTF-8 이 아닌 파일). 막힘(`blocked`)은 `CLAUDE_CONFIG_DIR` 가 거부될 때와 소유 기록을 읽을 수 없을 때만이다. 아무것도 쓰지 않는다.

결과: `schema_version`, `home`, `package_digest`, `source_sha256`, `flow_source_sha256`, `hosts`(설정 폴더마다 한 줄: `host`, `config_dir`, `file`(라우팅 파일), `preflight`, `blocked_reason`, `section`(라우팅 파일의 계획; 라우팅 파일이 `unchanged` 이고 흐름 파일만 바뀔 때는 `moai update` 가 이 줄만 읽으므로 흐름 파일의 동작을 싣는다), `flow`(흐름 파일의 계획), `old_block`), `not_supported`, `plan_digest`. 기본 설정 폴더는 늘 먼저 계획된다. `--config-dir` 는 절대 경로만 받고, 되풀이하면 폴더를 더하며, 같은 폴더는 한 번만 계획한다.

옛 `moai-routing` 블록(사용자 `CLAUDE.md` 안)은 더 이상 갱신하지 않는다. 블록이 있을 때만 `old_block` 줄이 나온다: `offered`(지울 수 있고 지금은 지우지 않음; 이 `CLAUDE.md` 를 가져다 쓰는 프로필 세션이 안내를 잃을 수 있다는 경고가 `detail` 에 있다), `remove`(`--remove-old-block` 을 줬고 블록이 기록된 기준과 같다), `waiting`(라우팅 파일이 아직 같은 폴더에 moai 것으로 놓이지 않았다), `kept`(기준 기록이 없거나 블록이 고쳐졌거나 `CLAUDE.md` 가 링크·UTF-8 아님; 사유가 `detail` 에 있다). 기준은 옛 상태 파일 `home-routing/claude.json` 이고 기본 설정 폴더에만 있다.

도구 폴더는 `src/host_home.rs` 하나가 정한다(SPEC-20261117-host-home-resolver): `HOME` 이 이 계정의 홈(비밀번호 데이터베이스, 시험은 `MOAI_ACCOUNT_HOME`)일 때 `CLAUDE_CONFIG_DIR` 이 `HOME/.claude` 를 대신한다(절대 경로만). 빈 값·상대 경로면 그 줄은 `blocked` 이고 사유가 변수 이름을 말한다(값은 되풀이하지 않음). 다른 홈(임시 HOME 등)은 변수와 상관없이 그 아래 폴더를 쓴다. 옮겨진 폴더의 `rules` 가 기본 폴더의 `rules` 로 가는 링크이면 따라가고, 다른 링크는 사유에 링크를 대는 `conflict` 다. 설정 폴더별 소유 기록은 `home-routing/rules-<폴더 경로 sha256 의 앞 12자>.json` 이다. Kiro·Cursor CLI 는 `not_supported` 에 사유와 함께 적고 쓰지 않는다.
