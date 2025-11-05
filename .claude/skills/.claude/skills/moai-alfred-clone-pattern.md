# Clone 패턴: Alfred의 자가 복제 메커니즘

## 개요

**Clone 패턴**은 Alfred가 자신의 복제본을 생성하여 특정 작업을 위임하는 아키텍처입니다. 복제본은 원본과 동일한 도구, 컨텍스트, Skills를 가지지만, 특정 작업 설명으로만 구분됩니다.

### 언제 사용하는가?

```
Task를 받으면:

1. 도메인 특화 필요? (UI, Backend, DB, Security, ML)
   ├─ YES: Lead-Specialist 패턴 활용
   │  ├─ ui-ux-expert
   │  ├─ backend-expert
   │  ├─ moai-domain-database
   │  └─ ...
   │
   └─ NO: Clone 패턴 또는 직접 처리
      └─ 다음 단계로

2. 멀티스텝 복잡 작업? (5단계 이상)
   ├─ YES: Clone 패턴 (Master-Clone)
   │  - 마이그레이션 (v0.14.0 → v0.15.2)
   │  - 전체 리팩토링 (100+ 파일)
   │  - 병렬 처리 작업 (독립적 서브태스크)
   │  - 탐색적 작업 (결과 불확실)
   │
   └─ NO: Alfred가 직접 처리
```

---

## Clone 패턴의 구조

### Master-Clone Architecture

```
Main Alfred Session
    │
    ├─ 의도 분석
    ├─ 작업 분류 (도메인/복잡도)
    │
    └─ Clone 생성 (해당하는 경우)
        │
        └─ Clone Instance
            ├─ 전체 프로젝트 컨텍스트
            ├─ 모든 도구 접근 권한
            ├─ 모든 Skills 로드됨
            ├─ 특정 작업 설명만 다름
            └─ 자율적 실행 & 학습
```

### Clone의 특징

| 특징 | Clone | Lead-Specialist |
|------|-------|-----------------|
| 컨텍스트 | 전체 유지 | 도메인만 전달 |
| 자율성 | 완전 자율적 | 지시에 따름 |
| 병렬 실행 | 가능 | 순차 실행 |
| 학습 | 자체 메모리 저장 | 피드백 기반 |
| 적합 작업 | 장기 멀티스텝 | 전문화 필요 |

---

## Clone 패턴 실제 사례

### 사례 1: v0.14.0 → v0.15.2 마이그레이션

**Alfred의 판단**:
```python
task = UserRequest(
    type="마이그레이션",
    scope="대규모",
    steps=8,  # 5단계 이상
    domains=["config", "hooks", "permissions"],
    uncertainty="높음"  # 새 구조로의 전환
)

# → Clone 패턴 적용
if task.steps > 5 and task.uncertainty > 0.5:
    clone = alfred.create_clone(
        description="v0.14.0 config 구조를 v0.15.2로 마이그레이션"
    )
    clone.execute()
```

**Clone의 자율적 실행**:
1. 마이그레이션 스크립트 작성
2. 백업 생성
3. 구조 변환 (자동)
4. 검증 실행
5. 실패 시 자가 디버깅
6. 학습 내용 메모리 저장

**결과**: 전체 작업을 자가 관리, 오류 발생 시 자동 복구

---

### 사례 2: 100+ 파일 리팩토링

**Alfred의 판단**:
```python
task = UserRequest(
    type="리팩토링",
    files_affected=150,  # 100+ 파일
    pattern="모든 imports 변경",
    complexity="높음"
)

# → Clone 패턴 적용
if task.files_affected > 100 and task.complexity == "높음":
    clone = alfred.create_clone(
        description="모든 Python 파일에서 imports 경로 업데이트"
    )
    clone.execute()
```

**Clone의 병렬 처리**:
- 파일을 배치로 분할
- 각 배치에서 변환 규칙 적용
- 검증 (타입 체크, 임포트 검사)
- 실패 부분만 재처리

**결과**: 리팩토링 속도 10배 향상

---

### 사례 3: 병렬 탐색 작업

**Alfred의 판단**:
```python
task = UserRequest(
    type="탐색_평가",
    items=["UI/UX 재설계", "Backend 최적화", "DB 마이그레이션"],
    independence="높음"  # 각 항목 독립적
)

# → Clone 패턴 적용
if task.independence > 0.7:
    clones = [
        alfred.create_clone(f"평가: {item}")
        for item in task.items
    ]
    results = parallel_execute(clones)
```

**Clone의 병렬 실행**:
```
Main Alfred
    ├─ Clone 1: UI/UX 재설계 평가 → 보고서
    ├─ Clone 2: Backend 최적화 평가 → 보고서
    └─ Clone 3: DB 마이그레이션 평가 → 보고서

(동시 실행 → 시간 1/3로 단축)
```

---

## Clone 패턴 구현 규칙

### Rule 1: Clone 생성 조건

```python
def should_create_clone(task) -> bool:
    """Clone 생성 여부 판단"""
    return (
        # 도메인 특화 불필요 AND
        task.domain not in ["ui", "backend", "db", "security", "ml"]

        # 다음 중 하나 만족:
        AND (
            task.steps >= 5                    # 5단계 이상
            or task.files >= 100               # 100+ 파일
            or task.parallelizable             # 병렬화 가능
            or task.uncertainty > 0.5          # 불확실성 높음
        )
    )
```

