# `repo-scans`

operation `repo_scans.judge` · 코드 `src/ops/repo_scans.rs`, 규칙 `src/repo_scans.rs` · Issue #2708(옛 `tools/repo_scans.py`, #2681)

`repo-scans [--refs TIP HEAD] [--repo DIR] [--table FILE]` 는 병합 전에 돌려야 할 저장소 전체 검사 시험을 고른다. 저장소의 검사 표 `tools/repo_scans.json` 에서, 한쪽은 어떤 검사의 규칙 파일을 바꾸고 다른 쪽은 그 검사의 범위 안 경로를 바꾼 경우를 찾는다(#2642·#2632 사고). `--refs` 를 주면 두 ref 의 merge-base 부터 `git diff --name-only` 로 변경 목록을 읽고, 두 ref 에 커밋된 표를 합쳐 판정한다(#2693 감사 B1). 어느 ref 에도 표가 없을 때와, `--refs` 없이 표준 입력 `{"pr": [...], "tip": [...]}` 을 줄 때는 대체 표(`--table`, 기본 `<repo>/tools/repo_scans.json`)를 쓴다.

결과는 `{remerge, hits, checks, reasons}` 이고, 옛 Python 도구가 출력하던 객체와 같다. `remerge: true` 이면 병합 결과에서 `checks` 를 먼저 돌린다. 내부 오류(잘못된 입력, 없는 ref, 깨진 표)는 언제나 `remerge: true` 와 표의 모든 검사로 끝난다. 표를 하나도 읽을 수 없으면 `checks` 는 비고 마지막 이유가 "hold the merge" 다. 이 경우는 병합하지 않는다. 옛 도구와 다른 점은 둘이다. 첫째, 모르는 옵션이나 `-` 로 시작하는 ref 는 사용법 오류(exit 2, `result` 없음)다. 둘째, 변경 경로를 `-z` 로 읽어 따옴표로 감싸지 않는다. Git 은 깨끗한 환경으로 부른다(`audit_delta::command`). 표준 입력과 표는 4 MiB 까지 읽는다. 아무것도 쓰지 않는다.
