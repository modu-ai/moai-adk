# Architecture Diagram: 마법사 UX 시스템 @DESIGN:WIZARD-UX

> **@DESIGN:WIZARD-UX** "3계층 아키텍처 기반 시각적 설계"

## 🏗️ 전체 시스템 아키텍처

### 1. 고수준 컴포넌트 다이어그램

```mermaid
graph TB
    %% 사용자 인터페이스 계층
    subgraph "UI Layer (Claude Code Terminal)"
        CLI["/moai:1-project 명령어"]
        Terminal["Rich Console 출력"]
        Progress["진행바 & 상태 표시"]
        Input["사용자 입력 처리"]
    end

    %% 비즈니스 로직 계층
    subgraph "Business Logic Layer"
        WizardCore["WizardCore"]
        QuestionEngine["QuestionEngine"]
        ValidationEngine["ValidationEngine"]
        StateManager["StateManager"]
    end

    %% 렌더링 계층
    subgraph "Rendering Layer"
        UIRenderer["UIRenderer"]
        ProgressRenderer["ProgressRenderer"]
        QuestionRenderer["QuestionRenderer"]
        SummaryRenderer["SummaryRenderer"]
    end

    %% 데이터 계층
    subgraph "Data Layer"
        SessionStore[".moai/indexes/state.json"]
        ConfigStore[".moai/config.json"]
        TemplateStore["프로젝트 템플릿"]
        UserPrefs["사용자 선호도"]
    end

    %% 에이전트 계층
    subgraph "Agent Integration"
        SteeringAgent["steering-architect"]
        SpecAgent["spec-manager"]
        TagAgent["tag-indexer"]
    end

    %% 연결 관계
    CLI --> WizardCore
    WizardCore --> QuestionEngine
    WizardCore --> ValidationEngine
    WizardCore --> StateManager

    QuestionEngine --> UIRenderer
    UIRenderer --> ProgressRenderer
    UIRenderer --> QuestionRenderer
    UIRenderer --> SummaryRenderer

    StateManager --> SessionStore
    StateManager --> ConfigStore
    WizardCore --> TemplateStore
    WizardCore --> UserPrefs

    UIRenderer --> Terminal
    Terminal --> Progress
    Terminal --> Input
    Input --> ValidationEngine

    WizardCore --> SteeringAgent
    WizardCore --> SpecAgent
    WizardCore --> TagAgent

    %% 스타일링
    classDef uiLayer fill:#e1f5fe
    classDef businessLayer fill:#fff3e0
    classDef renderLayer fill:#f3e5f5
    classDef dataLayer fill:#e8f5e8
    classDef agentLayer fill:#fff8e1

    class CLI,Terminal,Progress,Input uiLayer
    class WizardCore,QuestionEngine,ValidationEngine,StateManager businessLayer
    class UIRenderer,ProgressRenderer,QuestionRenderer,SummaryRenderer renderLayer
    class SessionStore,ConfigStore,TemplateStore,UserPrefs dataLayer
    class SteeringAgent,SpecAgent,TagAgent agentLayer
```

### 2. 데이터 흐름 다이어그램

```mermaid
sequenceDiagram
    participant User as 사용자
    participant CLI as Claude Code CLI
    participant Core as WizardCore
    participant QE as QuestionEngine
    participant VE as ValidationEngine
    participant UI as UIRenderer
    participant State as StateManager
    participant Agent as Agents

    User->>CLI: /moai:1-project
    CLI->>Core: 마법사 시작 요청

    Core->>State: 기존 세션 확인
    State-->>Core: 세션 상태 반환

    Core->>QE: 첫 질문 요청
    QE-->>Core: 질문 데이터

    Core->>UI: 진행바 + 질문 렌더링
    UI-->>CLI: Rich 마크업 출력
    CLI-->>User: 화면 표시

    User->>CLI: 답변 입력
    CLI->>VE: 입력 검증 요청

    alt 검증 성공
        VE-->>Core: 검증 성공
        Core->>State: 답변 저장
        Core->>QE: 키워드 감지 & 동적 질문
        QE-->>Core: 다음 질문 데이터
        Core->>UI: 진행바 업데이트 + 질문
    else 검증 실패
        VE-->>Core: 검증 실패 + 제안
        Core->>UI: 에러 메시지 + 재입력
        UI-->>CLI: 에러 화면 표시
    end

    loop 10단계 반복
        Note over User,Agent: 위 과정 반복
    end

    Core->>Agent: 프로젝트 생성 요청
    Agent-->>Core: 생성 완료
    Core->>UI: 완료 화면 렌더링
    UI-->>CLI: 성공 메시지 출력
```

