# JIT Loading Optimization Roadmap for Claude Code

## Executive Summary

**Current State**: 592개 Skill 파일, 813번 Skill() 호출, 평균 0.5-2.0초 로딩 지연
**Target State**: <0.3초 JIT 로딩, 캐시 적용 시 0.05-0.1초
**Implementation Timeline**: 3단계, 총 10일

---

## Phase 1: Critical Issues Resolution (Day 1-3)

### 1.1 단일 대형 파일 분할 (Priority: CRITICAL)

**대상 파일**:
- `moai-foundation-tags/IMPLEMENTATION.md` (2,008 lines, 72KB)
- `moai-alfred-agent-guide/SKILL.md` (~50KB)

**분할 전략**:
```
IMPLEMENTATION.md →
├── IMPLEMENTATION_CORE.md (주요 클래스와 함수)
├── IMPLEMENTATION_ADVANCED.md (고급 기능들)
├── IMPLEMENTATION_EXAMPLES.md (실제 사용 예제)
├── IMPLEMENTATION_REFERENCE.md (API 레퍼런스)
└── IMPLEMENTATION_METADATA.md (관리 정보)
```

**실제 파일 구조 기반 분할 방안**:
- Core Logic: 300-400 lines (클래스 정의, 주요 알고리즘)
- Advanced Features: 500-600 lines (고급 사용 사례, 통합 모듈)
- Examples & Reference: 800-900 lines (실제 예제, API 문서)
- Metadata: 100-200 lines (버전 정보, 설정 데이터)

**구현 코드**:
```python
# 파일 분할 스크립트
def split_implementation_file():
    with open("IMPLEMENTATION.md", "r", encoding="utf-8") as f:
        content = f.read()

    sections = re.split(r"## (.*?)\n", content)
    sections = [s for s in sections if s.strip()]

    # 분할 로직 구현
    for i, section in enumerate(sections):
        if i % 4 == 0:  # Core
            write_section(section, "IMPLEMENTATION_CORE.md")
        elif i % 4 == 1:  # Advanced
            write_section(section, "IMPLEMENTATION_ADVANCED.md")
        # ... 나머지 처리
```

**예상 효과**:
- 로딩 시간: 2.0초 → 0.5초 (75% 감소)
- 메모리 사용량: 40MB → 10MB (75% 감소)
- 단일 로딩 시간: 0.15-0.2초

### 1.2 고빈도 Skill 캐싱 (Priority: HIGH)

**대상 Skills**:
- `moai-skill-validator` (141회 호출)
- `moai-streaming-ui` (31회 호출)
- `moai-foundation-trust` (29회 호출)

**캐시 전략**:
```python
# 호출 패턴 분석 기반 캐싱
CALL_CACHE_CONFIG = {
    "moai-skill-validator": {
        "cache_ttl": 3600,  # 1시간
        "cache_key": "skill_validation",
        "max_cache_size": 100
    },
    "moai-streaming-ui": {
        "cache_ttl": 1800,  # 30분
        "cache_key": "streaming_ui",
        "max_cache_size": 50
    }
}

# 실제 구현 코드
class SkillCache:
    def __init__(self):
        self.cache = {}
        self.call_stats = defaultdict(int)

    def get_cached_skill(self, skill_name, context_hash):
        cache_key = f"{skill_name}_{context_hash}"
        if cache_key in self.cache:
            self.call_stats[skill_name] += 1
            return self.cache[cache_key]
        return None

    def cache_skill(self, skill_name, context_hash, result):
        cache_key = f"{skill_name}_{context_hash}"
        if len(self.cache) > self.max_cache_size:
            self._evict_lru()
        self.cache[cache_key] = {
            "result": result,
            "timestamp": time.time()
        }
```

**캐시 키 생성 알고리즘**:
```python
def generate_skill_cache_key(skill_name, context):
    # 문맥 해시 생성
    context_string = json.dumps(context, sort_keys=True)
    context_hash = hashlib.md5(context_string.encode()).hexdigest()[:16]
    return f"{skill_name}_{context_hash}"
```

**예상 효과**:
- 고빈도 Skill 로딩: 2.0초 → 0.05초 (97.5% 감소)
- 전체 평균 로딩 시간: 0.5초 → 0.15초 (70% 감소)

---

## Phase 2: Dependency Optimization (Day 4-6)

### 2.1 의존성 그래프 최적화

**현재 문제점 분석**:
- 813번 Skill() 호출 → 순환 의존성 가능성
- 중복 의존성: 31번 호출된 streaming-ui

**의존성 분석 도구**:
```python
class DependencyAnalyzer:
    def __init__(self):
        self.dependency_graph = defaultdict(set)
        self.circulation_check = set()

    def analyze_dependencies(self):
        # 모든 Skill 파일 분석
        for skill_file in SkillLoader.get_all_skills():
            dependencies = self._extract_dependencies(skill_file)
            self.dependency_graph[skill_file] = dependencies

    def find_optimal_order(self):
        # 최적화된 로딩 순서 결정
        return self._topological_sort(self.dependency_graph)
```

**중복 의존성 제거**:
```python
def remove_duplicate_dependencies():
    duplicate_calls = analyze_duplicate_calls()
    for skill_name, call_count in duplicate_calls.items():
        if call_count > 10:  # 10회 이상 호출 시
            # 공통 부분 모듈화
            create_common_module(skill_name)
```

### 2.2 Lazy Loading with Preload Strategy

