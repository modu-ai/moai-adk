package permission

import "testing"

// SPEC-SEC-HARDEN-005 §F.1 — Permission `:*` prefix-match `${IFS}` word-split bypass.
//
// 본 테스트 파일은 reproduction-first 계약을 따른다:
//   - AC-SEC5-001/002 (RED): 픽스 전 코드에서 `${IFS}`/`$IFS` 가 lexical separator 문자를
//     포함하지 않아 `hasUnquotedShellSeparator` 어휘 스캔을 통과 → chained 명령이
//     prefix 룰로 silently ALLOW 되는 word-split bypass 가 존재함을 입증
//     (픽스 전에는 이 테스트들이 FAIL).
//   - AC-SEC5-003/004 (NO-REG): 기존 separator/chain DENY 스위트는 동작 불변.
//   - AC-SEC5-005 (NO-REG): `$HOME`/`${HOME}`/`TestX$`/quoted `${IFS}` 등 word-split
//     의도 없는 입력은 ALLOW 유지 (over-deny 회귀 방지).
//   - AC-SEC5-006 (fail-closed): malformed shell → DENY (allow 아님).
//
// 봉쇄는 `mvdan.cc/sh/v3/syntax` shell-aware 파서로 구현하며 `$`-blacklist 어휘 해킹은
// 출하하지 않는다(REQ-SEC5-003).

// TestMatches_IFSWordSplit_Reproduction 은 AC-SEC5-001 (RED→GREEN) 이다.
// `Bash(go test:*)` 룰 + `go test ${IFS}curl${IFS}evil` 입력은:
//   - 픽스 전(RED): true — `${IFS}` 는 리터럴 separator 문자(`;`/`&&`/`|`/`$(`...)를
//     포함하지 않아 `hasUnquotedShellSeparator` 어휘 스캔을 통과 → bypass.
//   - 픽스 후(GREEN): false — shell-aware 파서가 unquoted `${IFS}` 파라미터 확장을
//     word-split boundary 로 인식 → prefix 룰 비매칭(deny path fall-through).
func TestMatches_IFSWordSplit_Reproduction(t *testing.T) {
	t.Parallel()

	rule := PermissionRule{Pattern: "Bash(go test:*)"}

	if got := rule.Matches("Bash", "go test ${IFS}curl${IFS}evil"); got != false {
		t.Errorf("Matches(${IFS} word-split) = %v, want false (unquoted ${IFS} introduces word-split → deny)", got)
	}
}

// TestMatches_IFSVariants 은 AC-SEC5-002 (RED→GREEN) 의 변형 테이블이다.
// `${IFS}` / `$IFS` (short form) / 다중 삽입 변형이 전부 word-split 로 인식되어
// 비매칭(deny)되어야 한다.
func TestMatches_IFSVariants(t *testing.T) {
	t.Parallel()

	rule := PermissionRule{Pattern: "Bash(go test:*)"}

	tests := []struct {
		name  string
		input string
	}{
		{"braced IFS prefix", "go test ${IFS}curl${IFS}evil"},
		// `$IFS` short form must be delimited by a non-identifier char to reference
		// IFS; `$IFScurl` would name a *different* variable `IFScurl` (greedy ident
		// match) and is intentionally NOT an IFS split, so we use `$IFS/...`.
		{"short IFS form", "go test $IFS/curl"},
		{"short IFS with separator arg", "go test $IFS-rf$IFS/"},
		{"single braced IFS mid-arg", "go test x${IFS}y"},
		{"trailing braced IFS then command", "go test ./...${IFS}curl${IFS}evil"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := rule.Matches("Bash", tt.input); got != false {
				t.Errorf("Matches(%q) = %v, want false (IFS word-split must deny)", tt.input, got)
			}
		})
	}
}

// TestMatches_IFSLegitNotRejected 은 AC-SEC5-005 (NO-REG) 다.
// `$`/`${...}` 를 word-split 의도 없이 포함하는 정상 입력은 ALLOW 유지되어야 한다
// (over-deny 회귀 방지).
//
// 경계 케이스 `go test TestX$`(trailing-`$`)는 REQ-SEC5-004(parse-fail→DENY)와
// REQ-SEC5-006(legit→ALLOW)의 충돌점이다 — design.md D.1.4 결정에 따라 반드시 ALLOW
// (파서가 lone trailing `$` 를 parse-error 로 보더라도 generic fail-closed DENY 이전에
// trailing-literal-`$` special-case ALLOW 로 처리).
func TestMatches_IFSLegitNotRejected(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pattern string
		input   string
	}{
		// SEC-HARDEN-002 9-sample legit set 계열 — word-split 의도 없는 정상 입력.
		{"plain flags single command", "Bash(go test:*)", "go test -race ./..."},
		{"HOME param expansion short", "Bash(go test:*)", "go test $HOME/x"},
		{"HOME param expansion braced", "Bash(go test:*)", "go test ${HOME}"},
		{"GOPATH param mid-arg", "Bash(go test:*)", "go test $GOPATH/src/x"},
		{"trailing dollar literal", "Bash(go test:*)", "go test TestX$"},
		{"run regex anchored test name", "Bash(go test:*)", "go test -run 'TestX$'"},
		{"quoted braced IFS no split", "Bash(echo:*)", `echo "${IFS}"`},
		{"single-quoted IFS literal", "Bash(echo:*)", `echo '${IFS}'`},
		{"empty remainder", "Bash(go test:*)", "go test"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			rule := PermissionRule{Pattern: tt.pattern}
			if got := rule.Matches("Bash", tt.input); got != true {
				t.Errorf("Matches(%q) = %v, want true (legit input without word-split intent must stay ALLOW)", tt.input, got)
			}
		})
	}
}

