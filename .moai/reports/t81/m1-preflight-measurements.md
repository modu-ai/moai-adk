# t81 (M1) plan-phase 선행 실측

worktree `WT-skills-canonical`. 아래는 SPEC 작성 전 오케스트레이터가 직접 관측한 값이다.

## A. 스킬 인벤토리 (카드 전제 정정)

| 집합 | 수 | 비고 |
|---|---|---|
| 로컬 `.claude/skills/` | 44 | 아래 두 집합의 합 |
| 템플릿 `internal/template/templates/.claude/skills/` | **34** | 배포 대상 |
| 로컬 전용 `hns-*` | 10 | dev-only, 템플릿 미러 금지 |

[정정 2026-08-22] 카드 본문의 "스킬 32개"도, 이 문서의 최초 "36개"도 틀렸다. 실제 배포 대상은 **34개**다 — `find ... -name SKILL.md | wc -l` 로 재측정. 최초의 46/36 은 `ls` 가 long-format 으로 별칭돼 헤더 행이 함께 세어진 결과다(집계 오류이지 관측 오류는 아니며, 델타 10 은 같은 편향을 공유해 영향받지 않았다). SPEC 은 34 기준.

로컬 전용 10개: hns-lsel-applier, hns-lsel-curator, hns-moaiadk-best-practices,
hns-moaiadk-dev-reference, hns-moaiadk-patterns, hns-oss-docs-i18n-rules,
hns-oss-docs-readme-sync, hns-oss-docs-structure-map, hns-oss-docs-verify,
hns-workflow-ci-loop.

`.agents/` 는 로컬·템플릿 **양쪽에 존재하지 않는다**(신규 생성 대상).

## B. [BLOCKER 급] `//go:embed` 는 심볼릭 링크를 **무음으로 버린다**

최소 재현(worktree 안에서 실행, 이후 삭제):

```
x/real/a.txt          (일반 파일)
x/real/linked.txt  -> ../../target.txt   (심볼릭 링크 파일)
x/linkdir          -> real                (심볼릭 링크 디렉터리)
//go:embed all:x
```

`fs.WalkDir` 결과:

```
ENTRY: .            err: <nil>
ENTRY: x            err: <nil>
ENTRY: x/real       err: <nil>
ENTRY: x/real/a.txt err: <nil>
```

`linked.txt` 와 `linkdir` **둘 다 사라졌고, 빌드 오류도 경고도 없다**.

**귀결**: `internal/template/templates/` 안에 심볼릭 링크를 두는 설계는 성립하지 않는다.
M0 이 확인한 "Codex 가 심볼릭 링크를 따라간다"는 **런타임(배포된 사용자 프로젝트) 사실**이고,
**빌드타임(임베드) 사실이 아니다**. 두 사실을 섞으면 무음으로 스킬 0개가 배포된다.

따라서 링크 생성 주체는 템플릿이 아니라 **배포기(`internal/template/deployer.go`)** 여야 한다.

## C. 배포기 현황

- `deployer.Deploy` 는 임베드 FS 를 walk 하며 `atomicWriteFile` 로 **일반 파일만** 쓴다. 심볼릭 링크 생성 경로가 **없다**.
- Windows 에서 `os.Symlink` 는 권한/개발자 모드가 필요하다 → **복사 폴백이 필수**(16개 언어 중립성과 별개로 OS 중립성 문제).

## D. [설계 제약] `moai update` 는 `.agents/` 를 청소하지 않는다

`ManagedCleanTargets` (deploy.go:53-79) 의 관리 대상 뿌리:

```
.claude/settings.json  .claude/commands/moai  .claude/agents/moai
.claude/skills/moai*(글롭)  .claude/rules/moai  .claude/output-styles/moai
.claude/hooks/moai       .moai/config
```

`.agents/` 는 **목록에 없다**. 즉 `.agents/skills/` 에 배포하면 이름이 바뀌거나 은퇴한 스킬이
사용자 프로젝트에 **영구 잔존**한다. 이는 가설이 아니라 이미 관측된 실패 형태다 —
M0 실측에서 `~/.codex/skills/` 에 2026-06-07 자 구 moai 스킬 **46개**가 그대로 남아 있었고,
그중 다수(`moai-lang-*`, `moai-platform-*`)는 현재 카탈로그에 존재하지 않는 이름이다.

**M1 은 배포뿐 아니라 청소 경로 등록까지 포함해야 한다.** 아니면 오염원을 하나 더 만든다.

## E. 사용자 실환경 오염원 (카드 note (1))

`~/.codex/skills/` 46개 — 2026-06-07 자. 사용자 홈이므로 **삭제는 본 SPEC 범위 밖**,
plan 에는 정리 방안(목록·판별 기준·명령)만 싣고 실행은 별도 승인.
`.system` 디렉터리는 Codex 자체 소유이므로 건드리면 안 된다.

## F. 회귀 금지 축 (카드 note (2))

M0 실측: Codex 는 `.claude/skills/` 를 **스캔하지 않는다**. 따라서 `.claude/skills/` 를
정본으로 두고 `.agents/skills/` 를 링크로 만드는 방향(역방향)도 Codex 쪽에서는 동작한다.
방향 선택은 SPEC 의 설계 결정 사항이며, 어느 쪽이든 Claude Code 의 `.claude/skills/`
경로 동작이 불변임을 검증하는 AC 가 필요하다.
