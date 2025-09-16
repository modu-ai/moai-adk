# Shell 규칙(요약)

- set -euo pipefail, IFS 안전 설정
- 변수/경로는 반드시 쿼팅("$VAR"), `--` 구분자 사용
- rm -rf 등 파괴 명령 제한, 경로 풀리솔브/확인 절차
- grep/find 대신 rg/fd 권장, xargs -0 사용
