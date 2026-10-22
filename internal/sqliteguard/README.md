# SQLite 물리 연결의 Windows 파일 보호

이 패키지는 쓰기용 homestate factory/profile lease DB와 factory backlog/GTD DB의 Windows 물리 연결에 파일 보호 경계를 추가한다. Unix `Open`은 기존 `sql.Open("sqlite", dsn)` 동작을 유지한다. 읽기 전용 factory/census 연결은 이 생성·보호 경로에 들어오지 않으며, 기존 mode=ro·생성 금지·checkpoint 금지 계약을 유지한다.

## 생성과 연결

Windows 파일은 `privatefile.OpenPinned`의 단일 `CreateFileW(OPEN_ALWAYS)`에서 보호 SA와 READ_DATA/no-DELETE-share를 함께 적용한다. CreateFile 성공 직후 검증 전에 ownership callback으로 핸들을 registry에 등록한다. 검증 오류·panic·정리 실패 때도 이 소유권을 보존한다. 새 파일에는 payload 이전에 current SID 소유의 protected DACL을 설치한다. 기존 파일은 같은 핸들에서 객체 종류·단일 링크·현재 소유자를 확인한 뒤 DACL을 좁히고 strict policy를 재조회한다. WRITE_DAC은 기존 자기 소유 객체를 좁히기 위한 접근권이며 owner/group 변경이나 privilege 활성화를 하지 않는다. 기존 파일을 truncate·replace·unlink하지 않는다.

새 관리 디렉터리만 private SA로 생성한다. 기존 leaf는 `CheckDirectory`로 검사하며 상속 DACL을 다시 적용하지 않는다. 기존 조상 디렉터리의 ACL을 수정하거나, foreign owner 디렉터리를 인수하거나, 기존 weak policy를 검사 통과를 위해 사후 복구하지 않는다.

file URI와 별도로 전달한 파일 경로가 일치하는지 확인한 뒤 parent의 8.3/case alias를 해소하고 URI와 guard 경로를 함께 정규화한다. 이 쓰기용 내부 API가 허용하는 URI query key는 `_pragma`, `_txlock`뿐이다. 다른 key와 mode=ro는 명시적으로 거부한다. 기존 일반 SQLite URI 전체와 동등한 API라는 주장은 하지 않는다.

각 physical Connect는 DB/WAL/SHM 세 객체를 먼저 보유한 뒤 원래 `sqlite.NewConnector`에 연결을 위임한다. 전역 driver 등록을 추가하지 않는다. pool에 반환하기 전에 같은 물리 연결의 public `FileControlPersistWAL("main", 1)` 결과와 readback을 검사한다. 재연결 때도 반복한다. Raw 사용자가 이후 FileControl로 설정을 변경할 수 있으므로 불변 설정이라고 주장하지 않는다.

## 종료 실패의 소유권

프로세스 registry는 정상 연결과 fault 연결 모두를 강하게 참조한다. 새 native 자원을 얻기 전에 전체 예약 수를 128개로 제한한다. 최대 384개 file pin의 상한이며, SQLite의 전체 stmt·메모리·Backup/ATTACH 목적지 핸들의 상한은 아니다. 현재 통합한 각 writable pool은 MaxOpenConns=1이며, 여러 pool은 총 128개 예약 범위에서 공존할 수 있다. 무제한 pool 동작과 동일하다는 주장은 하지 않는다.

정상 종료 경로는 **public driver.Close가 nil을 반환한 뒤에** pin을 해제한다. 이는 모든 OS 핸들 종료를 직접 확인하는 검사가 아니다. base.Connect 진입 뒤 opaque 실패, Close 오류·panic, 종료를 확인할 수 없는 Raw 자원은 명시적으로 보존한다. 이 경우 같은 canonical path의 새 admission을 거부하고 원래 error identity를 보존한다. `Status`와 안전한 로그에는 token·경로 digest·stage·보유 수만 표시한다. DSN·SQL·payload는 로그에 넣지 않는다. GC·TTL·자동 재시도·무조건 force release로 불확실한 객체를 해제하지 않는다. database/sql이 종료 오류를 버리는 경로에서도 registry와 진단은 남는다.

Raw stmt와 rows는 별도 계수한다. EOF만으로 rows를 종료 처리하지 않는다. prepared rows의 reset과 stmt finalize를 구분하고, 자식 Close는 한 번만 호출한다. raw 자식이 남으면 modernc의 close_v2/TLS 정리 전에 physical Close 자체를 보류하고 오류를 반환한다.

## 공개 표면과 명시적 한계

modernc 1.60.1의 declared Conn 메서드 24개, 표준 optional interface 9개, Stmt 메서드 6개와 Rows 메서드 8개를 보존한다. 원래 driver에 없는 NamedValueChecker 등을 추가하지 않는다. 원래 concrete conn의 embedded sync.Mutex에서 승격되는 Lock/TryLock/Unlock 세 메서드는 제공하지 않는다. 현재 저장소 사용은 발견되지 않았으며, 이를 실제 SQLite mutex인 것처럼 흉내내지 않는다.

NewBackup/NewRestore의 구체 반환형 `*sqlite.Backup`은 유지한다. 그러나 private 필드와 직접 호출되는 Finish/Commit을 계측할 수 없으므로 **진입 시점부터 sticky fault/retention**을 기록한다. 정상 Finish를 가정해 자동 해제하지 않고 subsequent physical Close도 보존 오류를 반환한다. 이는 기존 Raw 동작에 추가되는 명시적 보수 정책이다. 해당 API의 목적지 파일까지 보호한다고 주장하지 않는다.

