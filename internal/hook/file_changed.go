// Resolution: UPGRADE — MX re-scan for 16 supported language extensions on FileChanged.
// SPEC-V3R6-HOOK-ASYNC-EXPAND-001 M2 (REQ-HAE-001): MX delta scan + sidecar
// update execute in a background goroutine with a 5-second deadline. The main
// handler returns within ≤ 100 ms (p95 under 10-concurrent benchmark, AC-HAE-002).
package hook

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/modu-ai/moai-adk/internal/mx"
)

// Supported language extensions for MX tag scanning.
var supportedExtensions = map[string]bool{
	".go":    true,
	".py":    true,
	".ts":    true,
	".js":    true,
	".rs":    true,
	".java":  true,
	".kt":    true,
	".cs":    true,
	".rb":    true,
	".php":   true,
	".ex":    true,
	".exs":   true,
	".cpp":   true,
	".cc":    true,
	".cxx":   true,
	".h":     true,
	".hpp":   true,
	".scala": true,
	".r":     true,
	".dart":  true,
	".swift": true,
}

// asyncDeadline is the canonical deadline for hook handler side-effect
// goroutines per SPEC-V3R6-HOOK-ASYNC-EXPAND-001 REQ-HAE-005. After this
// duration the goroutine MUST self-cancel via its context.
const asyncDeadline = 5 * time.Second

// fileChangedHandler processes FileChanged events.
// It scans changed files for MX tag deltas asynchronously.
type fileChangedHandler struct {
	// wg tracks in-flight async side-effect goroutines. Tests use the
	// package-internal waitGroup() accessor + testutil.WaitForAsync to
	// deterministically await completion. Production callers MUST NOT
	// block on this WaitGroup.
	wg sync.WaitGroup
}

// NewFileChangedHandler creates a new FileChanged event handler.
func NewFileChangedHandler() Handler {
	return &fileChangedHandler{}
}

// waitGroup returns the handler's internal *sync.WaitGroup for use with
// testutil.WaitForAsync. Package-internal; not exposed via the Handler
// interface. Tests obtain a typed reference via a package-internal
// type-assert or via the typed constructor.
func (h *fileChangedHandler) waitGroup() *sync.WaitGroup {
	return &h.wg
}

// EventType returns EventFileChanged.
func (h *fileChangedHandler) EventType() EventType {
	return EventFileChanged
}

// Handle processes a FileChanged event. It returns within ≤ 100 ms (p95
// under 10-concurrent benchmark per AC-HAE-002) by deferring the MX delta
// scan + sidecar update to a background goroutine bounded by asyncDeadline.
//
// REQ-HAE-001: main return path completes synchronously with HookOutput{}.
// All side-effects (scan, sidecar update) execute in the spawned goroutine.
func (h *fileChangedHandler) Handle(_ context.Context, input *HookInput) (*HookOutput, error) {
	slog.Info("file changed externally",
		"session_id", input.SessionID,
		"file_path", input.FilePath,
		"change_type", input.ChangeType,
	)

	// Skip deleted files (synchronous fast-path — no side-effect work).
	if input.ChangeType == "deleted" {
		return &HookOutput{}, nil
	}

	// Check if file has supported extension (synchronous fast-path).
	ext := strings.ToLower(filepath.Ext(input.FilePath))
	if !supportedExtensions[ext] {
		slog.Debug("unsupported file extension for MX scan",
			"path", input.FilePath,
			"ext", ext,
		)
		return &HookOutput{}, nil
	}

	// REQ-HAE-001 async transition: MX scan + sidecar update run in a
	// background goroutine with a 5-second deadline. The main handler
	// returns immediately so the Claude Code main loop is unblocked.
	//
	// context.Background() decouples from the request ctx — request
	// cancellation MUST NOT cancel the side-effect (it runs to completion
	// or the asyncDeadline expires).
	asyncCtx, cancel := context.WithTimeout(context.Background(), asyncDeadline)
	h.wg.Add(1)
	go func() {
		defer cancel()
		defer h.wg.Done()
		h.runMXScan(asyncCtx, input)
	}()

	return &HookOutput{}, nil
}

