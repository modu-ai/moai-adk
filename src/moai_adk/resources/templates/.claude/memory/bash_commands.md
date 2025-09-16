# 자주 사용하는 Bash 명령어

> MoAI-ADK 기반 프로젝트의 표준 명령어 모음집

## ✅ 권장 도구와 안전 수칙(추가)

- 검색/목록: `rg`(ripgrep), `fd` 사용 권장 — 기존 `grep/find` 대비 빠르고 안전한 패턴 지정이 가능
- 쉘 안전: `set -euo pipefail`와 인자 quoting(`"$VAR"`) 습관화, `--` 구분자 사용
- 위험 명령 차단: 무심코 `rm -rf` 사용 금지. 꼭 필요한 경우 대상 경로를 풀리솔브 후 확인
- 네트워크: `curl -m <timeout> --retry <n>` 등 타임아웃/재시도 옵션 사용
- 권한: `sudo` 최소화, 민감 경로(`.env`, `.git/`, keys) 직접 조작 금지

예시(검색/파일 목록):
```bash
# files 목록
rg --files -g "*.ts" src | head -50

# 내용 검색(줄번호 포함)
rg -n "TODO|FIXME" src

# find 대체(fd)
fd -e py tests
```

## 🔨 빌드와 테스트

### Python 프로젝트
```bash
# 개발 환경 설정
python -m venv venv
source venv/bin/activate  # Linux/Mac
venv\Scripts\activate     # Windows

# 의존성 설치
python -m pip install -e .              # 개발 모드 설치
python -m pip install -r requirements.txt  # 운영 의존성
python -m pip install -r requirements-dev.txt  # 개발 의존성

# 테스트 실행
python -m pytest                        # 기본 테스트
python -m pytest --cov                  # 커버리지 포함
python -m pytest --cov --cov-report=html  # HTML 리포트
python -m pytest -v --tb=short          # 상세 출력

# 코드 품질
black .                                  # 코드 포매팅
isort .                                  # import 정렬
pylint src/                             # 정적 분석
mypy src/                               # 타입 체크

# 패키지 빌드
python -m build                         # 배포 패키지 생성
python -m twine upload dist/*           # PyPI 업로드
```

### Node.js 프로젝트
```bash
# 의존성 설치
npm install                             # 의존성 설치
npm ci                                  # CI용 클린 설치
npm audit fix                           # 보안 취약점 수정

# 개발 서버
npm run dev                             # 개발 서버 시작
npm run start                           # 프로덕션 서버
npm run watch                           # 파일 변경 감지

# 테스트
npm test                                # 테스트 실행
npm run test:watch                      # 감시 모드 테스트
npm run test:coverage                   # 커버리지 포함

# 빌드
npm run build                           # 프로덕션 빌드
npm run build:dev                       # 개발 빌드
npm run preview                         # 빌드 미리보기

# 코드 품질
npm run lint                            # ESLint 실행
npm run lint:fix                        # 자동 수정
npm run format                          # Prettier 포매팅
npm run type-check                      # TypeScript 타입 체크
```

## 🔍 디버깅과 로그

### 로그 확인
```bash
# 애플리케이션 로그
tail -f logs/app.log                    # 실시간 로그
grep "ERROR" logs/app.log               # 에러 로그 필터링
journalctl -u myservice -f              # systemd 서비스 로그

# 시스템 모니터링
top                                     # 프로세스 모니터링
htop                                    # 개선된 top
ps aux | grep python                   # 특정 프로세스 찾기
netstat -tulpn | grep :8000            # 포트 사용 확인
```

### 성능 분석
```bash
# Python 프로파일링
python -m cProfile -o profile.stats script.py
python -c "import pstats; p=pstats.Stats('profile.stats'); p.sort_stats('cumulative'); p.print_stats(10)"

# 메모리 사용량 확인
free -h                                 # 시스템 메모리
du -sh ./*                              # 디렉토리별 용량
df -h                                   # 디스크 사용량
```

## 🐳 Docker 관련

### 기본 Docker 명령어
```bash
# 이미지 빌드
docker build -t myapp:latest .          # 이미지 빌드
docker build --no-cache -t myapp .      # 캐시 없이 빌드

# 컨테이너 실행
docker run -d -p 8000:8000 myapp        # 백그라운드 실행
docker run -it --rm myapp /bin/bash     # 대화형 실행

# 관리 명령어
docker ps                               # 실행 중인 컨테이너
docker logs -f container_name           # 컨테이너 로그
docker exec -it container_name /bin/bash # 컨테이너 접속
docker system prune -a                  # 미사용 리소스 정리
```

