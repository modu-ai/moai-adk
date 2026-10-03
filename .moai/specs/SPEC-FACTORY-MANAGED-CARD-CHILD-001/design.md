---
id: SPEC-FACTORY-MANAGED-CARD-CHILD-001
title: "design.md — 설계 결정 D-1..D-6"
version: "0.1.0"
created: 2026-10-03
updated: 2026-10-03
author: GOOS (manager-spec)
tier: M
---

# design.md — 설계 결정

> 상태축 없음. 줄 번호는 기준 트리 `2b9e4a4d0` 를 **읽어서** 얻었다(실행 관측이 아니다). 실행으로 확인해야 하는 전제는 plan.md §C 의 측정 항목 P-1..P-4 로 넘겼다. 구현 이름은 run 단계가 더 나은 형태를 찾을 수 있도록 "기대 접점"으로만 적는다.

## D-1. 분기는 카드마다, 같은 술어로, 루프 안에서

**결정.** `runCodexFactoryLane` 이 카드 자식을 띄우는 한 곳(현행 `launchCodexCardSession` 호출, `codex_launcher.go:1000`)에서 카드마다 분기한다. 술어는 플레인 경로가 쓰는 두 함수 그대로다: `factoryManagedRequested(os.Environ()) && factoryLaunchEnabled(os.Environ())`(`codex_launcher.go:1182`, 정의 `factory_launch_pending.go:29-41`). 참이면 관리 소유자 이음새 `managedFactoryCodexLaunchFunc`(`:252-259`)를, 거짓이면 현행 `codexDirectLaunchFn` 경로(`:1038`)를 부른다.

**왜 이 술어인가.** 루프 프로세스의 환경에는 레인 스탬프가 이미 들어 있다: `enterFactoryLaneRun` 이 run id 를 싣고(루프가 `os.Getenv(config.EnvMoaiKanbanID)` 로 읽어 쓴다, `:948`), `enterFactoryLaneMode` 가 레인 라벨·역할 표식·레인 수를 `os.Setenv` 한다(`factory.go:797-799`). 그래서 스위치만 있으면 술어가 참이 된다. 술어를 새로 만들면 두 번째 해석이 생긴다 — 부모 SPEC 이 막으려 한 일이다(부모 D-7).

**카드마다 평가하는 이유.** 환경은 루프 중에 바뀌지 않지만, 분기를 한 곳에 두면 "카드마다 정확히 한 번"(REQ-CC-001)이 시험으로 직접 세어진다.

**기대 접점.** `launchCodexCardSession` 안(디렉터리·인수·환경을 이미 조립하는 곳)에서 갈라지는 쪽이 변경 면적이 작다. 직접 exec 문 쪽 줄은 한 글자도 바꾸지 않는다(REQ-CC-002).

**기각한 대안.**

- *루프 시작 때 한 번 갈라 두 개의 루프 본문을 둔다.* 루프의 lease·worktree·정지 논리가 둘로 복제되어 REQ-CC-003 을 시험으로 지키기 어렵다.
- *`runCodexLaunch` 의 분기를 재사용하도록 루프가 `runCodexLaunch` 를 부른다.* `runCodexLaunch` 는 초기화 제안 게이트·프로젝트 루트 해석·`-w` 해석을 앞에 두고(`:1101-1171`), 카드 자식 경로는 그 단계들을 건너뛴다. 합치면 카드 자식 경로의 현행 동작이 바뀐다.

## D-2. 프로세스 소유: 관리 소유자는 런처를 부모로 남긴다 — 그리고 직접 exec 문은 그렇지 않다

**읽은 사실.** 관리 소유자는 App Server 를 자식 프로세스로 띄우고(`managed_codex_factory.go:686-690`) 소유자 진입이 `session.Close()` 를 defer 로 건다(`:819`). 런처 PID 가 브로커 endpoint 소유자다(`:793-794`, `:806`). 한편 POSIX 의 직접 exec 문 기본 구현은 `syscall.Exec` 로 **프로세스를 교체**한다(`codex_direct_posix.go:23-57`, 교체 호출 `:53`). `launchCodexCardSession` 은 이 이음새를 부른다(`codex_launcher.go:1038`).

