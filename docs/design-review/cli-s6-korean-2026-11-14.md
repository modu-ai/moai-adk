# CLI 디자인 S6 — 한국어 도움말·사람용 메시지 설계 (2026-11-14)

Issue #2825. 바탕: #2799 D2(6864990761), 리더 확정 D3(6865936070, 결정 4(a) "영어 먼저, 한국어는 S6"), 제어 문자 결정(6867586011), clap `unicode` 판단을 S6 으로 미룬 결정(6867908522). 이 문서는 **설계만** 다룬다(코드 변경 없음, 리더 결정 D1). 새 crate 는 기본값으로 넣지 않는다(리더 결정 D2). 측정 기준: tip `08c4bb8582`(S2·S5 병합 뒤)와 열린 PR head — S1a #2824 `f72bc45e33`, S3 #2821 `a19db6c94d`, S4 #2818 `c076b21e52`.

## 0. 한 줄 요약

한국어는 **사람 화면(터미널)에서만** 나온다. stdout JSON 봉투, hook stderr, 상태줄, 그리고 파이프로 가는 모든 출력은 지금처럼 영어 바이트 그대로다. 언어는 `MOAI_LANG` → 프로젝트 프로필 `language.user_chat` → `LC_ALL`/`LC_MESSAGES`/`LANG` → 영어 순서로 고른다. 문구는 코드 안 표(새 crate 0)에 두고, 메시지마다 영어·한국어를 한 자리에 묶어 **빠진 번역이 컴파일 단계에서 드러나게** 한다.

## 1. 무엇을 옮기고 무엇을 두나

| 출력면 | 한국어로 | 근거 |
| --- | --- | --- |
| 도움말: 루트 묶음 제목·한 줄 설명·끝 안내(S1a `src/help.rs`) | 예 | 사람이 읽는 면 |
| 도움말: 하위 명령 긴 설명·인자 설명·예시(clap doc comment) | 단계적(S6e) | 149개 명령 × 여러 칸이라 양이 크다(3절) |
| clap 이 만드는 낱말(`Usage:`·`Arguments:`·`Options:`·`Print help`·`error:` 줄) | 단계적 | clap 의 `help_template`·`next_help_heading` 으로 머리는 바꿀 수 있지만 오류 문장은 clap 안에서 만들어진다(5절) |
| 오류 틀의 이름표와 `next:` 문장(S3 `ui::error_frame`·`error_next`) | 예 | 사람용 층 |
| 연산 오류 **메시지 본문**(소스의 `Err("…")` 833개) | **아니오**(S6 범위 밖) | 같은 문장이 봉투 `error.message` 에 실린다 — 봉투는 번역하지 않음(D3). 사람 화면에서는 영어 본문 + 한국어 이름표·다음 행동 |
| 요약 표 머리·합계·다음 행동(S4 `ui_summary::render`) | 예 | |
| 확인 질문·진행 줄(S5 `progress`) | 예 | |
| 토큰 단어 `ok`/`warning`/`error`/`next`(S2 `Token::word`) | 예 | 색 없이도 뜻을 전하는 단어라 같은 언어여야 함 |
| 배너 정체 줄 `◆ MoAI-ADK …`, `MoAI-ADK <버전>` | 아니오 | 제품 이름 |
| stdout JSON 봉투, `--version` 줄, hook stderr, 상태줄 | **아니오** | 기계가 읽는 바이트(D3·결정 2(a)) |
| 파이프로 가는 stderr·도움말 | **아니오(결정 S6-2)** | 레인·시험·스크립트가 영어 문장을 읽는다 |

## 2. 언어 고르기

### 2.1 순서(제안 기본값)

1. `MOAI_LANG`(값 `en`·`ko`; 그 밖의 값은 무시하고 다음 단계) — 한 번의 실행에서 바꾸는 명시 선택.
2. 프로젝트 프로필 `language.user_chat`(`src/language.rs` `ITEMS` 의 "사용자 대화") — 이 프로젝트에서 사용자가 고른 대화 언어. CLI 의 사람용 문장은 사용자에게 하는 말이라 이 항목이 뜻에 맞는다.
3. `LC_ALL` → `LC_MESSAGES` → `LANG`(POSIX 메시지 범주 순서, 비어 있는 값은 건너뜀)의 앞 두 글자가 `ko`.
4. 영어.

