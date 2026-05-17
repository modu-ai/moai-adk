package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	// @MX:WARN: [AUTO] bwrap --unshare-all용 user namespace 가용성은 OS 커널 설정에 의존
	// @MX:REASON: 일부 Linux 환경(컨테이너 내부, 오래된 커널)에서 사용자 네임스페이스가
	//             비활성화되어 있으면 bwrap이 실패할 수 있다. Available() 검사가 필수.
)

// BubblewrapBackend implements SandboxBackend for Linux using bwrap.
// It requires the bwrap binary in PATH and user namespace support in the kernel.
type BubblewrapBackend struct{}

// NewBubblewrapBackend returns a new BubblewrapBackend.
func NewBubblewrapBackend() *BubblewrapBackend {
	return &BubblewrapBackend{}
}

// Available reports whether bwrap is installed and usable on the current host.
func (b *BubblewrapBackend) Available() bool {
	_, err := exec.LookPath("bwrap")
	if err != nil {
		return false
	}

	// 최소 probe: bwrap --version 실행 (user-namespace 가용성 간접 검증)
	ctx, cancel := context.WithTimeout(context.Background(), bwrapProbeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bwrap", "--version")
	cmd.Env = []string{} // 최소 환경
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	// "bwrap N.M.P" 형식 확인 — 단순 존재 확인
	return len(out) > 0
}

// Exec runs cmd inside a bubblewrap sandbox with the given options.
//
// @MX:WARN: [AUTO] buildArgs는 보안 크리티컬 함수 — arg 순서/값 변경 시 sandbox 격리 파괴
// @MX:REASON: bwrap argument ordering matters: --unshare-all must precede all other flags;
//             --bind before --ro-bind; -- separator required before user command.
func (b *BubblewrapBackend) Exec(opts SandboxOptions, cmd []string) ([]byte, error) {
	if !b.Available() {
		return nil, ErrSandboxBackendUnavailable
	}
	if len(cmd) == 0 {
		return nil, fmt.Errorf("sandbox exec: empty command")
	}

	maxBytes := opts.MaxOutputBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxOutputBytes
	}

	// 프로파일 검증 (argument generation)
	baseArgs, err := GenerateBwrapArgs(opts)
	if err != nil {
		return nil, fmt.Errorf("sandbox: generate bwrap args: %w", err)
	}

	// 환경 변수 스크러빙
	env := ScrubEnv(os.Environ(), opts.EnvPassthrough)

	// 최종 args: baseArgs + -- + cmd
	allArgs := append(baseArgs, "--")
	allArgs = append(allArgs, cmd...)

	var buf bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), execTimeout)
	defer cancel()

	bwrapCmd := exec.CommandContext(ctx, "bwrap", allArgs...)
	bwrapCmd.Stdout = &limitedWriter{buf: &buf, limit: maxBytes}
	bwrapCmd.Stderr = bwrapCmd.Stdout
	bwrapCmd.Env = env

	runErr := bwrapCmd.Run()

	output := buf.Bytes()
	if int64(len(output)) >= maxBytes {
		// 출력이 잘림 — ErrSandboxOutputTruncated와 함께 반환
		return output[:maxBytes], fmt.Errorf("%w: output exceeded %d bytes",
			ErrSandboxOutputTruncated, maxBytes)
	}

	return output, runErr
}

// Profile returns the bwrap argument string that would be used for opts.
func (b *BubblewrapBackend) Profile(opts SandboxOptions) (string, error) {
	args, err := GenerateBwrapArgs(opts)
	if err != nil {
		return "", err
	}

	result := "bwrap"
	for _, a := range args {
		result += " " + a
	}
	result += " -- <cmd>"
	return result, nil
}
