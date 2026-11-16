# 원장 폴더: 소유자 전용 모드와 링크 거절 (Issue #3068 보강 C3·C4)

코드 `src/ledger/folders.rs`, 호출하는 곳 `src/ledger/open.rs`·`command.rs`·`recover.rs`. 원장 파일은 `<HOME>/.moai/projects/<project_id>/ledger.db` 이고, 그 위의 세 폴더 `<HOME>/.moai`·`projects`·`<project_id>` 가 여기서 다루는 대상이다.

## 소유자 전용 (C3)

- 이 코드가 **만드는** 폴더는 모드 0700 이다. 한 번에 여러 단계를 만들어도 한 단계씩 만들고, `mkdir` 뒤에 정확히 0700 으로 `chmod` 한다(umask 가 넓히지도 좁히지도 못한다). `<HOME>/.moai` 는 없을 때만 만든다. 이미 있는 `<HOME>/.moai` 는 다른 도구의 것이라 **절대 chmod 하지 않는다**.
- 이미 있는 `projects`·`projects/<id>` 가 진짜 폴더이고 현재 사용자(euid)의 것인데 그룹·기타 비트가 하나라도 있으면: `ledger-register-plan` 이 `would_tighten: [경로]` 로 미리 알리고(계획 digest 에 들어간다), `ledger-register-apply` 가 0700 으로 조이고 `tightened: [경로]` 로 알린다. 이미 등록된 원장이 있어도(`unchanged`) 같은 조임이 일어난다.
- 그런 폴더가 **다른 사용자의 것**이면 chmod 하지 않는다. 계획은 `action: blocked`, `problem: folder_not_ours` 이고 적용은 `folder_not_ours` 로 거절한다.
- `ledger-status` 는 읽기만 하며 `folder_modes: {"projects": "0755", "project": "0700"}`(8진 글자, 폴더가 없으면 null)을 보인다.
- 복구(`ledger-recover-*`)는 폴더를 만들지 않는다. `.recover.lock`·`.recovering`·새 원장은 0600 이다(남긴 `.damaged-*` 는 원래 파일의 모드).
- 새 원장 임시 파일 `.creating-*` 은 SQLite 가 한 바이트를 쓰기 전에 0600 으로 먼저 만든다.

## 링크 거절 (C4)

`.moai`·`projects`·`<id>` 중 하나가 심볼릭 링크이면(끊어진 링크도) 모든 원장 호출이 `path_link`(`ledger blocked: path_link: …`)로 거절하고, 링크 너머로는 읽지도 쓰지도 만들지도 않는다. 대상: `open_read_only`·`open_write`·`register`·`ledger-status`·`ledger-register-plan`/`-apply`·`stale_temp`와 복구의 `plan`·`apply`·`lock_recovery`·`build_recovering`·`set_aside_damaged`·`install_recovered`. 복구의 `target_state` 는 링크를 `unconfirmed`(손상 아님)로 분류해 `target_not_damaged` 로 거절한다.

HOME 자체와 그 위 폴더는 링크여도 된다(macOS 의 `/tmp`·`/var`). 검사하지 않는다.

- 계획: `action: blocked`, `problem: path_link`. `ledger-status`: `state: unavailable`, `problem: path_link`, 아무것도 읽지 않는다.
- 만들기·바꾸기 걸음(`register`, 복구의 각 걸음)은 `lstat` 만이 아니라 각 폴더를 `O_NOFOLLOW | O_DIRECTORY` 로 열어 `fstat` 의 (dev, inode) 가 새 `lstat` 과 같은지 다시 본다. 검사와 사용 사이에 바뀐 폴더는 `path_link` 이다.

### 남는 경쟁 (문서로 남긴다)

SQLite 는 데이터베이스를 **경로로** 연다. 마지막 확인 뒤, SQLite 의 `open`(또는 새 파일을 제자리에 놓는 하드 링크) 앞에 같은 사용자(또는 root)가 폴더를 링크로 바꾸면 이 코드는 보지 못한다. 폴더가 0700 이라 다른 사용자는 그 바꿔치기를 할 수 없다. 이 틈을 완전히 닫으려면 디렉터리 fd 기준 `openat` 이 필요한데 SQLite 가 지원하지 않는다.

## 새 낱말

| 낱말 | 뜻 |
| --- | --- |
| `path_link` (`LedgerError::PathLink`) | `.moai`·`projects`·`<id>` 가 링크(끊어진 것 포함)이거나 단계 사이에 바뀌었다. 종류 `blocked`. |
| `folder_not_ours` (`LedgerError::FolderNotOurs`) | `projects`·`projects/<id>` 가 다른 사용자의 것이다. 종류 `blocked`. |
| `lock_not_regular` (복구 거절 낱말) | `.recover.lock` 자리가 일반 파일이 아니다(파이프·폴더·장치·소켓·링크). 열지 않고 거절. |

## 일시 I/O 오류는 손상이 아니다 (R2)

- 헤더 100바이트 읽기: 짧으면(`UnexpectedEof`) 손상(`Unavailable`), 그 밖의 읽기·열기 오류는 `Io`.
- `meta.project_id` 조회: 표가 없다·행이 없다·데이터베이스가 아니다·깨졌다는 손상(`Unavailable`), 잠김은 `Busy`, 나머지는 `Io`.
- 복구의 `target_state` 는 `Io` 를 `unconfirmed` 로 분류하고 `target_not_damaged` 로 거절한다. 바꿔 끼우지 않는다.
