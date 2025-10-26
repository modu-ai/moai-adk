# moai-essentials-debug — 기술 레퍼런스

> **Version**: 2.1.0  
> **Last Updated**: 2025-10-27

이 문서는 moai-essentials-debug Skill의 상세한 기술 사양, 23개 언어별 디버거 설정, 컨테이너/분산 시스템 디버깅 가이드를 제공합니다.

---

## 23개 언어별 디버거 매트릭스

### Systems Programming

#### C/C++
**디버거**: GDB 14.x, LLDB 17.x, AddressSanitizer

**CLI 명령어**:
```bash
# GDB 기본 사용
gdb ./myapp
(gdb) break main
(gdb) run
(gdb) next
(gdb) print variable
(gdb) backtrace

# LLDB 사용
lldb ./myapp
(lldb) b main
(lldb) run
(lldb) n
(lldb) p variable
(lldb) bt

# AddressSanitizer (메모리 오류 감지)
gcc -fsanitize=address -g myapp.c -o myapp
./myapp
```

**VSCode launch.json**:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "C++ Debug",
      "type": "cppdbg",
      "request": "launch",
      "program": "${workspaceFolder}/build/myapp",
      "args": [],
      "stopAtEntry": false,
      "cwd": "${workspaceFolder}",
      "environment": [],
      "externalConsole": false,
      "MIMode": "gdb",
      "setupCommands": [
        {
          "description": "Enable pretty-printing",
          "text": "-enable-pretty-printing",
          "ignoreFailures": true
        }
      ]
    }
  ]
}
```

#### Rust
**디버거**: rust-lldb, rust-gdb, RUST_BACKTRACE

**CLI 명령어**:
```bash
# 백트레이스 활성화
RUST_BACKTRACE=1 cargo run
RUST_BACKTRACE=full cargo run  # 전체 백트레이스

# rust-lldb 사용
rust-lldb target/debug/myapp
(lldb) b main
(lldb) run

# rust-gdb 사용
rust-gdb target/debug/myapp
(gdb) break main
(gdb) run
```

**VSCode launch.json**:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Rust Debug",
      "type": "lldb",
      "request": "launch",
      "program": "${workspaceFolder}/target/debug/${workspaceFolderBasename}",
      "args": [],
      "cwd": "${workspaceFolder}",
      "env": {
        "RUST_BACKTRACE": "1"
      }
    }
  ]
}
```

#### Go
**디버거**: Delve 1.22.x

**CLI 명령어**:
```bash
# Delve 디버깅
dlv debug
(dlv) break main.main
(dlv) continue
(dlv) next
(dlv) print variable
(dlv) goroutines  # 고루틴 목록
(dlv) goroutine 1  # 특정 고루틴 전환

# 실행 중인 프로세스에 연결
dlv attach <pid>

# 테스트 디버깅
dlv test -- -test.run TestName
```

**VSCode launch.json**:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Go Debug",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}",
      "env": {},
      "args": []
    }
  ]
}
```

---

### JVM Ecosystem

#### Java
**디버거**: jdb, IntelliJ IDEA, Remote JDWP

**CLI 명령어**:
```bash
# jdb 사용
jdb -classpath . MyApp
> stop at MyClass:42
> run
> step
> print variable

# 원격 디버깅 활성화
java -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005 MyApp
```

**원격 디버깅 설정**:
```bash
# IntelliJ/VSCode 원격 디버거 연결
Host: localhost
Port: 5005
```

#### Kotlin
**디버거**: IntelliJ Kotlin Debugger, Coroutines Debugger

**코루틴 디버깅**:
```kotlin
// 코루틴 디버그 정보 활성화
kotlinOptions {
    freeCompilerArgs += ["-Xdebug"]
}
```

**IntelliJ 코루틴 뷰어**: View → Tool Windows → Kotlin Coroutines

#### Scala
**디버거**: IntelliJ Scala Plugin, sbt debug mode

**sbt 디버그 모드**:
```bash
# sbt 디버그 모드 실행
sbt -jvm-debug 5005 run
```

#### Clojure
**디버거**: CIDER, Cursive, REPL-based debugging

**CIDER 디버깅**:
```clojure
;; 브레이크포인트 설정
#break
(defn my-function [x]
  #break  ; 여기서 중단
  (+ x 1))
```

---

### Scripting Languages

#### Python
**디버거**: pdb, debugpy 1.8.0, pudb (TUI)

**CLI 명령어**:
```bash
# pdb 사용
python -m pdb script.py
(Pdb) break module.py:42
(Pdb) continue
(Pdb) next
(Pdb) print(variable)
(Pdb) where  # 스택 트레이스

