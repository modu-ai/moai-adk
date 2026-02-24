//go:build windows

package cli

import "syscall"

// initConsole enables UTF-8 console output on Windows by setting code page 65001.
// This prevents box-drawing and Unicode characters from appearing as mojibake
// in PowerShell or Command Prompt that default to Windows-1252 or CP437.
func initConsole() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleOutputCP := kernel32.NewProc("SetConsoleOutputCP")
	setConsoleCP := kernel32.NewProc("SetConsoleCP")

	const utf8CodePage = 65001
	setConsoleOutputCP.Call(utf8CodePage) //nolint:errcheck
	setConsoleCP.Call(utf8CodePage)       //nolint:errcheck
}
