package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	// @MX:WARN: [AUTO] sandbox-exec은 Apple에 의해 deprecated됨 (macOS 10.5 이후 작동)
	// @MX:REASON: Apple이 더 이상 sandbox-exec을 공식 지원하지 않으며, 향후 macOS 버전에서
	//             제거될 수 있다. v3.1+에서 App Sandbox entitlement 기반 대안 검토 예정.
)

// SeatbeltBackend implements SandboxBackend for macOS using sandbox-exec.
// It generates SBPL profiles and wraps commands with `sandbox-exec -p <profile>`.
type SeatbeltBackend struct{}

// NewSeatbeltBackend returns a new SeatbeltBackend.
func NewSeatbeltBackend() *SeatbeltBackend {
	return &SeatbeltBackend{}
}

// Available reports whether sandbox-exec is available at /usr/bin/sandbox-exec.
func (s *SeatbeltBackend) Available() bool {
	_, err := exec.LookPath("sandbox-exec")
	return err == nil
}

// Exec runs cmd inside a macOS seatbelt sandbox with the given options.
//
// @MX:WARN: [AUTO] execSandboxExec — SBPL profile는 exec 직전에 생성되며 파일로 저장되지 않음
// @MX:REASON: sandbox-exec의 -p flag는 인라인 프로파일을 받아들이므로 tmpfile 불필요.
//             그러나 프로파일이 매우 길면 arg list limit에 걸릴 수 있음.
//             현재 구현은 -p 사용; 필요시 -f (파일) 모드로 전환 가능.
func (s *SeatbeltBackend) Exec(opts SandboxOptions, cmd []string) ([]byte, error) {
	if !s.Available() {
		return nil, ErrSandboxBackendUnavailable
	}
	if len(cmd) == 0 {
		return nil, fmt.Errorf("sandbox exec: empty command")
	}

	maxBytes := opts.MaxOutputBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxOutputBytes
	}

	// SBPL 프로파일 생성
	profile, err := GenerateSBPL(opts)
	if err != nil {
		return nil, fmt.Errorf("sandbox: generate SBPL: %w", err)
	}

	// 환경 변수 스크러빙
	env := ScrubEnv(os.Environ(), opts.EnvPassthrough)

	// sandbox-exec -p <profile> <cmd...>
	execArgs := append([]string{"-p", profile}, cmd...)

	var buf bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), execTimeout)
	defer cancel()

	sbxCmd := exec.CommandContext(ctx, "sandbox-exec", execArgs...)
	sbxCmd.Stdout = &limitedWriter{buf: &buf, limit: maxBytes}
	sbxCmd.Stderr = sbxCmd.Stdout
	sbxCmd.Env = env

	runErr := sbxCmd.Run()

	output := buf.Bytes()
	if int64(len(output)) >= maxBytes {
		return output[:maxBytes], fmt.Errorf("%w: output exceeded %d bytes",
			ErrSandboxOutputTruncated, maxBytes)
	}

	return output, runErr
}

// Profile returns the SBPL profile that would be applied for opts.
func (s *SeatbeltBackend) Profile(opts SandboxOptions) (string, error) {
	return GenerateSBPL(opts)
}