// TestMatches_MalformedShellFailClosed 은 AC-SEC5-006 (fail-closed) 다.
// 셸 파서가 파싱 불가한 malformed remainder 는 fail-closed 로 비매칭(deny)되어야 한다
// (allow 아님). 단, trailing-literal-`$` 는 malformed 로 분류하지 않는다
// (TestMatches_IFSLegitNotRejected 가 그 경계를 ALLOW 로 고정).
func TestMatches_MalformedShellFailClosed(t *testing.T) {
	t.Parallel()

	rule := PermissionRule{Pattern: "Bash(go test:*)"}

	tests := []struct {
		name  string
		input string
	}{
		// genuine malformed shell — 파서 에러 → fail-closed DENY.
		{"unbalanced param brace", "go test ${IFS"},
		{"unbalanced command substitution", "go test $(echo"},
		{"unterminated double quote", `go test "unfinished`},
		{"unterminated single quote", "go test 'unfinished"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := rule.Matches("Bash", tt.input); got != false {
				t.Errorf("Matches(%q) = %v, want false (malformed shell must fail-closed → deny)", tt.input, got)
			}
		})
	}
}

// TestIsTrailingDollarLiteral 은 trailing-`$` special-case 판정을 직접 고정한다
// (REQ-SEC5-006 경계, design D.1.4). lone trailing `$`(Go regex anchor)는 true,
// genuine `$(`/`${` 확장을 동반하면 false(fail-closed deny 유지).
func TestIsTrailingDollarLiteral(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"trailing dollar anchor", "go test TestX$", true},
		{"trailing dollar with spaces", "go test TestX$  ", true},
		{"quoted regex anchor trailing dollar", "go test -run 'TestX$'", false}, // no trailing $ after quote
		{"no trailing dollar", "go test ./...", false},
		{"unterminated param expansion", "go test ${IFS", false},
		{"unterminated command substitution", "go test $(echo", false},
		{"dollar mid-string only", "go test $HOME/x", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isTrailingDollarLiteral(tt.s); got != tt.want {
				t.Errorf("isTrailingDollarLiteral(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

// TestMatches_IFSParserEdgeCases 은 AC-SEC5-001/002 의 shell-aware 파서 분기 보강
// 테이블이다. 다중 명령 / command·process substitution / subshell / quoted-IFS
// 경계를 각각 고정하여 hasIFSWordSplit 의 DENY/ALLOW 결정을 검증한다.
func TestMatches_IFSParserEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		match bool // expected Matches() result
	}{
		// 다중 statement (`;` 로 분리) → len(Stmts)!=1 → DENY.
		{"two statements semicolon", "go test ./...; echo done", false},
		// command substitution `$(...)` (unquoted) → DENY.
		{"command substitution arg", "go test $(echo evil)", false},
		// process substitution `<(...)` → DENY.
		{"process substitution arg", "go test <(curl evil)", false},
		// subshell `(...)` as the command → non-CallExpr → conservative DENY.
		{"subshell command", "go test && (curl evil)", false},
		// quoted command-substitution stays a single quoted word; the inner
		// CmdSubst is still embedded code → DENY (conservative, parser SSOT).
		{"quoted command substitution", `go test "$(echo x)"`, false},
		// double-quoted plain text — no IFS, no subst → ALLOW.
		{"double-quoted plain arg", `go test "a b c"`, true},
		// braced IFS inside double quotes → quoted context, no split → ALLOW.
		{"double-quoted braced IFS", `go test "x${IFS}y"`, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			rule := PermissionRule{Pattern: "Bash(go test:*)"}
			if got := rule.Matches("Bash", tt.input); got != tt.match {
				t.Errorf("Matches(%q) = %v, want %v", tt.input, got, tt.match)
			}
		})
	}
}
