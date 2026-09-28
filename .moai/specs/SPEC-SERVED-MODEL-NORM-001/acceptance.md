# Acceptance — SPEC-SERVED-MODEL-NORM-001

| AC | REQ | 검증 |
|---|---|---|
| AC-SMN-001 | REQ-SMN-001 | `TestServedModel_Classify/e_declared_context_suffix_id_served_bare_id_is_ok`, `/f_declared_context_suffix_alias_served_family_is_ok` → `ok` |
| AC-SMN-002 | REQ-SMN-002 | `TestServedModel_Classify/h_declared_inherit_without_profile_is_not_drift` → `unmapped` |
| AC-SMN-003 | REQ-SMN-003 | `TestServedModel_Classify/g_control_context_suffix_alias_served_glm_stays_drift` → `served_drift` |
| AC-SMN-004 | REQ-SMN-001, REQ-SMN-002, REQ-SMN-003 | 수리 바이너리의 `moai doctor --check "Served Model" --verbose` 출력에 `expected=…[1m]` 또는 `expected=inherit` 인 drift 행이 0건 |
