//go:build windows

package cli

import (
	"errors"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func configureClaudeAuditProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}

func runClaudeAuditProcess(cmd *exec.Cmd) error {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return err
	}
	var info windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION
	info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err := windows.SetInformationJobObject(
		job,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info)),
	); err != nil {
		_ = windows.CloseHandle(job)
		return err
	}

	var (
		closeOnce sync.Once
		closeErr  error
	)
	closeJob := func() error {
		closeOnce.Do(func() { closeErr = windows.CloseHandle(job) })
		return closeErr
	}
	defer closeJob()

	cmd.Cancel = func() error {
		jobErr := closeJob() // KILL_ON_JOB_CLOSE terminates the whole descendant tree.
		var processErr error
		if cmd.Process != nil {
			processErr = cmd.Process.Kill() // Covers cancellation before job assignment.
		}
		if jobErr != nil {
			return jobErr
		}
		if processErr != nil && !errors.Is(processErr, os.ErrProcessDone) {
			return processErr
		}
		return nil
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	process, err := windows.OpenProcess(
		windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE,
		false,
		uint32(cmd.Process.Pid),
	)
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return err
	}
	defer windows.CloseHandle(process)
	if err := windows.AssignProcessToJobObject(job, process); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return err
	}
	return cmd.Wait()
}
