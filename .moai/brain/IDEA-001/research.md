# Research — IDEA-001: MoAI 솔로 개발자용 SPEC 워크플로우 가시화 대시보드

> **Phase 3 산출물** · 작성일: 2026-05-04 · 입력 idea: "moai web 대시보드 (Go templ+HTMX, localhost)"
> Phase 1 Discovery 합의:
> - 페르소나: MoAI-ADK 솔로 개발자
> - 핵심 문제: SPEC 워크플로우(plan/run/sync) 진행 가시화
> - 성공 지표: 터미널/CLI 왕복 횟수 30%+ 감소
> - MVP: 단일 페이지 read-only · localhost · Go templ + HTMX

---

## Executive Summary

MoAI-ADK 솔로 개발자가 SPEC 워크플로우(plan → run → sync) 도중 빈번하게 수행하는 5종 CLI 왕복(`gh pr list`, `gh pr checks`, `git worktree list`, `git status`, `ls .moai/specs/`)을 단일 read-only 웹 페이지로 통합한다. Go `templ` 컴포넌트 + HTMX `hx-trigger="every Ns"` polling 패턴이 이 read-only 패러다임에 정확히 맞물린다. 외부 비교 대상으로 `gh-dash`(터미널)·`Headlamp`(웹·k8s 한정)·`gh-pr-dashboard`(GitHub 한정)가 있으나, **MoAI-ADK SPEC 모델·worktree·MX tag·TRUST 5라는 도메인 객체를 1차 시민으로 다루는 도구는 부재**하다. 즉, 본 idea의 niche는 명확하다. 컨텍스트 전환 비용 연구(Mark, RescueTime 등)는 도구 통합으로 최대 30% 생산성 회복이 가능함을 시사하며, 이는 사용자가 선택한 성공 지표(왕복 30%+ 감소)와 동일 수치 영역에 안착한다.

---

## 1. Existing Solutions Landscape

### 1.1 Terminal-First Dashboards (직접 경쟁군)

| 도구 | 도메인 | 본 idea와의 차별 |
|------|--------|------------------|
| `gh-dash` | GitHub PR/Issue 통합 TUI | 일반 GitHub 메타. SPEC ID·MoAI worktree·Phase 모델 인식 불가 |
| `lazygit` | git 인터랙티브 TUI | git 단일 도메인. SPEC/CI/PR 통합 뷰 부재 |
| `gitui` | lazygit 대안 (Rust) | 동일 — git only |
| `k9s` | Kubernetes TUI | 도메인 불일치 |
| `lazydocker` / `oxker` | Docker TUI | 도메인 불일치 |
| `btop` | 시스템 리소스 | 도메인 불일치 |

→ TUI 생태계는 성숙하나 **MoAI-ADK 도메인에 매핑되는 도구는 0건**.

### 1.2 Web-Based Developer Dashboards (간접 경쟁군)

| 도구 | 위치 | 본 idea와의 차별 |
|------|------|------------------|
| `Headlamp` (CNCF) | k8s 웹 UI | 도메인 불일치 (k8s) |
| `FreeLens` / `JET Pilot` | k8s 데스크톱 | 도메인 불일치 |
| `gh-pr-dashboard` | GitHub PR 통합 (CLI 확장) | GitHub만 — SPEC·MX·TRUST 부재 |
| `Graphite` | PR stack + dashboard | SaaS · GitHub PR 중심. 로컬 MoAI 워크플로우 미지원 |

### 1.3 Niche Validation

- 동일 도메인(SPEC-First DDD/TDD + worktree + MX tag + TRUST 5) read-only 가시화 도구는 **검색 결과상 존재하지 않음**.
- 가장 가까운 유사도: `Graphite`(PR pipeline 가시성) + `gh-dash`(터미널 GitHub 뷰) 조합 — 그러나 MoAI 도메인 객체 인식 불가.
- 결론: **빈 자리(green field) niche**. Solo developer + localhost + read-only 조합은 SaaS 경쟁자가 진입할 매력이 낮음(시장 작음, vendor lock-in 가치 없음).

---

## 2. Technical Stack Validation

### 2.1 Go templ + HTMX 생태계 성숙도 (2026 기준)

확인된 활성 레퍼런스:

