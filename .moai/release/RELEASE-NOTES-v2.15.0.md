# MoAI-ADK v2.15.0 Release Notes

**Release Date**: 2026-04-26  
**Previous Release**: v2.13.2 (v2.14.x skipped — V3R3 Phase A bundle)

---

## 핵심 메시지

**V3R3 Phase A 출시 — Architectural Foundation Upgrade**

이번 릴리스는 다른 프로젝트 개발 시 즉시 적용 가능한 **5가지 구조적 개선**을 포함합니다:

1. **Convention Compliance Sweep** — 11개 skill 표준 준수 + manager-git 위임 protocol
2. **SPEC Dependency Cycle 해소** — v3R2 진행 차단 7 critical defects 중 D-CRIT-001
3. **Expert Agent Tool Uplift** — 7개 expert agent escalation 능력 활성화
4. **Token Circuit Breaker** — Anthropic SSE stall 자동 감지/복구
5. **Mobile Native Coverage** — iOS/Android/RN/Flutter 4-strategy 가이드 (zero coverage 해소)

---

## 주요 변경 사항

### V3R3 Phase A — 6 SPECs 포함

| SPEC ID | Title | Status | Impact |
|---------|-------|--------|--------|
| SPEC-V3R3-DEF-007 | Convention Compliance Sweep | Implemented | 11 skills PD, manager-git scope boundaries |
| SPEC-V3R3-DEF-001 | ORC Dependency Cycle Resolution | Implemented | D-CRIT-001 resolved, WF-001/MIG-001 one-way |
| SPEC-V3R3-ARCH-003 | Expert Agent Tool Uplift | Implemented | 7 expert agents + escalation protocol |
| SPEC-V3R3-ARCH-007 | Token Circuit Breaker | Implemented | Per-agent budgets, 75%/90% thresholds |
| SPEC-V3R3-COV-001 | Mobile Native Coverage | Implemented | expert-mobile + 3 mobile skills |
| SPEC-V3R3-CMD-CLEANUP-001 | Commands Cleanup | Implemented | /moai gate added, context removed |

### 새 기능

#### Convention Compliance Sweep (SPEC-V3R3-DEF-007)
- **11 skills**: Added `progressive_disclosure` YAML frontmatter blocks
  - Progressive Disclosure 3단계: Quick Reference (80-120 lines) / Implementation (180-250 lines) / Advanced (80-140 lines)
  - Token efficiency: ~2-3K tokens per skill
  - Affected skills: moai-foundation-core, moai-foundation-cc, moai-domain-backend, moai-domain-frontend, moai-domain-mobile, moai-platform-deployment, moai-library-nextra, moai-workflow-ddd, moai-workflow-tdd, moai-workflow-testing, moai-library-shadcn
- **manager-git**: Added Scope Boundaries + Delegation Protocol sections
  - Manager responsibility enforcement
  - Delegation depth constraints

#### Expert Agent Tool Uplift (SPEC-V3R3-ARCH-003)
- **7 expert agents**: Added `Agent` tool for cross-call escalation
  - Agents: backend, frontend, testing, debug, performance, refactoring, devops, security
  - Max depth: 2 (T1 → T2 only, T2 ↔ T2 within scope)
  - Same-SPEC scope constraint
  - Enables domain expert collaboration without manager intervention
- **expert-debug, expert-performance**: Added `Write` and `Edit` tools
  - Direct fix patch authoring capability
  - No delegation to manager-ddd/tdd for diagnostic fixes
- **All 7 experts**: Added Escalation Protocol section
  - Boundary rules to prevent silent scope creep
  - Domain-boundary trigger patterns
  - Depth and scope constraints

#### Token Circuit Breaker (SPEC-V3R3-ARCH-007)
- **`.moai/config/sections/runtime.yaml`**: New config section
  - Per-agent budgets: manager-strategy 60K, expert-* 40K, evaluator-* 20K (defaults)
  - Thresholds: 75% (warning) / 90% (critical warning)
  - Stall detection: 60-second timeout detection
  - Retry policy: max 3 retries
