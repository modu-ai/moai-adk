# RC1 기준선 대조 (Issue #1985, 00-A)

이 문서는 Rust RC checkout과 Go 세대 `.claude` 자산의 기준선을 어떻게 측정했고 결과가 무엇인지 기록한다. 수치는 모두 측정 시각에 묶인 값이며, 상수로 옮겨 쓰지 않는다(REQ-00A-07). 자산의 의미 판정(keep/merge/replace 등)은 설계 페이지 14~18의 몫이고, 여기서는 바이트 차이만 다룬다.

## 산출물

| 파일 | 내용 | schema_version |
|---|---|---|
| `.moai-cli/manifests/baseline/rc1-baseline.json` | HEAD SHA, 현재 브랜치, worktree 목록, 원격 heads, 사용자 홈 항목의 존재 여부 | 1 |
| `.moai-cli/manifests/baseline/rc1-inventory.json` | tracked 파일 전체(`files`: mode, git object, sha256)와 변경·untracked 파일(`changes`), 측정 명령별 수치 | 1 |
| `.moai-cli/manifests/baseline/rc1-inventory.sha256` | shasum 호환 체크섬(`<hex>  <path>`), 일반 파일만 | — |
| `.moai-cli/manifests/baseline/claude-diff.json` | Go 기준선 `.claude`, Go template `.claude`, 현재 설치본 `.claude/` 세 집합의 경로별 분류 | 1 |
| `.moai-cli/manifests/baseline/preserve-before.*` | 작업 시작 전 보존 원장(읽기 전용으로 2026-11-08T07:50:52Z 수집). 수집기 코드보다 먼저 별도 커밋으로 기록했다 | — |
| `.moai-cli/manifests/baseline/preserve-excludes.json` | `.moai` 원장에서 뺀 Go 런타임 경로의 규칙과 개수 | 1 |

산출물의 경로는 모두 저장소 상대 경로다. 저장소 밖 worktree는 `<outside-project>`, 사용자 홈 항목은 `~/...`로만 적고 내용은 열거하지 않는다(REQ-00A-08·09).

## 재생성 방법

수집기는 아무것도 쓰지 않고 JSON 하나를 stdout으로 낸다. 파일 분리는 `jq`로 한다.

```sh
moai-cli baseline-inventory <PROJECT> --go-baseline-commit 4fa2aa9db2e0add0f1299e35407141b2b41c5ec0 > out.json
jq .baseline    out.json > rc1-baseline.json
jq .inventory   out.json > rc1-inventory.json
jq .claude_diff out.json > claude-diff.json
jq -j .checksums out.json > rc1-inventory.sha256   # -j: 마지막 빈 줄이 생기지 않도록
```

기준선 커밋이 프로젝트 object store에 없으면 `--go-repo <REPO>`로 다른 저장소를 지정한다. 커밋을 주지 않으면 `claude_diff.status`가 `not_collected`, 찾지 못하면 `commit_not_found`가 된다. 원격 조회에 실패하면 `remote_heads.status`가 `unavailable`이 되며, 오류 출력은 원격 URL을 담을 수 있어 기록하지 않는다.

## 수집 방식

- git은 `--no-optional-locks`로 호출해 index를 갱신하지 않는다. Go 기준선은 `cat-file -e`와 `ls-tree -r -z`로만 읽고 checkout·reset·fetch는 하지 않는다.
- 현재 설치본의 git blob id는 `hash-object --no-filters`(`-w` 없음)로 계산한다. 콘텐츠 해시는 SHA-256(`sha256:<64 lowercase hex>`)이고 git object(SHA-1)는 대조용으로만 함께 적는다.
- 심볼릭 링크는 따라가지 않고 링크 대상 문자열을 해시한다. 체크섬 파일에는 넣지 않는다(`shasum`은 링크를 따라가기 때문).
- 줄바꿈이나 `\`가 들어간 경로는 shasum 형식으로 그대로 적을 수 없어 체크섬 파일에서 빠진다. 대신 `files`/`changes`에는 남고, 빠진 경로는 `inventory.checksum_unrepresentable`에 나열한다(이번 측정에서는 0건).
- `status --porcelain=v2`는 v1과 달리 형식 안정성이 문서로 보장되지 않으므로, `#` 헤더와 모르는 레코드는 건너뛴다.
- 레인 worktree(`.claude/worktrees/`, `.moai-cli/worktrees/`, `.worktrees/`)는 측정 대상에서 빼고, 뺀 개수를 `excluded_changes`·`excluded_active_files`에 남긴다.

### 분류 규칙(`claude-diff.json`의 `rules`)

- 키: 각 집합의 `.claude` 디렉터리 기준 상대 경로. Go template은 `internal/template/templates/.claude/` 하나뿐임을 확인했다(template root가 둘 이상이면 `multiple_template_roots`로 멈춘다).
- class: 한 집합에만 있으면 `baseline-only`·`template-only`·`active-only`, 둘 이상에 있고 blob id가 하나면 `identical`, 아니면 `differing`.
- tracking(현재 checkout 기준): `tracked`·`ignored`·`untracked`·`absent`.
- origin: template 집합에 있으면 `generated`, 없으면 `source`.
- scope: 파일 이름에 `.local.`이 있거나 `.local`로 끝나면 `private`, 아니면 `shared`.

