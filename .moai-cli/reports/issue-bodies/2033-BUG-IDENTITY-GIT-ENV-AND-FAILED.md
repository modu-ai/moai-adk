<!-- moai-template:moai-bugfix@0.3.0; 2026-11-08 -->
```yaml
logical_key: BUG-IDENTITY-GIT-ENV-AND-FAILED
milestone: 2026.11.0-rc.1
parent_key: EP-02
work_kind: bugfix
related: [01-A, 00-A]
claim: {state: unclaimed}
evidence_status: not_run
```
# [RC1][결함] identity: git 환경 변수 주입 미차단과 `Git::Failed`를 `no_git`으로 보고하는 문제

## 출처
01-A #1988 독립 감사(agent-2)·레인 보고(agent-4) 2026-11-08 — PR #2032 비차단 지적 2건을 후속 결함 Issue로 발행(리더).

## 결함
1. `src/identity.rs`가 git을 부를 때 `GIT_DIR`·`GIT_WORK_TREE`·`GIT_COMMON_DIR`·`GIT_INDEX_FILE`·`GIT_NAMESPACE`는 제거하지만 `GIT_CONFIG_PARAMETERS`·`GIT_CONFIG_COUNT`/`GIT_CONFIG_KEY_*`/`GIT_CONFIG_VALUE_*`·`GIT_CEILING_DIRECTORIES`·`GIT_DISCOVERY_ACROSS_FILESYSTEM`은 상속된다. 호출자 환경이 `core.worktree` 등을 주입하면 `--show-toplevel` 결과(경계)가 바뀔 수 있다(코드 읽기 기준, 미실측).
2. `Git::Failed`(git은 있으나 실패: `safe.directory` 소유자 거부, 손상된 저장소 등)를 `no_git` "not inside a git repository"로 보고한다. 추정 id를 만들지 않는 것은 맞지만 사유가 관측과 다르다(관측하지 않은 사실 단정).

## 수정 방향
- 1: git 호출 전 위 변수들도 제거(`env_remove`), `LC_ALL=C` 유지. 시험: `GIT_CONFIG_COUNT=1 GIT_CONFIG_KEY_0=core.worktree GIT_CONFIG_VALUE_0=/tmp/x`를 주입해도 경계가 바뀌지 않음.
- 2: `Git::Failed`를 별도 kind `git_failed`(stderr 요약을 메시지에, 입력 경로 echo 금지)로 보고. 시험: `safe.directory` 거부를 fixture(다른 uid 흉내는 불가 → `-c safe.directory=` + `GIT_TEST_ASSUME_DIFFERENT_OWNER=1`)로 재현해 `git_failed`를 확인; 불가하면 not_run으로 표기.
- envelope(01-C) 통합 뒤라면 오류 객체 `kind`에 `git_failed` 추가(00-D 어휘 11종 + 1, docs 기록).

## 인수 기준
- AC-IDF-1: 환경 주입 시험 통과 · 결과: not_run
- AC-IDF-2: `git_failed` 구분 시험 통과(또는 재현 불가 시 not_run 명시) · 결과: not_run
- AC-IDF-3: 검사 묶음·독립 AUDIT PASS → 통합 · 결과: not_run

## 범위
- 변경 가능: `src/identity.rs`, `tests/product_cli.rs`, `docs/rust-product-path.md`(+ 01-C 통합 뒤 `docs/protocol-v1.md` kind 어휘, `src/contract.rs`의 `Kind::GitFailed` 추가와 classify 규칙 — r28, 리더 (b)안 지시 2026-11-08 10:0x에 따른 범위 확장; decisions 19:1x 행)
- 변경 금지: `src/agents.rs`, 공유 상수

## 의존성
- 01-A #1988 통합 뒤(blocked_by #1988)