### 3. 모듈 의존성 다이어그램

```mermaid
graph LR
    %% 핵심 모듈
    subgraph "Core Modules"
        WC[WizardCore<br/>세션 관리]
        QE[QuestionEngine<br/>질문 생성]
        VE[ValidationEngine<br/>입력 검증]
        SM[StateManager<br/>상태 저장]
    end

    %% UI 모듈
    subgraph "UI Modules"
        UR[UIRenderer<br/>렌더링 총괄]
        PR[ProgressRenderer<br/>진행바]
        QR[QuestionRenderer<br/>질문 UI]
        SR[SummaryRenderer<br/>요약 UI]
        ER[ErrorRenderer<br/>에러 UI]
    end

    %% 데이터 모듈
    subgraph "Data Modules"
        SS[SessionStore<br/>세션 저장]
        CS[ConfigStore<br/>설정 저장]
        TS[TemplateStore<br/>템플릿 관리]
        UP[UserPreferences<br/>사용자 설정]
    end

    %% 유틸리티 모듈
    subgraph "Utility Modules"
        KD[KeywordDetector<br/>키워드 감지]
        IS[InputSanitizer<br/>입력 정제]
        PM[PerformanceMonitor<br/>성능 모니터링]
        Logger[StructuredLogger<br/>구조화 로깅]
    end

    %% 의존성 관계
    WC --> QE
    WC --> VE
    WC --> SM
    WC --> UR

    QE --> KD
    VE --> IS
    SM --> SS
    SM --> CS
    SM --> TS
    SM --> UP

    UR --> PR
    UR --> QR
    UR --> SR
    UR --> ER

    VE --> Logger
    WC --> Logger
    QE --> PM
    UR --> PM

    %% 스타일링
    classDef coreModule fill:#ffeb3b,color:#000
    classDef uiModule fill:#2196f3,color:#fff
    classDef dataModule fill:#4caf50,color:#fff
    classDef utilModule fill:#9c27b0,color:#fff

    class WC,QE,VE,SM coreModule
    class UR,PR,QR,SR,ER uiModule
    class SS,CS,TS,UP dataModule
    class KD,IS,PM,Logger utilModule
```

## 🔄 에러 처리 및 복구 플로우

### 1. 에러 처리 전략

```mermaid
graph TD
    Start([사용자 입력]) --> Validate{입력 검증}

    Validate -->|성공| Process[답변 처리]
    Validate -->|실패| ErrorType{에러 유형 분석}

    ErrorType -->|빈 입력| EmptyError[빈 입력 에러]
    ErrorType -->|길이 부족| LengthError[최소 길이 에러]
    ErrorType -->|형식 오류| FormatError[형식 에러]
    ErrorType -->|보안 위험| SecurityError[보안 에러]

    EmptyError --> ShowHint[힌트 표시]
    LengthError --> ShowExample[예시 표시]
    FormatError --> ShowFormat[형식 가이드]
    SecurityError --> ShowSecurity[보안 경고]

    ShowHint --> Retry[재입력 요청]
    ShowExample --> Retry
    ShowFormat --> Retry
    ShowSecurity --> Retry

    Retry --> Validate

    Process --> Save{상태 저장}
    Save -->|성공| NextStep[다음 단계]
    Save -->|실패| StorageError[저장 에러]

    StorageError --> Backup[백업에서 복원]
    Backup --> Save

    NextStep --> End([완료])

    %% 스타일링
    classDef errorNode fill:#f44336,color:#fff
    classDef successNode fill:#4caf50,color:#fff
    classDef processNode fill:#2196f3,color:#fff

    class EmptyError,LengthError,FormatError,SecurityError,StorageError errorNode
    class Process,NextStep,End successNode
    class Validate,Save,Backup processNode
```

### 2. 세션 복구 메커니즘

