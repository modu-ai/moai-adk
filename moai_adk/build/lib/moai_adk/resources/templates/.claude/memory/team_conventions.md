# 팀 협업 규약

> MoAI-ADK 기반 프로젝트의 팀 협업 및 커뮤니케이션 규약

## 👥 팀 구성 및 역할

### 핵심 역할
- **Product Owner**: 제품 비전 및 백로그 관리
- **Tech Lead**: 기술 아키텍처 및 코드 품질 책임
- **Developer**: 기능 개발 및 테스트 구현
- **QA Engineer**: 품질 보증 및 테스트 자동화
- **DevOps Engineer**: 인프라 및 배포 파이프라인 관리

### 역할별 책임
| 역할 | 주요 책임 | MoAI-ADK 관련 업무 |
|------|-----------|-------------------|
| Product Owner | 요구사항 정의, 우선순위 결정 | @REQ 태그 생성, 수용기준 작성 |
| Tech Lead | 아키텍처 설계, 코드 리뷰 | @DESIGN 태그 관리, Constitution 5원칙 준수 (`@.moai/memory/constitution.md`) |
| Developer | 기능 구현, 단위 테스트 | @TASK 구현, TDD 적용 |
| QA Engineer | 테스트 계획, 품질 검증 | @TEST 태그 관리, 품질 게이트 운영 |
| DevOps Engineer | CI/CD, 모니터링 | @DEPLOY 태그 관리, Observability 구현 |

## 🗓️ 개발 프로세스

### 스프린트 계획 (2주 단위)
1. **Sprint Planning** (2시간)
   - 백로그 아이템 선정
   - @REQ 태그 검토 및 우선순위 조정
   - 개발 역량 기반 task 분배

2. **Daily Standup** (15분)
   - 어제 완료한 작업 (@TASK 상태 업데이트)
   - 오늘 진행할 작업
   - 블로커 및 도움 요청

3. **Sprint Review** (1시간)
   - 완성된 기능 데모
   - @FEATURE 태그 기반 결과 리뷰
   - 스테이크홀더 피드백 수집

4. **Retrospective** (1시간)
   - 프로세스 개선점 논의
   - MoAI Constitution 5원칙 준수 현황 점검 (`@.moai/memory/constitution.md`)
   - 다음 스프린트 액션 아이템

### 태스크 관리 워크플로우

```mermaid
flowchart TD
    A[@REQ 요구사항] --> B[@DESIGN 설계]
    B --> C[@TASK 작업분해]
    C --> D[개발자 할당]
    D --> E[@FEATURE 구현]
    E --> F[@TEST 테스트]
    F --> G[@REVIEW 리뷰]
    G --> H{승인?}
    H -->|Yes| I[@DEPLOY 배포]
    H -->|No| E
    I --> J[@MONITOR 모니터링]
```

## 💬 커뮤니케이션 규칙

### 회의 운영
- **정시 시작/종료**: 시간 엄수
- **아젠다 사전 공유**: 24시간 전 배포
- **액션 아이템 기록**: 책임자, 마감일 명시
- **회의록 공유**: 당일 내 팀 전체 공유

### 슬랙/디스코드 채널 구조
```
#general           - 전체 공지사항
#dev-discuss       - 기술 논의
#code-review       - 코드 리뷰 요청
#bug-reports       - 버그 신고 및 추적
#deployment        - 배포 알림
#monitoring        - 시스템 모니터링 알림
#random            - 자유 소통
```

### 메시지 작성 규칙
- **명확한 제목**: 핵심 내용 요약
- **태그 활용**: @REQ, @BUG 등 관련 태그 언급
- **긴급도 표시**: 🔴 긴급, 🟡 보통, 🟢 정보성
- **스레드 활용**: 관련 논의는 스레드로 정리

## 📋 문서화 규칙

### 필수 문서
- **README.md**: 프로젝트 개요 및 시작 가이드
- **API.md**: API 명세 및 예제
- **DEPLOYMENT.md**: 배포 가이드
- **TROUBLESHOOTING.md**: 문제 해결 가이드

### 문서 작성 표준
```markdown
# 문서 제목

## 개요
- 목적과 범위 명시

## 전제조건  
- 필요한 지식이나 도구

## 단계별 가이드
1. 첫 번째 단계
2. 두 번째 단계

## 트러블슈팅
- 자주 발생하는 문제와 해결책

## 참고자료
- 관련 링크 및 문서
```