# pudb TUI 디버거
python -m pudb script.py

# debugpy 원격 디버깅
python -m debugpy --listen 5678 script.py
```

**코드 내 브레이크포인트**:
```python
# 코드에서 pdb 시작
import pdb; pdb.set_trace()

# Python 3.7+ breakpoint()
breakpoint()
```

**VSCode launch.json**:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Python Debug",
      "type": "debugpy",
      "request": "launch",
      "program": "${file}",
      "console": "integratedTerminal",
      "justMyCode": false
    }
  ]
}
```

#### Ruby
**디버거**: debug gem (built-in), byebug, pry-byebug

**CLI 명령어**:
```bash
# Ruby 3.1+ 내장 디버거
ruby -r debug script.rb
debugger  # 코드 내

# byebug 사용
gem install byebug
# 코드 내에서
require 'byebug'
byebug

# pry-byebug 사용
gem install pry-byebug
# 코드 내에서
require 'pry-byebug'
binding.pry
```

#### PHP
**디버거**: Xdebug 3.3.x, phpdbg

**Xdebug 설정** (php.ini):
```ini
[xdebug]
zend_extension=xdebug
xdebug.mode=debug
xdebug.start_with_request=yes
xdebug.client_port=9003
```

**phpdbg 사용**:
```bash
phpdbg -e script.php
phpdbg> break script.php:42
phpdbg> run
```

#### Lua
**디버거**: ZeroBrane Studio, MobDebug

**MobDebug 사용**:
```lua
require("mobdebug").start()
-- 브레이크포인트
require("mobdebug").pause()
```

#### Shell (Bash)
**디버깅 모드**:
```bash
# 디버그 모드 실행
bash -x script.sh

# 스크립트 내 토글
set -x  # 디버그 모드 활성화
# ... 코드 ...
set +x  # 디버그 모드 비활성화
```

---

### Web & Mobile

#### JavaScript
**디버거**: Chrome DevTools, node --inspect

**Node.js 디버깅**:
```bash
# 디버그 모드 실행
node --inspect script.js
# 또는 첫 줄에서 중단
node --inspect-brk script.js

# Chrome DevTools 연결
chrome://inspect
```

**코드 내 브레이크포인트**:
```javascript
debugger;  // 여기서 중단
```

#### TypeScript
**디버거**: Chrome DevTools + Source Maps, VS Code

**tsconfig.json 소스맵 설정**:
```json
{
  "compilerOptions": {
    "sourceMap": true,
    "inlineSources": true
  }
}
```

**VSCode launch.json**:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "TypeScript Debug",
      "type": "node",
      "request": "launch",
      "program": "${workspaceFolder}/src/index.ts",
      "preLaunchTask": "tsc: build - tsconfig.json",
      "outFiles": ["${workspaceFolder}/dist/**/*.js"],
      "sourceMaps": true
    }
  ]
}
```

#### Dart/Flutter
**디버거**: Flutter DevTools, Hot Reload

**CLI 명령어**:
```bash
# Flutter 디버그 모드 실행
flutter run --debug

# DevTools 열기
flutter pub global activate devtools
flutter pub global run devtools
```

#### Swift
**디버거**: LLDB (Xcode), Instruments

**LLDB 사용**:
```bash
lldb MyApp.app
(lldb) breakpoint set --name viewDidLoad
(lldb) run
(lldb) po variable  # Print Object
```

---

### Functional & Concurrency

#### Haskell
**디버거**: GHCi debugger, Debug.Trace

**GHCi 디버깅**:
```haskell
-- GHCi에서
:break module.function
:trace expression
:step
:continue
```

**Debug.Trace 사용**:
```haskell
import Debug.Trace

myFunction x = trace ("x = " ++ show x) $ x + 1
```

#### Elixir
**디버거**: IEx debugger, :observer.start()

**IEx 디버깅**:
```elixir
# 코드 내
require IEx
IEx.pry()

# Observer 실행
:observer.start()
```

#### Julia
**디버거**: Debugger.jl, Infiltrator.jl

**Debugger.jl 사용**:
```julia
using Debugger

@enter myfunction(args)
# 또는
@bp  # 브레이크포인트
```

#### R
**디버거**: browser(), debug(), RStudio Debugger

**CLI 명령어**:
```r
# 함수 디버깅
debug(my_function)
my_function(args)

# 코드 내 브레이크포인트
browser()

