# TAG 시스템 모니터링 도구

> **용도**: TAG 시스템 건강도 및 메트릭 모니터링
>
> **대상**: 최종 사용자 및 개발자

---

## 📋 스크립트 목록

### `tag_health_monitor.py`

**목적**: TAG 시스템의 주간 건강 검사 및 메트릭 모니터링

**사용 대상**:
- 주간 TAG 시스템 정기 점검
- TAG 커버리지 추적
- 프로젝트 TAG 성숙도 모니터링
- 건강도 대시보드 생성

**모니터링 항목** (6가지):
1. **TAG 커버리지**: SPEC 대비 CODE/TEST/DOC 연계율
2. **고아 TAG**: 부모가 없는 TAG 탐지
3. **미충족 TAG**: 존재하지만 구현 미완료 TAG
4. **TAG 중복**: 동일 ID의 중복 TAG
5. **문서 동기화**: SPEC과 문서의 일관성
6. **정의 완성도**: TAG 메타데이터 완성도

**사용 방법**:
```bash
# 기본 건강 검사 (콘솔 출력)
python3 .moai/scripts/monitoring/tag_health_monitor.py

# HTML 대시보드 생성
python3 .moai/scripts/monitoring/tag_health_monitor.py --html

# 주간 리포트 (자동 저장)
python3 .moai/scripts/monitoring/tag_health_monitor.py --weekly

# Slack 알림 전송 (설정 필요)
python3 .moai/scripts/monitoring/tag_health_monitor.py --notify slack
```

**정상 실행 예시**:
```
🏥 TAG 시스템 건강 검사

📊 메트릭:
  ✅ TAG 커버리지: 95% (권장: 85%)
  ✅ 고아 TAG: 0개 (권장: 0)
  ⚠️ 미충족 TAG: 2개 (권장: 0)
  ✅ TAG 중복: 0개 (권장: 0)
  ✅ 문서 동기화: 98%
  ✅ 정의 완성도: 92%

📈 전체 건강도: 95/100 (매우 좋음)

💾 리포트: .moai/reports/health/tag_health_2025-11-13.json
```

---

## 🚀 주간 모니터링 워크플로우

```bash
# 매주 월요일 실행 (자동화 권장)
python3 .moai/scripts/monitoring/tag_health_monitor.py --weekly --notify slack
```

**출력 위치**:
- 콘솔: 즉시 결과 표시
- 파일: `.moai/reports/health/tag_health_YYYY-MM-DD.json`
- Slack: (설정 시) Slack 채널에 자동 알림

---

## 📊 메트릭 상세 설명

| 메트릭 | 설명 | 경고 기준 |
|--------|------|---------|
| **TAG 커버리지** | SPEC 문서의 TAG 정의 비율 | < 85% |
| **고아 TAG** | 부모가 없는 (@CODE가 있는데 @SPEC 없음) | > 0 |
| **미충족 TAG** | @SPEC 정의됨 but @CODE 미구현 | > 10% |
| **TAG 중복** | 동일 SPEC-ID로 여러 번 사용 | > 0 |
| **문서 동기화** | SPEC 수정 후 DOC 업데이트 여부 | < 90% |
| **정의 완성도** | TAG 메타데이터 필드 채움 | < 80% |

---

## ⚙️ Slack 알림 설정

`tag_health_monitor.py --notify slack` 사용 시:

1. **Slack Webhook URL 설정**:
   ```bash
   export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
   ```

2. **환경 변수 저장** (.env):
   ```bash
   echo "SLACK_WEBHOOK_URL=https://hooks.slack.com/..." >> .env
   ```

3. **자동 실행** (crontab):
   ```bash
   # 매주 월요일 09:00 실행
   0 9 * * 1 cd /path/to/project && python3 .moai/scripts/monitoring/tag_health_monitor.py --weekly --notify slack
   ```

---

## 📈 건강도 등급

| 점수 | 등급 | 상태 |
|------|------|------|
| 95-100 | 🟢 매우 좋음 | 유지 관리 필요 없음 |
| 85-94 | 🟡 좋음 | 경고 항목 확인 필요 |
| 75-84 | 🟠 주의 | 개선 작업 권장 |
| < 75 | 🔴 불량 | 즉시 개선 필요 |

---

## 🔗 관련 도구

- **TAG 중복 관리**: `.moai/scripts/validation/tag_dedup_manager.py`
- **TAG 시스템**: `.moai/specs/TAG-REFERENCE.md`

---

**마지막 업데이트**: 2025-11-13
**상태**: Production Ready
