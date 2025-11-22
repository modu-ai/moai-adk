---
name: moai-cloud-gcp-advanced
description: Advanced GCP architecture patterns, BigQuery optimization, and Kubernetes Engine best practices
version: 1.0.0
modularized: true
allowed-tools:
  - Read
  - Bash
  - WebFetch
last_updated: 2025-11-22
compliance_score: 75
auto_trigger_keywords:
  - advanced
  - cloud
  - database
  - gcp
category_tier: special
---

# Advanced GCP Cloud Architecture

## Quick Reference (30 seconds)

Google Cloud Platform(GCP)는 Kubernetes Engine(GKE), BigQuery 분석, 멀티 리전 배포,
AI/ML 서비스를 통해 엔터프라이즈급 클라우드 아키텍처를 제공합니다. 관리형 서비스,
자동 확장, 글로벌 인프라로 고가용성 분산 시스템을 구축할 수 있습니다.

**핵심 서비스** (November 2025):
- **Compute**: Compute Engine (VM), GKE (Kubernetes), Cloud Run (컨테이너), Cloud Functions (서버리스)
- **Storage**: Cloud Storage (객체), Persistent Disk (블록), Filestore (파일 공유)
- **Database**: Cloud SQL (관계), Cloud Spanner (글로벌), Firestore (NoSQL), Bigtable (시계열)
- **Analytics**: BigQuery (데이터 웨어하우스), Dataflow (스트림), Pub/Sub (메시징), Dataproc (Spark/Hadoop)
- **AI/ML**: Vertex AI (통합 플랫폼), AutoML, LLM API 등

---

## Implementation Guide

### 1. GKE (Google Kubernetes Engine) 아키텍처

**GKE Autopilot 설정**:
```yaml
apiVersion: container.googleapis.com/v1
kind: Cluster
metadata:
  name: my-autopilot-cluster
spec:
  location: us-central1
  autopilot:
    enabled: true
  releaseChannel:
    channel: REGULAR
  workloadIdentityConfig:
    workloadPool: "PROJECT_ID.svc.id.goog"
  maintenancePolicy:
    window:
      dailyMaintenanceWindow:
        startTime: "03:00"
        duration: "4h"
  binaryAuthorization:
    evaluationMode: PROJECT_SINGLETON_POLICY_ENFORCE
  podSecurityPolicy:
    enabled: true
```

**워크로드 배포**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-server
  template:
    metadata:
      labels:
        app: api-server
    spec:
      serviceAccountName: api-server-ksa
      containers:
      - name: api
        image: gcr.io/PROJECT_ID/api:v1
        ports:
        - containerPort: 8080
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
```

### 2. BigQuery 최적화 및 분석

**테이블 파티셔닝 및 클러스터링**:
```sql
CREATE TABLE `project.dataset.user_events`
PARTITION BY DATE(event_timestamp)
CLUSTER BY user_id, event_type, region
OPTIONS(
  partition_expiration_days=90,
  require_partition_filter=true,
  description="User activity events partitioned by date"
) AS
SELECT
  user_id,
  event_type,
  event_timestamp,
  region,
  properties
FROM raw_events
WHERE event_timestamp >= '2025-01-01';
```

**고성능 쿼리 패턴**:
```sql
-- 1. 파티션 필터 필수
SELECT user_id, COUNT(*) as event_count
FROM `project.dataset.events`
WHERE DATE(timestamp) BETWEEN '2025-11-01' AND '2025-11-21'
GROUP BY user_id;

-- 2. 클러스터 활용
SELECT *
FROM `project.dataset.events`
WHERE user_id = 'user_123'
  AND event_type = 'purchase'
LIMIT 1000;

-- 3. 구체화 뷰 (Materialized View)
CREATE MATERIALIZED VIEW daily_summary AS
SELECT
  DATE(timestamp) as date,
  event_type,
  COUNT(*) as count,
  SUM(value) as total_value
FROM `project.dataset.events`
WHERE DATE(timestamp) >= DATE_SUB(CURRENT_DATE(), INTERVAL 30 DAY)
GROUP BY date, event_type;
```

### 3. Cloud Run 서버리스 배포

**Dockerfile 최적화**:
```dockerfile
# Multi-stage build
FROM python:3.11-slim as builder

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

FROM python:3.11-slim

WORKDIR /app
COPY --from=builder /usr/local/lib/python3.11/site-packages /usr/local/lib/python3.11/site-packages
COPY app.py .

