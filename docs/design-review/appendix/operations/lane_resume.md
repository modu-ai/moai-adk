# `lane-resume`

operation `lane.resume`(SPEC-20261120-lane-limit-resume). 코드: `src/ops/lane_resume.rs`(명령과 판정). 화면 읽기와 상태·해제 순간 판정은 `lane-screen` 의 분류기(`moai_cli::lane_status::lane_screen` 의 `observe`·`reset_of`·`scan`)를, 글과 핸들 검사는 `lane-send` 의 조각(`check_text`·`check_handle`·`read_checked`·`Orca`·`ProcessOrca`)을 **읽기로** 재사용해요 — 그 두 파일은 고치지 않아요.

`lane-resume [--terminal HANDLE] --text TEXT [--enter] [--wait] [--now YYYY-MM-DDTHH:MM:SSZ]` 는 사용 한도("Usage limit reached … Your limit will reset at <시각>", 표기는 UTC+8일 수 있어요)로 멈춘 레인을 찾아, **해제 순간이 지났을 때** 이어서 한 줄을 보내는 도구예요. 2026-11-20 에 네 개의 glm 레인이 한도 앞에서 35분가량 서 있었고, 화면의 해제 시각을 누가 읽어 줘도 그 시각에 아무도 깨워 주지 않았던 일의 뒷마무리예요. `lane-send` 가 `usage_limit` 상태를 언제나 거절하는 것과 달리, 이 도구의 일은 바로 그 상태의 "해제 순간이 지난" 경우예요.

- **대상**: 기본은 `orca terminal list` 의 터미널을 하나씩 읽는 스캔(`lane-screen --scan` 과 같은 읽기)에서 `usage_limit` 레인 전부, `--terminal` 로는 그 터미널 하나만(목록을 읽지 않아요). 다른 상태의 레인은 대상이 아니에요.
- **해제 순간**: 화면 글자에서만, `lane-screen` 의 시간대 규칙 그대로예요(아는 시간대만 UTC 로 바꾸고, 시간대가 없거나 모르면 그 레인은 skipped). 봉투와 오류 글에 레인 이름과 함께 UTC 로 적어요.
- **보내기**: 해제 순간이 지난 레인만, 레인마다 이 실행에서 최대 한 줄. 보내기 직전에 화면을 한 번 더 읽고(`lane-send` 의 두 번 읽기), 상태가 바뀌었으면 보내지 않고 새 상태를 이름과 함께 알려요. 회복(`working`·`idle`)은 보낼 일이 없다는 뜻이고, 확인 창(`permission_prompt` 등)은 refused 로 거절해요 — 이 도구도 확인 창에는 절대 넣지 않아요.
- **waiting 과 `--wait`**: 해제 순간이 미래인 레인은 waiting 으로 이름과 UTC 순간이 오고 아무것도 보내지 않아요. `--wait` 을 주면 가장 이른 미래 순간까지 **한 번만** 잠들고(상한 8시간 — 5시간 창과 여유), 깨어서 멈춤 레인을 다시 읽어 판정해요. 상한을 넘으면 잠들지 않고 waiting 으로 돌아가요(다시 실행하는 것은 레인의 깨우기 주기 몫이에요).
- **skipped**: 읽지 못한 레인(Orca 답이 아니거나 읽기가 실패)과 해제 순간을 못 읽은 레인은 이유를 적어 skipped 로 알리고 다른 레인은 계속해요.
- **글·핸들 검사**(`orca` 를 부르기 전에): `lane-send` 와 같아요. 비어 있지 않고, `-` 로 시작하지 않고, 2000자 이하이고, 제어·줄 구분자·양방향 문자가 없어야 해요. 핸들은 영문자·숫자·`.`·`_`·`-` 128자 이하, 영문자나 숫자로 시작. 오류 글의 접두는 `lane resume refused:` 예요.
- **Enter**: `--enter` 를 줄 때만 보내요.
- **보내기 답**: `lane-send` 와 같은 계약이에요. 답이 JSON 객체이고 `ok` 가 `true` 일 때만 보냈다고 해요. `ok: false` 면 "answered ok:false" 와 Orca 의 오류 한 줄(제어 문자 없이 160자까지), 답이 글자뿐이거나 `ok` 가 없으면 "not an Orca answer" 와 "글이 레인에 닿았을 수 있어요" 로 실패해요.
- **상태 없음**: 이 도구는 아무 파일도 쓰지 않아요. 중복 막기는 두 가지 몫이에요: 한 실행은 멈춘 레인당 한 줄뿐이고, 보낸 줄로 깨어난 레인은 화면을 다시 그려 다음 실행이 멈춤을 찾지 못해요. 같은 멈춤에 줄을 두 번 넣고 싶지 않으면 같은 인수로 되풀이하지 않는 것이 호출자의 규칙이에요.
- **종료 코드**: 한 줄이라도 보냈으면 0, 아무것도 못 보냈으면 0 이 아니고 오류 글에 모든 레인의 결과(waiting·skipped·refused·회복)가 있어요.
- **`orca` 호출**: `MOAI_ORCA_BIN`(없으면 PATH 의 `orca`), 호출마다 15초(`MOAI_ORCA_TIMEOUT_SECS` 로 1–300초) — `lane-screen`·`lane-send` 와 같아요.

성공하면 결과에 `sent`(참), `lines`(보낸 레인마다 `terminal`·`state`·`evidence`·`reset`·`enter`·`orca`), 그리고 `waiting`·`refused`·`recovered`·`skipped` 목록이 있어요. 각 목록의 항목에는 `action` 이 붙어요. 이 도구는 한도를 풀지 않고, 확인 창에 대신 답하지 않고, 보낸 줄이 무시되면 다시 보내지도 않아요.
