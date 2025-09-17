# SQL 규칙

## ✅ 필수
- 모든 쿼리는 파라미터 바인딩 사용, 동적 문자열 연결 금지
- 인덱스 전략을 읽기 패턴/카디널리티 기준으로 설계하고 쿼리 계획(EXPLAIN) 검증
- 마이그레이션은 idempotent, 업/다운 스크립트와 롤백 절차 문서화
- 트랜잭션 경계를 명확히 하고 잠금/교착상태 모니터링을 추가

## 👍 권장
- 규약 기반 명명(스키마/테이블/컬럼), snake_case 유지, ENUM보단 reference 테이블 선호
- 데이터 품질: NOT NULL/DEFAULT, CHECK 제약 조건 적극 활용, GDPR/PII 마스킹 전략 포함
- 배치/ETL 작업은 Job 로그/재시도 전략과 SLA 정의

## 🚀 확장/고급
- 파티셔닝/샤딩/리플리케이션 전략을 ADR에 기록, 읽기/쓰기 분리 적용
- 시뮬레이션 데이터/샌드박스 환경 구축, 데이터 테스트(Assertions, dbt tests)
- DB Observability(Performance Schema, pg_stat_statements 등)와 경보 시스템 연동