# 디버그 모드 해제
undebug(my_function)
```

---

### Enterprise & Data

#### C#
**디버거**: Visual Studio Debugger, Rider, vsdbg

**vsdbg 사용** (Linux/macOS):
```bash
# vsdbg 설치
curl -sSL https://aka.ms/getvsdbgsh | bash /dev/stdin -v latest -l ~/.vsdbg
```

**VSCode launch.json**:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": ".NET Core Debug",
      "type": "coreclr",
      "request": "launch",
      "program": "${workspaceFolder}/bin/Debug/net8.0/MyApp.dll",
      "args": [],
      "cwd": "${workspaceFolder}",
      "stopAtEntry": false
    }
  ]
}
```

#### SQL
**디버거**: EXPLAIN ANALYZE, pg_stat_statements

**PostgreSQL 디버깅**:
```sql
-- 쿼리 플랜 분석
EXPLAIN ANALYZE SELECT * FROM users WHERE id = 1;

-- 느린 쿼리 추적
CREATE EXTENSION pg_stat_statements;
SELECT * FROM pg_stat_statements ORDER BY total_exec_time DESC LIMIT 10;
```

---

## 컨테이너 디버깅 완전 가이드

### Docker 디버깅

#### 기본 디버깅 패턴
```bash
# 1. 실행 중인 컨테이너에 접속
docker exec -it <container_name> /bin/sh

# 2. 로그 확인
docker logs <container_name>
docker logs -f <container_name>  # 실시간

# 3. 컨테이너 상태 확인
docker inspect <container_name>
docker stats <container_name>
```

#### 언어별 원격 디버깅

**Java (JDWP)**:
```dockerfile
# Dockerfile
ENV JAVA_TOOL_OPTIONS='-agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=*:5005'
EXPOSE 5005
```

```bash
# 컨테이너 실행
docker run -p 5005:5005 -p 8080:8080 myapp
```

**Python (debugpy)**:
```dockerfile
# Dockerfile
RUN pip install debugpy
ENV DEBUGPY_ENABLE=true
EXPOSE 5678
```

```python
# app.py
import debugpy

if os.environ.get('DEBUGPY_ENABLE'):
    debugpy.listen(("0.0.0.0", 5678))
    print("Debugger listening on port 5678")
```

```bash
# 컨테이너 실행
docker run -p 5678:5678 -p 8000:8000 myapp
```

**Node.js (--inspect)**:
```dockerfile
# Dockerfile
EXPOSE 9229
CMD ["node", "--inspect=0.0.0.0:9229", "app.js"]
```

```bash
# 컨테이너 실행
docker run -p 9229:9229 -p 3000:3000 myapp
```

**Go (Delve)**:
```dockerfile
# Dockerfile
RUN go install github.com/go-delve/delve/cmd/dlv@latest
EXPOSE 2345
CMD ["dlv", "debug", "--headless", "--listen=:2345", "--api-version=2", "--accept-multiclient"]
```

```bash
# 컨테이너 실행
docker run -p 2345:2345 -p 8080:8080 myapp
```

#### 멀티스테이지 빌드 디버깅
```dockerfile
# 디버그 스테이지
FROM golang:1.22 AS debug
RUN go install github.com/go-delve/delve/cmd/dlv@latest
COPY . .
CMD ["dlv", "debug", "--headless", "--listen=:2345"]

# 프로덕션 스테이지
FROM golang:1.22 AS production
COPY . .
RUN go build -o app
CMD ["./app"]
```

```bash
# 디버그 빌드
docker build --target debug -t myapp:debug .
docker run -p 2345:2345 myapp:debug
```

---

### Kubernetes 디버깅

#### 기본 디버깅 커맨드
```bash
# 1. Pod 로그 확인
kubectl logs <pod-name>
kubectl logs -f <pod-name>  # 실시간
kubectl logs <pod-name> -c <container-name>  # 특정 컨테이너
kubectl logs <pod-name> --previous  # 이전 컨테이너 로그

# 2. Pod에 접속
kubectl exec -it <pod-name> -- /bin/sh
kubectl exec -it <pod-name> -c <container-name> -- /bin/bash

# 3. Pod 상태 확인
kubectl describe pod <pod-name>
kubectl get pod <pod-name> -o yaml
```

#### 포트 포워딩 (디버거 연결)
```bash
# Java JDWP
kubectl port-forward pod/<pod-name> 5005:5005

# Python debugpy
kubectl port-forward pod/<pod-name> 5678:5678

# Node.js --inspect
kubectl port-forward pod/<pod-name> 9229:9229

# Go Delve
kubectl port-forward pod/<pod-name> 2345:2345
```

