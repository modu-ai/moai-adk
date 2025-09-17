# Shell 규칙

## ✅ 필수
- 스크립트 시작 시 `set -euo pipefail` 및 안전한 IFS 설정(`IFS=$'\n\t'`)
- 변수/경로는 반드시 `"$VAR"` 로 감싸고, 위험 명령은 `--` 구분자와 경로 확인 절차 수행
- `rm -rf` 등 파괴 명령은 금지 또는 승인 절차(경고/대상 출력) 후 실행
- grep/find 대신 `rg`, `fd` 사용, `xargs -0`/`readarray` 등 안전한 파이프 사용

## 👍 권장
- shellcheck로 정적 분석, shfmt로 포맷팅, shebang은 목적에 맞게(`#!/usr/bin/env bash`)
- 함수/스크립트는 POSIX 호환을 고려하고, 장비별 차이는 `uname`/feature detect로 분기
- 로그/에러는 표준 출력과 분리(`exec 3>&1 4>&2` 패턴 등)

## 🚀 확장/고급
- 복잡한 로직은 Python/Go 등으로 이관, Shell은 orchestration 수준 유지
- tmux/screen/atop 등 운영 도구와 연계, CLI argument parser(shellopt, getopts) 사용
- CI 스크립트에는 dry-run/verbose 옵션 제공, 실행 전 `set -x`를 조건부로 활성화