| 레퍼런스 | 적합도 | 주요 가치 |
|----------|--------|----------|
| `Piszmog/go-htmx-template` | ★★★★★ | Go 1.25+ native CSRF · sqlc · Tailwind via go tools · `air` hot reload · go-tw |
| `josephspurrier/gohtmxapp` | ★★★★ | `localhost:8080/dashboard` 직진 — `make watch` 한 줄 |
| `webdevfuel/go-htmx-data-dashboard` | ★★★ | data-heavy dashboard 패턴 (단, Meili+Postgres 의존성은 본 idea엔 과잉) |
| `alekLukanen/go-templ-htmx-example-app` | ★★★ | 인증 흐름 포함 — 본 idea(localhost-only)엔 불필요 |
| Alberto Jaen Medium 가이드 | ★★★★ | templ proxy + Air 동시 운용 패턴 — DX 측면 유의미 |

→ Go templ + HTMX 조합은 2026년 시점 **mature dev tooling 갖춘 일급 스택**. 복잡한 인증/DB 미사용 시 학습 비용 ≈ 1일 수준.

### 2.2 핵심 패턴: templ Fragments + HTMX Polling

**templ Fragments (partial rendering)** — read-only dashboard에 정확히 부합:

```templ
@templ.Fragment("worktree-list") {
  <div id="worktree-list" hx-get="/api/worktrees" hx-trigger="every 5s">
    // worktree rows
  </div>
}
```

서버 측 `templ.Handler(Page(state), templ.WithFragments("worktree-list"))` 호출로 **부분 갱신만** 응답 가능 → 전체 페이지 재렌더 불필요, 대역폭·CPU 모두 효율적.

**HTMX 폴링** — `hx-trigger="every Ns"` 한 줄로 자동 갱신:

```html
<div hx-get="/api/spec-status" hx-trigger="every 5s" hx-swap="innerHTML">
  Loading...
</div>
```

서버가 폴링 중지를 원하면 HTTP 286 응답으로 stop signal 전송 가능. 본 idea는 단일 페이지 read-only이므로 SSE/WebSocket 도입 불필요 — **polling만으로 충분**.

### 2.3 권장 스타터 구성 (Phase 4 Lean Canvas 입력용)

- 베이스: `Piszmog/go-htmx-template` (Go 1.25+ native CSRF, air hot reload)
- 컴포넌트 라이브러리: `templUI`(/templui/templui) — Tailwind 기반 미리 제작된 UI 부품 (선택, 학습 가속 시)
- 폴링 주기: 5초 (worktree·git status), 10초 (PR·CI 상태) — 경계는 Phase 6 SPEC 분해 시 결정
- 데이터 소스: 모두 read-only · 로컬 명령(`git`, `gh`, `find`, filesystem). 외부 DB 불필요

---

## 3. Productivity Research (왕복 30%+ 감소 가설 검증)

### 3.1 컨텍스트 전환 비용

- Dr. Gloria Mark (UC Irvine): **인터럽션 후 원래 작업 복귀까지 평균 23분 15초**.
- RescueTime 분석: 8시간+ 근무에도 실제 집중 작업은 **2시간 48분**. 나머지는 컨텍스트 전환·복구.
- 인터럽션된 작업은 **소요 시간 2배 + 오류 2배**.

### 3.2 도구 통합으로 회복 가능한 생산성

- "도구 전환 마찰을 줄이면 생산성의 최대 **30%** 회복 가능" — 본 idea 사용자가 선택한 성공 지표(왕복 30%+ 감소)와 정확히 일치.
- Asana 사례: Graphite 스택형 PR 도입 후 **엔지니어당 주 7시간 회복 + 21% 코드 출하 증가**.

### 3.3 본 idea 적용 가능성

MoAI-ADK 솔로 개발자가 1시간 작업 중 평균 수행하는 추정 CLI 호출 (베이스라인 가설):

| 명령 | 빈도/시간 (추정) | 목적 |
|------|------------------|------|
| `git status` | 5-10회 | 변경 확인 |
| `gh pr list` / `gh pr view N` | 3-5회 | PR 진척 |
| `gh pr checks N` | 3-5회 | CI 상태 |
| `git worktree list` | 2-4회 | 활성 worktree |
| `ls .moai/specs/` | 2-4회 | SPEC 목록 |
| **합계** | **15-28회** | — |

대시보드 도입 시 위 5종 모두 단일 페이지에서 자동 폴링으로 surface → CLI 명령 호출 자체가 사라짐. **이론적 감소율 70-90%**, 실측 30%+ 목표는 보수적 추정으로서 합리적.