```mermaid
graph TB
    SessionStart([마법사 시작]) --> CheckExisting{기존 세션 확인}

    CheckExisting -->|없음| NewSession[새 세션 생성]
    CheckExisting -->|있음| SessionType{세션 상태}

    SessionType -->|진행중| ResumeOption{복원 선택}
    SessionType -->|완료| CompletedSession[완료된 세션]
    SessionType -->|중단| AbandonedSession[중단된 세션]
    SessionType -->|오류| CorruptedSession[손상된 세션]

    ResumeOption -->|복원| ValidateSession{세션 검증}
    ResumeOption -->|새로시작| NewSession

    ValidateSession -->|유효| RestoreSession[세션 복원]
    ValidateSession -->|무효| RepairSession[세션 수리]

    RepairSession -->|수리성공| RestoreSession
    RepairSession -->|수리실패| BackupRestore[백업 복원]

    BackupRestore -->|성공| RestoreSession
    BackupRestore -->|실패| NewSession

    CompletedSession --> ShowResults[완료 결과 표시]
    AbandonedSession --> ResumeOption
    CorruptedSession --> RepairSession

    NewSession --> InitializeWizard[마법사 초기화]
    RestoreSession --> ContinueWizard[마법사 계속]

    InitializeWizard --> WizardFlow[마법사 진행]
    ContinueWizard --> WizardFlow
    ShowResults --> WizardFlow

    WizardFlow --> Complete([완료])

    %% 스타일링
    classDef startNode fill:#4caf50,color:#fff
    classDef decisionNode fill:#ff9800,color:#fff
    classDef errorNode fill:#f44336,color:#fff
    classDef processNode fill:#2196f3,color:#fff

    class SessionStart,Complete startNode
    class CheckExisting,SessionType,ResumeOption,ValidateSession decisionNode
    class CorruptedSession,RepairSession,BackupRestore errorNode
    class NewSession,RestoreSession,InitializeWizard,ContinueWizard,WizardFlow processNode
```

## 🎨 UI 컴포넌트 구조

### 1. 렌더링 컴포넌트 계층

```mermaid
graph TD
    %% 최상위 렌더러
    subgraph "Top Level Renderer"
        MasterRenderer[MasterRenderer<br/>전체 화면 조율]
    end

    %% 메인 컴포넌트들
    subgraph "Main Components"
        HeaderRenderer[HeaderRenderer<br/>🗿 타이틀 & 브랜딩]
        ProgressRenderer[ProgressRenderer<br/>진행바 & 단계 표시]
        QuestionRenderer[QuestionRenderer<br/>질문 & 입력 폼]
        SummaryRenderer[SummaryRenderer<br/>설정 요약 패널]
        FooterRenderer[FooterRenderer<br/>도움말 & 단축키]
    end

    %% 서브 컴포넌트들
    subgraph "Sub Components"
        ProgressBar[ProgressBar<br/>시각적 진행바]
        StepIndicator[StepIndicator<br/>단계 인디케이터]
        QuestionText[QuestionText<br/>질문 텍스트]
        InputField[InputField<br/>입력 필드]
        ValidationMsg[ValidationMsg<br/>검증 메시지]
        HintPanel[HintPanel<br/>힌트 & 예시]
        ConfigPreview[ConfigPreview<br/>설정 미리보기]
        ActionButtons[ActionButtons<br/>액션 버튼들]
    end

    %% 유틸리티 컴포넌트들
    subgraph "Utility Components"
        ColorScheme[ColorScheme<br/>색상 관리]
        LayoutManager[LayoutManager<br/>레이아웃 계산]
        ResponsiveHandler[ResponsiveHandler<br/>화면 크기 대응]
        AnimationEngine[AnimationEngine<br/>애니메이션 효과]
    end

    %% 연결 관계
    MasterRenderer --> HeaderRenderer
    MasterRenderer --> ProgressRenderer
    MasterRenderer --> QuestionRenderer
    MasterRenderer --> SummaryRenderer
    MasterRenderer --> FooterRenderer

    ProgressRenderer --> ProgressBar
    ProgressRenderer --> StepIndicator

    QuestionRenderer --> QuestionText
    QuestionRenderer --> InputField
    QuestionRenderer --> ValidationMsg
    QuestionRenderer --> HintPanel

    SummaryRenderer --> ConfigPreview
    SummaryRenderer --> ActionButtons

    %% 유틸리티 연결
    MasterRenderer --> ColorScheme
    MasterRenderer --> LayoutManager
    MasterRenderer --> ResponsiveHandler
    MasterRenderer --> AnimationEngine

    %% 스타일링
    classDef masterClass fill:#673ab7,color:#fff
    classDef mainClass fill:#3f51b5,color:#fff
    classDef subClass fill:#2196f3,color:#fff
    classDef utilClass fill:#009688,color:#fff

    class MasterRenderer masterClass
    class HeaderRenderer,ProgressRenderer,QuestionRenderer,SummaryRenderer,FooterRenderer mainClass
    class ProgressBar,StepIndicator,QuestionText,InputField,ValidationMsg,HintPanel,ConfigPreview,ActionButtons subClass
    class ColorScheme,LayoutManager,ResponsiveHandler,AnimationEngine utilClass
```

### 2. 상태 기반 UI 전환

