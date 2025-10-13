# SPEC-CLI-001 구현 계획

## 우선순위별 마일스톤

### 1차 목표: CLI 프레임워크 구조
- [ ] moai_adk/cli/ 디렉토리 생성
- [ ] __main__.py 진입점 작성
- [ ] cli/main.py 그룹 명령어 작성
- [ ] ASCII 로고 추가

### 2차 목표: 4개 핵심 명령어 구현
- [ ] init 명령어 (프로젝트 초기화)
- [ ] doctor 명령어 (시스템 진단)
- [ ] status 명령어 (상태 조회)
- [ ] restore 명령어 (백업 복원)

### 3차 목표: Rich 출력 개선
- [ ] 성공/실패 아이콘 추가
- [ ] 색상 스타일 적용
- [ ] 진행 상황 표시 (Progress Bar)

### 최종 목표: 통합 테스트
- [ ] 각 명령어 실행 검증
- [ ] 에러 처리 검증
- [ ] 도움말 메시지 검증

---

## 기술적 접근 방법

### Click vs Commander 비교

| 기능 | Commander (TS) | Click (Python) |
|------|----------------|----------------|
| 명령어 정의 | `.command()` | `@cli.command()` |
| 인자 | `.argument()` | `@click.argument()` |
| 옵션 | `.option()` | `@click.option()` |
| 서브명령어 | `.command()` | `@cli.group()` |

### Rich 활용 전략
- Console: 전역 인스턴스 1개 사용
- Table: doctor 명령어에서 결과 표시
- Progress: 파일 복사 등 긴 작업

---

## 아키텍처 설계 방향

### 명령어 분리
```
cli/
├── main.py           # CLI 그룹 및 로고
├── commands/
│   ├── init.py       # init 명령어
│   ├── doctor.py     # doctor 명령어
│   ├── status.py     # status 명령어
│   └── restore.py    # restore 명령어
└── utils.py          # 공통 유틸리티
```

### 의존성 주입
- 각 명령어는 core/ 모듈만 호출
- CLI는 프레젠테이션 계층만 담당

---

## 리스크 및 대응 방안

### 리스크 1: Click 학습 곡선
- **문제**: Click은 Commander와 구조가 다름
- **대응**: 공식 문서 예제 참조, 데코레이터 패턴 익히기

### 리스크 2: Rich 출력 복잡도
- **문제**: 터미널별 색상 지원 차이
- **대응**: no_color 옵션 제공, fallback to plain text

### 리스크 3: 진입점 충돌
- **문제**: moai 명령어가 이미 존재할 수 있음
- **대응**: 설치 시 충돌 확인, 대체명 고려 (moai-adk)
