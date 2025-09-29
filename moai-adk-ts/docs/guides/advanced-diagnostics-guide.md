# MoAI-ADK 고급 진단 기능 사용자 가이드

**최종 업데이트**: 2025-09-29
**버전**: v0.0.1
**태그**: @GUIDE:ADVANCED-DIAGNOSTICS-001 @DOCS:USER-GUIDE-001

## 개요

MoAI-ADK의 고급 진단 기능은 개발 환경의 성능과 상태를 종합적으로 분석하여 최적화 권장사항을 제공합니다. 기본 시스템 검증을 넘어서 실제 개발 생산성에 영향을 미치는 요소들을 심층 분석합니다.

## 🚀 빠른 시작

### 기본 사용법

```bash
# 고급 진단 실행 (기본 성능 메트릭만)
moai doctor --advanced

# 전체 기능 활성화
moai doctor --advanced --include-benchmarks --include-recommendations --include-environment-analysis --verbose
```

### 1분 완전 진단

가장 포괄적인 진단을 원한다면:

```bash
moai doctor --advanced \
  --include-benchmarks \
  --include-recommendations \
  --include-environment-analysis \
  --verbose
```

**예상 소요 시간**: 30초 ~ 1분
**출력**: 성능 메트릭, 벤치마크 결과, 최적화 권장사항, 환경 분석, 건강도 점수

## 📊 진단 결과 이해하기

### 시스템 건강도 점수

고급 진단의 핵심은 **0-100점의 건강도 점수**입니다:

```
🎯 System Health Score:
  85/100 - GOOD
```

#### 점수 해석

| 점수 범위 | 등급 | 상태 | 권장 조치 |
|-----------|------|------|-----------|
| **90-100점** | Excellent | 최적 상태 | 현재 상태 유지 |
| **70-89점** | Good | 양호 | 경미한 최적화 권장 |
| **50-69점** | Fair | 보통 | 개선 작업 필요 |
| **0-49점** | Poor | 문제 있음 | 즉시 조치 필요 |

### 성능 메트릭 읽기

```
📊 Performance Metrics:
  CPU Usage: 45.2%
  Memory Usage: 68% (5,461MB/8,192MB)
  Disk Usage: 72% (285GB/394GB)
  Network Latency: 23ms
```

#### 메트릭 기준값

| 메트릭 | 우수 | 양호 | 주의 | 위험 |
|--------|------|------|------|------|
| **CPU 사용률** | <40% | 40-60% | 60-80% | >80% |
| **메모리 사용률** | <50% | 50-70% | 70-85% | >85% |
| **디스크 사용률** | <80% | 80-90% | 90-95% | >95% |
| **네트워크 지연** | <50ms | 50-100ms | 100-200ms | >200ms |

### 벤치마크 결과 분석

```
🏃 Benchmark Results:
  ✅ File I/O: 92/100 (156ms)
  ✅ CPU Operations: 88/100 (234ms)
  ⚠️  Memory Allocation: 65/100 (445ms)
  ❌ JSON Processing: 45/100 (890ms)
```

#### 벤치마크 성능 기준

| 벤치마크 | 우수 (90-100점) | 양호 (70-89점) | 보통 (50-69점) | 개선 필요 (<50점) |
|----------|-----------------|----------------|----------------|-------------------|
| **File I/O** | >150MB/s | 100-150MB/s | 50-100MB/s | <50MB/s |
| **CPU Ops** | >2M ops/s | 1-2M ops/s | 0.5-1M ops/s | <0.5M ops/s |
| **Memory** | <5ms GC | 5-10ms GC | 10-20ms GC | >20ms GC |
| **JSON** | >20MB/s | 10-20MB/s | 5-10MB/s | <5MB/s |

## 🔧 최적화 권장사항 활용하기

### 권장사항 우선순위

고급 진단은 심각도별로 권장사항을 제시합니다:

```
💡 Top Recommendations:
  1. 🚨 Critical: High memory usage detected (95%)
     Close unnecessary applications or upgrade RAM

  2. ❌ Error: Disk space critically low (97%)
     Free up disk space by cleaning temporary files

  3. ⚠️  Warning: CPU usage consistently high (78%)
     Consider closing background applications

  4. ℹ️  Info: Node.js version outdated
     Upgrade to Node.js 20.x for better performance
```

### 권장사항 구현 가이드

#### 🚨 Critical Issues (즉시 조치)

**높은 메모리 사용률 (>85%)**
```bash
# 메모리 사용량 확인
htop  # Linux/macOS
taskmgr  # Windows

# Node.js 메모리 사용량 최적화
export NODE_OPTIONS="--max-old-space-size=4096"
```

**디스크 공간 부족 (>90%)**
```bash
# 임시 파일 정리
npm cache clean --force
yarn cache clean
rm -rf node_modules/.cache

# 대용량 파일 찾기
du -sh * | sort -rh | head -10
```

#### ❌ Error Issues (빠른 해결 필요)

**Node.js 버전 문제**
```bash
# Node.js 최신 LTS 설치
# macOS
brew install node

# Linux (Ubuntu/Debian)
curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
sudo apt-get install -y nodejs

# Windows
choco install nodejs
```

**Git 설정 문제**
```bash
# Git 사용자 정보 설정
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# Git 성능 최적화
git config --global core.preloadindex true
git config --global core.fscache true
```

#### ⚠️ Warning Issues (개선 권장)

**CPU 사용률 높음 (60-80%)**
```bash
# 백그라운드 프로세스 확인
ps aux | grep -v grep | sort -k 3 -r | head -10  # Linux/macOS
wmic process get name,processid,percentprocessortime  # Windows

# VS Code 확장 프로그램 최적화
# Settings > Extensions > 불필요한 확장 비활성화
```

**패키지 매니저 최적화**
```bash
# Bun 설치 (98% 성능 향상)
curl -fsSL https://bun.sh/install | bash

# npm 대신 Bun 사용
bun install  # npm install 대신
bun run build  # npm run build 대신
```

#### ℹ️ Info Issues (선택적 최적화)

**TypeScript 컴파일 최적화**
```json
// tsconfig.json
{
  "compilerOptions": {
    "incremental": true,
    "composite": true,
    "skipLibCheck": true
  }
}
```

**Vitest 성능 향상**
```javascript
// vitest.config.ts
export default defineConfig({
  test: {
    pool: 'threads',
    poolOptions: {
      threads: {
        singleThread: true
      }
    }
  }
});
```

## 🛠️ 환경 분석 결과 해석

### 개발 도구 상태

```
🛠️ Development Environments:
  ✅ Node.js 20.10.0 - optimal
  👍 npm 10.2.3 - good
  ⚠️  TypeScript 4.9.5 - warning
  ❌ Python 3.8.0 - poor
```

#### 환경 상태 의미

| 상태 | 아이콘 | 의미 | 조치 |
|------|--------|------|------|
| **optimal** | ✅ | 최신 안정 버전 | 유지 |
| **good** | 👍 | 사용 가능한 버전 | 선택적 업그레이드 |
| **warning** | ⚠️ | 업그레이드 권장 | 가능한 빨리 업그레이드 |
| **poor** | ❌ | 호환성 문제 | 즉시 업그레이드 필요 |

### 환경별 최적화 가이드

#### TypeScript 환경 최적화

```bash
# TypeScript 최신 버전 설치
npm install -g typescript@latest

# 프로젝트별 TypeScript 버전 관리
npm install --save-dev typescript@^5.0.0

# 타입 체크 성능 향상
tsc --noEmit --incremental
```

#### Python 개발 환경 최적화

```bash
# Python 최신 버전 설치
# macOS
brew install python@3.12

# Linux (Ubuntu/Debian)
sudo apt update
sudo apt install python3.12 python3.12-venv

# Virtual Environment 설정
python3.12 -m venv venv
source venv/bin/activate  # Linux/macOS
venv\Scripts\activate  # Windows
```