이 규칙들은 기계적인 정의일 뿐이며 자산의 계승 여부를 판단하지 않는다.

## tree digest (`tree_digest`)

`moai_cli::tree_digest(root, excludes) -> Result<String, String>`은 00-A가 소유하는 트리 digest다. 05-C #2003, 07-D #2011은 이 함수를 호출만 한다.

- 형식 `moai-tree-v1`: 상대 경로를 바이트 순으로 정렬하고, 항목마다 종류 바이트(`f` 파일, `l` 링크), u64 BE 경로 길이, 경로, u64 BE 내용 길이, 내용(파일 바이트 또는 링크 대상)을 이어 붙여 SHA-256을 낸다.
- 빈 디렉터리는 반영되지 않는다. `excludes`는 root 기준 경로이며 그 경로 자체와 하위가 빠진다. 특수 파일, 읽을 수 없는 파일, non-UTF-8 경로는 오류로 드러낸다.
- `claude-diff.json`의 `active_tree_digest`는 현재 설치본 `.claude/`(`worktrees/` 제외)의 값이다.

## 측정 결과 (2026-11-08T08:22:41Z, HEAD a6c67ba2d2750c42e0d46b9c5a626db38dc0b555)

| 항목 | 값 | 측정 명령 |
|---|---|---|
| tracked 파일 | 116 | `git ls-files -s -z` |
| 변경된 tracked 파일 | 24 | `git status --porcelain=v2 -z --untracked-files=all` |
| untracked 파일 | 50 | 같은 명령 |
| 체크섬 행 | 166 | tracked·변경 행 중 일반 파일 |
| worktree | 3 | `git worktree list --porcelain -z` |
| Go 기준선 `.claude` | 600 | `git ls-tree -r -z 4fa2aa9… -- .claude` |
| Go template `.claude` | 512 | 같은 명령, `internal/template/templates` |
| 현재 설치본 `.claude/` | 186 | 파일 시스템 walk(링크 비추종, worktree 제외 0) |
| 참고: Issue 본문의 template 513 | — | moai-adk-go 작업 트리 `find` 값. 위 512는 기준선 커밋의 `ls-tree` 값이라 기준이 다르다 |
| 분류 합계 | 654 = identical 479 + baseline-only 89 + template-only 52 + differing 32 + active-only 2 | `claude-diff.json` |

변경·untracked 파일 수는 이 시점의 primary checkout 작업 트리를 반영한다. 다른 세션이 작업 중이면 값이 바뀌며, 이 수치는 위 시각에만 유효하다.

## 보존 원장과 제외 규칙 (00A-AC-3)

보존 원장은 작업 전(2026-11-08T07:50:52Z)에 읽기만으로 만들었고, 수집기 코드보다 먼저 단독 커밋(64e4459e0)으로 남겼다.

- `preserve-before.claude.sha256`: primary `.claude/` 186파일
- `preserve-before.moai.sha256`: primary `.moai/` 중 Go 런타임 파일을 뺀 268파일
- `preserve-before.go-dirty.sha256`: moai-adk-go 작업 트리에서 수정 상태인 24파일(`status --porcelain`으로 실측)

`.moai/`는 실행 중인 Go 훅이 계속 쓰기 때문에 전체를 전후 비교할 수 없다. 사용자 결정(docs/decisions-2026-11-08.md)에 따라 `.moai/logs/`·`.moai/state/`·`.moai/db/`와 `.moai/**/*.jsonl`(Go 훅이 덧붙여 쓰는 로그)을 뺐다. 원래 339행에서 71행(logs 25, state 43, db 0, jsonl 3)을 지웠을 뿐 다시 해시하지 않았으므로, 남은 268행도 작업 전 측정값이다. 규칙과 개수는 `preserve-excludes.json`에 있고, 원래 339행은 64e4459e0에서 볼 수 있다.

## 검증 명령

```sh
# AC-2: 측정한 checkout(primary)에서 실행
shasum -a 256 -c --strict --status .moai-cli/manifests/baseline/rc1-inventory.sha256
# AC-3: primary 루트에서 앞의 두 개, moai-adk-go 루트에서 세 번째
shasum -a 256 -c --strict --status .moai-cli/manifests/baseline/preserve-before.claude.sha256
shasum -a 256 -c --strict --status .moai-cli/manifests/baseline/preserve-before.moai.sha256
shasum -a 256 -c --strict --status <primary>/.moai-cli/manifests/baseline/preserve-before.go-dirty.sha256
# AC-5
jq -e '([.rows[]|select(.class==null)]|length)==0 and ([.rows[].class]|length)==.total' .moai-cli/manifests/baseline/claude-diff.json
# AC-8
cargo test --offline --locked --test product_cli tree_digest_stable_sensitive_and_excludes
```

체크섬 파일은 측정 시각의 primary 작업 트리 바이트를 담는다. 같은 커밋의 깨끗한 checkout에서 돌리면 측정 당시 수정돼 있던 파일은 FAILED로 나오는 것이 정상이다.
