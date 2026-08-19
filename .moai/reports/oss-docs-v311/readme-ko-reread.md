# README.ko.md 정본 전수 재독 — v3.1.1 델타 감사

- 대상: `README.ko.md` (755줄, t47 ko-canonical 스킬레톤) — en/ja/zh는 같은
  구조(755줄 동일)이므로 동일 갭이 4-로케일에 그대로 적용됨.
- 방법: 전수 정독 + 배치 기능 대조. 모든 지적에 줄번호(README.ko.md 기준).

## 1. 낡은 사실 (outdated) — 정정 필요

### 1-1. MCP 서버 도구 수·그룹 수 [최우선]

- **L275-289** (`### MCP 서버`): "**다섯 그룹으로 묶인 17개 MoAI 도구**" —
  배치에서 glm_task 위임 도구군 4개가 추가돼 실제는 **여섯 그룹 21개**.
  - L277 본문 "17개", L279-285 그룹 테이블에 "GLM 위임" 그룹 없음.
  - 소스 근거: `internal/cli/mcp_server.go` 21개 등록(검증 기록은
    commit-analysis.md 참조). 그룹 테이블에 다음 행 추가 필요:
    `| GLM 위임 | glm_task, glm_job_status, glm_job_result, glm_job_cancel | GLM(z.ai) 백그라운드 작업 |`
- en/ja/zh README 동일 문장("17 tools"/"五个组" 등) 동시 정정 필요.

### 1-2. statusline 예시·표 (L440-462)

- **L443-445** 예시 3줄이 옛 레이아웃. 실제(v3.1.1)는:
  - 첫째 줄: `🤖 모델 │ 🧠 effort │ ♻️ 캐시 │ 🔅 v │ 🗿 v │ ⏳ │ 💬` (유사하나
    cc버전 표기 `🔅 cc v…` 형태 가능)
  - 셋째 줄: `📁 … │ 📡 owner/name | 🅱️ branch … │ 📫 +0 M6 ?0 │ 📋 … │ 💌 PR …`
    — 리포 아이콘 `📡`(기존 🔀), repo|branch 조인이 ASCII `|`, git 상태가
    💾→**📫 사서함 계열**(📬 staged / 📫 modified / 📪 untracked / 📭 clean).
  - **넷째 줄(조건부) 세션 라인**: `🏷️ <세션명> │ 👤 <에이전트> │ 🔄 TODO: <picked> / <queued> │ 🔀 <이슈> / <PR>` — 현재 예시에 아예 없음.
- **L448-460** 요소 표: `🔀 리포` 행 → `📡 리포`로, `💾 git 상태` → `📫`로,
  세션 라인 4종(🏷️·👤·🔄·🔀) 행 추가 필요.

### 1-3. 웹콘솔 설정 탭 수 (L360, L670)

- L360 "10개 탭" / L670 "10-탭 설정" → 실제 **11개** (Cross-Session 탭 추가,
  `internal/web/schemaform.go consoleTabs()` 근거). 나열에 `Cross-Session` 추가.

### 1-4. `moai todo` 동사 목록 (L310)

- "`moai todo <add|list|next|done>`" → 실제 표면은
  `<add|list|next|done|unpick>` + bare(=list) + 2단어 이상 자연어 add.
  문장 하나로 갱신.

### 1-5. Release 배지 (L24)

- `Release-v3.1.0` 배지 — v3.1.1 태그 시 릴리즈 프로세스가 갱신하는 항목.
  본 감사에서는 지적만(임의 변경 금지 — 버전 SSOT는 릴리즈 동기화 소관).

## 2. 누락 기능 (missing) — 추가 필요

### 2-1. 팩토리 모드 `-k <N>` [최우선 — v3.1.1 대표 기능]

- **L40-108** "v3.1 새 기능 — 칸반 모드" 절 전체가 3역할 칸반만 다룬다.
  같은 `-k` 토큰으로 들어가는 **팩토리 모드**(`moai cc -k <N>` 리드 +
  `moai cc -k <N> --name worker-<i>` 워커, 기본 8, A/B클래스 통째·C클래스
  직렬 3단계, `-f`는 은퇴)가 한 줄도 없음.
- 권장: "시작하기"(L54) 아래 또는 별도 H3 `### 팩토리 모드 — 번호 붙은 워커`로
  진입 명령 3줄 + 리드 루프 한 문단. 세부는 docs-site kanban-mode.md로 링크.

### 2-2. `moai graph` (L654-672 CLI 명령표)

- L656-670 CLI 명명표(13개)와 L672 "전체 36개 커맨드" — `moai graph`,
  `moai tokens`, `moai memory` 미포함. "36개" 숫자의 산출 근거도 흐려짐
  (v3.1.1 신규 최상위 커맨드 3개 추가).
- 권장: 표에 `moai graph <build|query>` 행 추가(레인 건전성 교차검사),
  `moai tokens record`·`moai memory`는 "전체 N개" 안내 문구로 처리 가능.
  정확 커맨드 수는 릴리즈 시점 `moai --help`로 확정 필요(Gaps G6).

### 2-3. `moai update` 데이터 보호 (선택)

- L661 "`moai update` | 최신 버전으로 업디트 (자동 롤백 지원)" — v3.1.1에서
  관리 뿌리 정리 전 미관리 파일을 `.moai-backups/…/pre-clean/`에 백업하는
  안전장치가 들어왔다. "삭제 전 백업" 여섯 글자 보탤 가치(사용자 신뢰 직결).

### 2-4. Windows install.bat (선택)

- L213-217 설치 — PowerShell만 안내. cmd.exe용 `install.bat`가 새로 출시됐으나
  README 수준에서는 "cmd.exe 미지원"(L259) 서술과 충돌하지 않게
  설치 스크립트 3종(install.sh/ps1/bat) 정도만 언급 가능. 우선순위 낮음.

## 3. 구조 관찰 (structure) — 변경 불필요, 참고

- 12개 H2 + 스위처 헤더(L11-16) + 4-로케일 755줄 패리티 모두 정상
  (t47 착지 상태 유지).
- L640 유틸리티 커맨드 섹션 나열에 `todo` 빠짐 — docs-site
  utility-commands 섹션엔 moai-todo 페이지가 있으니 나열에 추가하는 게
  일치(경미).
- L633 "12개 섹션" — docs-site 실제 12 섹션과 일치(유지).
- 세 핵심 표현(L122 "세 가지 핵심 (three axes)") — t59 용어 카드가
  다듬은 표현으로 보임, 유지.

## 4. 드래프트 반영 범위

`drafts/README.ko.md`에 1-1 ~ 1-4(버전 배지 제외) + 2-1 + 2-2(표 1행 +
문구) + L640 `todo` 추가를 반영한 전체 파일. 1-5(배지)와 2-3/2-4(선택 항목
일부)는 authoring-plan 참조.