**함의 (읽기 기반 가설 H-1, plan.md P-1 에서 측정).** 직접 exec 문이 그대로 `syscall.Exec` 이면 POSIX 에서 레인 루프는 첫 카드 자식이 뜨는 순간 런처 프로세스가 교체되어 "다음 카드로 이어진다"가 시험 이음새(스텁)에서만 성립한다. 루프 주석은 반대로 적고 있다 — "no process replacement happens on this path … the launcher stays the parent across every card"(`:911-913`). 이 SPEC 은 H-1 의 참거짓에 기대지 않는다: REQ-CC-002 는 직접 exec 문에 넘기는 `*exec.Cmd` 의 동일성만 고정하고, REQ-CC-009 는 **관리 소유자가 반환한 뒤** 루프가 이어짐을 고정한다. 관리 경로에서는 런처가 부모로 남으므로 이어짐이 실제로 성립한다. H-1 이 참이면 plan.md 의 열린 질문 Q4 로 리더에게 올린다(이 SPEC 의 범위 밖 결함 후보).

## D-3. 앵커 락과 레인 claim 스탬프는 하지 않는다

**읽은 사실.**

- 앵커 락 `codexWorktreeAnchorLock` 의 호출은 `runCodexLaunch` 안 세 곳뿐이다: 관리 분기(`codex_launcher.go:1186`), spawn 문(`:1232`), 직접 문(`:1254`). 카드 자식 경로 `launchCodexCardSession`(`:1015-1039`)에는 없다.
- `stampCodexLaneClaim`(`codex_factory.go:241-265`)은 spawn 문(`codex_launcher.go:286`)과 Windows 직접 문(`codex_direct_windows.go:61`)에서만 불린다. 그 스탬프는 레인 claim 의 PID 를 *자식* PID 로 옮긴다("A spawned launcher exits immediately; its lane claim must follow the Codex process", `codex_factory.go:239-240`).
- 레인 claim 은 루프가 `resolveFactoryLaneName`(`factory.go:849-872`)으로 이미 `os.Getpid()` 에 잡아 둔다. 관리 경로에서 런처는 끝까지 부모이고 그 PID 가 브로커 endpoint 소유자이므로 옮길 이유가 없다.

**결정.** 둘 다 하지 않는다 — 현행 카드 자식 경로와 같다. 관리 분기를 `runCodexLaunch` 안에서 쓰는 `-w` 앵커 락은 카드 자식에 대응물이 없다(카드 워크트리 점유는 lease 와 카드 레코드가 운반한다).

**되돌림 조건.** 카드 워크트리에 두 번째 쓰기 세션이 들어오는 사고가 관측되면 락을 더하는 별도 카드로 다룬다. 이 SPEC 에서 더하면 "현행과 같다"는 REQ-CC-006 이 깨진다.

## D-4. 카드 자식 인수·환경·디렉터리의 조립

**디렉터리.** 소유자에게 카드 워크트리 `wt` 를 `dir` 로 넘긴다. 소유자는 이를 `session.dir` 에 두고(`:805`) App Server 프로세스 `cmd.Dir`(`:688`)에 쓴다. 스레드 cwd 도 같은 디렉터리라는 것은 기존 시험이 이미 단언한다(`TestManagedCodexOwnerUsesLaunchDir`, `managed_codex_factory_test.go:349-403`, 로그 `server-cwd`·`thread-cwd`).

**인수.** 소유자의 인수 해석은 `-c`/`--config` 와 `-m`/`--model` 만 받고 나머지는 `managed codex launch does not support <arg>` 로 거부한다(`:573-600`; 시험 `TestManagedCodexOptions` 가 `--profile` 거부를 이미 단언). 직접 exec 문의 현행 인수는 `-C <wt>` 와 로컬 지침 쌍 `-c developer_instructions=<json>`(`codex_launcher.go:1020`, 생산자 `:132-135`, 쌍을 만드는 줄 `:168`)이다. 관리 경로에서는 **`-C` 를 빼고** 로컬 지침 쌍만 `[bin, "-c", "developer_instructions=…"]` 형태로 넘긴다(소유자 규약: argv[0] 은 프로그램 이름, `:575` 가 1부터 훑는다). 크기 검사(`checkCodexInstructionSize`, `:1022-1026`)는 그대로 둔다 — 초과 시 소유자를 부르지 않고 오류가 현행 처리(stderr 한 줄, 루프 계속)로 간다.