ENV PORT=8080
EXPOSE 8080

CMD ["python", "app.py"]
```

**Cloud Run 배포**:
```bash
gcloud run deploy my-service \
  --source . \
  --platform managed \
  --region us-central1 \
  --memory 512Mi \
  --cpu 1 \
  --allow-unauthenticated \
  --set-env-vars DATABASE_URL="postgresql://..."
```

### 4. 보안 및 네트워크

**VPC Service Controls**:
```yaml
apiVersion: accesscontextmanager.cnrm.cloud.google.com/v1beta1
kind: AccessContextManagerServicePerimeter
metadata:
  name: my-perimeter
spec:
  title: "Production Perimeter"
  description: "Secure access control"
  perimeterType: PERIMETER_TYPE_REGULAR
  status:
    resources:
    - "projects/PROJECT_NUMBER"
    restrictedServices:
    - "storage.googleapis.com"
    - "bigquery.googleapis.com"
```

**Workload Identity 설정**:
```bash
# GCP 서비스 계정 생성
gcloud iam service-accounts create my-workload

# GKE 서비스 계정 생성
kubectl create serviceaccount my-ksa

# Binding
gcloud iam service-accounts add-iam-policy-binding \
  my-workload@PROJECT_ID.iam.gserviceaccount.com \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:PROJECT_ID.svc.id.goog[NAMESPACE/my-ksa]"
```

### 5. 비용 최적화

**예산 알람**:
```bash
gcloud billing budgets create \
  --billing-account BILLING_ID \
  --display-name "Monthly Budget" \
  --budget-amount 1000 \
  --threshold-rule amount=80,percent \
  --threshold-rule amount=100,percent
```

**Committed Use Discounts (CUD)**:
```bash
# 1년 약정으로 최대 40% 할인
gcloud compute commitments create my-commitment \
  --plan="one-year" \
  --resources="cpus=8,memory=32" \
  --region=us-central1
```

---

## Best Practices

### ✅ DO
- **관리형 서비스 사용**: Cloud SQL, Cloud Run, App Engine
- **Workload Identity 활용**: 서비스 계정 키 불필요
- **VPC Service Controls**: 데이터 반출 방지
- **Cloud CDN**: 글로벌 콘텐츠 배포
- **오토스케일링**: HPA, Vertical Pod Autoscaler
- **로그 및 모니터링**: Cloud Logging, Cloud Monitoring
- **이미지 최적화**: 멀티 스테이지 빌드, 최소 기본 이미지

### ❌ DON'T
- 서비스 계정 키를 코드에 저장
- VPC Service Controls 무시 (보안 위험)
- Security Command Center 알람 무시
- 기본 네트워크 사용 (Custom VPC 필수)
- 무제한 확장 (예산 초과)
- 컨테이너 이미지 최적화 무시

---

## Context7 Integration

### Related GCP Services & Libraries
- [Google Cloud Compute](/gcp/compute): Virtual machines and containers
- [Google Kubernetes Engine](/gcp/gke): Managed Kubernetes service
- [Google Cloud SQL](/gcp/cloudsql): Managed relational databases
- [Google Cloud Firestore](/gcp/firestore): Real-time NoSQL database
- [Google BigQuery](/gcp/bigquery): Data warehouse and analytics
- [Google Vertex AI](/gcp/vertex-ai): Unified ML platform

### Official GCP Documentation
- [GCP Cloud Architecture Center](https://cloud.google.com/architecture)
- [GKE Best Practices](https://cloud.google.com/kubernetes-engine/docs/best-practices)
- [BigQuery Documentation](https://cloud.google.com/bigquery/docs)
- [Cloud Security Best Practices](https://cloud.google.com/security/best-practices)

### Related Modularized Skills
- `moai-cloud-aws-advanced` - AWS equivalent services and patterns
- `moai-domain-devops` - CI/CD and deployment automation
- `moai-security-api` - API security implementation
- `moai-essentials-perf` - Performance optimization techniques

---

## Works Well With

- `moai-cloud-aws-advanced` (멀티 클라우드 비교)
- `moai-domain-devops` (GCP CI/CD 파이프라인)
- `moai-domain-database` (데이터 모델링)
- `moai-domain-security` (클라우드 보안)

---

**Version**: 2.0.0
**Last Updated**: 2025-11-22
**Status**: Fully Modularized
**Official Reference**: https://cloud.google.com/architecture/framework