### Docker Compose
```bash
# 서비스 실행
docker-compose up -d                    # 백그라운드 실행
docker-compose up --build              # 빌드 후 실행
docker-compose logs -f service_name     # 특정 서비스 로그

# 서비스 관리
docker-compose ps                       # 서비스 상태
docker-compose restart service_name     # 서비스 재시작
docker-compose down                     # 서비스 중지 및 제거
```

## 🔧 시스템 관리

### 파일 작업
```bash
# 파일 검색
fd -e py                                 # (권장) 파일 찾기
fd -e log -X rm -f                       # (주의) 삭제 전 목록 확인 권장
rg -n "filename"                         # (권장) 빠른 파일 검색

# 텍스트 처리
rg -n "TODO" src/                        # (권장) 재귀 검색
rg -n "error|Error|ERROR" logs/          # (권장) 정규식 검색
sed 's/old_text/new_text/g' file.txt    # 텍스트 치환
awk '{print $1}' file.txt               # 컬럼 추출
```

### 프로세스 관리
```bash
# 백그라운드 실행
nohup python app.py &                   # 터미널 종료 후에도 실행
screen -S session_name                  # 스크린 세션 생성
tmux new-session -s session_name        # tmux 세션 생성

# 프로세스 종료
kill -9 PID                             # 강제 종료
killall python                          # 모든 python 프로세스 종료
pkill -f "python app.py"                # 패턴으로 프로세스 종료
```

## 📊 모니터링과 성능

### 시스템 성능
```bash
# CPU 및 메모리
iostat -x 1                             # I/O 통계
sar -u 1 10                             # CPU 사용률
vmstat 1                                # 가상 메모리 통계

# 네트워크
ss -tuln                                # 소켓 상태
iftop                                   # 네트워크 사용량
curl -w "@curl-format.txt" http://url   # HTTP 응답 시간 측정
```

### 데이터베이스 관련
```bash
# PostgreSQL
psql -U user -d database                # 데이터베이스 접속
pg_dump database > backup.sql           # 백업
psql -U user -d database < backup.sql   # 복원

# MySQL/MariaDB
mysql -u user -p database               # 데이터베이스 접속
mysqldump -u user -p database > backup.sql  # 백업
mysql -u user -p database < backup.sql  # 복원

# Redis
redis-cli                               # Redis CLI
redis-cli --scan --pattern "*key*"     # 키 패턴 검색
```

## 🚀 배포 관련

### 서버 배포
```bash
# SSH 접속
ssh user@server                         # 기본 접속
ssh -i keyfile.pem user@server          # 키 파일 사용
scp file.txt user@server:/path/         # 파일 복사

# 서비스 관리 (systemd)
sudo systemctl start myservice          # 서비스 시작
sudo systemctl enable myservice         # 부팅 시 자동 시작
sudo systemctl status myservice         # 서비스 상태
sudo systemctl reload myservice         # 설정 다시 로드
```

### 환경 설정
```bash
# 환경 변수
export API_KEY="your-key"               # 환경 변수 설정
echo $API_KEY                           # 환경 변수 확인
printenv | grep API                     # API 관련 환경 변수

# crontab 작업
crontab -e                              # crontab 편집
crontab -l                              # crontab 목록
# 매일 오전 2시 실행: 0 2 * * * /path/to/script.sh
```

## ⚡ 성능 최적화 팁

### Python 최적화
```bash
# 프로파일링
python -m timeit "code_to_test"         # 코드 실행 시간 측정
python -m memory_profiler script.py     # 메모리 사용량 분석

# 캐시 최적화
python -O script.py                     # 최적화 모드 실행
python -OO script.py                    # 고급 최적화
```

### Node.js 최적화
```bash
# 성능 분석
node --prof app.js                      # 프로파일링
node --prof-process isolate-*.log       # 프로파일 분석

# 메모리 최적화
node --max-old-space-size=4096 app.js   # 메모리 제한 설정
```

---

**참고**: 이 명령어들은 MoAI-ADK의 표준 워크플로우에 최적화되어 있습니다. 프로젝트별로 필요에 따라 수정하여 사용하세요.

**마지막 업데이트**: 2025-09-12  
**버전**: v0.1.12