**관측 한계.** 이 쌍이 App Server 의 *스레드* 개발자 지침으로 적용되는지는 읽기로도 실행으로도 확인하지 않았다(소유자는 이 값을 App Server 명령행 `-c` 로 붙인다: `managedCodexAppServerArgs` `:619-623`). 소유자가 인수를 받아 명령행에 싣는다는 사실까지만 이 SPEC 이 고정한다(AC-CC-004). 적용 여부는 문서의 "알려진 한계"에 미관측으로 적는다.

**환경.** 현행 카드 자식 환경 `codexCardLaunchEnv(label, cardID)`(`:1052-1061`; 11키 청소 `codexChildEnv` `:600-618` 위에 역할 표식·라벨 두 운반체·백엔드 `gpt`·카드 id)에 **run id 한 키(`MOAI_KANBAN_ID`)** 를 더한다. 이유: 소유자가 run id 를 환경에서 읽고 없으면 `factory managed session requires a factory run id` 로 거부하며(`:789-791`) 브로커 등록도 같은 키와 `MOAI_FACTORY_WORKER` 를 쓴다(`factory_launch_pending.go:46-52`). `MOAI_FACTORY_WORKERS` 는 청소가 지우고 현행처럼 더하지 않는다(루프 중 값 `0` 은 `factory.go:799`; 주석 `:1049-1050`). 값은 루프가 `:948` 에서 읽은 `runID`.

**기록되는 차이 (현행 카드 환경과 다른 두 곳).** (a) 현행 카드 환경은 run id 를 **싣지 않는다**(`TestSD_AC003_CodexRelaunchPerCard` 가 `MOAI_KANBAN_ID == ""` 를 단언, `factory_m5_test.go:154`) — 관리 경로의 자식은 싣는다. (b) 소유자가 App Server 환경에 `MOAI_SESSION_PID=<런처 PID>` 를 더한다(`:798`, `launch_session_pid.go:32-45`, 청소가 지우는 키 목록 `codex_launcher.go:601` 의 `drop` 맵) — 현행 청소는 그 키를 지운다. 둘 다 소유자 설계의 귀결이며(플레인 관리 경로는 `os.Environ()` 전체를 넘긴다, `codex_launcher.go:1191`) 이 SPEC 이 소유자를 바꾸지 않으므로 문서에 차이로 적는다. **대안**: 소유자 환경과 자식 환경을 분리해 run id 를 소유자에게만 주기 — 소유자 진입 시그니처 변경이라 REQ 제약(소유자 불변)에 어긋나 기각.

**`-d` 디버그.** 관리 경로는 현행 플레인 관리 분기와 같게 `RUST_LOG` 주입을 건너뛴다(`codexApplyDebugEnv` 를 호출하지 않는다). 문서의 기존 문장("관리되는 Codex 런치는 … `RUST_LOG` 주입을 건너뛴다")이 이미 이를 적고 있다. 루프의 사전 디버그 덤프(`:963-965`)는 첫 세션 전에 이미 나가므로 영향 없다.

## D-5. 관리 세션의 끝과 루프 연속

**읽은 사실.** 드라이버는 운영자 줄이 `/exit` 또는 `/quit` 일 때 `nil` 로 반환하고(`managed_factory_session.go:320-322`), 치명 오류에서 오류로 반환하고(`:357-367`), **stdin EOF 에서는 반환하지 않는다** — `absorb` 가 EOF 를 만나면 `inputs = nil` 로 두고 `false` 를 돌려줄 뿐이다(`:315-319`). 소유자 진입은 반환 직전에 `session.Close()` 로 App Server 를 내린다(`:819`).

**결정.** 드라이버의 끝 조건을 바꾸지 않는다(범위 밖). 루프는 소유자가 **반환했을 때** 현행 처리로 이어진다: 오류면 `codex lane: card <id> session: <오류>` 한 줄을 stderr 에 쓰고(`:1004` 와 같은 형태) 다음 `factoryNextLeaseOnce` 로 간다.

