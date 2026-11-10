# 증분 색인의 generation 보존 계약

대상: #2182(C6), `index-build`·`index-query`. watcher·실호스트 설치·generation 정리는 포함하지 않는다.

- generation 키는 base commit과 정렬한 추적 파일 표(path·digest·byte 수)의 내용으로만 정한다. `built_at`과 현재 시각은 키에 넣지 않는다. 같은 내용의 실제 CLI 빌드를 1.1초 간격으로 실행해도 ID·포인터·기존 manifest 바이트가 같고 `action: unchanged`여야 한다. 이 시험은 초 단위 시각을 키에 넣는 M5s 회귀도 검출한다.
- `current.json`의 `generation_id`는 경로에 결합하기 전에 `^gen-[0-9a-f]{16}$` 형식을 만족해야 한다. 절대 경로·상위 이동·길이 불일치·대문자·비16진수 이름은 `unavailable`과 `index pointer unreadable`로 보고한다. 질의는 포인터나 외부 파일을 변경하지 않는다.
- 기존 generation 디렉터리는 재사용 전에 manifest를 해석하고 schema version·generation ID·파일 수 상한·현재 base commit·전체 파일 표를 확인한다. 손상된 manifest를 `unchanged`로 승인하거나 덮어쓰지 않는다. 실패하면 기존 포인터를 유지하고 실패 표식을 남겨 `partial`을 보고한다(REQ-C6-01). manifest를 복구한 뒤 정상 빌드하면 실패 표식이 해소된다.
- `git rm` 없이 추적 파일을 파일시스템에서 지웠어도 `deleted`로 처리한다. 실제 파일 부재와 권한·읽기 실패를 구분하며, 삭제 뒤에도 정상적으로 다시 빌드할 수 있다.
- 심볼릭 링크는 `symlink-sha256:` 이름공간으로 target 문자열을 기록한다. 일반 파일의 `sha256:`와 구분되므로 같은 바이트 길이·문자열이어도 파일→링크 전환은 `modified`가 되고 generation이 바뀐다. 이전 generation은 수정하지 않는다.
- 모든 빌드 쓰기는 기존 `apply::target`의 대상·상위 symlink 거부 규칙을 따른다. 실패 표식도 예외가 아니며, 안전하지 않은 표식/상위 경로를 통해 소스나 외부 파일을 덮어쓰지 않는다. 새 staging 파일은 `create_new`로 생성한다.
- 동일 내용의 동시 게시가 rename에서 경합하면 먼저 게시된 manifest의 schema·ID·base·전체 파일 표를 검증한 뒤 `unchanged`로 재사용한다. 다른 내용이나 손상된 winner는 실패로 남긴다. 이번 staging은 orphan으로 보존하며 자동 삭제하지 않는다.
- 읽기는 문자열 경로의 `lstat` 뒤 같은 경로를 다시 여는 방식에 의존하지 않는다. project directory handle에서 각 상위 구성 요소를 no-follow로 열고, 최종 파일도 no-follow로 연 뒤 그 handle의 metadata를 검사하고 같은 handle에서 hash를 계산한다. 정적 상위 symlink는 `deleted`이고, 열린 상위 폴더가 rename·symlink로 교체돼도 바깥 파일을 열지 않는다. leaf symlink의 target 문자열 digest는 유지한다.
- 추적 파일이 일반 디렉터리로 바뀌면 삭제로 처리한다. Git index의 mode가 `160000`인 항목만 submodule commit으로 기록하며, FIFO 등 지원하지 않는 특수 파일은 열지 않고 실제 원인으로 실패한다.
- 경계는 추적 파일 16,384개, 개별 파일 16,777,216 bytes, Git stdout 4,194,304 bytes다. 각 N은 허용하고 N+1은 거부한다. 추적 파일 수 초과는 `invalid_input`·`retryable: false`이며, Git stdout 초과는 command 실패가 아니라 output limit 원인으로 보고한다.
- untracked 목록도 16,384개를 넘으면 조용히 자르지 않는다. 일반 질의는 `unavailable`·`changes: null`, 기존 실패 표식이 있는 질의는 `partial`을 유지하되 알 수 없는 변경 목록을 빈 배열로 대신하지 않는다.
- 원자 포인터 회귀는 기존 포인터의 hard link가 이전 바이트를 유지하고 새 경로가 새 generation을 가리키는지 확인한다. 제자리 덮어쓰기와 copy-overwrite는 이 reader 보존 계약을 위반한다.
- pointer·manifest·실패 표식도 directory-relative no-follow IO를 사용하고 일반 파일인지 확인한다. Unix에서는 FIFO 교체 때도 open이 기다리지 않도록 `O_NONBLOCK`을 사용한다. cache JSON은 33,554,432 bytes까지 읽으며 N+1은 거부한다. 읽을 수 없는 실패 표식을 무시하고 `fresh`로 승인하지 않는다.