#### Ephemeral 컨테이너 (K8s 1.23+)
```bash
# 디버그 도구가 있는 임시 컨테이너 추가
kubectl debug -it <pod-name> --image=busybox --target=<container-name>

# 또는 디버그 도구 이미지 사용
kubectl debug -it <pod-name> --image=nicolaka/netshoot --target=<container-name>
```

#### 네트워크 디버깅
```bash
# 네트워크 정책 확인
kubectl get networkpolicies

# Service 엔드포인트 확인
kubectl get endpoints <service-name>

# DNS 확인 (Pod 내부에서)
kubectl exec -it <pod-name> -- nslookup <service-name>
```

#### 리소스 디버깅
```bash
# 리소스 사용량 확인
kubectl top pod <pod-name>
kubectl top node <node-name>

# 이벤트 확인
kubectl get events --sort-by='.lastTimestamp'
kubectl get events --field-selector involvedObject.name=<pod-name>
```

---

## 분산 추적 (Distributed Tracing)

### OpenTelemetry 1.24.0+ 설정

#### Python 설정
```python
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter

# Tracer Provider 설정
trace.set_tracer_provider(TracerProvider())
tracer = trace.get_tracer(__name__)

# OTLP Exporter 설정
otlp_exporter = OTLPSpanExporter(
    endpoint="http://localhost:4317",
    insecure=True
)
trace.get_tracer_provider().add_span_processor(
    BatchSpanProcessor(otlp_exporter)
)

# 사용 예제
with tracer.start_as_current_span("my-operation"):
    # 작업 수행
    pass
```

#### TypeScript/Node.js 설정
```typescript
import { NodeTracerProvider } from '@opentelemetry/sdk-trace-node';
import { BatchSpanProcessor } from '@opentelemetry/sdk-trace-base';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-grpc';

// Tracer Provider 설정
const provider = new NodeTracerProvider();
const exporter = new OTLPTraceExporter({
  url: 'http://localhost:4317',
});

provider.addSpanProcessor(new BatchSpanProcessor(exporter));
provider.register();

// 사용 예제
import { trace } from '@opentelemetry/api';

const tracer = trace.getTracer('my-service');
const span = tracer.startSpan('my-operation');
// 작업 수행
span.end();
```

#### Go 설정
```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Tracer Provider 설정
exporter, _ := otlptracegrpc.New(ctx,
    otlptracegrpc.WithEndpoint("localhost:4317"),
    otlptracegrpc.WithInsecure(),
)

tp := sdktrace.NewTracerProvider(
    sdktrace.WithBatcher(exporter),
)
otel.SetTracerProvider(tp)

// 사용 예제
tracer := otel.Tracer("my-service")
ctx, span := tracer.Start(ctx, "my-operation")
defer span.End()
```

---

### Prometheus 2.48.x 통합

#### Python (prometheus-client 0.19.0)
```python
from prometheus_client import Counter, Histogram, Gauge, start_http_server

# 메트릭 정의
request_count = Counter('http_requests_total', 'Total HTTP requests')
request_duration = Histogram('http_request_duration_seconds', 'HTTP request duration')
active_connections = Gauge('active_connections', 'Active connections')

# 사용 예제
request_count.inc()
with request_duration.time():
    # 작업 수행
    pass
active_connections.set(42)

# 메트릭 서버 시작 (포트 8000)
start_http_server(8000)
```

#### Go (prometheus/client_golang)
```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    requestCount = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total HTTP requests",
    })
    
    requestDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
        Name: "http_request_duration_seconds",
        Help: "HTTP request duration",
    })
)

func init() {
    prometheus.MustRegister(requestCount)
    prometheus.MustRegister(requestDuration)
}

// 메트릭 엔드포인트
http.Handle("/metrics", promhttp.Handler())
```

---

### Cloud Debugger 통합

#### AWS X-Ray
```python
from aws_xray_sdk.core import xray_recorder
from aws_xray_sdk.ext.flask.middleware import XRayMiddleware

app = Flask(__name__)
XRayMiddleware(app, xray_recorder)

# 커스텀 서브세그먼트
@xray_recorder.capture('my_function')
def my_function():
    # 작업 수행
    pass
```

#### GCP Cloud Debugger
```python
try:
    import googleclouddebugger
    googleclouddebugger.enable()
except ImportError:
    pass
```

---

## 성능 프로파일링

### CPU 프로파일링

#### Python (cProfile, py-spy)
```bash
# cProfile
python -m cProfile -o output.prof script.py
python -m pstats output.prof

# py-spy (프로덕션 안전)
py-spy record -o profile.svg -- python script.py
py-spy top --pid <pid>
```

