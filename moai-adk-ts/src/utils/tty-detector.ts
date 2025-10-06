// @CODE:INIT-001 | SPEC: SPEC-INIT-001.md | TEST: __tests__/utils/tty-detector.test.ts
// Related: @CODE:CLI-INIT-001, @SPEC:INIT-001

/**
 * @file TTY detection utility for non-interactive environment support
 * @author MoAI Team
 * @tags @CODE:INIT-001:TTY
 */

/**
 * Check if TTY is available for interactive prompts
 * @returns True if both stdin and stdout have TTY, false otherwise
 * @description
 * Detects whether the current process is running in a TTY (interactive terminal) environment.
 * Returns false for:
 * - Claude Code (no TTY)
 * - CI/CD environments (no TTY)
 * - Docker containers (no TTY)
 * - Piped/redirected I/O
 *
 * Returns true for:
 * - Regular terminal sessions
 * - SSH sessions
 * - tmux/screen sessions
 */
export function isTTYAvailable(): boolean {
  try {
    // Check if both stdin and stdout are TTY
    // process.stdin.isTTY and process.stdout.isTTY are true in interactive terminals
    // They are undefined in non-interactive environments
    return process.stdin.isTTY === true && process.stdout.isTTY === true;
  } catch (_error) {
    // If any exception occurs (e.g., permission issues), safely fallback to false
    // This ensures the program doesn't crash and defaults to non-interactive mode
    return false;
  }
}