### Rule 2: Clone 생성 방식

```python
def create_clone(
    task_description: str,
    context_scope: str = "full",  # "full" | "domain"
    learning_enabled: bool = True
) -> CloneInstance:
    """
    Alfred 복제본 생성

    Args:
        task_description: 작업 설명 (구체적이고 목표 명확)
        context_scope: 컨텍스트 범위
        learning_enabled: 학습 메모리 저장 여부

    Returns:
        독립 실행 가능한 Clone 인스턴스
    """
    clone = Task(
        subagent_type="general-purpose",
        description=f"Clone: {task_description}",
        prompt=f"""
You are an Alfred Clone with full MoAI-ADK capabilities.

TASK: {task_description}

CONTEXT:
- Full project context loaded
- All .moai/ configuration available
- All 55 Skills accessible
- Same tools as Main Alfred
- Same TRUST 5 principles enforced

EXECUTION:
1. Plan your approach
2. Execute with transparency
3. Document decisions via @TAG
4. Create PR if modifications needed
5. Log learnings to clone-memory

SUCCESS CRITERIA:
- TRUST 5 principles maintained
- @TAG chain integrity preserved
- All tests passing
- PR ready for review

You have full autonomy. Main Alfred will review your output only.
"""
    )
    return clone
```

### Rule 3: Clone 학습 저장

```python
# Clone이 작업 완료 후:

def save_learning(task_type: str, learnings: dict):
    """Clone의 학습을 메모리에 저장"""
    memory_file = Path(".moai/memory/clone-learnings.json")

    learnings_db = json.loads(memory_file.read_text())
    learnings_db[task_type].append({
        "timestamp": now(),
        "success": True/False,
        "approach_used": "...",
        "pitfalls_discovered": [...],
        "optimization_tips": [...]
    })

    memory_file.write_text(json.dumps(learnings_db, indent=2))
```

---

## Clone 패턴의 장점

### 1. 컨텍스트 격리 없음
```
Lead-Specialist: 도메인만 전달 → 전체 그림 못 봄
Clone: 전체 컨텍스트 → "왜"를 이해하고 결정
```

### 2. 에이전트 자율성 최대화
```
Lead-Specialist: "이렇게 하세요" (명령)
Clone: "이것을 해결하세요" (목표) → 자율 판단
```

### 3. 병렬 처리로 확장성
```
Lead-Specialist: 순차 실행만 가능
Clone: 여러 Clone 동시 실행 가능
```

### 4. 자가 학습 및 개선
```
Clone이 마이그레이션 하면서:
- 발견한 문제 패턴
- 효과적이었던 방식
- 우회할 함정들

다음 마이그레이션에 활용
```

---

## Clone 패턴 사용 체크리스트

### Clone 생성 전
- [ ] 도메인 특화 필요 없나?
- [ ] 작업이 멀티스텝(5+) 또는 대규모(100+ 파일)?
- [ ] 작업이 독립적이거나 병렬화 가능?
- [ ] Clone에게 충분한 컨텍스트 제공했나?
- [ ] 성공 기준이 명확한가?

### Clone 실행 중
- [ ] Clone이 자율적으로 판단하고 실행?
- [ ] 오류 발생 시 자가 디버깅?
- [ ] 진행 상황이 투명하게 기록?

### Clone 완료 후
- [ ] 학습 내용이 메모리에 저장?
- [ ] PR이 생성되고 검증 가능?
- [ ] 다음 유사 작업의 템플릿 업데이트?

---

## Clone 패턴 vs Lead-Specialist 비교

### Clone 패턴을 선택하는 경우

```
✅ "프로젝트 전체를 마이그레이션하세요"
   → Clone: 전체 컨텍스트로 최적 경로 찾음

✅ "100+ 파일에서 imports 업데이트"
   → Clone: 병렬 처리로 1시간에 완료

✅ "다음 달 아키텍처 개선 계획 탐색"
   → Clone: 불확실성 높은 작업도 자율적 탐색

❌ "React 컴포넌트 디자인"
   → ui-ux-expert (도메인 특화)

❌ "Python 성능 최적화"
   → backend-expert (도메인 특화)
```

---

## FAQ

**Q: Clone이 실수하면?**
A: Clone은 오류에 대응하여 자가 디버깅합니다. 최악의 경우 Main Alfred가 개입하거나 되돌립니다.

**Q: Clone이 얼마나 자율적인가?**
A: 완전 자율적입니다. TRUST 5 원칙과 @TAG 체계만 강제합니다.

**Q: 비용 문제는?**
A: Clone도 토큰 사용합니다. 하지만 병렬 처리로 전체 시간이 단축되어 효율적입니다.

**Q: Lead-Specialist와 함께 사용 가능?**
A: 가능합니다. 예) Clone이 전체 작업 계획 → Specialist가 특정 부분 구현

---

**참조**:
- CLAUDE.md: 🔄 Alfred의 하이브리드 아키텍처
- Skill("moai-alfred-workflow"): 4단계 워크플로우
- Skill("moai-alfred-agent-guide"): 19명 팀 멤버 상세