### 코드 문서화
- **인라인 주석**: 복잡한 로직 설명
- **Docstring**: 함수/클래스 설명
- **ADR**: 아키텍처 결정 기록 작성

## 🔄 코드 리뷰 프로세스

### Pull Request 규칙
1. **브랜치 명명**: `feature/MOAI-123-add-user-auth`
2. **PR 제목**: `[MOAI-123] Add user authentication`
3. **설명 템플릿**:
```markdown
## 변경사항
- 주요 변경내용 요약

## 관련 태그
- @FEATURE:AUTH-001
- @TEST:AUTH-001

## 테스트 계획
- [ ] 단위 테스트 작성
- [ ] 통합 테스트 실행
- [ ] 수동 테스트 시나리오

## 체크리스트
- [ ] 코딩 표준 준수
- [ ] 테스트 커버리지 80% 이상
- [ ] 문서 업데이트
```

### 리뷰어 책임
- **24시간 내 리뷰**: 업무일 기준
- **건설적 피드백**: 개선 방향 제시
- **코드 품질 검증**: 표준 준수 확인
- **보안 검토**: 취약점 점검

### 리뷰 우선순위
1. 🔴 **긴급**: 핫픽스, 보안 이슈
2. 🟡 **높음**: 핵심 기능, API 변경
3. 🟢 **보통**: 일반 기능, 리팩토링

## 🚀 배포 및 릴리스

### 브랜치 전략 (Git Flow)
```
main (production)
├── release/v1.2.0
├── develop (integration)
│   ├── feature/user-auth
│   ├── feature/payment
│   └── feature/notifications
└── hotfix/critical-bug-fix
```

### 릴리스 프로세스
1. **Feature Freeze**: 기능 개발 중단
2. **Release Branch**: develop에서 분기
3. **QA Testing**: 전체 시나리오 테스트
4. **Production Deploy**: main 브랜치 병합
5. **Post-Deploy Monitoring**: 24시간 모니터링

### 롤백 절차
```bash
# 1. 이전 버전으로 롤백
kubectl rollout undo deployment/app --to-revision=2

# 2. 상태 확인
kubectl rollout status deployment/app

# 3. 로그 확인
kubectl logs -f deployment/app
```

## 📊 품질 메트릭

### 팀 성과 지표
| 메트릭 | 목표값 | 측정 방법 |
|--------|--------|----------|
| 코드 리뷰 시간 | < 24시간 | GitHub PR metrics |
| 테스트 커버리지 | > 80% | Coverage 도구 |
| 버그 발생률 | < 5% | Issue tracking |
| 배포 성공률 | > 95% | CI/CD 로그 |

### 개인 성과 평가
- **코드 품질**: 리뷰 패스율, 버그 발생률
- **협업**: 리뷰 참여도, 도움 제공 횟수
- **학습**: 새 기술 도입, 지식 공유
- **리더십**: 멘토링, 프로세스 개선

## 🎯 분쟁 해결

### 기술적 의견 충돌
1. **데이터 수집**: 각 접근법의 장단점 정리
2. **POC 실시**: 가능한 경우 프로토타입 구현
3. **팀 투표**: 민주적 의사결정
4. **Tech Lead 결정**: 최종 기술 판단

### 우선순위 갈등
1. **Product Owner 확인**: 비즈니스 우선순위 재확인
2. **임팩트 분석**: 각 옵션의 영향도 평가
3. **이해관계자 논의**: 관련자 의견 수렴
4. **합의 도출**: Win-Win 솔루션 모색

## 🔒 보안 및 컴플라이언스

### 정보 보안
- **코드 리포지토리**: 2FA 필수 설정
- **API 키 관리**: 환경변수로 분리
- **접근 권한**: 최소 권한 원칙
- **보안 스캔**: 주간 의존성 취약점 점검

### 개인정보 처리
- **데이터 최소화**: 필요한 정보만 수집
- **암호화**: 민감 데이터 암호화 저장
- **접근 로그**: 개인정보 접근 기록
- **삭제 정책**: 보관 기간 준수

---

**마지막 업데이트**: 2025-09-15  
**버전**: v0.1.12  
**승인자**: Tech Lead, Product Owner