// runMXScan executes the MX tag scan + sidecar update for the changed file.
// It MUST be called from a goroutine; the deadline-aware ctx is honored by
// downstream MX scanner / manager calls when they support context.
//
// Errors are logged and swallowed — async side-effects never propagate
// failure to the main handler response (REQ-HAE-005 design intent).
//
// TOCTOU note (SPEC-SEC-HARDEN-005 §F.3, OPT-SEC5-001 — non-gating): the scan
// resolves the project root and then reads the changed file / writes its sidecar;
// a concurrent adversarial process could in principle race a path swap between
// resolution and the subsequent read/write (check-vs-use window). This TOCTOU
// window is out of scope under the offline single-process threat model (the hook
// runs as a single process on the user's own machine) per SEC-HARDEN-003/004
// §F.1 precedent. No code-behavior change.
func (h *fileChangedHandler) runMXScan(ctx context.Context, input *HookInput) {
	// Respect deadline cancellation defensively.
	select {
	case <-ctx.Done():
		slog.Warn("file_changed async: context cancelled before scan",
			"path", input.FilePath,
			"error", ctx.Err(),
		)
		return
	default:
	}

	// SPEC-SEC-HARDEN-003 C-F1 봉쇄 (additive guard): hook stdin JSON에서 온
	// input.FilePath / input.CWD는 공격자 영향이 가능하므로, 스캔·사이드카 쓰기
	// 전에 해소된 프로젝트 루트 내부로 봉쇄한다. 위반 시 fail-closed(로그 후
	// early return) — 비동기 side-effect 실패는 hook 응답에 전파되지 않는다
	// (REQ-SEC3-004 / REQ-HAE-005 보존; main handler는 고정 빈 payload 반환).
	//
	// 루트 해소는 중앙화된 resolveProjectRootFromInputOrEnv만 사용한다(B7 canonical;
	// env 인라인 조회 금지, NFR-SEC3-002).
	root := resolveProjectRootFromInputOrEnv(input, "runMXScan")
	if root == "" {
		// 루트를 해소할 수 없으면 봉쇄 불가 → 사이드카 작업 스킵(fail-closed).
		slog.Warn("file_changed async: project root unresolved, skipping MX scan",
			"path", input.FilePath,
		)
		return
	}

	// REQ-SEC3-002: input.FilePath가 해소된 루트를 탈출하면 스캔 없이 거부한다.
	if !pathContainedIn(root, input.FilePath) {
		slog.Warn("file_changed async: FilePath escapes project root, skipping MX scan (containment)",
			"path", input.FilePath,
			"project_root", root,
		)
		return
	}

	// SPEC-SEC-HARDEN-004 F2 (REQ-SEC4-004): lexical pathContainedIn은 symlink를
	// 따라가지 않으므로, root 안에 lexically 존재하지만 root 밖을 가리키는 symlink
	// (root/innocent.go → /secret/secret.go)가 위 가드를 통과한다. ScanFile은
	// os.ReadFile로 링크를 추종하므로 root 밖 내용(MX-tag 텍스트)을 읽어 사이드카에
	// 기록할 수 있다(CWE-61 읽기 증폭). 이를 봉쇄하기 위해 EvalSymlinks로 실경로를
	// 해소해 root 봉쇄를 재검사한다(resolve-recheck; symlink-skip이 아님 — 정상
	// in-root symlink 스캔은 보존). EvalSymlinks error(not-exist 포함)는 스캔 대상이
	// 없거나 해소 불가이므로 fail-closed early return(NFR-SEC4-004).
	//
	// 비교 base(root)도 EvalSymlinks로 정규화한다: root가 symlinked 경로(예: macOS
	// /var → /private/var)면 leaf만 해소한 resolvedPath와 정규화 안 된 root를
	// 비교 시 false escape가 발생한다(M1 F1 watch-item-2와 동일 정규화 요구).
	// root 해소 실패 시(존재하지 않는 등)는 원래 root로 폴백한다 — 정규화는
	// best-effort이며 실패해도 leaf 해소 + lexical 비교가 그대로 적용된다.
	resolvedPath, rerr := filepath.EvalSymlinks(input.FilePath)
	if rerr != nil {
		slog.Warn("file_changed async: FilePath unresolvable, skipping MX scan (containment)",
			"path", input.FilePath,
			"error", rerr,
		)
		return
	}
	normRoot := root
	if resolvedRoot, err := filepath.EvalSymlinks(root); err == nil {
		normRoot = resolvedRoot
	}
	if !pathContainedIn(normRoot, resolvedPath) {
		slog.Warn("file_changed async: FilePath symlink escapes project root, skipping MX scan (containment)",
			"path", input.FilePath,
			"resolved", resolvedPath,
			"project_root", normRoot,
		)
		return
	}

	// Scan file for MX tags.
	scanner := mx.NewScanner()
	tags, err := scanner.ScanFile(input.FilePath)
	if err != nil {
		slog.Warn("failed to scan file for MX tags",
			"path", input.FilePath,
			"error", err,
		)
		return
	}

	// Compare with existing sidecar tags and update index.
	// This tracks MX tag deltas across file edits for monitoring and validation.
	//
	// REQ-SEC3-003: 사이드카 stateDir은 raw input.CWD가 아니라 해소된 루트에서
	// 유도하며, 그 대상이 루트를 탈출하지 않음을 쓰기 전에 봉쇄 검증한다.
	stateDir := filepath.Join(root, ".moai", "state")
	if !pathContainedIn(root, stateDir) {
		slog.Warn("file_changed async: sidecar stateDir escapes project root, skipping sidecar update (containment)",
			"path", input.FilePath,
			"state_dir", stateDir,
			"project_root", root,
		)
		return
	}
	manager := mx.NewManager(stateDir)
	if _, uerr := manager.UpdateFile(input.FilePath, tags); uerr != nil {
		slog.Warn("failed to update MX sidecar",
			"path", input.FilePath,
			"error", uerr,
		)
	}

	// Format summary message and log for observability.
	// Note: in async mode the SystemMessage is NOT returned to Claude Code
	// (REQ-HAE-001 design — main response is fixed empty payload). The
	// message is logged at info level for observability.
	if msg := h.formatTagDelta(input.FilePath, tags); msg != "" {
		slog.Info("mx tag delta (async)", "summary", msg)
	}
}

// formatTagDelta creates a summary message for MX tag changes.
func (h *fileChangedHandler) formatTagDelta(filePath string, tags []mx.Tag) string {
	if len(tags) == 0 {
		return ""
	}

	// Count tags by kind
	counts := make(map[mx.TagKind]int)
	for _, tag := range tags {
		counts[tag.Kind]++
	}

	var sb strings.Builder
	sb.WriteString(filePath)
	sb.WriteByte(':')
	for kind, count := range counts {
		fmt.Fprintf(&sb, " %s=%d", kind, count)
	}

	return "MX tag delta on " + sb.String()
}

// pathContainedIn reports whether target resolves to a path inside root
// (root itself counts as contained). It is the C-F1 root-relative containment
// guard for SPEC-SEC-HARDEN-003: both paths are made absolute and cleaned, then
// filepath.Rel is used to detect a `..` escape. Cross-platform via filepath
// (NFR-SEC3-005). Fail-closed: any resolution error returns false (NFR-SEC3-004).
func pathContainedIn(root, target string) bool {
	if root == "" || target == "" {
		return false
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(absRoot, absTarget)
	if err != nil {
		return false
	}
	// rel == "." means target == root; a leading ".." (or bare "..") escapes root.
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return false
	}
	return true
}
