# Terraform 규칙

## ✅ 필수
- `terraform fmt`, `terraform validate`, `tflint`, `tfsec` 를 파이프라인에 포함
- 모듈화와 버전 고정(`required_version`, provider version constraints), 변수 타입/기본값 명시
- 상태 파일은 원격 backend(S3+Lock, GCS 등)로 관리하고, 민감정보는 state 및 코드에 금지
- 변경은 `terraform plan` 리뷰 후 `apply`, `destroy` 는 환경 보호/승인 절차 필요

## 👍 권장
- Workspace/환경(staging/prod) 분리, env-vars/TF_VAR로 비밀 관리, Terragrunt 고려
- Module output 최소화, 데이터 소스 캐싱, `for_each`/`count` 사용 시 주석으로 의도 명시
- drift 감지를 위해 scheduled plan, cost estimation 도구와 연동(Infracost 등)

## 🚀 확장/고급
- Policy as Code(OPA/sentinel)로 거버넌스 자동화, Checkov 추가
- GitOps(Atlantis/Spacelift) 파이프라인 구성, ChatOps 승인 플로우
- 모듈 레지스트리/버전 전략 수립, 문서 자동화(tfplugindocs, terraform-docs)
