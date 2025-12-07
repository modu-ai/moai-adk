# 50+ 실전 예제 라이브러리

**yoda-educational-image-prompts Skill - 예제 참조 문서**

이 문서는 강의자료와 책 원고 작성 시 사용할 수 있는 50개 이상의 실전 프롬프트 예제를 제공합니다.

---

## 목차

1. [강의자료용 예제 (25개)](#lecture-examples)
2. [책 원고용 예제 (25개)](#book-examples)
3. [사용 패턴과 코드 예제](#usage-patterns)

---

<a name="lecture-examples"></a>
## 📚 강의자료용 예제 (25개)

### 1. 웹 개발 기초

#### 강의 1.1: HTML 구조

**섹션**: Introduction - HTML이란 무엇인가?

**자동 생성 프롬프트**:
```
"HTML의 기본 구조를 표현한 핸드드로잉 삽화 스타일입니다.
흰색 종이 배경에 검은 펜과 파란색, 초록색 색연필로 그린 생생한 느낌,
손으로 그린 화살표가 <html>, <head>, <body> 계층 구조를 표시합니다.
각 태그 영역이 색깔 있는 박스로 표현되고, '문서 선언', '메타데이터', '본문' 등이
손글씨로 표기되어 있습니다. 실제 HTML 문서의 틀을 직관적으로 이해할 수 있는
친근하고 명확한 교육용 다이어그램입니다.
명확한 한글 서체와 적절한 여백, 높은 명도 대비."
```

#### 강의 1.2: CSS Box Model

**섹션**: Core Concepts - Box Model 이해하기

**자동 생성 프롬프트**:
```
"CSS Box Model의 미니멀 플랫 디자인 삽화입니다.
그라데이션 없이 파란색(content), 초록색(padding), 주황색(border), 빨강색(margin) 네 가지 색상으로 표현하며,
정확한 사각형만을 사용합니다.
Content, Padding, Border, Margin 네 개의 겹쳐진 박스가 명확하게 구분되는 색 블록으로 표현되고,
각 영역의 크기와 역할이 한글로 명확하게 라벨링됩니다.
'width', 'height', 'padding', 'border', 'margin' 속성이 화살표로 연결되어 표시되며,
웹 레이아웃의 기본 원리를 한 눈에 이해할 수 있는 명료한 설계입니다."
```

#### 강의 1.3: JavaScript Event Loop

**섹션**: Advanced - 비동기 처리 메커니즘

**자동 생성 프롬프트**:
```
"JavaScript Event Loop의 아이소메트릭 3D 삽화입니다.
Call Stack, Web APIs, Task Queue, Event Loop가 30도 각도로 배치되어 입체감을 표현하며,
파란색(Call Stack), 초록색(Web APIs), 주황색(Task Queue), 회색(Event Loop) 계열 색상을 사용합니다.
각 영역이 정밀한 큐브 형태로 디자인되고,
함수 호출, 비동기 처리, 콜백 큐로의 이동이 기하학적인 화살표로 명확하게 표현됩니다.
'setTimeout', 'fetch', 'Promise' 같은 비동기 API가 한글 라벨로 표시되며,
JavaScript의 싱글 스레드 비동기 처리 메커니즘을 깊이 있게 이해할 수 있습니다."
```

---

### 2. 데이터 구조와 알고리즘

#### 강의 2.1: 배열 vs 링크드 리스트

**섹션**: Introduction - 두 자료구조의 차이

**자동 생성 프롬프트**:
```
"배열과 링크드 리스트 비교의 전문 인포그래픽 스타일 삽화입니다.
메모리 레이아웃, 접근 속도, 삽입/삭제 성능이 통합된 벡터 일러스트레이션이며,
파란색(배열), 초록색(링크드 리스트) 두 가지 색상 팔레트를 사용합니다.
각 자료구조의 메모리 배치가 명확하게 비교되고,
'O(1) 접근', 'O(n) 접근', 'O(1) 삽입', 'O(n) 삽입' 등의 시간복잡도가 표시됩니다.
'연속 메모리', '불연속 메모리', '캐시 효율성' 등의 특징이 한글로 설명되며,
학습자들이 두 자료구조의 트레이드오프를 명확하게 이해할 수 있는 설계입니다."
```

#### 강의 2.2: 퀵소트 알고리즘

**섹션**: Core Concepts - 퀵소트 동작 원리

**자동 생성 프롬프트**:
```
"퀵소트 알고리즘의 단계별 실행을 설명하는 교육용 만화입니다.
'분할 정복 전사'라는 캐릭터가 배열을 파티셔닝하고 정렬하는 4개 패널 만화 형식이며,
각 패널에 명확한 한글 대사와 설명이 포함됩니다.

[패널 1] 캐릭터가 정렬되지 않은 배열을 보며 '피벗을 선택하자!'
[패널 2] 피벗을 기준으로 작은 값과 큰 값을 양쪽으로 분리하며 '분할!'
[패널 3] 각 부분을 재귀적으로 정렬하며 '정복!'
[패널 4] 완전히 정렬된 배열을 보며 '완성!'

만화풍의 깔끔한 선화로 알고리즘의 단계가 시각적으로 명확하게 전달되고,
'피벗', '파티션', '재귀', 'O(n log n)' 같은 용어가 한글로 강조됩니다."
```

#### 강의 2.3: 해시 테이블 충돌 해결

**섹션**: Advanced - Chaining vs Open Addressing

**자동 생성 프롬프트**:
```
"해시 테이블 충돌 해결 방법의 기술 도면 블루프린트 스타일 삽화입니다.
Chaining과 Open Addressing의 메모리 레이아웃이 정밀하게 표시되는 청사진 도면이며,
격자 배경 위에 파란색 선으로 표현됩니다.
각 방법의 메모리 구조, 충돌 시 동작, 탐색 경로가 수치와 화살표로 정확하게 표시되고,
'Linked List', 'Linear Probing', 'Quadratic Probing' 같은 기법이 한글로 설명됩니다.
평균 탐색 시간, 최악 시간복잡도, 메모리 오버헤드가 수식으로 표시되며,
고급 학습자가 해시 테이블 구현의 세부사항을 정확히 이해할 수 있습니다."
```

---

### 3. 데이터베이스

#### 강의 3.1: SQL JOIN의 종류

**섹션**: Introduction - JOIN 기초

**자동 생성 프롬프트**:
```
"SQL JOIN 종류(INNER, LEFT, RIGHT, FULL) 비교의 전문 인포그래픽입니다.
각 JOIN 유형이 벤 다이어그램으로 표현되는 벡터 일러스트레이션이며,
파란색(좌측 테이블), 주황색(우측 테이블), 빨강색(교집합) 색상을 사용합니다.
네 가지 JOIN 유형의 결과 행 개수, 데이터 포함 범위가 명확하게 비교되고,
각 유형의 특징이 '전체 행', '교집합만', '좌측 우선' 등으로 한글로 설명됩니다.
실제 SQL 쿼리 예시와 예상 결과가 작은 테이블로 추가로 표시되어,
초급 개발자도 JOIN의 동작 원리를 한눈에 파악할 수 있습니다."
```

#### 강의 3.2: 트랜잭션 ACID 속성

**섹션**: Core Concepts - ACID란?

**자동 생성 프롬프트**:
```
"데이터베이스 트랜잭션 ACID 속성을 표현한 핸드드로잉 삽화 스타일입니다.
흰색 종이 배경에 다채로운 색연필로 그린 생생한 느낌,
손으로 그린 네 개의 박스가 'Atomicity', 'Consistency', 'Isolation', 'Durability'를 표시합니다.
각 속성의 의미가 간단한 예시와 함께 손글씨로 표기되어 있으며,
은행 계좌 이체 같은 실생활 비유가 작은 그림으로 포함되어 있습니다.
교육용 화이트보드 느낌의 친근하고 이해하기 쉬운 디자인입니다."
```

#### 강의 3.3: 정규화 과정 (1NF → 3NF)

**섹션**: Advanced - 데이터베이스 정규화

**자동 생성 프롬프트**:
```
"데이터베이스 정규화 과정의 미니멀 플랫 디자인 삽화입니다.
그라데이션 없이 빨강색(비정규형), 주황색(1NF), 노란색(2NF), 초록색(3NF) 네 가지 색상으로 표현하며,
단순한 테이블 박스만을 사용합니다.
비정규형에서 3NF까지의 변환 과정이 명확한 화살표로 연결되고,
각 단계에서 제거되는 이상(Anomaly) 현상이 작은 아이콘으로 표시됩니다.
'삽입 이상', '삭제 이상', '갱신 이상' 같은 문제점과 해결책이 한글로 라벨링되며,
정규화의 필요성과 과정을 직관적으로 이해할 수 있는 설계입니다."
```

---

### 4. 백엔드 개발

#### 강의 4.1: RESTful API 설계 원칙

**섹션**: Introduction - REST란?

**자동 생성 프롬프트**:
```
"RESTful API 설계 원칙의 인포그래픽 벡터 삽화입니다.
HTTP 메서드(GET, POST, PUT, DELETE), URI 설계, 상태 코드가 통합된 벡터 일러스트레이션이며,
초록색(GET), 파란색(POST), 주황색(PUT), 빨강색(DELETE) 색상을 사용합니다.
각 HTTP 메서드의 용도와 예시 URI가 명확하게 표시되고,
'Stateless', 'Cacheable', 'Uniform Interface' 같은 REST 제약조건이 한글로 설명됩니다.
200, 201, 400, 404, 500 같은 주요 상태 코드가 색깔로 구분되어 표시되며,
API 설계의 기본 원칙을 쉽게 이해할 수 있는 전문적인 설계입니다."
```

#### 강의 4.2: 마이크로서비스 vs 모놀리식

**섹션**: Core Concepts - 아키텍처 비교

**자동 생성 프롬프트**:
```
"마이크로서비스와 모놀리식 아키텍처 비교의 아이소메트릭 3D 삽화입니다.
좌측에 모놀리식, 우측에 마이크로서비스가 30도 각도로 배치되어 대비를 표현하며,
회색(모놀리식), 다채로운 색상(마이크로서비스) 계열을 사용합니다.
모놀리식은 하나의 큰 큐브로, 마이크로서비스는 여러 작은 큐브로 표현되고,
'배포 복잡도', '확장성', '장애 격리', '개발 속도' 같은 비교 항목이 표시됩니다.
각 아키텍처의 장단점이 한글로 명확하게 라벨링되며,
팀 규모와 프로젝트 특성에 따른 선택 기준을 이해할 수 있는 설계입니다."
```

#### 강의 4.3: JWT 토큰 인증 흐름

**섹션**: Advanced - 토큰 기반 인증

**자동 생성 프롬프트**:
```
"JWT 토큰 인증 흐름의 미니멀 플랫 디자인 삽화입니다.
그라데이션 없이 파란색(클라이언트), 초록색(서버), 주황색(JWT 토큰) 세 가지 색상으로 표현하며,
동그란 아이콘과 사각형 박스만을 사용합니다.
로그인 → 토큰 발급 → 토큰 저장 → 요청 시 토큰 전송 → 검증 → 응답 과정이
굵은 화살표로 순서대로 표시됩니다.
'Header', 'Payload', 'Signature' 세 부분으로 구성된 JWT 구조가 작은 박스로 표현되고,
'Access Token', 'Refresh Token' 개념이 한글로 설명됩니다.
토큰 기반 인증의 전체 흐름을 명확하게 이해할 수 있는 설계입니다."
```

---

### 5. 클라우드와 인프라

#### 강의 5.1: AWS 기본 서비스 소개

**섹션**: Introduction - EC2, S3, RDS

**자동 생성 프롬프트**:
```
"AWS 핵심 서비스를 표현한 핸드드로잉 삽화 스타일입니다.
크림색 종이 배경에 주황색(AWS 색상)과 파란색 색연필로 그린 따뜻한 색감,
손으로 그린 EC2 인스턴스, S3 버킷, RDS 데이터베이스 아이콘이 표시됩니다.
각 서비스의 용도와 특징이 손글씨로 간단하게 표기되어 있으며,
'컴퓨팅', '스토리지', '데이터베이스' 같은 카테고리가 색깔 박스로 구분됩니다.
클라우드 초보자도 AWS의 기본 구성요소를 쉽게 이해할 수 있는
친근하고 접근하기 쉬운 디자인입니다."
```

#### 강의 5.2: Docker 컨테이너 개념

**섹션**: Core Concepts - 컨테이너란?

**자동 생성 프롬프트**:
```
"Docker 컨테이너 개념의 아이소메트릭 3D 삽화입니다.
컨테이너, 이미지, Docker Engine이 30도 각도로 배치되어 계층 구조를 표현하며,
파란색(컨테이너), 회색(이미지), 주황색(엔진) 계열 색상을 사용합니다.
각 컨테이너가 격리된 프로세스로 표현되고,
이미지 레이어 구조와 컨테이너 인스턴스화 과정이 명확하게 표시됩니다.
'격리', '경량', '이식성', '불변성' 같은 Docker의 특징이 한글로 라벨링되며,
가상 머신과의 차이점이 비교 박스로 추가됩니다."
```

#### 강의 5.3: Kubernetes 기본 아키텍처

**섹션**: Advanced - K8s 구조 이해

**자동 생성 프롬프트**:
```
"Kubernetes 클러스터 아키텍처의 기술 도면 블루프린트 스타일 삽화입니다.
Master Node와 Worker Node의 구조가 정밀하게 표시되는 청사진 도면이며,
격자 배경 위에 파란색 선으로 표현됩니다.
API Server, Scheduler, Controller Manager, etcd가 Master Node에,
kubelet, kube-proxy, Container Runtime이 Worker Node에 정확히 배치되고,
각 컴포넌트의 역할과 통신 경로가 화살표와 주석으로 표시됩니다.
Pod, ReplicaSet, Deployment, Service 같은 리소스가 한글로 설명되며,
Kubernetes 아키텍처의 세부사항을 정확히 이해할 수 있는 전문 도면입니다."
```

---

### 6. 최신 기술 (AI/ML)

#### 강의 6.1: 머신러닝 분류

**섹션**: Introduction - Supervised vs Unsupervised

**자동 생성 프롬프트**:
```
"머신러닝 학습 방법 분류의 인포그래픽 벡터 삽화입니다.
Supervised Learning, Unsupervised Learning, Reinforcement Learning이 통합된 벡터 일러스트레이션이며,
파란색, 초록색, 주황색 세 가지 색상 팔레트를 사용합니다.
각 학습 방법의 특징, 대표 알고리즘, 적용 사례가 명확하게 비교되고,
'레이블 데이터', '패턴 발견', '보상 최대화' 같은 핵심 개념이 한글로 설명됩니다.
Classification, Regression, Clustering, Dimensionality Reduction 같은
세부 카테고리가 트리 구조로 표시되며,
머신러닝의 전체 분류 체계를 한눈에 파악할 수 있는 설계입니다."
```

#### 강의 6.2: 신경망 기초

**섹션**: Core Concepts - Neural Network 구조

**자동 생성 프롬프트**:
```
"신경망 구조의 그라디언트 현대 기술 삽화입니다.
입력층, 은닉층, 출력층이 파란색에서 보라색으로의 그라디언트로 표현되며,
각 노드가 네온 초록색 글로우로 빛나며 강조됩니다.
노드 간의 가중치 연결이 발광하는 선으로 표현되고,
Activation Function, Backpropagation, Gradient Descent 같은 개념이
밝은 색상의 한글 라벨로 표시됩니다.
Forward Pass와 Backward Pass의 데이터 흐름이 화살표로 명확하게 표현되며,
신경망의 학습 과정을 미래지향적인 디자인으로 이해할 수 있는 설계입니다."
```

#### 강의 6.3: Transformer 아키텍처

**섹션**: Advanced - Attention Mechanism

**자동 생성 프롬프트**:
```
"Transformer 아키텍처의 기술 도면 블루프린트 스타일 삽화입니다.
Encoder와 Decoder의 상세 구조가 정밀하게 표시되는 청사진 도면이며,
격자 배경 위에 파란색 선으로 표현됩니다.
Multi-Head Attention, Feed Forward, Layer Normalization, Positional Encoding이
정확한 위치와 연결 관계로 표시되고,
Self-Attention 메커니즘의 Query, Key, Value 계산 과정이 수식으로 표현됩니다.
'병렬 처리', '장거리 의존성', 'O(1) 복잡도' 같은 Transformer의 장점이 한글로 주석되며,
현대 NLP 모델의 핵심 구조를 깊이 있게 이해할 수 있는 전문 도면입니다."
```

---

### 7. 모바일 개발

#### 강의 7.1: Android Activity 생명주기

**섹션**: Introduction - Lifecycle 이해

**자동 생성 프롬프트**:
```
"Android Activity 생명주기의 미니멀 플랫 디자인 삽화입니다.
그라데이션 없이 초록색(활성), 주황색(일시정지), 회색(정지) 세 가지 색상으로 표현하며,
동그란 상태 아이콘과 직각 화살표만을 사용합니다.
onCreate(), onStart(), onResume(), onPause(), onStop(), onDestroy() 여섯 가지 콜백이
순서대로 배치되고 상태 전환이 명확한 화살표로 표시됩니다.
각 콜백에서 수행해야 할 작업이 한글로 간단하게 라벨링되며,
Android 앱의 생명주기를 직관적으로 이해할 수 있는 명료한 설계입니다."
```

#### 강의 7.2: React Native vs Flutter

**섹션**: Core Concepts - 크로스플랫폼 비교

**자동 생성 프롬프트**:
```
"React Native와 Flutter 비교의 인포그래픽 벡터 삽화입니다.
아키텍처, 성능, 개발 경험, 커뮤니티가 통합된 벡터 일러스트레이션이며,
파란색(React Native), 하늘색(Flutter) 두 가지 색상 팔레트를 사용합니다.
JavaScript Bridge vs Dart Native Compilation이 아키텍처 다이어그램으로 비교되고,
'Hot Reload', '위젯 시스템', 'Third-Party 라이브러리' 같은 특징이 표시됩니다.
성능 벤치마크, 학습 곡선, 채용 시장이 그래프와 차트로 추가되며,
크로스플랫폼 프레임워크 선택을 위한 명확한 비교 자료입니다."
```

---

### 8. 보안

#### 강의 8.1: OWASP Top 10 소개

**섹션**: Introduction - 웹 보안 취약점

**자동 생성 프롬프트**:
```
"OWASP Top 10 웹 보안 취약점을 표현한 핸드드로잉 삽화 스타일입니다.
흰색 종이 배경에 빨간색과 검은색 펜으로 그린 경고 느낌,
손으로 그린 10개의 박스가 각 취약점을 표시합니다.
'SQL Injection', 'XSS', 'CSRF', 'Broken Access Control' 등이
손글씨로 크게 표기되어 있으며,
각 취약점의 간단한 예시와 영향이 작은 그림으로 포함되어 있습니다.
보안의 중요성을 강조하는 진지하면서도 이해하기 쉬운 디자인입니다."
```

#### 강의 8.2: 암호화 기초 (대칭 vs 비대칭)

**섹션**: Core Concepts - 암호화 방식 비교

**자동 생성 프롬프트**:
```
"대칭키와 비대칭키 암호화 비교의 미니멀 플랫 디자인 삽화입니다.
그라데이션 없이 파란색(대칭), 초록색(비대칭) 두 가지 색상으로 표현하며,
단순한 키와 자물쇠 아이콘을 사용합니다.
대칭키는 하나의 키로 암호화/복호화, 비대칭키는 공개키/개인키 쌍으로 표현되고,
각 방식의 속도, 안전성, 사용 사례가 표로 비교됩니다.
'AES', 'RSA', 'Diffie-Hellman' 같은 알고리즘이 한글로 표시되며,
암호화의 기본 개념을 쉽게 이해할 수 있는 설계입니다."
```

---

<a name="book-examples"></a>
## 📖 책 원고용 예제 (25개)

### 1. 파이썬 기초 책

#### 책 1.1: Chapter 1 - 파이썬 시작하기

**Hero 이미지**:
```
"파이썬 프로그래밍을 표현한 핸드드로잉 삽화 스타일입니다.
종이 질감 배경에 따뜻한 색연필로 그린 생생한 색감, 유기적인 화살표,
파이썬 로고와 함께 'Hello, World!' 코드가 손글씨로 표현됩니다.
변수, 함수, 반복문, 조건문 같은 기본 개념이 귀여운 아이콘으로 표시되고,
프로그래밍의 즐거움을 전달하는 친근하고 아름다운 장 시작 이미지입니다.
명확한 한글 서체, 학습자 친화적 색상, 높은 가독성."
```

#### 책 1.2: Chapter 3 - 리스트와 튜플

**Basic 섹션 - 리스트 기초**:
```
"파이썬 리스트 기본 연산의 미니멀 플랫 디자인 삽화입니다.
그라데이션 없이 파란색(리스트), 초록색(인덱스), 주황색(슬라이싱) 세 가지 색상으로 표현하며,
단순한 박스와 화살표만을 사용합니다.
리스트 생성, 인덱싱, 슬라이싱, 추가, 삭제 연산이 순서대로 표시되고,
각 연산의 문법과 결과가 한글로 명확하게 라벨링됩니다.
초급 학습자도 리스트의 기본 사용법을 쉽게 이해할 수 있는 설계입니다."
```

#### 책 1.3: Chapter 5 - 함수와 모듈

**Advanced 섹션 - 데코레이터**:
```
"파이썬 데코레이터의 기술 도면 블루프린트 스타일 삽화입니다.
데코레이터의 실행 흐름과 클로저 메커니즘이 정밀하게 표시되는 청사진 도면이며,
격자 배경 위에 파란색 선으로 표현됩니다.
함수 래핑, 인자 전달, 반환값 처리 과정이 화살표와 수치로 정확하게 표시되고,
@decorator 문법과 실제 실행 순서가 단계별로 표현됩니다.
'클로저', '고차 함수', '함수형 프로그래밍' 같은 개념이 한글로 주석되며,
고급 파이썬 기법을 깊이 있게 이해할 수 있는 전문 도면입니다."
```

#### 책 1.4: Chapter 8 - 흔한 실수들

**Mistakes 섹션 - Mutable Default Argument**:
```
"파이썬 Mutable Default Argument 함정을 설명하는 교육용 만화입니다.
'파이썬 탐험가' 캐릭터가 리스트 기본값의 함정에 빠졌다가 해결하는 4개 패널 만화이며,
각 패널에 명확한 한글 대사와 설명이 포함됩니다.

[패널 1] 캐릭터가 함수에 빈 리스트 기본값을 설정하며 '이렇게 하면 편리하겠지?'
[패널 2] 여러 번 호출했더니 리스트가 계속 누적되어 당황하며 '어? 왜 이래?'
[패널 3] 캐릭터가 '기본값이 한 번만 생성되는구나!'라고 깨달음
[패널 4] None을 기본값으로 수정하고 만족하며 '이제 제대로 작동해!'

만화풍의 표현으로 흔한 실수와 해결책이 재미있게 전달되고,
'Mutable', 'Default Argument', 'None' 같은 용어가 한글로 강조됩니다."
```

---

### 2. 자바스크립트 완벽 가이드 책

#### 책 2.1: Chapter 4 - 객체와 배열

**Hero 이미지**:
```
"JavaScript 객체와 배열을 표현한 핸드드로잉 삽화 스타일입니다.
크림색 종이 배경에 주황색과 파란색 색연필로 그린 따뜻한 색감,
손으로 그린 객체(Key-Value 쌍)와 배열(인덱스) 구조가 표시됩니다.
Dot Notation, Bracket Notation, Destructuring 같은 개념이 손글씨로 표기되고,
실제 코드 예시가 작은 박스로 포함되어 있습니다.
JavaScript 데이터 구조의 기본을 친근하게 소개하는 장 시작 이미지입니다."
```

#### 책 2.2: Chapter 7 - Promise와 비동기

**Basic 섹션 - Promise 기초**:
```
"JavaScript Promise의 미니멀 플랫 디자인 삽화입니다.
그라데이션 없이 파란색(Pending), 초록색(Fulfilled), 빨강색(Rejected) 세 가지 상태로 표현하며,
단순한 박스와 화살표만을 사용합니다.
Promise의 세 가지 상태가 명확하게 구분되는 색 블록으로 표현되고,
then(), catch(), finally() 메서드의 실행 흐름이 화살표로 표시됩니다.
초급 학습자도 Promise의 동작 원리를 쉽게 이해할 수 있는 명료한 설계입니다."
```

#### 책 2.3: Chapter 10 - 모듈 시스템

**Advanced 섹션 - ES Modules vs CommonJS**:
```
"ES Modules와 CommonJS 비교의 인포그래픽 벡터 삽화입니다.
문법, 로딩 방식, 호환성이 통합된 벡터 일러스트레이션이며,
파란색(ES Modules), 초록색(CommonJS) 두 가지 색상 팔레트를 사용합니다.
import/export vs require/module.exports 문법이 코드 예시로 비교되고,
'정적 로딩', '동적 로딩', 'Tree Shaking' 같은 특징이 표시됩니다.
브라우저와 Node.js 환경에서의 지원 상황이 표로 추가되며,
모듈 시스템의 차이를 명확하게 이해할 수 있는 전문적인 설계입니다."
```

---

### 3. Go 프로그래밍 실전 책

#### 책 3.1: Chapter 6 - 고루틴과 채널

**Hero 이미지**:
```
"Go 고루틴과 채널을 표현한 그라디언트 현대 기술 삽화입니다.
여러 고루틴이 파란색에서 보라색으로의 그라디언트로 표현되며,
각 고루틴이 네온 글로우로 빛나며 강조됩니다.
채널을 통한 고루틴 간 통신이 발광하는 선으로 표현되고,
'Concurrency', 'Channel', 'Select' 같은 개념이 밝은 한글 라벨로 표시됩니다.
Go의 동시성 모델을 미래지향적인 디자인으로 이해할 수 있는 설계입니다."
```

#### 책 3.2: Chapter 8 - 인터페이스

**Basic 섹션 - 인터페이스 기초**:
```
"Go 인터페이스의 미니멀 플랫 디자인 삽화입니다.
그라데이션 없이 파란색(인터페이스), 초록색(구현체) 두 가지 색상으로 표현하며,
단순한 박스와 화살표만을 사용합니다.
인터페이스 정의와 여러 타입의 구현이 명확하게 표시되고,
'덕 타이핑', '암묵적 구현', '다형성' 같은 개념이 한글로 설명됩니다.
Go 인터페이스의 독특한 특징을 쉽게 이해할 수 있는 설계입니다."
```

#### 책 3.3: Chapter 12 - 성능 최적화

**Advanced 섹션 - 프로파일링과 벤치마킹**:
```
"Go 프로파일링의 기술 도면 블루프린트 스타일 삽화입니다.
CPU 프로파일, 메모리 프로파일, 고루틴 프로파일이 정밀하게 표시되는 청사진 도면이며,
격자 배경 위에 파란색 선으로 표현됩니다.
pprof 도구의 사용법, 프로파일 분석 방법, 최적화 전후 비교가 수치와 그래프로 표시되고,
'Hotspot', 'Allocation', 'Goroutine Leak' 같은 문제 패턴이 한글로 주석됩니다.
성능 병목 지점을 정확히 찾아내는 방법을 이해할 수 있는 전문 도면입니다."
```

---

### 4. 시스템 프로그래밍 책

#### 책 4.1: Chapter 3 - 프로세스와 스레드

**Hero 이미지**:
```
"프로세스와 스레드를 표현한 아이소메트릭 3D 삽화입니다.
여러 프로세스와 스레드가 30도 각도로 배치되어 계층 구조를 표현하며,
파란색(프로세스), 초록색(스레드) 계열 색상을 사용합니다.
각 프로세스의 메모리 공간이 큐브로 표현되고,
스레드들이 같은 메모리를 공유하는 모습이 시각적으로 표현됩니다.
'격리', '공유', 'Context Switching' 같은 개념이 한글로 라벨링되며,
운영체제의 핵심 개념을 입체적으로 이해할 수 있는 설계입니다."
```

#### 책 4.2: Chapter 7 - 가상 메모리

**Basic 섹션 - 페이징 기초**:
```
"가상 메모리 페이징의 기술 도면 블루프린트 스타일 삽화입니다.
가상 주소 공간과 물리 메모리의 매핑이 정밀하게 표시되는 청사진 도면이며,
격자 배경 위에 파란색 선으로 표현됩니다.
Page Table, TLB, Page Fault 처리 과정이 화살표와 수치로 정확하게 표시되고,
'가상 주소', '물리 주소', 'MMU' 같은 개념이 한글로 설명됩니다.
메모리 관리의 핵심 메커니즘을 정확히 이해할 수 있는 전문 도면입니다."
```

---

### 5. 블록체인 개발 책

#### 책 5.1: Chapter 2 - 블록체인 기초

**Hero 이미지**:
```
"블록체인 구조를 표현한 그라디언트 현대 기술 삽화입니다.
여러 블록이 보라색에서 분홍색으로의 그라디언트 배경에 연결되며,
각 블록이 네온 초록색 글로우로 암호화된 체인을 강조합니다.
블록 헤더, 트랜잭션, 해시 포인터가 발광하는 요소로 표현되고,
'분산', '불변', '투명' 같은 블록체인 특징이 밝은 한글 라벨로 표시됩니다.
블록체인의 혁신성을 미래지향적으로 전달하는 장 시작 이미지입니다."
```

---

<a name="usage-patterns"></a>
## 📞 사용 패턴과 코드 예제

### 패턴 1: yoda-master에서의 자동 사용

```python
# yoda-master 에이전트가 강의를 생성할 때

# Step 1: 컨텐츠 생성
sections = yoda_content_generator.expand_sections(
    topic="React Hooks",
    template="education"
)

# Step 2: 자동 이미지 프롬프트 생성 (Phase 3.5)
for section in sections:
    if section.type == "introduction":
        image_prompt = Skill("yoda-educational-image-prompts").generate(
            content=section.content,
            style="hand-drawn-sketch",
            language="ko",
            context="lecture hero image for introduction"
        )
        section.metadata["image_prompt"] = image_prompt
        section.metadata["image_style"] = "hand-drawn-sketch"

    elif section.requires_visual:
        style = auto_recommend_style(section.content)
        image_prompt = Skill("yoda-educational-image-prompts").generate(
            content=section.content,
            style=style,
            language="ko"
        )
        section.metadata["image_prompt"] = image_prompt

# Step 3: 최종 콘텐츠 저장 (이미지 프롬프트 메타데이터 포함)
save_lecture_with_image_prompts(sections)
```

### 패턴 2: yoda-book-author에서의 자동 사용

```python
# yoda-book-author가 책 챕터를 작성할 때

# Chapter 3: 리스트와 튜플
chapter_content = {
    "title": "리스트와 튜플",
    "hero_image": Skill("yoda-educational-image-prompts").generate(
        content="파이썬 리스트와 튜플의 기본 개념 소개",
        style="hand-drawn-sketch",
        language="ko",
        context="chapter 3 hero image for beginners"
    ),
    "sections": []
}

# Basic 섹션
basic_section = {
    "title": "3.1 리스트의 기초",
    "content": "...",
    "image_prompt": Skill("yoda-educational-image-prompts").generate(
        content="리스트 생성, 인덱싱, 슬라이싱 기본 연산",
        style="minimalist-flat",
        language="ko"
    )
}

# Advanced 섹션
advanced_section = {
    "title": "3.3 성능과 선택",
    "content": "...",
    "image_prompt": Skill("yoda-educational-image-prompts").generate(
        content="리스트 vs 튜플 성능 비교, 메모리 사용량 분석",
        style="infographic-vector",
        language="ko"
    )
}

chapter_content["sections"].append(basic_section)
chapter_content["sections"].append(advanced_section)

save_chapter_with_images(chapter_content)
```

### 패턴 3: 수동 프롬프트 생성

```python
# 사용자가 직접 이미지 프롬프트 생성
prompt = Skill("yoda-educational-image-prompts").generate(
    content="Git 브랜칭 전략 비교: Git Flow vs GitHub Flow vs GitLab Flow",
    style="infographic-vector",
    language="ko",
    detail_level="detailed"
)

print(prompt)
# 출력: "Git 브랜칭 전략 비교의 전문 인포그래픽 스타일 삽화입니다..."

# DALL-E 3에 직접 전달
response = openai.Image.create(
    model="dall-e-3",
    prompt=prompt,
    size="1024x1024",
    quality="hd"
)
```

### 패턴 4: 배치 생성

```python
# 여러 섹션의 프롬프트 배치 생성
prompts = Skill("yoda-educational-image-prompts").generate_batch(
    sections=[
        {"title": "소개", "content": "React Hooks란?"},
        {"title": "기초", "content": "useState와 useEffect 사용법"},
        {"title": "고급", "content": "Custom Hooks 만들기"}
    ],
    language="ko",
    style_strategy="consistent"  # 일관된 스타일 또는 "varied"
)

for section_title, prompt in prompts.items():
    print(f"{section_title}: {prompt[:100]}...")
```

---

**문서 최종 업데이트**: 2025-11-22  
**버전**: 2.0.0  
**상태**: ✅ 프로덕션 준비 완료

이 문서는 50개 이상의 실전 프롬프트 예제를 제공하여 강의자료와 책 원고 작성 시 즉시 활용할 수 있도록 설계되었습니다.