```mermaid
stateDiagram-v2
    [*] --> Initializing

    Initializing --> LoadingSession : 세션 로드
    LoadingSession --> SessionFound : 기존 세션 발견
    LoadingSession --> NewSession : 새 세션 생성

    SessionFound --> ResumePrompt : 복원 여부 확인
    ResumePrompt --> RestoringSession : 복원 선택
    ResumePrompt --> NewSession : 새로 시작

    RestoringSession --> QuestionDisplay : 복원 완료
    NewSession --> QuestionDisplay : 첫 질문 표시

    QuestionDisplay --> WaitingInput : 사용자 입력 대기
    WaitingInput --> ValidatingInput : 입력 검증 중

    ValidatingInput --> ValidationError : 검증 실패
    ValidatingInput --> ProcessingAnswer : 검증 성공

    ValidationError --> QuestionDisplay : 에러 표시 후 재입력

    ProcessingAnswer --> SavingState : 답변 저장
    SavingState --> UpdatingProgress : 진행 상황 업데이트

    UpdatingProgress --> CheckingCompletion : 완료 여부 확인
    CheckingCompletion --> QuestionDisplay : 다음 질문
    CheckingCompletion --> GeneratingProject : 모든 질문 완료

    GeneratingProject --> ProjectComplete : 프로젝트 생성 완료
    ProjectComplete --> [*] : 마법사 종료

    %% 에러 상태들
    LoadingSession --> SessionError : 세션 로드 실패
    SavingState --> StateError : 상태 저장 실패
    GeneratingProject --> ProjectError : 프로젝트 생성 실패

    SessionError --> NewSession : 새 세션으로 복구
    StateError --> QuestionDisplay : 임시 저장 후 계속
    ProjectError --> GeneratingProject : 재시도
```

## 🔧 성능 최적화 아키텍처

### 1. 렌더링 최적화

```mermaid
graph LR
    %% 입력 처리
    UserInput[사용자 입력] --> InputBuffer[입력 버퍼]
    InputBuffer --> Debouncing[디바운싱<br/>300ms]

    %% 검증 파이프라인
    Debouncing --> QuickValidation[빠른 검증<br/>< 50ms]
    QuickValidation --> ValidationCache[검증 캐시]
    ValidationCache --> FullValidation[전체 검증<br/>< 200ms]

    %% 렌더링 파이프라인
    FullValidation --> RenderQueue[렌더링 큐]
    RenderQueue --> ComponentCache[컴포넌트 캐시]
    ComponentCache --> VirtualDOM[가상 DOM]
    VirtualDOM --> DiffEngine[차이점 계산]
    DiffEngine --> ActualRender[실제 렌더링]

    %% 성능 모니터링
    ActualRender --> PerformanceLog[성능 로그]
    PerformanceLog --> MetricsCollection[메트릭 수집]

    %% 피드백 루프
    MetricsCollection --> AdaptiveThrottling[적응형 스로틀링]
    AdaptiveThrottling --> InputBuffer

    %% 스타일링
    classDef inputClass fill:#4caf50,color:#fff
    classDef validationClass fill:#ff9800,color:#fff
    classDef renderClass fill:#2196f3,color:#fff
    classDef monitorClass fill:#9c27b0,color:#fff

    class UserInput,InputBuffer,Debouncing inputClass
    class QuickValidation,ValidationCache,FullValidation validationClass
    class RenderQueue,ComponentCache,VirtualDOM,DiffEngine,ActualRender renderClass
    class PerformanceLog,MetricsCollection,AdaptiveThrottling monitorClass
```

### 2. 메모리 관리 전략

```mermaid
graph TB
    %% 메모리 풀
    subgraph "Memory Pools"
        SessionPool[세션 풀<br/>최대 5개]
        ComponentPool[컴포넌트 풀<br/>최대 20개]
        CachePool[캐시 풀<br/>최대 50MB]
    end

    %% 가비지 컬렉션
    subgraph "Garbage Collection"
        GCTrigger[GC 트리거<br/>70% 사용률]
        CompactMemory[메모리 압축]
        WeakReferences[약한 참조 정리]
    end

    %% 캐시 관리
    subgraph "Cache Management"
        LRUCache[LRU 캐시<br/>최근 사용 우선]
        TTLCache[TTL 캐시<br/>시간 기반 만료]
        SizeBasedEviction[크기 기반 제거]
    end

    %% 메모리 모니터링
    subgraph "Memory Monitoring"
        MemoryTracker[메모리 추적기]
        LeakDetector[메모리 누수 감지]
        AllocationProfiler[할당 프로파일러]
    end

    %% 연결 관계
    SessionPool --> GCTrigger
    ComponentPool --> GCTrigger
    CachePool --> GCTrigger

    GCTrigger --> CompactMemory
    GCTrigger --> WeakReferences

    CompactMemory --> LRUCache
    WeakReferences --> TTLCache
    LRUCache --> SizeBasedEviction

    SessionPool --> MemoryTracker
    ComponentPool --> MemoryTracker
    CachePool --> MemoryTracker

    MemoryTracker --> LeakDetector
    MemoryTracker --> AllocationProfiler

    %% 스타일링
    classDef poolClass fill:#4caf50,color:#fff
    classDef gcClass fill:#f44336,color:#fff
    classDef cacheClass fill:#2196f3,color:#fff
    classDef monitorClass fill:#ff9800,color:#fff

    class SessionPool,ComponentPool,CachePool poolClass
    class GCTrigger,CompactMemory,WeakReferences gcClass
    class LRUCache,TTLCache,SizeBasedEviction cacheClass
    class MemoryTracker,LeakDetector,AllocationProfiler monitorClass
```