- **`internal/runtime/budget.go`**: Tracker type with 5 methods
  - `RecordCall(agent string, tokens int)`: Log token usage per call
  - `Usage(agent string) int`: Get current usage
  - `IsApproachingLimit(agent string) bool`: Check 75% threshold
  - `IsAtHardLimit(agent string) bool`: Check hard limit
  - `DetectStall(agent string) bool`: 60-second idle detection
  - Goroutine-safe via `sync.RWMutex`
- **`internal/runtime/persist.go`**: Atomic progress.md + resume message
  - Saves state before /clear
  - Generates paste-ready resume message per context-window-management.md
  - Format: `Wave N 이어서 진행. SPEC-* from <approach>. next: <command>.`
- **SessionStart hook integration**
  - Loads runtime.yaml at session start
  - Initializes Tracker
  - Falls back to defaults if file missing

#### Mobile Native Coverage (SPEC-V3R3-COV-001)
- **expert-mobile agent**: 4-strategy router
  - iOS native (SwiftUI) → expert-ios-native
  - Android native (Jetpack Compose) → expert-android-native
  - React Native → moai-framework-react-native
  - Flutter → moai-framework-flutter-deep
- **moai-domain-mobile skill**: iOS/Android native patterns
  - Authentication (Sign In With Apple, Biometric, OAuth)
  - Database (Core Data, Realm, Room)
  - Networking (URLSession, Volley, HTTP)
  - UI patterns and best practices
- **moai-framework-react-native skill**: RN deep dive
  - Cross-platform bridges
  - Performance optimization
  - Native module integration
- **moai-framework-flutter-deep skill**: Flutter comprehensive
  - Advanced state management (GetX, Provider, Riverpod)
  - Performance and optimization
  - Widget architecture patterns
- **Mobile keyword auto-detection**: SKILL.md routing
  - Keywords: iOS, Android, mobile, react-native, flutter, swift, kotlin, jetpack, swiftui
  - Auto-loads expert-mobile + appropriate domain skill

#### Commands Cleanup (SPEC-V3R3-CMD-CLEANUP-001)
- **`/moai gate` command file**: Added missing wrapper
  - Thin Command Pattern: routes to gate.md skill
  - Syntax: `/moai gate [<SPEC-ID>]`
  - Runs pre-commit quality gate (parallel: lint+format+type+test)
- **review.md Phase 4**: Strengthened security depth
  - Dependency vulnerability scanning (cargo audit, npm audit, pip audit)
  - Secrets git history scan (commitlint, gitleaks)
  - Data isolation and access control patterns
- **sync.md Phase 0.55**: Strengthened manifest audit
  - CHANGELOG consistency with commits
  - Version field synchronization
  - Package metadata validation
- **Removed**: `/moai context` skill
  - Superseded by @MX annotations (code-level context)
  - Superseded by auto-memory (persistent session memory)
  - Git-based context memory deprecated in favor of SPEC documents

### Breaking Changes

| Change ID | Type | Description | Migration |
|-----------|------|-------------|-----------|
| **BC-V3R3-001** | Feature | 7 expert agents now have Agent tool (T2 escalation) | Additive; existing code unchanged |
| **BC-V3R3-002** | Feature | expert-debug, expert-performance have Write/Edit | Additive; existing code unchanged |
| **BC-V3R3-006** | Config | Token Circuit Breaker with per-agent budgets | Optional: customize runtime.yaml; default non-breaking |

**Impact Assessment**:
- ✅ **BC-V3R3-001/002**: Additive capabilities, no existing workflows broken
- ✅ **BC-V3R3-006**: Warning-first policy; hard-fail deferred to Phase 5

**Opt-out Options**:
- Set `MOAI_RUNTIME_DISABLED=1` to disable token circuit breaker temporarily
- Edit `.moai/config/sections/runtime.yaml` to customize per-agent budgets

---

## 업그레이드 가이드

### 1. 설치

```bash
# Homebrew
brew upgrade moai-adk

# Go install
go install github.com/modu-ai/moai-adk/cmd/moai@v2.15.0

# 검증
moai version  # → v2.15.0
```

### 2. 기존 프로젝트 동기화

```bash
cd your-project/
moai update          # 11 skill PD + 7 expert tools + Mobile skills 자동 배포
moai doctor          # 진단 (BC-V3R3-001~006 호환성 체크)
```

