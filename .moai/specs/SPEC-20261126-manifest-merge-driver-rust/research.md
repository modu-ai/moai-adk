# 조사: git 머지 드라이버가 받는 것과 해야 하는 것

공식 문서: https://git-scm.com/docs/gitattributes 의 "Defining a custom merge driver" 절. 2026-11-26T01:20:46Z 에 `curl -sSL --fail` 로 읽었고, HTML 바이트의 sha256 은 `3c2a1c79e123a33d7968c8725c361340ae9106232d13de4ddabbb9ecdc794be7` 이에요. 로컬 git 은 2.54.0 (Apple Git-157)이에요.

| 질문 | 인용 |
| --- | --- |
| 드라이버는 어디에 정의하나 | "The definition of a merge driver is done in the .git/config file, not in the gitattributes file" |
| 인자는 | "These three tokens are replaced with the names of temporary files that hold the contents of these versions when the command line is built." (`%O`, `%A`, `%B`; `%P` 는 "the pathname in which the merged result will be stored") |
| 결과의 자리와 종료값의 뜻 | "The merge driver is expected to leave the result of the merge in the file named with %A by overwriting it, and exit with zero status if it managed to merge them cleanly, or non-zero if there were conflicts." |
| 죽었을 때 | "When the driver crashes (e.g. killed by SEGV), it is expected to exit with non-zero status that are higher than 128, and in such a case, the merge results in a failure (which is different from producing a conflict)." |
| 공통 조상끼리의 머지 | "The merge.*.recursive variable specifies what other merge driver to use when the merge driver is called for an internal merge between common ancestors, when there are more than one. When left unspecified, the driver itself is used for both internal merge and the final merge." |

이 인용이 정하는 것: `%A` 를 건드리지 않고 128 미만의 0 아닌 값으로 끝나면 우리 쪽만 보이는 충돌이 돼요. 그래서 거절은 먼저 git 의 글 머지를 `%A` 에 써요. Rust 의 panic 은 101 로 끝나고 git 은 이를 충돌로 읽으므로, 작업은 `catch_unwind` 안에서 돌려요. 드라이버는 파일 내용을 임시 파일 세 개로만 받고, 문서는 작업 폴더에 대해 아무것도 약속하지 않아요. 이전 드라이버의 머리글에는 git 2.54 측정이 적혀 있었어요: git 은 머지·rebase·cherry-pick·`merge-tree` 에서 합친 파일을 작업 폴더에 쓰기 전에 드라이버를 불러요(`merge-tree` 는 쓸 작업 폴더가 없어요). 그러므로 그 시점에는 디스크의 파일에서 해시를 읽을 수 없어요.

## 이 SPEC 에서 직접 잰 것 (git 2.54.0, 임시 저장소)

| 사실 | 근거 |
| --- | --- |
| `.gitattributes` 가 가리키는 드라이버 이름이 설정에 없으면 git 은 자기 글 머지를 해요. 결과 바이트가 드라이버를 아무것도 등록하지 않은 머지와 같아요 | `tests/manifest_merge_driver_cli.rs` 의 `registered_a_clone_with_only_the_old_driver_line_runs_no_python` |
| `git merge` 에서 드라이버는 **머지하는 체크아웃(체크아웃된 브랜치)의 `.gitattributes`** 가 골라요. 들어오는 쪽의 속성 줄은 그 머지의 드라이버를 바꾸지 않아요 | 같은 파일의 `registered_the_attributes_of_the_checkout_that_merges_decide_the_driver` (두 방향). `merge-tree` 에서는 재지 않았어요 |
| 지금 설치된 `moai` 4.2648.53-dev 는 이 명령을 몰라 `manifest-merge-driver --help` 가 2 로 끝나고, 등록 줄은 git 글 머지로 떨어져요 | 임시 저장소 위험 경로 실행(2026-11-26) |

## 읽지 않은 것·재지 않은 것

- "작업 폴더를 쓰기 전에 부른다"는 순서는 새로 재지 않았어요. 해시를 작업 폴더에서 읽으면 합친 해시가 우리 쪽 해시와 달라 이 SPEC 의 테스트가 실패해요.
- GitHub 서버의 머지: 커스텀 드라이버를 돌리지 않는다는 말은 저장소 문서의 것이고, 여기서 GitHub 문서를 다시 읽지 않았어요.
- Linux·Windows 의 git 동작.