### 3.4 측정 방법 (성공 지표 운용)

- 단순 측정: shell history 비교 (대시보드 도입 전/후 1주씩 grep `gh|git worktree|moai status`)
- 정확 측정: zsh/bash hook으로 명령 카운트 자동 기록 (옵션, 본 idea MVP 외 별도 SPEC 후보)

---

## 4. Competitive Positioning

```
                            웹 (browser)
                                 │
            Headlamp ──────┐     │     ┌────── (없음)
                           │     │     │
          k8s 도메인 ──────┼─────┼─────┼────── MoAI 도메인
                           │     │     │
                gh-dash ───┘     │     └─── ★ IDEA-001 (본 제안)
                                 │
                          터미널 (TUI)
```

본 idea의 좌표: **(MoAI 도메인) × (웹·localhost)** — 경쟁자 부재.

`gh-dash`/`lazygit` 사용자가 본 도구로 100% 이주하지 않더라도, MoAI-ADK 도메인 객체(SPEC ID, Phase, MX tag, TRUST 5)에 한정해서는 본 도구가 유일한 가시화 채널이 된다.

---

## 5. Risks & Open Questions

### Risk R1 — read-only 한계의 사용자 만족도
| | |
|---|---|
| 가설 | 사용자는 "보기만 가능 + 명령 복사 가능" UI에 만족할 것 |
| 위험 | 사용자가 1주일 사용 후 "버튼 클릭으로 `moai sync` 실행"을 강하게 요구 가능 |
| 완화 | MVP에 "Quick Action Launcher" (명령 텍스트 + 복사 버튼)만 포함. write 액션은 별도 SPEC으로 격리 |

### Risk R2 — 폴링 주기 vs 시스템 부하
| | |
|---|---|
| 가설 | 5초 polling으로 충분 (localhost 단일 사용자) |
| 위험 | `gh pr checks` rate limit (GitHub API hourly cap) — 다수 PR + 짧은 polling 시 위협 |
| 완화 | 폴링 주기 분리 (worktree·git: 5s / PR·CI: 30s) + 캐시 + 사용자 설정 가능 |

### Risk R3 — 학습 비용 (Go templ 신규 도입)
| | |
|---|---|
| 가설 | moai-adk-go 코드베이스가 이미 Go이므로 templ 학습 비용 1일 |
| 위험 | templ proxy + Air 이중 운용 DX 복잡도 (Alberto Jaen 가이드 참조) |
| 완화 | `Piszmog/go-htmx-template` 한 줄 `make run` 으로 캡슐화 — 진입 장벽 낮음 |

### Open Question OQ1
브랜드 통합 수준 — `.moai/project/brand/visual-identity.md`의 컬러 팔레트·타이포를 dashboard에도 반영할지? Phase 7 핸드오프 패키지에서 결정.

### Open Question OQ2
CG/GLM 모드(tmux 환경) 사용자도 동일 dashboard에 접근 가능해야 하는지? 또는 solo Claude Code 사용자만? — Phase 6 SPEC 분해 시 보조 SPEC 후보.

---

## Sources

### Tutorials & Templates (Go templ + HTMX)

