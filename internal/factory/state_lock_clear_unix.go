//go:build !windows

// state_lock_clear_unix.go — the Unix stale-lock clear is gated out
// (SPEC-KANBAN-BOARD-001 REQ-KB-023, review finding F5 containment).
//
// state_lock_unix.go's own header records that the Unix substrate holds
// flock(2) on an open descriptor which the kernel releases on process exit,
// so a killed holder leaves an inert artifact — the fourteen orphaned
// spec-close-*.lock files in this repository's .moai/state/ are the worked
// example. The clear therefore exists for the Windows substrate, where the
// artifact IS the lock and a killed holder blocks every subsequent mutation
// permanently. Gating the clear to Windows removes the Unix-side window the
// M1 implementation opened (acquire flock, THEN record owner identity) —
// that gap is unreachable when the kernel drops the lock on exit — and
// touches the documented AP-29 residual not at all (it is a Windows-substrate
// residual and the Windows clear keeps its re-read mitigation).
//
// The gate is the absence of a Unix clear: the board-bound no-op that used to
// report it went with the board (SPEC-LAUNCHER-ENTRY-FLAGS-001 M6), and the
// integration lock's mutation lock mirrors this gate and its reason
// (integration_lock_mutation_unix.go).
package factory