- 카탈로그는 `en`·`ko` 두 벌. 프로필이 `ja`·`zh`(허용 코드)를 고르면 영어로 쓰고, 경고는 하지 않는다(사람 화면에 매번 경고를 찍지 않음 — 언어 설정 화면 `language-setup` 이 이미 알려 준다).
- **어느 출력에나 같은 판정 하나**: `ui` 에 `fn ui_lang(stream) -> Lang` 을 두고, 그 스트림이 터미널이 아니면 언제나 `En`(결정 S6-2).
- 프로필을 찾는 곳: 실행한 폴더에서 `state_root` 로 프로젝트 상태 폴더를 찾고 기존 `read_profile` 로 읽는다. 프로젝트 밖(HOME 의 `moai update` 등)에서는 2단계를 건너뛴다. 읽기 실패는 조용히 다음 단계(사람 화면의 언어 때문에 명령이 실패하면 안 됨).

### 2.2 갈림길 — 리더 결정 요청

| # | 질문 | 선택지 | 기본안 | 되돌림 |
| --- | --- | --- | --- | --- |
| S6-1 | 프로필의 어느 항목을 읽나 | (가) 기존 `user_chat` (나) 새 항목 `cli_messages` — `ITEMS`·질문·검증·관리 구간 문안이 늘고 프로필 스키마가 바뀜 (다) 프로필을 읽지 않음(환경만) | (가) | 예 |
| S6-2 | 파이프로 가는 출력도 한국어로 하나 | (가) 터미널일 때만 한국어, 파이프는 늘 영어 (나) 파이프도 선택 언어 | (가) — 레인·시험이 영어 문장을 grep 하고, D2 1절의 "사람용 꾸밈은 터미널에서만"과 같은 규칙 | 예 |
| S6-3 | 명시 선택 수단 | (가) 환경 `MOAI_LANG` 만 (나) 전역 플래그 `--lang` 도 — S1 의 전역 플래그(`--color`)와 함께 149개 도움말 전부에 한 줄 | (가), 플래그는 S1 이 `--color` 를 달 때 같이 판단 | 예 |
| S6-4 | 한국어 확인 질문의 답 | (가) `y`/`yes` 만(지금과 같음), 질문 문장만 한국어 `[y/N]` 그대로 (나) `예`·`네` 도 승인 | (가) — 한글 입력기 상태와 상관없이 같은 키, 기본 아니오 유지 | 예 |

## 3. 문구를 어디에 두나

### 3.1 선택지

| | 내용 | 새 crate | 빠진 번역 | 판단 |
| --- | --- | --- | --- | --- |
| (가) 코드 안 표 | `src/ui_text.rs` 에 `enum Msg { … }` 와 `fn text(msg, lang) -> &'static str` 한 `match` — 메시지마다 `(영어, 한국어)` 가 한 줄에 같이 있음 | 0 | **컴파일 오류**(`match` 가 모든 변형을 요구) | 기본안 |
| (나) 카탈로그 파일 + 우리 파서 | `locales/en.toml`·`ko.toml` 을 `include_str!` 로 넣고 시작 때 읽음 | 0 (`toml_edit` 이미 있음) | 시험으로만 잡힘(실행 때) | 번역가가 코드 없이 고칠 수 있음 — 지금 번역은 레인이 하므로 이점이 작음 |
| (다) Fluent / gettext | `fluent-bundle` 등 | 여러 개 | 실행 때 | D2(새 crate 기본 없음)와 어긋남 |

**선례**: 이미 같은 방식이 있다 — `src/language.rs` `Item { label_en, label_ko }`·`CODES`, `src/agents.rs` 관리 구간의 영어·한국어 두 벌. (가)는 그 방식을 사람용 출력 전체로 넓히는 것이다.

### 3.2 (가)의 모양

```rust
pub enum Lang { En, Ko }
pub enum Msg { SummaryAction, SummaryNote, WritesNothing, Next, StillWaiting, ConfirmRemove, /* … */ }
pub fn text(msg: Msg, lang: Lang) -> &'static str {
    use Msg::*;
    match (msg, lang) {
        (WritesNothing, Lang::En) => "writes nothing",
        (WritesNothing, Lang::Ko) => "쓰는 것 없음",
        // …
    }
}
```