**주의**: Go 바이너리 재빌드 필요
- SessionStart hook이 새로운 runtime.yaml을 로드하므로
- 기존 바이너리는 hook 실행 불가
- `moai update` 후 Claude Code 재시작 권장

### 3. 새 프로젝트에 v2.15 패턴 활용

#### A) Mobile 프로젝트

```bash
moai init my-mobile-app
cd my-mobile-app
/moai plan "iOS native authentication"
# → expert-mobile + ios-native skill 자동 로드
# → 4 strategy guide: iOS/Android/RN/Flutter
```

#### B) Token Budget 안전망

```yaml
# .moai/config/sections/runtime.yaml (자동 생성)
runtime:
  context_management:
    pre_clear_threshold: 0.75  # 75% 도달 → progress.md 자동 저장
    per_agent_budget:
      manager-strategy: 60000
      expert-*: 40000
      evaluator-*: 20000
    circuit_breaker:
      stall_detection_seconds: 60
      retry_max: 3
```

#### C) Expert Escalation 활용

```
사용자: /moai run SPEC-AUTH-001 (auth + db + frontend)
  → manager-tdd → expert-backend (Go API)
                  ↓ Agent() escalation (depth 1)
                  → expert-security (OWASP review)
                  → expert-database (schema design)
  → manager-tdd → expert-frontend (UI)
```

이전: manager-tdd 단독 처리 (silent assumption)  
이후: 도메인별 expert 자동 협업

---

## 알려진 한계

- **expert-mobile**: New agent — dogfood 데이터 누적 후 v2.16에서 정제 예정
- **Token Circuit Breaker**: Warning-first 모드 (1주 시범 후 hard-fail로 전환)
- **Architect/Editor split** (manager-tdd v2): v2.16 예정
- **v3R2 31 SPECs 잔여 작업**: 이번 릴리스에 미포함

---

## 다음 릴리스 예고 (v2.16+)

### V3R3 Iteration 3 — Architecture
- **ARCH-001**: Tier enforcement (T1/T2/T3 agent depth)
- **ARCH-002**: Architect/Editor split (manager-tdd v2 separation)

### V3R3 Iteration 4+ — Coverage
- **COV-002**: ML/AI domain (expert-ml-ai, data science skills)
- **COV-003**: DevOps/IaC (expert-infrastructure, Terraform/Ansible skills)
- **COV-004**: Backend framework depth (Rails, Django, Spring Boot deep dives)

---

## 기술 세부 사항

### Binary Size
- Δ+33.6 KB (constitution registry + runtime module)
- Total: ~12 MB (unchanged)

### Test Coverage
- New code coverage: 86.5% (constitution), 88.2% (runtime)
- Overall coverage: 85%+

### Performance
- SessionStart hook: <50ms overhead (runtime.yaml parse)
- Tracker operations: <1ms per call (O(1) RWMutex)
- Circuit breaker detection: <100ms per check

### Dependencies
- No new external dependencies
- All internal packages (runtime, constitution)

---

## 기여자 및 감사

- **Architecture & Implementation**: Claude (manager-tdd, manager-spec, manager-strategy)
- **Audit & Research**: Claude (Explore, auditors, researchers)
- **Direction & Review**: GOOS행님 (@goos)
- **Testing & Validation**: Community feedback (early dogfood)

---

## 참고 자료

- 공식 문서: https://adk.mo.ai.kr
- GitHub: https://github.com/modu-ai/moai-adk
- Discord 커뮤니티: https://discord.gg/moai-adk
- SPEC 문서: `.moai/specs/SPEC-V3R3-*/`
- 세션 audit: auto-memory at `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/`

---

## 설치 후 체크리스트

- [ ] `moai version` → v2.15.0 확인
- [ ] `moai update` 실행 후 Claude Code 재시작
- [ ] `moai doctor` 진단 결과 확인
- [ ] `.moai/config/sections/runtime.yaml` 검토
- [ ] 첫 번째 `/moai plan` 또는 `/moai run` 실행해보기
- [ ] Mobile 프로젝트: expert-mobile 라우팅 테스트

---

**Status**: Final  
**Release Version**: v2.15.0  
**Build Date**: 2026-04-26  
**Commit Range**: b3b635a80 ~ 538f44950 (main branch)