#### Java 개발 환경 설정

```bash
# OpenJDK 최신 LTS 설치
# macOS
brew install openjdk@21

# Linux (Ubuntu/Debian)
sudo apt install openjdk-21-jdk

# Windows
choco install openjdk
```

## 📈 성능 벤치마크 개선 전략

### File I/O 성능 최적화

**SSD 최적화**
```bash
# macOS: TRIM 활성화 확인
sudo trimforce enable

# Linux: SSD 최적화 설정
echo 'deadline' | sudo tee /sys/block/sda/queue/scheduler
```

**Node.js I/O 최적화**
```javascript
// 비동기 I/O 활용
import { promises as fs } from 'fs';

// 동기식 (느림)
const data = fs.readFileSync('large-file.txt');

// 비동기식 (빠름)
const data = await fs.readFile('large-file.txt');
```

### CPU 성능 최적화

**Node.js 워커 스레드 활용**
```javascript
// worker-thread.js
import { Worker, isMainThread, parentPort } from 'worker_threads';

if (isMainThread) {
  const worker = new Worker(__filename);
  worker.postMessage({ task: 'heavy-computation' });
} else {
  parentPort.on('message', ({ task }) => {
    // CPU 집약적 작업 수행
    const result = performHeavyComputation();
    parentPort.postMessage(result);
  });
}
```

**프로세스 우선순위 조정**
```bash
# Linux/macOS: 프로세스 우선순위 높이기
nice -n -10 node app.js

# Windows: 우선순위 조정
wmic process where name="node.exe" CALL setpriority "high priority"
```

### 메모리 성능 최적화

**Node.js 가비지 컬렉션 튜닝**
```bash
# V8 플래그 설정
export NODE_OPTIONS="--max-old-space-size=8192 --optimize-for-size"

# 가비지 컬렉션 로그 활성화
node --trace-gc app.js
```

**메모리 누수 감지**
```javascript
// 메모리 사용량 모니터링
setInterval(() => {
  const used = process.memoryUsage();
  console.log('Memory Usage:', {
    rss: Math.round(used.rss / 1024 / 1024) + 'MB',
    heapTotal: Math.round(used.heapTotal / 1024 / 1024) + 'MB',
    heapUsed: Math.round(used.heapUsed / 1024 / 1024) + 'MB'
  });
}, 5000);
```

## 🎯 실전 활용 시나리오

### 시나리오 1: 개발 환경 초기 설정

**목표**: 새로운 머신에서 최적의 개발 환경 구축

```bash
# 1. 기본 진단으로 현재 상태 확인
moai doctor

# 2. 고급 진단으로 성능 기준선 설정
moai doctor --advanced --include-benchmarks --verbose

# 3. 권장사항 기반 환경 최적화
# (권장사항 따라 도구 설치/업그레이드)

# 4. 최적화 후 재진단
moai doctor --advanced --include-benchmarks --include-recommendations --verbose
```

### 시나리오 2: 성능 저하 문제 해결

**목표**: 갑자기 느려진 개발 환경의 원인 분석

```bash
# 1. 전체 진단 실행
moai doctor --advanced --include-benchmarks --include-recommendations --include-environment-analysis --verbose

# 2. 건강도 점수 확인 (70점 미만이면 문제)
# 3. Critical/Error 권장사항 우선 해결
# 4. 벤치마크 결과에서 성능 병목 지점 식별
# 5. 환경 분석에서 버전 충돌 확인
```

### 시나리오 3: 팀 환경 표준화

**목표**: 팀 전체의 개발 환경을 일관성 있게 유지

```bash
# 1. 표준 환경에서 진단 결과 생성
moai doctor --advanced --include-environment-analysis > standard-report.txt

# 2. 팀원들에게 동일한 진단 실행 요청
moai doctor --advanced --include-environment-analysis > my-report.txt

# 3. 차이점 분석 및 표준화 작업
diff standard-report.txt my-report.txt
```

