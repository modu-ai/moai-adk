# 배정·진행 기록 `assignments.json` 스키마 (DR-03 #2139)

`.moai-cli/manifests/assignments.json`은 어떤 Issue가 아직 배정되지 않았고, 어떤 Issue를 어느 레인이 어느 단계까지 진행했는지 한 파일에 모은 기록이다. 지금까지는 리더가 `.moai-cli/manifests/github-issues.json`의 `dispatch_journal`·`session_notes`에 손으로 적었다. 이 문서는 그 기록의 형식과 갱신 규칙을 정한다. 리더가 이 기록으로 감독하는 절차는 `AGENTS.local.md` § "Assignment record and supervision"에 있다.

**로컬 파일은 lock이 아니다. 소유권은 Issue의 claim 댓글이 정한다.** 이 파일은 같은 호스트 안의 캐시이자 쓰기 순서를 맞추는 보조 기록일 뿐이다. 이 파일에 이름이 적혀 있다고 해서 그 Issue를 쓸 권리가 생기지 않는다. 파일과 원격이 다르면 원격이 정본이다.

## 형식

```json
{
  "schema_version": 1,
  "refreshed_at": "<UTC RFC 3339>",
  "unassigned": [<issue 번호>, ...],
  "assigned": [
    {
      "issue": <번호>,
      "agent": "<claim 댓글의 session 값, 예: agent-5-767ceb>",
      "claim_ref": "<issue>#<comment-id>",
      "epoch": <정수>,
      "stage": "claimed | in_progress | pr_open | audited | merged | lost",
      "worktree": "<저장소 기준 상대 경로 또는 null>",
      "pr": <PR 번호 또는 null>,
      "updated_at": "<UTC RFC 3339>"
    }
  ]
}
```

| 필드 | 뜻 |
| --- | --- |
| `schema_version` | 형식 버전. 지금은 `1`이다. |
| `refreshed_at` | `unassigned`를 원격에서 마지막으로 다시 계산한 시각(UTC). |
| `unassigned` | 열려 있고 claim 댓글이 없는 Issue 번호. 오름차순. |
| `assigned[].issue` | claim 댓글이 있는 Issue 번호. |
| `assigned[].agent` | claim을 낸 세션. claim 댓글의 `session=` 값을 그대로 옮긴다(대개 `agent-N-<ref>` 형태지만 세션 UUID일 수도 있다). |
| `assigned[].claim_ref` | `<issue>#<comment-id>`. 커밋 trailer `Moai-Claim:`과 같은 값이다. |
| `assigned[].epoch` | claim 댓글의 `epoch=` 값. |
| `assigned[].stage` | 진행 단계. 아래 표를 따른다. |
| `assigned[].worktree` | 레인 worktree의 저장소 기준 상대 경로(`.moai-cli/worktrees/<n>-<key>`). 개인 절대 경로는 쓰지 않는다. 확인하지 못했으면 `null`. |
| `assigned[].pr` | 열린 PR 번호. 없으면 `null`. |
| `assigned[].updated_at` | 이 항목을 마지막으로 고친 시각(UTC). |

| `stage` | 근거 |
| --- | --- |
| `claimed` | claim 댓글은 있고 원격 브랜치·PR은 없다. |
| `in_progress` | 원격 브랜치가 있고 PR은 아직 없다. |
| `pr_open` | `Refs #<issue>` PR이 열려 있다. |
| `audited` | 그 PR의 현재 head에 `moai/audit` status가 있다. |
| `merged` | PR이 `v2026.11.0`에 병합되었다. |
| `lost` | 같은 Issue에서 더 낮은 claim 댓글 id에 밀린 claim이다. 기록으로만 남긴다. |

## 갱신 규칙

1. `unassigned`는 매번 원격에서 새로 계산한 스냅샷이다. 누구나 다시 계산해 쓸 수 있고, 다른 값과 충돌하면 원격을 다시 읽은 값이 이긴다. 계산 명령은 `gh issue list --state open --milestone <milestone> --json number`와 각 Issue의 claim 댓글 조회(`gh api repos/<owner>/<repo>/issues/<n>/comments --jq '.[] | select(.body|test("moai-claim:")) | .id'`)다.
2. `assigned` 항목은 claim 댓글 id가 있을 때만 추가한다. 디스패치 메시지나 로컬 worktree가 있다는 사실만으로는 추가하지 않는다.
3. 같은 Issue에 claim 댓글이 둘 이상이면 가장 낮은 댓글 id가 이긴다. 진 쪽은 `stage: lost`로 남긴다. lease 만료로 자동 인수하지 않으며, 인수는 운영자가 정한다.
4. `stage`는 위 표의 근거를 원격에서 읽어 정한다. 레인 보고나 idle 알림만으로 올리지 않는다.
5. 불일치를 발견하면 원격 값으로 고치고, 무엇이 달랐는지는 해당 Issue 댓글에 남긴다(이 파일은 이력을 쌓지 않는다).

## 초기 파일의 범위

초기 파일(2026-11-09 작성)은 열린 RC1 Issue 중 #1975·#1984·#2014와 milestone `design-docs-r2` 전부(#2137~#2143)를 담는다. 같은 시각에 열려 있던 #2016(04-D)은 금지 카드라서 넣지 않았다. #2141·#2143은 로컬 worktree는 있었지만 claim 댓글이 없어서 규칙 2에 따라 `unassigned`에 두었다. 원격 대조 명령과 출력은 PR 본문에 기록한다.
