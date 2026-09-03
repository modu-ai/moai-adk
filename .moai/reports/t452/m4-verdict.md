# t452 M4 — 쓰기 자세 판정

카드: t452 · SPEC-CODEX-SKILL-LOADER-001 · BASELINE_SHA `c529b2e4aaf5148aee7e6c67649bf392837bbb06`

## 판정 — AC-CSL-011 = **해당 없음** (빈 스윕)

AC-CSL-011 의 스윕 대상은 `$BASELINE_SHA` 이후 변경된 Go 소스다. 그 집합이 **비었다**:

```
$ git diff --name-only c529b2e4aaf5148aee7e6c67649bf392837bbb06 -- '*.go'
(출력 없음)
$ git diff --name-only c529b2e4aaf5148aee7e6c67649bf392837bbb06 -- '*.go' | wc -l
0
```

이 회차의 전체 변경 파일 집합(추적 파일):

```
$ git diff --name-only c529b2e4aaf5148aee7e6c67649bf392837bbb06
.claude/settings.json
.moai/specs/SPEC-CODEX-SKILL-LOADER-001/acceptance.md
internal/template/agentemit/agents-codex.yaml
internal/template/catalog.yaml
internal/template/templates/.codex/agents/moai/sync-auditor.toml
```

Go 파일이 하나도 없다. **피연산자 한쪽이 비면 교집합의 0 은 아무것도 주장하지 않는다** —
"사용자 계층 쓰기를 도입하지 않았다"를 PASS 로 올리면 빈 스윕의 초록을 판정으로
승격시키는 것이다. 그래서 **해당 없음**으로 기록한다.

(참고: `.claude/settings.json` 은 이 SPEC 의 작업이 아니며 커밋하지 않는다 —
`m3-verdict.md` 의 귀속 절 참조.)

## 셀렉터 생존 확인 — 0 매치의 침묵으로 판정하지 않기 위해

대상 집합이 비어 있어 판정에는 쓰이지 않지만, 셀렉터 자체가 살아 있음을 보인다.
저장소 Go 트리 전체에 같은 표현을 돌린 결과:

```
$ grep -rEn "os\.WriteFile|os\.Create|os\.MkdirAll|os\.Remove" --include='*.go' internal/ | wc -l
4771
$ grep -rn "skills\.config" --include='*.go' internal/ pkg/ cmd/ | wc -l
44
$ grep -rn "\.codex/skills" --include='*.go' internal/ pkg/ cmd/ | wc -l
0
```

앞의 둘이 실제로 매치하므로 셀렉터는 죽어 있지 않다. 셋째의 0 은 그 사실 위에서만
읽을 수 있고, 그마저도 이 SPEC 의 변경 집합에 대한 판정이 아니다(집합이 비었으므로).

## Gaps

- **`skills.config` 44 매치는 선행 작업의 표면이다.** 파일은
  `internal/cli/doctor_codex.go` · `doctor_codex_test.go` · `internal/codexwiring/skills.go`
  · `skills_test.go` 넷이며, **이 회차에 하나도 편집하지 않았다**(위 변경 집합에 없다).
  그 코드가 사용자 계층에 실제로 **쓰는지**는 이 회차가 감사하지 않았다 —
  AC-CSL-011 의 주어가 아니기 때문이다. 그 축이 필요하면 별개 카드다.
- 셀렉터는 Go 소스만 본다. 이 회차가 더한 프로브 픽스처(`.toml`/`SKILL.md`)와 증거
  문서는 실행 가능한 쓰기 경로가 아니므로 대상에서 제외했다.

## Residual risk

- 이 회차의 프로브는 격리 `CODEX_HOME=/tmp/t452-m2/codexhome` 안에서만 돌았고, 개발
  저장소의 `~/.codex` 불변은 AC-CSL-010 이 해시 대조로 따로 판정한다. 다만 그 판정은
  `config.toml` 해시와 `skills/` 1 단계 목록 두 축만 본다 — 그 밖의 파일이 변했을
  가능성은 배제되지 않는다.