#### Go (pprof)
```go
import _ "net/http/pprof"

// HTTP 서버에 pprof 엔드포인트 자동 추가
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

```bash
# CPU 프로파일 수집
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 프로파일 분석
go tool pprof -http=:8080 cpu.prof
```

#### Rust (flamegraph)
```toml
[dependencies]
pprof = { version = "0.13", features = ["flamegraph"] }
```

```bash
# Flamegraph 생성
cargo flamegraph
```

#### Java (JFR)
```bash
# JFR 활성화
java -XX:+FlightRecorder -XX:StartFlightRecording=duration=60s,filename=recording.jfr MyApp

# JFR 파일 분석 (JDK Mission Control)
jmc recording.jfr
```

---

### 메모리 프로파일링

#### Python (memory_profiler, tracemalloc)
```python
from memory_profiler import profile

@profile
def my_function():
    # 작업 수행
    pass
```

```bash
python -m memory_profiler script.py
```

**tracemalloc (내장)**:
```python
import tracemalloc

tracemalloc.start()
# 작업 수행
snapshot = tracemalloc.take_snapshot()
top_stats = snapshot.statistics('lineno')

for stat in top_stats[:10]:
    print(stat)
```

#### Go (pprof heap)
```bash
# 힙 프로파일 수집
go tool pprof http://localhost:6060/debug/pprof/heap

# 메모리 할당 추적
go tool pprof http://localhost:6060/debug/pprof/allocs
```

#### C/C++ (Valgrind massif)
```bash
# Massif 실행
valgrind --tool=massif ./myapp

# Massif 시각화
ms_print massif.out.<pid>
```

#### Rust (heaptrack)
```bash
# heaptrack 실행
heaptrack ./target/release/myapp

# 결과 분석
heaptrack_gui heaptrack.myapp.<pid>.gz
```

---

## Advanced 디버깅 기법

### 조건부 브레이크포인트

**Python (pdb)**:
```python
# 조건부 브레이크포인트
(Pdb) break script.py:42, x > 100
```

**GDB**:
```bash
(gdb) break main.c:42 if x > 100
```

**LLDB**:
```bash
(lldb) breakpoint set --file main.c --line 42 --condition 'x > 100'
```

---

### 역방향 디버깅 (Time Travel Debugging)

**GDB 역방향 실행**:
```bash
(gdb) target record-full
(gdb) continue
# 오류 발생 후
(gdb) reverse-continue  # 역방향 실행
(gdb) reverse-step
```

**WinDbg Time Travel Debugging** (Windows):
```bash
# TTD 트레이스 수집
ttd.exe -out trace.run myapp.exe

# 트레이스 디버깅
windbg -z trace.run
```

---

### 동적 계측 (Dynamic Instrumentation)

**DTrace (macOS/FreeBSD)**:
```bash
# 함수 호출 추적
sudo dtrace -n 'pid$target::my_function:entry { printf("Called with arg=%d", arg0); }' -p <pid>
```

**SystemTap (Linux)**:
```bash
# 함수 호출 추적
stap -e 'probe process("/path/to/binary").function("my_function") { println("Called") }'
```

**BPF/eBPF (Linux)**:
```bash
# bpftrace 사용
bpftrace -e 'uprobe:/path/to/binary:my_function { printf("Called\n"); }'
```

---

## Best Practices 요약

### 1. 디버거 선택
- 언어에 적합한 디버거 사용 (Python → debugpy, Go → Delve)
- 프로덕션 환경에서는 안전한 프로파일러 사용 (py-spy, async-profiler)

### 2. 로깅 전략
- 구조화된 로깅 사용 (JSON 형식)
- 로그 레벨 적절히 설정 (DEBUG, INFO, WARNING, ERROR)
- 분산 시스템에서는 Correlation ID 추가

### 3. 성능 고려
- 디버그 심볼 포함 (-g 플래그)
- 소스맵 생성 (TypeScript, JavaScript)
- 프로파일링 전 워밍업 수행

### 4. 보안
- 프로덕션 디버그 포트는 방화벽으로 보호
- 민감 정보 로그에 기록 금지
- 디버그 모드는 환경 변수로 제어

### 5. 자동화
- CI/CD 파이프라인에 디버그 빌드 추가
- 자동 스택 트레이스 수집 (Sentry, Rollbar)
- 메트릭 자동 수집 (Prometheus, Datadog)

---

**End of Reference** | moai-essentials-debug v2.1.0
