# t1259 — AC-IFU-031 원격 CI 판정

## Claim

**PASS.** `SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001`의 AC-IFU-031을 `origin/develop`의 `d42ccbc6c9048d39fee281f425f566130dc34f9e`에서 확인했다. 이 head는 t1259 브랜치 HEAD `cdae544a31701caed3139465f267f19560ae5d7a`와 로컬 병합 SHA `e11afe3358d3aa3c687b3313a1aca77b40497c9f`를 모두 포함한다.

## Evidence

```text
$ git merge-base --is-ancestor cdae544a31701caed3139465f267f19560ae5d7a origin/develop; echo $?
0
$ git merge-base --is-ancestor e11afe3358d3aa3c687b3313a1aca77b40497c9f origin/develop; echo $?
0
$ git rev-parse origin/develop
d42ccbc6c9048d39fee281f425f566130dc34f9e
```

GitHub Actions의 같은 head에 대해 확인한 결과:

| 실행 | Run ID | 상태 | 판독 |
|---|---:|---|---|
| `CI` | `36374288896` | `completed / success` | `Test (ubuntu-latest)`, `Race Test`, 브라우저 가드, Windows·Ubuntu·macOS 통합 테스트 모두 `completed / success`; 빌드·린트·헌법 검사도 성공 |
| `SPEC Lint` | `36374288881` | `completed / success` | `spec-lint` 잡 성공 |
| `docs i18n parity check` | `36374288911` | `completed / success` | 로그를 직접 읽음: `strict=false`, `Errors:   0`, `Warnings: 0`, `OK: all 4 locales pass parity, frontmatter, H1, and glossary checks.` |

```text
$ gh run view 36374288896 --repo modu-ai/moai-adk --json headSha,status,conclusion
{"conclusion":"success","headSha":"d42ccbc6c9048d39fee281f425f566130dc34f9e","status":"completed"}
$ gh run view 36374288881 --repo modu-ai/moai-adk --json headSha,status,conclusion
{"conclusion":"success","headSha":"d42ccbc6c9048d39fee281f425f566130dc34f9e","status":"completed"}
$ gh run view 36374288911 --repo modu-ai/moai-adk --json headSha,status,conclusion
{"conclusion":"success","headSha":"d42ccbc6c9048d39fee281f425f566130dc34f9e","status":"completed"}
```

## Baseline-attribution

2026-09-28에 로컬 `develop` 워크트리에서 `origin/develop`을 읽고, 위 세 Actions 실행의 `headSha`를 각각 조회했다. 이 판정은 **`d42ccbc6c9048d39fee281f425f566130dc34f9e` 한 트리**에만 귀속한다. 문서 검사 로그는 실행 `36374288911`의 완료된 로그에서 판독했다.

## Gaps

- 카드 t1259는 이 보고서 작성 시점에 여전히 `picked`다. 현재 goal 계약에는 `done` 권한이 없어 큐를 변경하지 않았다.
- 이 보고서는 AC-IFU-031의 원격 실행 결과를 다룬다. 다른 인수 기준의 로컬·감사 근거는 SPEC `progress.md`와 별도 감사 보고서에 있다.

## Residual-risk

이후 `develop` 변경에는 별도 CI 결과가 필요하다. 이 head의 성공을 미래 변경의 검증으로 재사용하지 않는다. 문서 i18n 워크플로의 `strict=false`는 CI 성공 여부와 별개인 운영 설정이지만, 이번 실행 로그의 오류와 경고는 모두 0건이었다.
