# Message durability 시험 격리 (#2471)

## 범위와 관측

base는 `a2f59b960dbc0907d7b9376b8d0bb1110710a15a`이다.
[Issue #2471](https://github.com/modu-ai/moai-adk/issues/2471)의 원 실패는
run43078726787 / macos-arm64 job130725352748의 libtest247개 중
`audit_new_ancestors_synced_before_post` 한 건이다.
첫 invoke 이후 같은 고유 프로젝트를 다시 호출할 때
`Unavailable: message lock wait: another invocation holds the advisory lock`가 났다.
원 step 로그 byte SHA256은
`63e9d948febf6779da6fc4504c6bd8636e9c9fc5408b31127839b510b54bb7f6`이다.

각 시험의 TempDir는 이미 고유하고 PROBE도 thread-local이었다.
원인은 같은 이름의 임시 경로나 공유 Rust probe라는 가정으로 확정할 수 없었다.
공유되는 것은 병렬 libtest 프로세스의 open-file descriptor table이다.
이웃 시험의 자식 생성과 message lock 수명이 겹치면 자식이 lock의 열린 reference를
상속할 수 있다. 부모 File drop만으로 마지막 reference가 닫힌다고 가정할 수 없으며,
그 사이의 두 번째 try_lock은 제품의 정상 정책대로 즉시 Unavailable을 반환한다.

기존 CI 로그에는 특정 자식 PID/producer 시험 이름이 없어 그 이름은 미확정이다.
아래 제어 실험은 descriptor 상속이 실제 message fixture에서 같은 오류를 발생시키고
프로세스 격리가 그 간섭을 제거함을 입증한다. 자연 flake 재현율과 구분한다.

## 수리

`src/message.rs`의 `cfg(all(test, unix))` 여섯 case만 변경한다.
원래 case 본문과 fsync/ACK/restart 단언은 helper 함수에 그대로 보존했다.
각 named test는 현재 unit-test executable에서 exact-name 한 case를
`--test-threads=1`로 실행한다. 자식별 고유 fixture와 TLS probe는 그대로 유지한다.

부모는 child exit, 실제 1-case 선택, case 완료 marker와 부모와 다른 PID를 확인한다.
fixture 생성 시에도 해당 case의 worker marker를 요구한다.
marker 환경 변수는 Command.env로 자식에게만 전달하며 전역 set_var는 없다.
worker 안에는 다른 unit case나 process-spawning 이웃 시험이 없으므로,
새 message lock을 보유하는 descriptor table을 이웃과 공유하지 않는다.

제품 함수·try_lock 정책·대기/재시도·sleep은 변경하지 않았다.
test module 이전 제품 prefix의 byte SHA256은 base와 같다.
실험용 fork/pipe 주입 코드는 후보에 포함하지 않는다.
전역 mutex는 다른 module의 process 생성과 조율하지 못하므로 채택하지 않았다.
전체 suite 직렬화와 workflow 변경도 채택하지 않았다.

## 재현과 측정

실행 host는 native macOS arm64, Rust1.94.1이다.
격리한 HOME/TMPDIR와 이 worktree의 한 CARGO_TARGET_DIR를 사용했다.
실행 파일은 반복 시작 전에 복사하고 SHA256으로 고정했다.

| 대상 | 방법 | N | 실패 |
| --- | --- | ---: | ---: |
| base, 자연 반복 | 전체 lib247개, test-threads32 | 100 | 0 |
| kernel 수명 witness | 두 pipe로 부모 close/child reference 보유 순서 고정 | 100 | 100회 WouldBlock; child 종료 후 전부 재획득 |
| base, 실제 audit fixture 제어 | POST 중 root process에 상속 window 주입, exact case/threads32 | 100 | 100 (같은 lock wait 오류) |
| 수리본, 같은 제어 overlay | exact case/threads32, worker에서는 이웃 자체가 없는 모델 | 100 | 0 |
| 최종 후보, 자연 반복 | 전체 lib247개, test-threads32 | 100 | 0 |

자연 base100회 실패0을 원 CI 재현 성공으로 세지 않는다.
[제어 overlay 원문·실행 절차](https://github.com/modu-ai/moai-adk/issues/2471#issuecomment-6804780402)는
원본 fixture의 POST 분기에만 실험 hook을 넣고 두 pipe로 순서를 보장한다.
sleep이나 확률적 지연을 사용하지 않는다.
프로세스 격리 worker에는 이웃 시험을 실행하지 않는다는 모델이 적용된다.
제어 수리 실행 파일 SHA256:
`678a17f7fd849a105a8e74cad535fad90ce0cae9f854bfd798699d2d0b126aa0`.

빌드 실패 뒤 구 바이너리 실행과 pin이 보장되지 않은 중간 실험은 통계에서 제외했다.
자연 반복·제어 반복·kernel witness의 결과와 source snapshot은 별도 기록이다.

같은 자연 반복을 실행하는 방법:

```sh
export CARGO_HOME=/tmp/moai-product-cargo
export CARGO_TARGET_DIR=/tmp/moai-target-your-own-lane
cargo test --offline --locked --lib --no-run
# Use the src/lib.rs executable path printed by the successful build.
case_bin="<printed libtest executable>"
mkdir -p "$CARGO_TARGET_DIR/2471-repeat"
cp "$case_bin" "$CARGO_TARGET_DIR/2471-repeat/unit-tests"
shasum -a 256 "$CARGO_TARGET_DIR/2471-repeat/unit-tests"
failed=0
for n in {1..100}; do
  RUST_MIN_STACK=8388608 "$CARGO_TARGET_DIR/2471-repeat/unit-tests" --test-threads=32 \
    > "$CARGO_TARGET_DIR/2471-repeat/$n.log" 2>&1 || failed=$((failed + 1))
done
echo "N=100 failed=$failed"
test "$failed" -eq 0
```

격리 제거 변이는 fixture worker 요구에서, 잘못된 child exact filter 변이는
실제 1-case 선택 요구에서 검출됐다. 두 복사본 모두 compile0, 부모 selected1,
runtime FAIL101이다. 같은 standalone module 구성의 원본 positive control은 통과했다.
제품 source를 바꾸거나 compile error/선택0을 변이 검출로 세지 않았다.

## 공유 자원 전수 표

| 범위·고정 revision | 자원과 소비자 | 공유/간섭 판단 및 처리 |
| --- | --- | --- |
| src/message.rs durability 여섯 case | project 및 ACK sender TempDir, outbox/inbox/r1.json | case별 고유 경로. 같은 case의 반복/재시작만 의도적으로 재사용; 보존 |
| 같은 module | PROBE created/synced/fail | thread_local RefCell. process 전역 공유 아님; worker에서 초기화/기존 reset 보존 |
| 같은 module | Remote project/comments/actor/posts/calls/hidden_kind | case-local Cell/RefCell, 실제 network 호출 없음; 보존 |
| 같은 module | Time 및 환경 | 고정 Clock/no-op sleep. 전역 cwd/HOME/env 수정 없음. 새 worker marker는 자식 전용 |
| 같은 libtest process | message .lock의 open-file reference, 이웃의 git/Command 생성 | descriptor 상속 window가 실제 간섭 지점; 여섯 case 모두 새 single-case process로 격리 |
| tests/message_cli.rs (base) | 고유 temp_root/Remote, 두 advisory-lock 경계 시험 | global static/env/process spawn 없음. 의도적으로 held .lock인 시험은 즉시 Unavailable을 계속 검증; 제품 정책 보존 |
| src/dispatch.rs (S2 head9d212a9d463e3b13b05173194126e6b95158efcf) | S1 호출, Clock/Transport 인자 | base에는 아직 없는 파일. 자체 static/thread_local/test module 없음; 별도 pin으로 읽었고 이 PR에 합치지 않음 |
| tests/dispatch_cli.rs (같은 S2 head) | case별 temp_root/Remote/고정 Clock; native_cli_check_envelope의 단일 CLI spawn | 경로/상태는 고유하지만 같은 integration binary 안의 spawn은 같은 종류의 descriptor 간섭 후보. 자연 실패는 미관측. 미병합 S2 범위이므로 수정하지 않고 리더에게 후속 조율 대상으로 보고 |

부모 libtest의 다른 case 전체를 공유 경로로 단정하지 않는다.
예를 들어 claude의 version-cleanup case와 git을 실행하는 identity/index 등의
Command 생성은 가능한 producer이며, 원 CI의 특정 producer를 식별한 증거는 아니다.
S2의 잠재 간섭 후보도 이번 tip flake의 실제 producer로 주장하지 않는다.

## 공식 근거와 채택/기각

아래 원문은 2026-11-11T18:49:21–22Z에 HTTP200으로 실제 읽고 byte hash를 기록했다.

| 출처 | byte SHA256 | 채택·적용 |
| --- | --- | --- |
| [Apple XNU flock(2)](https://raw.githubusercontent.com/apple-oss-distributions/xnu/main/bsd/man/man2/flock.2) | 54195927032ace18230a4f926d9edfbb672fc8706267be00727d203a01b6d616 | dup/fork reference와 nonblocking 동작의 근거; native pipe witness로 확인 |
| [Rust1.94.1 Unix File](https://raw.githubusercontent.com/rust-lang/rust/1.94.1/library/std/src/sys/fs/unix.rs) | bdd7dbe59f56df067504141fd6c04f46601f8eb7af39c672921eb2c7e4bc12df | 이 toolchain의 try_lock→flock 적용 확인. 현재 stable 문서의 다른 버전을 대신 적용하지 않음 |
| [Rust1.94.1 Unix process](https://raw.githubusercontent.com/rust-lang/rust/1.94.1/library/std/src/sys/process/unix/unix.rs) | b1cbb22bf9cfa57be9968e46137f8289e36948289a3f55948d6aadbe422cc302 | fork/posix_spawn 경로 존재 확인; 모든 spawn이 같은 방식이라고 단정하지 않음 |
| [Rust Book: test 실행](https://doc.rust-lang.org/book/ch11-02-running-tests.html) | 5594a0101b0b307b34e3b3d8dc992f2cd36d13cd8f2eeed3ebc8fd297cd81019 | 기본 병렬·test-threads 제어 근거. blanket 직렬화 대신 case process 격리 채택 |

Linux/Windows hosted 결과와 fresh-clone 검사는 이 로컬 측정에서 not_run이다.
정확-head CI와 독립 감사는 PR에서 별도로 확인한다.
