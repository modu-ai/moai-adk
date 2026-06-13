// SPEC-SEC-HARDEN-002 M1 — validateSpecID 공유 sanitizer 헬퍼 테스트.
//
// 재현 우선(reproduction-first): 헬퍼가 존재하기 전에는 미정의 심볼 참조로
// 컴파일이 실패한다(RED). 헬퍼 추가 후 GREEN. 정상 SPEC-ID는 수락,
// path-traversal 입력(`..`, 경로 구분자, 절대 경로)은 거부함을 검증한다.
package cli

import "testing"

func TestValidateSpecID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		specID  string
		wantErr bool
	}{
		// 정상 canonical SPEC-ID — 수락 (AC-SEC2-M1-002)
		{name: "legitimate canonical spec id", specID: "SPEC-SEC-HARDEN-002", wantErr: false},
		{name: "legitimate short spec id", specID: "SPEC-AUTH-001", wantErr: false},

		// `..` traversal — 거부 (AC-SEC2-M1-001)
		{name: "dotdot traversal", specID: "../../../../tmp/evil", wantErr: true},
		{name: "embedded dotdot", specID: "SPEC-..-001", wantErr: true},

		// 경로 구분자 — 거부 (AC-SEC2-M1-003)
		{name: "forward slash separator", specID: "foo/bar", wantErr: true},
		{name: "backslash separator", specID: "foo\\bar", wantErr: true},

		// 절대 경로 — 거부 (AC-SEC2-M1-001 / 003)
		{name: "absolute path unix", specID: "/etc/passwd", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateSpecID(tt.specID)
			if tt.wantErr && err == nil {
				t.Fatalf("validateSpecID(%q) = nil, want non-nil error", tt.specID)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("validateSpecID(%q) = %v, want nil", tt.specID, err)
			}
		})
	}
}
