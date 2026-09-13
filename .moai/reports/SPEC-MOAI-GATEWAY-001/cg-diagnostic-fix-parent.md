# CG 폐기 진단 — 부모 실행 파일 재검증

## Claim

실제 working binary의 `cg`, `cg --help`, `cg --spawn -f 2`가 각각 exit 1, 빈 stdout, stderr의 이전 안내 한 번을 반환했다. 처음 관측한 무출력 오류는 같은 subprocess 경계에서 해결되었다. hybrid apply는 정확한 명령으로 capability 오류를 반환하고 원본·백업을 바꾸지 않았다.

## Evidence

부모는 각 명령을 Python `subprocess.run(..., capture_output=True, text=True, timeout=15)`로 실행했다. `tempfile.TemporaryDirectory`의 프로젝트에 `llm:\n  team_mode: cg\n`만 쓰고, PATH `/usr/bin:/bin`, 그 프로젝트의 HOME·MOAI_HOME만 전달했다. stdout/stderr/exit와 설정 원문 및 백업 부재를 직접 판정했다. 앞 RED는 `/tmp/gateway-cg-binary-gates-red-20260911.json`, 최종 원문은 `/tmp/gateway-cg-binary-gates-green-20260911.json`이다.

최종 빌드 명령:

```sh
GOCACHE=/tmp/gateway-foundation-cache go build -o /tmp/moai-gateway-working-20260911 ./cmd/moai
```

module stat cache 쓰기 권한 경고가 있었으며 최종 exit 0, 후속 출력 없음이었다. 빌드 뒤 바이너리 SHA-256은 `93f685aa9425ac48b179607548e1c8c07a09d4cdbb12715fb1597dc78eb97360`이다.

원래 RED:

```json
{"args":["cg"],"exit":1,"stdout":"","stderr":"","source":"llm:\n  team_mode: cg\n","backups_created":false}
```

최종 JSON 중 세 CG 조합은 각각 exit 1, stdout `""`, pass `true`였으며 stderr 원문은 동일했다:

```text
moai cg is retired; run moai migrate cg to preview an explicit teammate-role migration
```

`migrate cg --target claude-glm --apply`도 exit 1, stdout `""`, pass `true`였다. stderr 오류 본문은 다음과 같았다:

```text
Claude-Glm migration is unavailable until teammate routing capability is verified.
```

전체 결과는 `"all_pass": true`, Python 실행은 exit 0이었다. 이 검증은 처음 smoke의 부정확한 hybrid consent flag를 사용하지 않는다. temporary project는 context manager 종료로 정리되었다.

## Baseline-attribution

2026-09-11 WT `moai-proxy-unified`, branch `WT-unified-gateway`, HEAD `81c1d58f9`. 수리는 root early guard에서 기존 `moaiErrorHandler`를 한 번 호출하는 변경이다. 작성자의 main-equivalent helper RED/GREEN·race 증거는 `pre-live-ac-map.md` 후반에 있으며, 이 보고서는 부모의 새 working binary 측정이다.

## Gaps

실제 Claude·tmux·provider를 실행하지 않았다. 자식 실행 0의 내부 counter는 작성자의 기존 회귀 시험 범위이며 이 subprocess 결과만으로 별도 OS 감시를 했다고 주장하지 않는다. 원격 CI·release binary·Windows 실행 검증은 아니다.

## Residual-risk

기존 renderer와 별도의 조기 경로가 유지되므로 미래 오류 표시 정책 변경 시 같은 실행 파일 시험을 보존해야 한다. 이 작은 진단 수리로 전체 gateway·CG-RETIRE 관련 AC를 완료 처리하지 않는다.