### 시나리오 4: CI/CD 환경 최적화

**목표**: 지속적 통합 환경의 성능 최적화

```bash
# GitHub Actions에서 사용
- name: System Diagnostics
  run: |
    npm install -g moai-adk
    moai doctor --advanced --include-benchmarks

# 성능 기준선 설정
- name: Performance Check
  run: |
    HEALTH_SCORE=$(moai doctor --advanced --include-benchmarks | grep "Health Score" | cut -d: -f2 | cut -d/ -f1)
    if [ $HEALTH_SCORE -lt 70 ]; then
      echo "Performance threshold not met: $HEALTH_SCORE/100"
      exit 1
    fi
```

## 📋 트러블슈팅

### 일반적인 문제와 해결책

#### 문제: 진단이 느리게 실행됨

**원인**: 벤치마크 실행 시 시스템 리소스 부족
**해결책**:
```bash
# 벤치마크 없이 실행
moai doctor --advanced --include-recommendations

# 타임아웃 설정
timeout 30s moai doctor --advanced --include-benchmarks
```

#### 문제: 권한 에러 발생

**원인**: 시스템 정보 접근 권한 부족
**해결책**:
```bash
# macOS: 터미널 전체 디스크 접근 권한 부여
# 시스템 환경설정 > 보안 및 개인정보보호 > 전체 디스크 접근 권한

# Linux: sudo 없이 시스템 정보 접근
sudo chmod +r /proc/meminfo /proc/cpuinfo
```

#### 문제: 벤치마크 결과가 일관되지 않음

**원인**: 백그라운드 프로세스의 간섭
**해결책**:
```bash
# 시스템 안정화 후 재실행
sleep 30 && moai doctor --advanced --include-benchmarks

# 여러 번 실행하여 평균값 확인
for i in {1..3}; do
  echo "Run $i:"
  moai doctor --advanced --include-benchmarks
  sleep 10
done
```

### 고급 디버깅

#### 상세 로그 활성화

```bash
# 환경 변수 설정
export MOAI_LOG_LEVEL=debug
export MOAI_VERBOSE=true

# 진단 실행
moai doctor --advanced --verbose 2>&1 | tee diagnostics.log
```

#### 성능 프로파일링

```bash
# Node.js 성능 프로파일링
node --prof $(which moai) doctor --advanced

# 프로파일 분석
node --prof-process isolate-*.log > profile.txt
```

## 📚 추가 리소스

### 관련 명령어

- `moai doctor --help`: 모든 진단 옵션 확인
- `moai doctor --list-backups`: 백업 디렉토리 관리
- `moai status`: 프로젝트 상태 확인
- `moai init --help`: 프로젝트 초기화 옵션

### 외부 도구 연동

#### VS Code 통합

```json
// .vscode/tasks.json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "System Diagnostics",
      "type": "shell",
      "command": "moai",
      "args": ["doctor", "--advanced", "--verbose"],
      "group": "test",
      "presentation": {
        "echo": true,
        "reveal": "always",
        "focus": false,
        "panel": "new"
      }
    }
  ]
}
```

#### Git Hook 통합

```bash
# .git/hooks/pre-commit
#!/bin/bash
echo "Running system diagnostics..."
HEALTH_SCORE=$(moai doctor --advanced | grep "Health Score" | cut -d: -f2 | cut -d/ -f1 | tr -d ' ')

if [ "$HEALTH_SCORE" -lt 50 ]; then
    echo "Warning: System health score is low ($HEALTH_SCORE/100)"
    echo "Consider running optimization before committing."
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi
```

---

**참고 문서**:
- [CLI Commands API](../api/cli-commands.md)
- [Diagnostics System API](../api/diagnostics-system.md)
- [Architecture Guide](../architecture/diagnostics-architecture.md)