- 끼워 넣는 값은 `{name}` 자리표시 한 가지만(순서가 언어마다 달라도 됨). 끼워 넣는 값은 데이터라 **Painter 를 거쳐 `ui::visible`** 로 제어 문자를 거른다(6867586011). 번역 문장 자체는 상수이고, 상수에 제어 문자가 없음을 시험으로 고정한다.
- 영어 문장은 지금 코드의 문자열을 그대로 옮긴다 — 영어 골든 바이트가 바뀌지 않는다.

## 4. 폭(한글 2칸)

- 표·자르기: S4 가 들인 `unicode-width`(0.2.2)로 이미 한글 2칸을 센다(`ui_summary::width`, `cut_middle`/`cut_end`). 한국어 머리(`동작`·`경로`·`메모`)도 같은 함수로 맞춘다.
- S1a 루트 도움말은 우리 렌더러(`help::render_groups`·`help::wrap`)가 그린다 → `wrap` 이 글자 수 대신 `unicode-width` 로 세면 된다(새 crate 0).
- 하위 명령 도움말(clap 이 그림)에 한국어 긴 설명을 넣는 때가 clap `unicode` 기능을 다시 판단할 때다(6867908522).

| # | 질문 | 선택지 | 기본안 | 되돌림 |
| --- | --- | --- | --- | --- |
| S6-5 | 하위 명령 도움말의 한국어 줄바꿈 | (가) clap `unicode`(+`unicase`, 새 crate 1) (나) 하위 명령도 `help_template` + 우리 `wrap` 으로 그림(새 crate 0, S1a 루트와 같은 방식) (다) 하위 명령 도움말은 S6 동안 영어로 둠 | 지금은 (다), S6e 착수 때 (나) | 예 |

## 5. clap 이 만드는 낱말

- 머리(`Usage:`·`Arguments:`·`Options:`): `help_template` 와 `next_help_heading` 으로 바꿀 수 있다 — S1a 가 루트에 이미 `help_template` 을 쓴다(`help::configure`).
- `-h, --help` 의 설명 `Print help`: `disable_help_flag` 뒤 같은 플래그를 우리 설명으로 다시 단다.
- clap 오류 문장(`error: the following required arguments were not provided:` 등): clap 안에서 만들어진다. S3 가 이미 사용 오류를 오류 틀로 다시 그리므로(`ui::error_frame`, 빠진 인자 이름은 `detail` 로 받음) 한국어 화면에서는 **틀의 첫 줄을 오류 종류별 우리 문장**으로 바꾸고 clap 이 준 인자 이름만 쓴다. 봉투의 `error.message` 는 그대로.

## 6. 시험

- **언어별 골든**: `tests/fixtures/ui/{en,ko}/…golden`. 한국어 골든은 `MOAI_LANG=ko` + 가짜 터미널(pty)로. 영어 골든은 지금 파일 그대로 — **회귀 없음**(같은 바이트).
- **파이프는 영어**: `MOAI_LANG=ko`·프로필 `user_chat: ko` 로도 파이프의 stdout·stderr 가 영어 바이트와 같음(결정 S6-2).
- **봉투 바이트**: 위 조합 모두에서 stdout 이 파이프·pty·언어와 상관없이 같음.
- **빠진 번역**: (가)는 컴파일이 막는다. 더해, 모든 `Msg` 를 도는 시험이 `text(m, Ko)` 가 비어 있지 않고 `{name}` 자리표시 집합이 영어와 같은지 본다.
- **폭**: 한국어 골든의 모든 줄이 `COLUMNS=60`·`80` 에서 `unicode-width` 로 폭 안(복사해야 하는 digest 줄은 예외 — S4 와 같은 규칙).
- **제어 문자**: 한국어 문장에 끼운 데이터에 ESC·BEL·C1 을 넣어도 화면에 날 바이트가 없음(S4 의 시험과 같은 모양).
- **언어 고르기**: 순서 표(2.1)를 단위 시험 표로 — `MOAI_LANG` 이 프로필을 이김, 빈 `LC_ALL` 은 건너뜀, `ja` 프로필은 영어, 프로필 읽기 실패는 조용히 다음.

## 7. S1–S5 사람용 문구 위치 (AC-2)