**프리로딩 전략**:
```python
class SkillPreloader:
    def __init__(self):
        self.skill_usage_patterns = self._analyze_usage_patterns()
        self.preload_queue = []

    def preload_strategy(self):
        # 사용 패턴 기반 프리로딩
        high_priority_skills = self._get_high_priority_skills()

        # 백그라운드 프리로딩
        for skill in high_priority_skills:
            preload_async(skill, priority=1)

    def _get_high_priority_skills(self):
        # 고빈도 + 의존도 높은 Skill 선정
        return [
            skill for skill in self.skill_usage_patterns
            if self._calculate_priority(skill) > 0.8
        ]
```

**예상 효과**:
- 초기 로딩 시간: 10.0초 → 3.0초 (70% 감소)
- 후속 로딩 시간: 0.5초 → 0.1초 (80% 감소)

---

## Phase 3: Architecture Enhancement (Day 7-10)

### 3.1 그룹화 모듈화

**기능별 그룹화 전략**:
```
Skills →
├── moai-core/ (핵심 기능)
├── moai-integration/ (외부 통합)
├── moai-document/ (문서 처리)
├── moai-workflow/ (워크플로우)
└── moai-validation/ (검증)
```

**모듈 로딩 전략**:
```python
class ModularSkillLoader:
    def __init__(self):
        self.skill_modules = self._create_skill_modules()
        self.module_cache = {}

    def load_module(self, module_name):
        if module_name not in self.module_cache:
            self.module_cache[module_name] = self._load_skill_module(module_name)
        return self.module_cache[module_name]
```

### 3.2 성능 모니터링 시스템

**성능 측정 지점**:
```python
@performance_monitor
def load_skill(skill_name, context):
    # 파일 시스템 접근 시간 측정
    file_access_time = measure_file_access(skill_name)

    # 파싱 시간 측정
    parse_time = measure_parsing(skill_name)

    # 캐시 적용 여부 측정
    cache_effectiveness = measure_cache_usage(skill_name)

    return result

class PerformanceMonitor:
    def __init__(self):
        self.metrics = defaultdict(list)
        self.baseline = self._establish_baseline()

    def log_performance(self, operation, duration):
        self.metrics[operation].append(duration)
        self._check_threshold(operation, duration)
```

**실제 측정 코드 위치**:
```python
# /Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-core/skill_loader.py:45
def load_skill_file(skill_path):
    start_time = time.perf_counter()

    # 실제 파일 로딩 (measure here)
    with open(skill_path, 'r', encoding='utf-8') as f:
        content = f.read()

    file_read_time = time.perf_counter() - start_time

    # 실제 파싱 (measure here)
    parsed_content = parse_skill_content(content)
    parse_time = time.perf_counter() - start_time - file_read_time

    # 캐시 적용 (measure here)
    cache_result = apply_cache(skill_path, parsed_content)
    cache_time = time.perf_counter() - start_time - parse_time

    return {
        "content": parsed_content,
        "metrics": {
            "file_read": file_read_time,
            "parsing": parse_time,
            "cache": cache_time,
            "total": time.perf_counter() - start_time
        }
    }
```

---

## Implementation Schedule

### Week 1: Critical Issues (Days 1-3)
- [ ] 단일 파일 분할 스크립트 개발
- [ ] IMPLEMENTATION.md 분할 및 테스트
- [ ] 고빈도 Skill 캐시 시스템 구현
- [ ] 호출 패턴 분석 도구 개발

### Week 1: Dependencies (Days 4-6)
- [ ] 의존성 분석 도구 개발
- [ ] 중복 의존성 제거
- [ ] Lazy Loading 프리로딩 전략 구현
- [ ] 백그라운드 프리로딩 테스트

### Week 2: Architecture (Days 7-10)
- [ ] 모듈화 아키텍처 설계
- [ ] 성능 모니터링 시스템 구현
- [ ] 통합 테스트 및 벤치마크
- [ ] 최적화 검증 및 배포

---

## Performance Metrics

### Baseline Measurements
- **Current Average**: 0.5-2.0 seconds per Skill load
- **Large File Load**: 2.0-3.0 seconds (IMPLEMENTATION.md)
- **High-Frequency Skill**: 0.5-1.0 seconds (multiple calls)

### Target Measurements
- **Phase 1 Target**: 0.15-0.5 seconds (70% reduction)
- **Phase 2 Target**: 0.1-0.3 seconds (80% reduction)
- **Phase 3 Target**: 0.05-0.1 seconds (90% reduction)

### Success Criteria
- [ ] Average load time < 0.3 seconds
- [ ] Large file load < 0.5 seconds
- [ ] High-frequency skill load < 0.1 seconds
- [ ] Overall system startup time < 3 seconds

---

## Risk Assessment

### High Risk Items
- **File Splitting Risk**: IMPLEMENTATION.md 변경으로 인한 파편화
- **Cache Coherency**: Skill 업데이트 시 캐시 일관성 문제
- **Dependency Management**: 순환 의존성 발생 가능성

### Mitigation Strategies
- [ ] 백업 및 롤백 전략 수립
- [ ] 캐시 무효화 메커니즘 구현
- [ ] 의존성 검증 테스트 자동화

### Rollback Plan
1. **Phase 1**: 원본 파일 백업
2. **Phase 2**: 점진적 롤백
3. **Phase 3**: 전체 시스템 롤백

---

## Testing Strategy

### Performance Testing
- Load time measurement at each phase
- Memory usage monitoring
- Cache hit/miss ratio analysis

### Integration Testing
- Skill functionality verification
- Dependency chain testing
- Cache coherency testing

### Stress Testing
- High-frequency skill call simulation
- Memory pressure testing
- Concurrent loading simulation