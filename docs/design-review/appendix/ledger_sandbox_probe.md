# 원장 위치 샌드박스 쓰기 측정 (`tools/sandbox_write_probe.py`)

Issue #3068 P8 · AC9: "원장 위치가 샌드박스 안에서 쓰기 가능한지 측정한 기록이 있다." 이 문서는 도구, 실행 명령, 실제로 잰 결과, 그 결과가 말해 주지 않는 것을 적는다. 실제 `~/.moai` 에 대한 측정은 하지 않았다(아래 "실제 HOME 측정은 하지 않았다").

## 무엇을 재나

원장은 `<HOME>/.moai/projects/<project id>/ledger.db`(SQLite, `journal_mode=DELETE`)를 쓰고, 손상된 사본을 하드 링크로 남기고, `.recover.lock` 을 `flock` 으로 잠그고, 파일을 rename 한다. 샌드박스(`codex sandbox`, 감사 sub-agent)가 이 파일 연산을 막으면 원장은 열리지 않는다. 도구는 같은 연산을 호출자가 고른 임시 폴더 안에서 한 단계씩 해 보고 단계마다 성공 여부와 errno 를 JSON 으로 낸 뒤 만든 것을 모두 지운다.

| 단계 | 하는 일 |
|---|---|
| `mkdir_tree` | `<base>/.moai/projects/<id>` 를 모드 0700 으로 만든다 |
| `create_file` | `ledger.db` 를 `O_CREAT\|O_EXCL`, 모드 0600 으로 만든다 |
| `sqlite_write` | 그 파일을 SQLite 로 열어 `PRAGMA journal_mode=DELETE`, 표 만들기, 행 넣기, commit |
| `sqlite_read` | 새 연결을 `mode=ro` URI 로 열어 행을 읽는다 |
| `hard_link` | `ledger.db` → `ledger.db.damaged-probe`, 링크 수 2 확인 |
| `rename` | `ledger.db.next` 를 만들고 `ledger.db.recovering` 을 그 위로 rename(덮어쓰기), 내용 확인 |
| `advisory_lock` | `.recover.lock` 을 만들고 `flock(LOCK_EX\|LOCK_NB)` 후 해제 |
| `fsync_dir` | 프로젝트 폴더를 읽기 전용으로 열어 `fsync` |
| `cleanup` | 만든 파일을 하나씩 `unlink`, 폴더를 하나씩 `rmdir`(재귀 삭제 없음). 못 지운 것은 `left_behind` |

앞 단계가 실패하면 그 결과가 필요한 단계는 `ok: false`, `errno: null`, `error: "skipped: …"` 로 적고, 나머지는 계속한다. `cleanup` 은 항상 돈다. 도구가 실행됐으면 일부 단계가 실패해도 exit 0 이고, 거절(exit 2)이면 아무것도 만들지 않는다. 환경 변수를 읽지 않고(홈은 `pwd` 의 uid 항목), 네트워크를 쓰지 않는다. 거절 낱말과 JSON 모양은 `tools/sandbox_write_probe.py` 머리말이 정본이고 `tests/sandbox_write_probe_test.py` 가 지킨다.

## 명령

```
# 비교군: 샌드박스 없이
B=$(mktemp -d /tmp/p8probe.XXXXXX)
/usr/bin/python3 -I tools/sandbox_write_probe.py --base "$B"; rmdir "$B"

# 샌드박스 안에서 (프로파일 이름은 :workspace, :read-only)
B=$(mktemp -d /tmp/p8probe.XXXXXX)
codex sandbox -P :workspace -C "$B" -- /usr/bin/python3 -I "$PWD/tools/sandbox_write_probe.py" --base "$B"
rmdir "$B"

# 거부된 연산의 이름까지 보려면
codex sandbox -P :read-only --log-denials -C "$B" -- /usr/bin/python3 -I "$PWD/tools/sandbox_write_probe.py" --base "$B"
```

`--base` 는 비어 있는 새 폴더여야 한다(`.moai` 가 이미 있으면 거절). 심볼릭 링크, `/`, 실제 홈과 그 `.moai`·`.moai-cli`·`.claude`·`.codex`·`go` 폴더도 거절한다.

**감사 sub-agent 의 경우.** sub-agent 는 스스로 `codex sandbox` 를 실행할 수 없다. 부모 세션이 같은 스크립트를 같은 명령으로 돌리고 JSON 을 sub-agent 의 보고에 붙여 넣는다. sub-agent 자신의 샌드박스를 잰 것이 아니라 같은 프로파일 이름으로 `codex sandbox` 아래에서 잰 것이다.

## 측정 결과

날짜 `date -u` 2026-11-16 05:33Z–05:35Z(비교군 05:33:46Z) · macOS 27.0.1 (26A434) · `codex-cli 0.169.0` · `/usr/bin/python3` 3.9.6 · uid 501 · 기준 폴더는 매번 `mktemp -d /tmp/p8probe.XXXXXX` 로 만든 빈 폴더이고, 끝난 뒤 `rmdir` 했다(모두 성공, 즉 빈 폴더였다). 스크립트는 작업 폴더 `tools/` 의 것을 그대로 읽었다(`:read-only` 도 스크립트 경로를 읽을 수 있어 복사하지 않았다).

