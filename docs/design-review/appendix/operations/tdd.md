# `tdd`

operation `tdd` · 코드 `src/ops/tdd.rs`, `src/tdd.rs` · 카드 C-86 3조각, SPEC-20261117-green-test-lock

`tdd` 는 이 worktree 의 TDD 한 바퀴를 기록하고, GREEN·REFACTOR 동안 tool-preflight hook(규칙 A6)이 시험 파일 쓰기를 막게 한다. 기록 `tdd.json` 과 로그 `tdd-log.jsonl` 은 worktree 자기 git 폴더의 `moai/` 아래에 있어 커밋되지 않는다. 가지는 SPEC 작업 가지 `spec/<slug>-<yyyymmdd>` 여야 한다(옛 꼴 `spec/<yyyymmdd>-<slug>` 와 끝의 수리 접미사 `-r<N>` 도 읽는다). 이름이 두 SPEC id 로 읽히면(`spec/20261117-demo-20261118`) `.moai/specs/` 에 그중 하나의 폴더만 있을 때를 빼고 거절하고, 이유에 두 id 를 적는다.

| 동작 | 쓰기 | 조건 | 결과 |
| --- | --- | --- | --- |
| `tdd red --check <cmd> …` | 기록·로그 | 열린 바퀴가 없음, head 커밋 마지막 문단에 `Moai-Phase: red`, head 커밋이 시험 경로를 바꿈, 모든 `--check`(1~8개, 셸 없이 argv)가 assertion 실패로만 떨어짐(`run::red_confirmed`) | `phase: "green"`, `record`(`spec`, `red_commit`, `locked`: RED 커밋이 바꾼 시험 경로, `checks`) |
| `tdd green` | 기록·로그 | 기록이 GREEN, 기록한 check 가 head 에서 모두 통과 | `phase: "refactor"`, 잠금은 그대로 |
| `tdd reopen --reason <text>` | 기록 지움·로그 | 열린 바퀴, 이유 한 줄 1~500자 | 잠금 해제, 로그에 이유 |
| `tdd done` | 기록 지움·로그 | 열린 바퀴 | 바퀴 끝, 로그는 남음 |
| `tdd status` | 없음 | — | `phase`, `record`, `log`, `problem`(로그의 마지막 사건이 `red`·`green` 인데 기록이 없거나 읽히지 않으면 그 사실을 적은 문장, 아니면 `null`; 잠금은 없음) |

시험 경로는 `src/tdd.rs` 의 언어별 표 `TEST_PATTERNS`(감지 언어 13개 + 공통 `fixtures`·`golden`)와 프로필의 `test_paths`(최대 32개, 상대 경로, `..` 없음)다. 프로젝트마다 자기 언어의 행과 공통 행만 쓴다(`TestPaths`): 언어는 프로필 `detected.languages`, 없으면 루트 manifest 파일(코드 파일은 훑지 않음)이고, 둘 다 없으면 모든 행이다. 그래서 python·rust 인 이 저장소에서 `templates/spec/*.md` 는 시험 경로가 아니다(Ruby `spec/` 행은 Ruby 프로젝트에만). `run` 의 oracle 판단도 같은 좁힘을 쓴다. 단계가 `green`·`refactor` 가 아닌 기록은 잠그지 않고 `status` 에서 단계가 없다. 기록·로그·`moai` 폴더가 심볼릭 링크면 거부하고, 기록은 임시 파일에 쓴 뒤 이름을 바꾼다. 기록이 없거나 깨지면 hook 은 아무것도 막지 않는다. 관문 강제(SPEC D-8)는 사용자 결정을 기다린다.
