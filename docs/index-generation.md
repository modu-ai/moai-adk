# 증분 색인의 generation 보존 계약

대상: #2182(C6), `index-build`·`index-query`. watcher·실호스트 설치·generation 정리는 포함하지 않는다.

- generation 키는 base commit과 정렬한 추적 파일 표(path·digest·byte 수)의 내용으로만 정한다. `built_at`과 현재 시각은 키에 넣지 않는다. 같은 내용의 실제 CLI 빌드를 1.1초 간격으로 실행해도 ID·포인터·기존 manifest 바이트가 같고 `action: unchanged`여야 한다. 이 시험은 초 단위 시각을 키에 넣는 M5s 회귀도 검출한다.
- `current.json`의 `generation_id`는 경로에 결합하기 전에 `^gen-[0-9a-f]{16}$` 형식을 만족해야 한다. 절대 경로·상위 이동·길이 불일치·대문자·비16진수 이름은 `unavailable`과 `index pointer unreadable`로 보고한다. 질의는 포인터나 외부 파일을 변경하지 않는다.
- 기존 generation 디렉터리는 재사용 전에 manifest를 해석하고 schema version·generation ID·파일 수 상한·현재 base commit·전체 파일 표를 확인한다. 손상된 manifest를 `unchanged`로 승인하거나 덮어쓰지 않는다. 실패하면 기존 포인터를 유지하고 실패 표식을 남겨 `partial`을 보고한다(REQ-C6-01). manifest를 복구한 뒤 정상 빌드하면 실패 표식이 해소된다.

시험: `tests/index_cli.rs`의 `identical_rebuild_across_seconds_keeps_generation_and_manifest`, `pointer_generation_id_is_validated_before_path_resolution`, `damaged_reusable_manifest_keeps_previous_pointer_and_marks_partial`. 기존 freshness·원자 포인터 시험도 유지한다.

기존 읽기 권한 실패 fixture는 Unix permission bit를 사용하므로 `cfg(unix)`로 한정한다. Windows 빌드 job이 이 Unix 전용 준비를 컴파일하지 않게 하며, Windows에서 해당 실패 경로를 실행했다는 뜻은 아니다.

AC-1 FRESHNESS fixture의 권위는 리더가 지정한 #2346(C10a)이다. 이번 감사 수리는 위 경계에 한정하며, C10a 병합 뒤 fixture 정합은 후속으로 남긴다. 이 문서나 로컬 시험 통과가 C10a 연동·독립 감사·다른 OS 실행을 증명하지 않는다.