시험: `tests/index_cli.rs`의 `identical_rebuild_across_seconds_keeps_generation_and_manifest`, `pointer_generation_id_is_validated_before_path_resolution`, `damaged_reusable_manifest_keeps_previous_pointer_and_marks_partial`. 기존 freshness·원자 포인터 시험도 유지한다.

기존 읽기 권한 실패 fixture는 Unix permission bit를 사용하므로 `cfg(unix)`로 한정한다. Windows 빌드 job이 이 Unix 전용 준비를 컴파일하지 않게 하며, Windows에서 해당 실패 경로를 실행했다는 뜻은 아니다.

AC-1 FRESHNESS fixture의 권위는 리더가 지정한 #2346(C10a)이다. 이번 감사 수리는 위 경계에 한정하며, C10a 병합 뒤 fixture 정합은 후속으로 남긴다. 이 문서나 로컬 시험 통과가 C10a 연동·독립 감사·다른 OS 실행을 증명하지 않는다.

## Handle 기반 IO dependency 결정 — 2026-11-10

감사 6765650536의 실제 parent-swap 재현 때문에 std의 path-based `File::open`은 채택하지 않는다. 이미 lock에 있는 `rustix 1.1.4`의 filesystem API는 Unix용이어서 그것만으로 Windows 경로까지 같은 모델로 구현하지 않는다. `cap-std =4.0.3`와 `cap-fs-ext =4.0.3`를 정확히 고정해 directory-relative IO와 component별 `open_dir_nofollow`·leaf `follow(No)`를 사용한다. Unix의 `O_NONBLOCK` 상수에는 이미 lock에 있던 `libc =0.2.189`를 direct target dependency로 선언한다. 자체 unsafe syscall·플랫폼 숫자 상수·경로 재검사 fallback을 만들지 않기 위한 선택이다.

공식 API를 읽고 2026-11-10 확인한 source와 crates.io 원본 package checksum:

