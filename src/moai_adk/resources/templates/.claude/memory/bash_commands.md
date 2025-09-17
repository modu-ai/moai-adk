# Bash 명령어(요약)

자주 쓰는 CLI/배포/데이터베이스 명령 전체 목록은 `@.moai/memory/operations.md` 5장에서 확인하세요. 여기에는 최소한의 핵심 명령만 남겨둡니다.

- 검색: `rg -n "패턴" src`, `fd -e py tests`
- 테스트: Python `pytest --cov`, Node `pnpm test`, 공통 `pre-commit run --all-files`
- 컨테이너: `docker build -t $PROJECT_NAME:latest .`, `docker compose up -d`
- 배포: `ssh user@host`, `sudo systemctl restart service`, `kubectl rollout undo deployment/app`

### 참조
- 운영 메모 전문: `@.moai/memory/operations.md`