원래 driver의 global ConnectionHook은 wrapper 생성 전에 unwrapped conn을 받는다. 그 hook이 미종료 stmt/rows를 보관하면 이 wrapper의 계수에 들어오지 않는 경계가 있다. 현재 저장소에서 hook 등록 사용은 발견되지 않았으나, 임의 외부 hook의 자원까지 관측한다고 주장하지 않는다. 그 조합은 별도 계측·정책·native 검증이 필요하다.

기존 namespace의 조상 rename/reparse 경합, 협조하지 않는 외부 SQLite 연결, ATTACH/VACUUM 목적지와 임의 Raw 리소스의 전체 수명까지 이 세 file pin으로 보장하지 않는다. Linux 회귀·Windows 교차 컴파일은 실제 Windows native 실행의 대체 증거가 아니다. 이전 합성 lifetime 실험의 통과도 생산 경로의 원자 생성·재연결·종료 오류 검증과 구분한다.

## VFS 종료 관측의 남은 경계

현재 의존성 modernc.org/sqlite v1.60.1에서 `lib/sqlite.go`의 `_sqlite3OsClose`는 VFS xClose의 반환값을 버린다. `_sqlite3PagerClose`와 `_sqlite3BtreeClose` 역시 하위 파일 종료 오류를 public Close 결과로 전달하지 않는 경로가 있다. Windows `_winClose`는 CloseHandle 실패를 오류로 만들 수 있으므로 **public Close nil만으로 모든 kernel 핸들의 종료를 입증할 수 없다**. 이 소스 경로는 실제 하위 CloseHandle 실패가 이번 실행에서 발생했다는 관측을 뜻하지 않는다.

registry의 fault retention은 wrapper에 전달된 오류·panic과 관측한 미종료 자원에 적용한다. 하위 VFS가 감춘 종료 오류를 native pin이 실제로 감지하거나 방지하는지는 아직 검증되지 않은 Gap이다. ATTACH/VACUUM 출력의 숨은 연결 종료도 별도 경계이며, 주 DB의 세 pin으로 출력 파일까지 보호했다고 주장하지 않는다. 이 패치는 실제 Windows 생산 회귀와 해당 실패 경계 검토를 위한 변경이며 보안 수리 완료 판정이 아니다.

## backlog와 relocation 통합 범위

`openBacklogEngine`과 `openGTDDB`의 writable 연결이 같은 물리 guard를 사용한다. queue lock 이전에도 기존 ancestor 거부를 실행한 뒤 관리 namespace를 준비하므로 기본 Windows owner로 부모를 먼저 만드는 경로를 피한다. Unix에서는 기존 owner의 write bit가 없는 디렉터리를 chmod로 다시 쓰기 가능하게 만들지 않는다. 기존 shared lock 검사와 실제 잠금·release 처리는 유지한다.

논리 migration/relocation의 sibling DB는 privatefile.CreateTemp로 payload 전에 생성한다. 기존 parity/readback·quarantine·rollback 원본·대상 no-overwrite 확인·atomic publication 순서는 유지한다. readonly reader의 query-only SQL, 기존 DB만 열기와 live WAL을 checkpoint하지 않는 경로는 변경하지 않는다. 기존 DB의 schema preflight도 이전 순서를 유지하며, 임의 legacy readonly 연결의 모든 파일 수명까지 관측한다는 주장은 하지 않는다.

`BackupGTDStore`의 VACUUM INTO 출력은 별도 함수 소유의 `privatefile.CreateExclusive`로 생성한다. CREATE_NEW/O_EXCL과 Windows protected current-SID SA를 첫 SQL 쓰기 전에 적용하고, READ_DATA/no-DELETE-share 핸들을 SQL 실행 동안 유지하며 같은 파일 identity와 strict policy를 확인한다. 전체 DB를 복제하는 기존 VACUUM SQL은 유지한다. 기존 출력은 빈 파일이어도 거부하고 SQL 실패가 남긴 새 partial output은 private 상태로 보존한다. 함수가 소유한 FD는 정상·오류 반환 때 한 번 Close하며 오류를 전달한다. 이 경로는 새 registry 한도나 영구 pin을 추가하지 않는다. public Exec nil과 직접 FD Close는 SQLite 내부 VFS 핸들의 모든 종료를 입증하지 않는다.

`RestoreGTDStore`는 원본을 스트리밍으로 private staging에 복사하고 Sync·Close한 뒤 `atomicfile.RenameNoReplace`로 같은 볼륨의 빈 이름에만 게시한다. Linux renameat2(RENAME_NOREPLACE), Darwin renameatx_np(RENAME_EXCL), Windows MoveFileEx(flags0)를 사용하며 unsupported kernel/filesystem에서 replace fallback을 하지 않는다. 기존 integrity 검증과 rollback 원본, 추가 테이블까지 포함하는 전체 DB snapshot을 유지한다. 조상 경로 rename/reparse 경합은 별도 경계다. native 회귀는 strict policy와 capture-time file ID, weak 기존 namespace의 무변경 거부를 검사하도록 작성했으며 Linux 실행·교차 컴파일만으로 Windows 통과를 주장하지 않는다.
