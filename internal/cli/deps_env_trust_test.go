package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// SPEC-SEC-HARDEN-005 §F.2 — update source env-trust allowlist.
//
// 본 테스트 파일은 reproduction-first 계약을 따른다:
//   - AC-SEC5-008/009 (RED): 픽스 전 EnsureUpdate 가 non-https / allowlist 외 host 의
//     MOAI_UPDATE_URL 로도 update checker 를 구성함을 입증(adversarial source 도달).
//   - AC-SEC5-010 (RED): URL-shaped MOAI_RELEASES_DIR 도 검증 없이 사용됨.
//   - AC-SEC5-011 (NO-REG): env 미설정 default 경로는 api.github.com checker 정상 구성.
//
// 검증 함수는 update source 3종 env(MOAI_UPDATE_SOURCE/MOAI_UPDATE_URL/
// MOAI_RELEASES_DIR)로 한정한다(REQ-SEC5-011) — .env.glm / WSL2-PATH 미확장.
//
// 이 테스트들은 process env(os.Getenv)를 통한 EnsureUpdate 동작을 검사하므로
// 병렬 실행하지 않는다(t.Setenv + 비-parallel).

// TestEnsureUpdate_RejectsNonHTTPSUpdateURL 은 AC-SEC5-008 (RED→GREEN) 이다.
// non-https scheme 의 MOAI_UPDATE_URL 은 fail-closed 거부되어야 하고 update checker 가
// 구성되어선 안 된다. 픽스 전: checker 구성(err==nil) → 이 테스트 FAIL.
func TestEnsureUpdate_RejectsNonHTTPSUpdateURL(t *testing.T) {
	t.Setenv(config.EnvUpdateURL, "http://evil.example/repos/x/releases")

	d := &Dependencies{}
	err := d.EnsureUpdate()
	if err == nil {
		t.Fatal("EnsureUpdate must fail-closed for non-https MOAI_UPDATE_URL, got nil error")
	}
	if d.UpdateChecker != nil {
		t.Error("EnsureUpdate must NOT construct an update checker for a rejected source")
	}
	if !strings.Contains(err.Error(), config.EnvUpdateURL) {
		t.Errorf("error should reference %s, got: %v", config.EnvUpdateURL, err)
	}
}

// TestEnsureUpdate_RejectsDisallowedHost 은 AC-SEC5-009 (RED→GREEN) 이다.
// https 이지만 allowlist 외 host 인 MOAI_UPDATE_URL 은 fail-closed 거부되어야 한다.
func TestEnsureUpdate_RejectsDisallowedHost(t *testing.T) {
	t.Setenv(config.EnvUpdateURL, "https://evil.example/repos/modu-ai/moai-adk/releases")

	d := &Dependencies{}
	err := d.EnsureUpdate()
	if err == nil {
		t.Fatal("EnsureUpdate must fail-closed for disallowed host, got nil error")
	}
	if d.UpdateChecker != nil {
		t.Error("EnsureUpdate must NOT construct an update checker for a disallowed host")
	}
}

// TestEnsureUpdate_RejectsURLShapedReleasesDir 은 AC-SEC5-010 (RED→GREEN) 이다.
// MOAI_UPDATE_SOURCE=local 일 때 MOAI_RELEASES_DIR 가 URL-shaped 이면 fail-closed
// 거부되어야 한다(로컬 경로여야 함).
func TestEnsureUpdate_RejectsURLShapedReleasesDir(t *testing.T) {
	t.Setenv(config.EnvUpdateSource, "local")
	t.Setenv(config.EnvReleasesDir, "https://evil.example/releases")

	d := &Dependencies{}
	err := d.EnsureUpdate()
	if err == nil {
		t.Fatal("EnsureUpdate must fail-closed for URL-shaped MOAI_RELEASES_DIR, got nil error")
	}
	if d.UpdateChecker != nil {
		t.Error("EnsureUpdate must NOT construct a local checker for a URL-shaped releases dir")
	}
}

// TestEnsureUpdate_DefaultPathNoRegression 은 AC-SEC5-011 (NO-REG) 이다.
// update 관련 env 미설정 시 canonical api.github.com 기반 update checker 가 정상
// 구성되어야 한다(회귀 없음). t.Setenv 로 3종 env 를 명시적으로 비워 격리한다.
func TestEnsureUpdate_DefaultPathNoRegression(t *testing.T) {
	t.Setenv(config.EnvUpdateSource, "")
	t.Setenv(config.EnvUpdateURL, "")
	t.Setenv(config.EnvReleasesDir, "")

	d := &Dependencies{}
	err := d.EnsureUpdate()
	if err != nil {
		t.Fatalf("EnsureUpdate default path must succeed (no env override), got: %v", err)
	}
	if d.UpdateChecker == nil {
		t.Error("EnsureUpdate default path must construct an UpdateChecker")
	}
	if d.UpdateOrch == nil {
		t.Error("EnsureUpdate default path must construct an UpdateOrch")
	}
}

// TestEnsureUpdate_AcceptsCanonicalHTTPSUpdateURL 은 REQ-SEC5-010 의 명시 override
// 통과 케이스다. https + allowlist host(api.github.com) 인 MOAI_UPDATE_URL 은 통과해야
// 한다(정상 override 가 거부되지 않음을 고정).
func TestEnsureUpdate_AcceptsCanonicalHTTPSUpdateURL(t *testing.T) {
	t.Setenv(config.EnvUpdateURL, "https://api.github.com/repos/modu-ai/moai-adk/releases")

	d := &Dependencies{}
	err := d.EnsureUpdate()
	if err != nil {
		t.Fatalf("EnsureUpdate must accept canonical https api.github.com URL, got: %v", err)
	}
	if d.UpdateChecker == nil {
		t.Error("EnsureUpdate must construct an UpdateChecker for an allowed source")
	}
}
