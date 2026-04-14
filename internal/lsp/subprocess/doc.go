// Package subprocess provides lifecycle management for language server subprocesses.
//
// It contains two main types:
//
//   - Launcher: spawns a language server binary using a ServerConfig and returns
//     a LaunchResult with isolated stdio pipes (REQ-LC-004, REQ-LC-005).
//
//   - Supervisor: monitors a running subprocess for crashes, supports graceful
//     shutdown via Signal, and forced termination via Kill (REQ-LC-005, REQ-LC-031).
//
// Binary lookup follows the system PATH. When the requested binary is not
// found, Launcher returns ErrBinaryNotFound; the caller is expected to log
// and skip the server (warn_and_skip pattern — no panics).
package subprocess