- [go-templ-htmx-example-app (alekLukanen)](https://github.com/alekLukanen/go-templ-htmx-example-app)
- [Tailbits — Setting up HTMX and Templ for Go](https://tailbits.com/blog/setting-up-htmx-and-templ-for-go)
- [gohtmxapp (josephspurrier)](https://github.com/josephspurrier/gohtmxapp)
- [DEV.to — How To Build a Web Application with HTMX and Go (Calvin McLean)](https://dev.to/calvinmclean/how-to-build-a-web-application-with-htmx-and-go-3183)
- [Go-Blueprint Docs — HTMX and Templ](https://docs.go-blueprint.dev/advanced-flag/htmx-templ/)
- [go-htmx-data-dashboard (webdevfuel)](https://github.com/webdevfuel/go-htmx-data-dashboard)
- [go-htmx-template (Piszmog)](https://github.com/Piszmog/go-htmx-template)
- [Making A Dashboard With HTMX & Go (Stacy L, Medium)](https://medium.com/@hhartleyjs/making-a-dashboard-with-htmx-go-50820d7ddedb)
- [Go + HTMX + Templ + Tailwind: Complete Project Setup & Hot Reloading (Alberto Jaen, Medium)](https://medium.com/ostinato-rigore/go-htmx-templ-tailwind-complete-project-setup-hot-reloading-2ca1ba6c28be)
- [How to build a fullstack application with Go, Templ, and HTMX (FullstackWriter)](https://fullstackwriter.dev/post/how-to-build-a-fullstack-application-with-go-templ-and-htmx?category=Golang)

### Terminal Dashboard Comparison (gh-dash, lazygit, k9s)

- [gh-dash 공식 문서](https://www.gh-dash.dev/)
- [gh-dash on Terminal Trove](https://terminaltrove.com/gh-dash/)
- [lazygit GitHub repo](https://github.com/jesseduffield/lazygit)
- [Best Terminal Tools for Developers in 2026 (DEV)](https://dev.to/raxxostudios/best-terminal-tools-for-developers-in-2026-4jn1)
- [The Best TUI Apps for Linux Developers 2026](https://www.thetechbasket.com/best-tui-apps/)
- [Essential CLI/TUI Tools for Developers (freeCodeCamp)](https://www.freecodecamp.org/news/essential-cli-tui-tools-for-developers/)

### Web Dashboard Alternatives (Kubernetes context)

- [Kubernetes Dashboard Alternatives in 2026 (DEV — Alexandre Vazquez)](https://dev.to/alexandrev/kubernetes-dashboard-alternatives-in-2026-best-web-ui-options-after-official-retirement-4e02)
- [Kubernetes Dashboard Alternatives 2026 — Headlamp, FreeLens & Best Web UIs](https://alexandre-vazquez.com/kubernetes-dashboard-alternatives-2026/)
- [OpenLens Is Dead: 13 Kubernetes Dashboard Alternatives 2025](https://nadimtuhin.com/blog/kubernetes-dashboard-alternatives)

### Productivity & Context Switching Research

- [Mitigating Context Switching in Software Development (Jellyfish)](https://jellyfish.co/library/developer-productivity/context-switching/)
- [Reduce Context Switching | Developer Productivity (GitScrum Docs)](https://docs.gitscrum.com/en/best-practices/minimize-context-switching-developer-tools/)
- [Reducing context switching in development workflows (Graphite)](https://graphite.com/guides/reducing-context-switching-development-workflows)
- [gh-pr-dashboard: A Software Measurement Tool for Enhanced Developer Productivity (devactivity.com)](https://devactivity.com/insights/boost-developer-productivity-with-gh-pr-dashboard-your-unified-pr-command-center/)
- [The Hidden Cost of Context Switching for Developers (Crownest)](https://www.crownest.dev/blog/hidden-cost-context-switching-developers)
- [The Context-Switching Problem: Why I Built a Tracker That Lives in My Terminal (DEV)](https://dev.to/tejas1233/the-context-switching-problem-why-i-built-a-tracker-that-lives-in-my-terminal-4dpe)
- [Context Switching is Killing Your Productivity (Software.com DevOps Guides)](https://www.software.com/devops-guides/context-switching)
- [The True Cost of Context Switching in Developer Workflows (Axolo Blog)](https://axolo.co/blog/p/cost-context-switching-developer-workflow)
- [Context switching: How to reduce productivity killers (Atlassian)](https://www.atlassian.com/work-management/project-management/context-switching)
- [Context Switching: The Silent Killer of Developer Productivity (Hatica)](https://www.hatica.io/blog/context-switching-killing-developer-productivity/)

### Library Documentation (Context7)

- [a-h/templ — Fragments for HTMX partial updates](https://github.com/a-h/templ/blob/main/docs/docs/03-syntax-and-usage/19-fragments.md)
- [a-h/templ — Layout & forms](https://github.com/a-h/templ/blob/main/docs/docs/03-syntax-and-usage/11-forms.md)
- [bigskysoftware/htmx — hx-trigger every (polling)](https://github.com/bigskysoftware/htmx/blob/master/www/content/attributes/hx-trigger.md)
- [bigskysoftware/htmx — Hypermedia APIs vs Data APIs (table polling pattern)](https://github.com/bigskysoftware/htmx/blob/master/www/content/essays/hypermedia-apis-vs-data-apis.md)
- [bigskysoftware/htmx — Paris 2024 Olympics network automation (progress bar polling)](https://github.com/bigskysoftware/htmx/blob/master/www/content/essays/paris-2024-olympics-htmx-network-automation.md)