| source | version / checksum | 채택 근거 |
| --- | --- | --- |
| [cap-std Dir](https://docs.rs/cap-std/4.0.3/cap_std/fs/struct.Dir.html), [File](https://docs.rs/cap-std/4.0.3/cap_std/fs/struct.File.html), [registry metadata](https://crates.io/api/v1/crates/cap-std/4.0.3) | 4.0.3 / `c1ec78e242cfa2cfe276807ac2ecc00315a6c97786977414bcd1c3963b6c91b8` | 열린 directory authority와 handle metadata/Read, std File 변환 |
| [DirExt](https://docs.rs/cap-fs-ext/4.0.3/cap_fs_ext/trait.DirExt.html), [OpenOptionsFollowExt](https://docs.rs/cap-fs-ext/4.0.3/cap_fs_ext/trait.OpenOptionsFollowExt.html), [registry metadata](https://crates.io/api/v1/crates/cap-fs-ext/4.0.3) | 4.0.3 / `56ff379b70af8e08307a8f65e7040c7301cb4a572538ade16b4984f0da77847f` | component/leaf no-follow 옵션 |

두 package는 edition 2021, `Apache-2.0 WITH LLVM-exception OR Apache-2.0 OR MIT`이고 조회 시 yanked가 아니다. registry에 MSRV가 선언되지 않아 compiler 적용 여부는 실제 빌드 증거로 기록한다. Context7은 cap library를 정확히 resolve하지 못해 공식 rustdoc/registry를 읽었으며, Rustix의 Unix-only filesystem 제한은 해당 API 문서와 대조했다. Windows도 같은 capability/no-follow API 경로를 사용하되 Unix `O_NONBLOCK`만 cfg로 분리한다. Windows runtime은 여기서 측정하지 않았고, CI compile과 runtime 증거를 구분한다.

state byte cap은 초기 검토의 8 MiB가 아니라 32 MiB다. 허용되는 4 MiB Git 경로 표는 JSON control-character escaping 때문에 최대 6배로 커질 수 있고 파일별 metadata도 추가된다. 정상 producer가 만든 manifest·실패 reason을 8 MiB guard가 새로 거부하지 않도록 이 여유를 둔다. 실제 literal 33,554,432/33,554,433 경계와 ±1 변이는 별도 검증 대상이다.

## F3·state FIFO 수리의 실제 측정

Rust `1.94.1 (e408947bf)`·`aarch64-apple-darwin`에서 새 dependency를 포함해 실제 컴파일했다. `cargo test --offline --locked --lib index::handle_io_tests`는 3/0, `cargo test --offline --locked --test index_cli`는 22/0이다(`artifact://315`). parent-open 전과 열린 parent handle 뒤 leaf-open 전의 교체를 각각 재현하며, repository 내부/외부 symlink와 두 깊이의 parent를 검사한다. leaf의 metadata/open 사이 교체도 거부한다. FIFO CLI 회귀는 자기 child에 8초 deadline을 두고 pointer·marker·manifest query, manifest 재사용 build와 실패 marker 쓰기를 검사한다.

실제 binary에 builder 3개를 동시에 실행하고 parent를 원래 폴더/외부 symlink로 반복 교체했다. 각 half-cycle에 10 ms를 요청했으며 실제 완료 cycle 수는 아래와 같다. 반환된 generation뿐 아니라 게시된 모든 manifest도 검사했다. timeout·잘못된 JSON·probe error는 각 실행에서 모두 0이다.

| 요청 시간(s) | 실제 시간(s) | build 시도 | 성공 | 거부 | 교체 cycle | 외부 digest 반환/게시 |
| --- | --- | --- | --- | --- | --- | --- |
| 30 | 30.057 | 848 | 847 | 1 | 1,029 | 0 / 0 |
| 20 | 20.127 | 601 | 601 | 0 | 695 | 0 / 0 |
| 20 | 20.128 | 327 | 326 | 1 | 434 | 0 / 0 |

Windows backend는 `cap-primitives 4.0.3`의 handle-relative `NtCreateFile` 경로(`OBJECT_ATTRIBUTES.RootDirectory`)와 no-follow reparse-point 옵션을 사용한다. Unix의 `O_NONBLOCK`은 Windows 코드에 포함되지 않는다. Windows runtime 검증은 `not_run`이며, cross-target compile 결과와 혼동하지 않는다. 이 문서의 로컬 증거는 독립 감사가 아니다.

`cargo check --offline --locked --all-targets --target x86_64-pc-windows-msvc`는 설치돼 있던 Windows standard library로 실행해 exit 0이다(`artifact://322`, 49.97 s). 이는 MSVC target 전체 source의 type/compile check이며 Windows binary의 link·runtime 시험은 아니다. 기존 `claim_verify_cli`·`forge_cli`·`git_env_scrub`·`sources_cli`의 cfg 관련 unused/dead-code warning은 남아 있으며 범위 밖 파일을 고치거나 warning을 억제하지 않았다.

### 변이 검증과 원복

초기 parent-follow 변이는 absolute symlink가 capability 경계 자체에서 거부돼 살아남았다. 따라서 내부 alias fixture를 현재 parent handle 아래의 relative symlink로 강화했고, 정상 baseline 3/0을 다시 확인한 뒤 아래 변이를 각각 독립 적용했다. production 환경변수 hook은 없다.

| 변이 | 선택 / 실패 | exit |
| --- | --- | --- |
| ambient `File::open(project.join(p))`로 회귀 | 1 / 1 | 101 |
| parent `open_dir_nofollow` → `open_dir` | 1 / 1 | 101 |
| leaf `follow(No)` → `follow(Yes)` | 1 / 1 | 101 |
| `O_NONBLOCK` → 0 | 3 / 2(query·실패 marker write deadline), 기존 source FIFO 1 통과 | 101 |
| manifest file cap 16,385 | 1 / 1 | 101 |
| manifest file cap 16,383 | 1 / 1 | 101 |
| state byte cap 33,554,433 | 1 / 1 | 101 |
| state byte cap 33,554,431 | 1 / 1 | 101 |

각 변이 뒤 원래 source를 복구했다. 최종 `src/index.rs` SHA256은 `174116dd24262ae7a7915663e37216115435831c055186c7202cab6414635609`로 확인했다. literal cache-read 경계는 manifest 16,384/16,385개와 세 state 파일 각각 33,554,432/33,554,433 bytes다. marker producer도 같은 byte 경계에서 초과 쓰기가 이전 실패 표식을 보존하는지 실제 query로 검사한다.