**귀결(문서에 적는다).** (i) stdin 이 EOF 이거나 입력이 없는 비대화형 레인에서는 관리 카드 세션이 `/exit`·`/quit`·치명 오류 없이는 끝나지 않는다 — 열린 질문 Q2. (ii) 세션이 끝나도 카드는 lease 만료까지 leased 로 남는다(현행 주석 `:1001-1003` 과 같다).

## D-6. 운영자 입력: 프로세스 수명 단일 펌프와 세션별 닫을 수 있는 어댑터

**문제.** 소유자는 세션마다 `readManagedOperatorInput(in, lines)` 고루틴을 띄우고(`managed_factory_session.go:314`) 그 고루틴은 `bufio.Scanner.Scan()` 에서 `in` 을 읽으며 막힌다(`:290-296`). 세션이 `/exit` 로 끝나도 그 고루틴은 살아 있고, `in` 이 프로세스의 `os.Stdin` 이면 **다음 운영자 줄을 먼저 읽어 끝난 세션의 채널로 보낸다** — 레인 루프에서 다음 카드 세션은 그 줄을 영영 못 받는다. 플레인 경로는 세션이 하나뿐이고 프로세스가 곧 끝나 이 문제가 없었다.

**결정.** 프로세스 수명의 **펌프 하나**가 `os.Stdin` 을 읽고, 세션마다 **닫을 수 있는 `io.Reader` 어댑터**를 소유자에게 stdin 으로 넘긴다. 어댑터는 (a) 펌프가 읽은 조각을 받아 `Read` 로 돌려주고, (b) 세션이 끝나면 `Close` 되어 막힌 `Read` 를 즉시 EOF 로 풀며, (c) 어댑터가 닫힌 뒤 펌프가 읽은 조각은 **다음 어댑터가 붙을 때까지 펌프가 쥐고 있다가 그 어댑터에 준다**(읽어 둔 조각은 한 번에 한 조각, 닫힌 어댑터에는 절대 가지 않는다). 펌프는 첫 어댑터가 만들어질 때 처음 시작한다 — 스위치가 꺼진 경로는 `os.Stdin` 을 건드리지 않는다.

**기대 접점.** 소유자 진입 이음새의 기본 구현 `managedFactoryCodexLaunchFunc`(`codex_launcher.go:257-259`)가 `runManagedFactoryCodex(..., os.Stdin)` 대신 `runManagedFactoryCodex(..., sharedOperatorInput.attach())` 를 부르고 반환 때 어댑터를 닫는다. 이렇게 두면 소유자 시그니처(`runManagedFactoryCodex(..., stdin io.Reader)` 는 이미 stdin 을 인수로 받는다, `:787`)와 드라이버를 바꾸지 않고, 플레인 경로도 같은 이음새라 한 세션짜리 동작이 같다. 새 코드는 새 파일 하나에 둔다(`managed_operator_input.go`) — t1408·t1459 가 건드릴 `managed_codex_factory.go` 와의 텍스트 충돌을 피한다.

**기각한 대안.**

- *각 세션에 `os.Stdin` 을 그대로 넘긴다(현행 방식).* 위의 입력 도둑질.
- *세션마다 `os.Pipe` 를 만들고 부모가 복사한다.* 복사 고루틴이 같은 문제를 한 겹 위로 옮긴다 — 끝난 세션의 파이프로 줄을 쓰고 만다.
- *드라이버가 `ctx` 를 받아 읽기를 취소한다.* 드라이버 시그니처 변경(범위 밖). 표준 라이브러리의 `os.Stdin` 읽기는 취소할 수 없으므로 어댑터가 필요하다.
- *EOF 를 세션 종료로 본다.* 드라이버 끝 조건 변경(범위 밖), 플레인 경로의 현행 동작(EOF 이후 계속)을 바꾼다.

**알려진 비대칭.** 펌프가 읽은 한 조각이 **끝난 세션이 아닌 다음 세션**에 전달되므로, 운영자가 `/exit` 를 두 번 치면 두 번째가 다음 카드 세션을 즉시 끝낼 수 있다. 이는 대화형 터미널 입력의 본질적 한계이며 문서에 적는다.

## 이 설계가 만들지 않는 것

새 환경 변수, 새 설정 키, 새 CLI 플래그, 새 브로커 스키마, 새 의존. 스위치는 부모가 정한 `MOAI_FACTORY_MANAGED` 하나다.