| 실행 | `-C` | exit | `all_ok` | 실패 단계 | `left_behind` |
|---|---|---|---|---|---|
| 비교군(샌드박스 없음) | - | 0 | true | 없음 | `[]` |
| `-P :workspace` | 기준 폴더 | 0 | true | 없음 | `[]` |
| `-P :workspace` | 이 worktree | 0 | true | 없음 | `[]` |
| `-P :read-only` | 기준 폴더 | 0 | **false** | `mkdir_tree` errno 1 (`Operation not permitted`) | `[]` |
| `-P :read-only --log-denials` | 기준 폴더 | 0 | **false** | `mkdir_tree` errno 1 | `[]` |
| `-P :danger-no-sandbox` | - | 돌리지 않음 | - | - | - |

- `:read-only` 에서 `mkdir_tree` 가 EPERM 으로 실패하고, 그 결과가 필요한 7단계는 `skipped: …`(errno null), `cleanup` 은 ok 였다. 도구는 죽지 않았고 exit 0 이었다.
- `--log-denials` 는 거부한 연산을 이름으로 댔다. 이 도구가 만든 것은 `file-write-create /private/tmp/p8probe.n2GeO9/.moai` 한 줄이고, 나머지 줄(`sysctl-read kern.bootargs`, `mach-lookup com.apple.logd`, `file-write-data /dev/dtracehelper` 등)은 파이썬 시작 소음이다. `:read-only` 의 stderr 에는 `python3: warning: confstr() failed … DARWIN_USER_TEMP_DIR` 경고도 나왔다(결과와 무관).
- `:danger-no-sandbox`: `codex sandbox --help` 가 `-P` 에 쓸 수 있는 이름을 나열하지 않으므로 돌리지 않았다.
- 모든 실행의 `python` 은 3.9.6, `writes_outside_base` 는 false(도구가 `<base>` 밖을 쓰지 않는다는 구조상의 사실이며, 샌드박스가 그것을 검증해 주지는 않는다).

## 한계

- `:workspace` 프로파일에서 `/tmp` 아래 임시 폴더는 쓰기 가능 범위 안이다(`-C` 를 기준 폴더로 하든 이 worktree 로 하든 같았다). 그래서 위 통과가 말하는 것은 **원장이 쓰는 파일 연산(SQLite, 하드 링크, rename, flock, fsync)이 샌드박스 안에서 동작한다**는 것뿐이다. **실제 `<HOME>/.moai` 가 쓰기 가능하다는 뜻이 아니다.**
- `:read-only` 의 실패는 "쓰기 가능 범위 밖에서는 첫 `mkdir` 부터 막힌다"는 것을 보여 줄 뿐, `:workspace` 에서 `<HOME>/.moai` 가 막히는지 아닌지는 보여 주지 않는다.
- 이 기계 한 대, 이 `codex` 버전, 이 프로파일 이름 두 개의 측정이다. 다른 호스트(Claude Code 샌드박스, 감사 sub-agent 의 별도 환경)의 값은 아니다.

## 실제 HOME 측정은 하지 않았다 (Q4, 보호됨) — NOT RUN

실제 `~/.moai/projects/.probe-*` 에 쓰는 측정은 계획(Issue #3068 계획 댓글 6896939813, AC9·Q4)에서 사용자 확인이 있어야 하는 보호 항목이다. 이번 작업은 실행하지 않았다. 이 도구는 실제 홈과 그 `.moai` 를 가리키면 `real_home` 으로 거절하므로(설계상), 그대로는 그 측정을 할 수 없다. 나중에 사용자가 허락하면 쓸 최소한의 형태(프로브 폴더 하나를 만들고 바로 지운다. `.moai/projects` 가 이미 있어야 하며 없으면 ENOENT 로 끝난다):

```
python3 -I tools/home_entry_count.py --home <real home> > BEFORE.json
codex sandbox -P :workspace -C <worktree> -- /usr/bin/python3 -I -c \
  'import os,sys; p=sys.argv[1]; os.mkdir(p, 0o700); os.rmdir(p)' \
  <real home>/.moai/projects/.probe-<id>
python3 -I tools/home_entry_count.py --home <real home> --compare BEFORE.json   # exit 0 이어야 한다
```

전제: (1) 사용자가 그 홈 경로와 프로브 id 를 이름으로 확인한다, (2) `home_entry_count.py` 의 전·후 비교가 exit 0(`.moai` 바로 아래 항목 수·이름 불변)이다, (3) 프로브 폴더가 실제로 지워졌다(`.probe-<id>` 가 남으면 실패로 보고한다). 전체 단계를 실제 홈에서 돌리려면 이 도구에 명시적 옵션을 더하는 별도 변경이 먼저 필요하다. 이 문서는 그 옵션을 약속하지 않는다.

## 샌드박스가 쓰기를 막으면

원장 코드는 `create_dir_all` 이나 파일 쓰기가 실패하면 `LedgerError::Io` 를 `io` 코드로 보고한다(EACCES·EPERM 은 원문 메시지에 남는다). 대체 위치로 옮겨 쓸지, 읽기 전용으로만 동작할지, 명령을 거절할지는 이 문서가 정하지 않고 다음 단계의 Issue 가 정한다.
