# `forge-audit-status`

operation `forge.audit_status` · 코드 `src/ops/forge_audit_status.rs` · #2173

`forge-audit-status PROJECT --permission-reference R` 는 정확한 SHA 에 `moai/audit` commit status 를 올린다. stdin 의 typed JSON. `MOAI_AUDIT_APP_TOKEN` 이 있으면 감사 App 토큰을 쓰고, 없으면 경고한다. `src/main.rs` 에서 그대로 옮김(#2834 R2b).
