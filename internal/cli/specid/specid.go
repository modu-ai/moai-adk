// Package specid는 CLI SPEC-ID 공유 sanitizer를 제공하는 leaf package다.
//
// SPEC-SEC-HARDEN-002 M1/M2a — ValidateSpecID는 모든 CLI SPEC-ID 경계
// (spec view/status/close)에서 사용하는 단일 검증 헬퍼다.
// leaf package에 두는 이유: internal/cli가 internal/cli/worktree를 import하므로
// worktree는 cli를 import할 수 없다(import cycle). 본 package는 cli와 worktree
// 양쪽에서 import 가능하고 둘 중 어느 것도 import하지 않으므로 cycle이 없다.
// 검증 로직은 이 package에 단 한 번만 정의된다(REQ-SEC2-M1-004).
package specid

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// CanonicalSpecIDShapeLiteral is the anchored SPEC-ID shape regexp source.
//
// SPEC-SPEC-LINT-ID-ARG-001 (REQ-SLI-006, REQ-SLI-009) — it is a COPY of the
// strict pattern that lives, unexported, in internal/spec (lint.go). The copy
// exists because that symbol is lowercase and therefore unimportable, and
// exporting it would edit internal/spec, which this SPEC's radius forbids
// (REQ-SLI-008). specid_shape_test.go compares this literal, byte for byte,
// against the literal read out of the internal/spec source, so the two
// drifting apart is detected rather than assumed.
//
// Two properties are load-bearing and must not be relaxed:
//   - both ends are anchored (^…$), so it cannot match a substring of a path;
//   - neither "." nor a path separator appears in any character class.
//
// Both exist to keep this from behaving like the LOOSER, same-named symbol
// specIDPattern in internal/cli/spec_status.go:18, which is unanchored on
// purpose — it scrapes SPEC-IDs out of git commit messages. Reusing that one
// as an argument discriminator misreads a path as an ID; the risk has a name,
// same-name/different-meaning reuse drift, and spec.md §H records it.
const CanonicalSpecIDShapeLiteral = `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`

var canonicalSpecIDShape = regexp.MustCompile(CanonicalSpecIDShapeLiteral)

// HasCanonicalSpecIDShape reports whether s has the canonical SPEC-ID shape.
//
// This is a SHAPE question, and it is deliberately a different question from
// the one ValidateSpecID answers: that function is a security sanitizer that
// rejects exactly three things (absolute path, "..", path separators) and says
// nothing about whether the remainder looks like a SPEC-ID at all. A caller
// deciding "is this argument an ID or a path?" needs this function; a caller
// about to build a filesystem path out of an ID needs that one.
//
// The name is deliberately unlike specIDPattern (spec_status.go:18), so that
// a reader reaching for "the SPEC-ID regexp" cannot pick the loose one by
// accident (REQ-SLI-006).
func HasCanonicalSpecIDShape(s string) bool {
	return canonicalSpecIDShape.MatchString(s)
}

// ValidateSpecID는 specID에 path-traversal 문자("..", "/" 또는 "\", 절대 경로)가
// 포함되어 있으면 구조화된 검증 에러를 반환한다. 정상 canonical SPEC-ID(예:
// "SPEC-SEC-HARDEN-002")는 nil을 반환한다.
//
// 호출 위치는 CLI args[0] 경계 — filepath.Join이나 os.MkdirAll/WorktreeProvider.Add
// 같은 경로 구성/생성 sink에 specID가 도달하기 전에 호출되어야 한다. 기존
// validateSkillID(internal/cli/update_archive.go)를 모델로 한다.
//
// @MX:NOTE: [AUTO] SPEC-SEC-HARDEN-002 M1 — 모든 CLI SPEC-ID 경계의 단일 sanitizer (leaf package). validateSkillID(update_archive.go) 모델. ".."/경로구분자/절대경로 거부 후 filepath.Join 도달 차단.
func ValidateSpecID(specID string) error {
	// 절대 경로 거부
	if filepath.IsAbs(specID) {
		return fmt.Errorf("SPEC-ID must be a simple identifier, not an absolute path: %q", specID)
	}
	// ".." 거부
	if strings.Contains(specID, "..") {
		return fmt.Errorf("SPEC-ID must not contain '..': %q", specID)
	}
	// 경로 구분자("/" 또는 "\") 거부 — 크로스 플랫폼 (validateSkillID와 동일 패턴)
	if strings.ContainsAny(specID, "/\\") {
		return fmt.Errorf("SPEC-ID must not contain path separators: %q", specID)
	}
	return nil
}