| 단계 | 파일·함수(줄은 측정 기준 ref) | 문구 |
| --- | --- | --- |
| S1a (#2824, `f72bc45e33`) | `src/help.rs` `GROUPS`(20행~) 묶음 제목·명령 한 줄 설명, `LAUNCHER_LINE`(212), `render_groups`(251) 끝 안내 두 줄, `configure`(325)의 `help_template`(346) | 루트·`help --all` 화면 |
| S1 공통 | `src/main.rs` 명령 정의의 doc comment(`about`·`long_about`·인자 설명), clap 기본 낱말 | 하위 명령 도움말 |
| S2 (병합) | `src/ui.rs` `Token::word`(270–272: `warning`·`next` 등) | 상태 단어 |
| S2 (병합) | `src/banner.rs` `IDENTITY_REST`(49)·`render_line`(132) | 옮기지 않음(제품 이름) |
| S3 (#2821, `a19db6c94d`) | `src/ui.rs` `error_frame`(340: `error:`·`cause:`·`next:`), `src/error_next.rs` `NEXT_PHRASES`(13)·`next_action`(39–44: `run \`… --help\` to see …`) | 오류 틀 |
| S4 (#2818, `c076b21e52`) | `src/ui_summary.rs` `render`: `ACTION`(311)·`NOTE`(321)·이름 칸 머리(`PATH`/`HOST`/…)·`… N more`(345)·`writes nothing`(357)·`done`(360)·`next:`(372)·`<same arguments>`(374) | 요약 표 |
| S5 (병합) | `src/progress.rs` `ask`(102: `[y/N]`, 답 108)·`StateNotes::observe`(85: `still waiting:`) | 질문·진행 |
| S5 (병합) | `src/main.rs` `key-remove`(2037 `key-remove deletes 1 file`, 2041 `Remove it?`, 2044 `next: run with --yes …`)·`update`(2473 `moai-cli update preview:`, 2474 `Apply this update?`) | 확인 화면 |
| S5 (병합) | `src/index.rs` 764·769·789(3단계), `src/gate_loop.rs` 120(`checking pull request`), `src/forge/mod.rs` 304(`reading page`), `src/forge/wait.rs` 187·190(`ci on …: … runs completed`) | 진행 줄 |
| 그 밖(tip) | `src/main.rs` 1530 인자 없는 실행의 사용법 줄 | 사용법 — S3 오류 틀과 함께 |

## 8. 다음 구현 카드 (AC-3)

순서는 의존 관계대로 — 앞 카드가 병합돼야 다음이 그 위에 선다.

| 카드 | 내용 | 기다리는 것 | 크기(대략) |
| --- | --- | --- | --- |
| S6a | `src/ui_text.rs`(`Lang`·`Msg`·`text`), `ui::ui_lang`(2.1 순서, 터미널 아니면 영어), 시험(순서 표·빠진 번역·자리표시·파이프 영어). 화면 변화는 `MOAI_LANG=ko`/프로필 `ko` + 터미널일 때만 | S6-1~S6-3 결정 | 작음(~300줄) |
| S6b | S2·S4·S5 문구를 `text()` 로: 토큰 단어, 요약 표 머리·합계·다음, 질문·진행 줄. 한국어 골든 | S6a, #2818 병합 | 중간 |
| S6c | S3 오류 틀 이름표·`next:` 문장·사용 오류 첫 줄(5절) | S6a, #2821 병합 | 작음 |
| S6d | S1a 루트·`help --all` 묶음 제목과 한 줄 설명 30+개, `help::wrap` 을 `unicode-width` 로 | S6a, #2824(S1a)·S1b 병합 | 중간 |
| S6e | 하위 명령 도움말(긴 설명·인자·예시)을 루트 묶음 명령부터. S6-5 (나)로 그리기 | S6d, S6-5 결정 | 큼 — 묶음별로 쪼갬 |

## 9. 지키는 것 (모든 S6 카드 공통)

- stdout JSON 봉투·`--version`·hook stderr·상태줄 바이트 불변.
- 파이프로 가는 출력은 영어(결정 S6-2 의 기본안).
- 끼운 데이터는 Painter → `ui::visible`.
- 영어 골든 회귀 없음, 한국어 골든은 pty 로.
- 새 crate 0(S6-5 (가)를 고르지 않는 한).