## 📊 모니터링 및 관찰성

### 1. 실시간 메트릭 수집

```mermaid
graph TD
    %% 메트릭 소스
    subgraph "Metric Sources"
        UserActions[사용자 액션]
        SystemPerf[시스템 성능]
        RenderTime[렌더링 시간]
        ValidationTime[검증 시간]
        ErrorEvents[에러 이벤트]
    end

    %% 메트릭 수집기
    subgraph "Metric Collectors"
        ActionCollector[액션 수집기]
        PerfCollector[성능 수집기]
        ErrorCollector[에러 수집기]
        TimingCollector[타이밍 수집기]
    end

    %% 메트릭 처리
    subgraph "Metric Processing"
        Aggregator[집계기]
        Correlator[상관관계 분석]
        Alerting[알림 시스템]
        Dashboard[대시보드]
    end

    %% 연결 관계
    UserActions --> ActionCollector
    SystemPerf --> PerfCollector
    RenderTime --> TimingCollector
    ValidationTime --> TimingCollector
    ErrorEvents --> ErrorCollector

    ActionCollector --> Aggregator
    PerfCollector --> Aggregator
    ErrorCollector --> Aggregator
    TimingCollector --> Aggregator

    Aggregator --> Correlator
    Correlator --> Alerting
    Correlator --> Dashboard

    %% 스타일링
    classDef sourceClass fill:#4caf50,color:#fff
    classDef collectorClass fill:#2196f3,color:#fff
    classDef processClass fill:#ff9800,color:#fff

    class UserActions,SystemPerf,RenderTime,ValidationTime,ErrorEvents sourceClass
    class ActionCollector,PerfCollector,ErrorCollector,TimingCollector collectorClass
    class Aggregator,Correlator,Alerting,Dashboard processClass
```

## 🚀 배포 및 확장성 고려사항

### 1. 모듈화 전략

| 모듈 | 책임 | 독립성 | 확장성 |
|------|------|--------|--------|
| **WizardCore** | 세션 관리, 전체 흐름 제어 | High | Medium |
| **QuestionEngine** | 질문 생성, 동적 분기 | High | High |
| **ValidationEngine** | 입력 검증, 보안 검사 | High | High |
| **UIRenderer** | 화면 렌더링, 사용자 경험 | High | Medium |
| **StateManager** | 데이터 저장, 상태 관리 | Medium | High |

### 2. 성능 목표 및 제약사항

| 지표 | 목표 | 제약사항 | 모니터링 |
|------|------|----------|----------|
| **응답 시간** | < 500ms | 네트워크 지연 | 실시간 |
| **메모리 사용** | < 100MB | 시스템 RAM | 연속 모니터링 |
| **CPU 사용률** | < 20% | 백그라운드 실행 | 주기적 체크 |
| **완료율** | > 85% | 사용자 이탈 | 세션 추적 |

---

## 🔗 연관 태그 시스템

**@DESIGN:WIZARD-UX**와 연결된 주요 태그들:
- **@REQ:WIZARD-UX-001** → 요구사항 추적
- **@DATA:WIZARD-UX** → 데이터 모델 연결
- **@TASK:ARCH-IMPL** → 구현 작업 연결
- **@TEST:ARCH-VALIDATION** → 아키텍처 검증 테스트
- **@PERF:WIZARD-METRICS** → 성능 모니터링

---

> **@DESIGN:WIZARD-UX**를 통해 이 아키텍처 설계가 전체 시스템에서 완벽하게 추적됩니다.
>
> **3계층 아키텍처와 모듈형 설계로 확장성과 유지보수성을 보장합니다.**