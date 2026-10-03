# t1474 판정서 — GitHub #1731 무편집 전체 폼 저장의 바이트 불변성

- 카드: t1474 · 이슈: GitHub #1731 · 브랜치: `WT-web-fullform-noop-save`
- 기준 트리: 로컬 `develop` `b5815ca80` (origin/develop `3f3ebb763` 위에 fast-forward 흡수)
- 카드 등급: **A로 착수 → B로 전환** (새 테스트가 팁에서 잔여 결함을 재현함)

## Claim

1. 수리 전 팁(`b5815ca80`)에서 무편집 전체 폼 저장은 **바이트 불변이 아니었다**. user / quality / workflow / git-strategy / git-convention은 불변이지만 llm.yaml에 두 가지가 생겼다: 손대지 않은 `glm.effort` 블록(4티어 모두 `max`)과 `profile: medium`.
2. 이슈가 보고한 `harness: claude` 줄은 팁에서 **재현되지 않았다**.
3. 수리(`2f7829a0b`) 후 같은 테스트가 통과하고, 두 가드는 각각 필요하다 — 하나만 끄면 그 가드가 막던 변경만 되살아난다.
4. 수리 후 `internal/web/...`, `internal/settings/...` 전 테스트가 통과한다.

## Evidence

**RED — 수리 전 팁**: `go test -count=1 -run TestFullFormNoEditSaveIsByteIdentical -v ./internal/web/` → exit 1 (원문: `red-run.txt`, 커밋 `7b8458642`)

```
    save_lossless_test.go:333: llm.yaml changed by a no-edit save (GitHub #1731)
        --- after ---
        ...
          glm:
            models:
              high: glm-5.3 # inline note
            effort:
              fable: max
              high: max
              low: max
              medium: max
          profile: medium
--- FAIL: TestFullFormNoEditSaveIsByteIdentical (0.13s)
```

이 출력에서 실패한 파일은 llm.yaml 하나뿐이다. 다른 다섯 파일은 `before == after`였고, 저장 뒤 새 섹션 파일도 생기지 않았다.

**원인 두 가지**
- `llm.glm.effort.<tier>`: `radioEffectiveValue`(internal/web/schemaform.go)가 키가 없을 때 라디오를 필드 Default(`max`)로 미리 선택해 렌더한다. `applyTypedEdits`의 no-op 게이트는 그 제출값을 LoadRaw 현재값 `""`과 비교했고, 다르다고 판정해 effort 블록을 새로 넣었다.
- `llm.profile`: 프로필이 비어 있으면 셀렉터가 `agentFMPerfTierDefault`(medium)에 체크된 채 렌더된다(fieldsets.templ). 그 값을 제출하면 `applyPerfTierEdits`가 `profile: medium`을 써서, 프로필 없음(일반 상속)이 명시적 열 지정으로 바뀌었다.

**수리** (`2f7829a0b`, 18줄 추가)
- `internal/settings/sectionapply.go` `applyTypedEdits`: `cur == "" && f.Default != "" && next == f.Default`이고 디스크에 키가 없으면 건너뛴다.
- `internal/web/handlers.go` `handleSave`: 저장된 프로필이 비어 있고 제출값이 `agentFMPerfTierDefault`이면 `perfTier = ""`(보존)로 둔다.

**GREEN**: 같은 명령 → `--- PASS: TestFullFormNoEditSaveIsByteIdentical (0.48s)` / `ok github.com/modu-ai/moai-adk/internal/web 1.417s`

**음성 대조** (가드를 하나씩 `&& false`로 임시로 끄고 실행한 뒤 백업에서 복원. 복원 후 `grep -c "&& false"` 결과는 두 파일 모두 0)
- 프로필 가드만 끔 → exit 1, llm.yaml에 `profile: medium`만 다시 생김 (`negctl-profile-guard-off.txt`)
- effort 가드만 끔 → exit 1, llm.yaml에 `effort:` 블록(4티어 `max`)만 다시 생김 (`negctl-effort-guard-off.txt`)

**패키지 재측정**: `go test -count=1 ./internal/web/... ./internal/settings/...` → exit 0 (`pkg-run.txt`)

```
ok  	github.com/modu-ai/moai-adk/internal/web	142.379s
ok  	github.com/modu-ai/moai-adk/internal/settings	1.800s
ok  	github.com/modu-ai/moai-adk/internal/settings/agentfm	0.408s
ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	0.411s
```

`gofmt -l internal/web internal/settings` → 출력 없음 · `go vet ./internal/web/... ./internal/settings/...` → exit 0

## Baseline-attribution

모든 측정은 이번 실행에서 이 워크트리(`.claude/worktrees/agent-ad22b849134cd7072`)를 대상으로 했다. RED는 `b5815ca80`에 테스트만 더한 트리(`7b8458642`)에서, GREEN·음성 대조·패키지 재측정은 수리 트리(`2f7829a0b`와 같은 내용)에서 측정했다. 테스트 바이너리는 `go test`가 매번 이 트리에서 새로 컴파일한다. 설치된 `moai` 빌드는 판정에 쓰지 않았다.

## Gaps

- 실제 브라우저로 재현하지 않았다. 제출 폼은 렌더된 HTML에서 `extractBrowserSubmission`으로 재구성했고, 클라이언트 JS가 제출 직전에 값을 바꾸는 경우는 이 재구성에 반영되지 않는다.
- 픽스처는 이슈에 나온 키에 맞춰 손으로 만들었다. 실제 3.1.2 프로젝트의 sections 전체(32개 파일)를 재현하지는 않았다.
- `golangci-lint`는 돌리지 않았다(지시 범위가 vet·gofmt).
- 전체 스위트 판정은 CI 몫이다(원격 develop push 뒤에 확정되며, 보고 시점에는 PENDING).

## Residual-risk

- 두 가드는 렌더 기본값과 같은 제출을 무편집으로 본다. 그래서 키가 없거나 프로필이 비어 있는 상태에서는 기본값(`max` / `medium`)을 **명시적으로 디스크에 고정할 방법이 콘솔에 없다**. 화면에는 이미 그 값이 선택된 것으로 보이므로 화면 상태와는 맞는다. 다만 런타임에서는 빈 값과 의미가 다르다: 빈 effort 슬롯은 덮어쓰기가 없다는 뜻이고, 빈 프로필은 일반 상속이다. 이 차이를 고정하려면 별도 결정이 필요하다.
- 같은 형태의 결함은 Default를 선언한 다른 typed 라디오에도 생길 수 있다. 이번 가드는 typed 경로 전체에 적용된다. seam 경로(workflow 등)의 라디오 기본값 처리는 이번 픽스처의 workflow.yaml이 불변으로 나온 범위까지만 관측했다.
