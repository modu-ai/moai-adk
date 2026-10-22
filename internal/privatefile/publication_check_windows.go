package privatefile

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

func checkPublicationHandleNative(f *os.File) error {
	h := windows.Handle(f.Fd())
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		return err
	}
	if info.NumberOfLinks != 1 {
		return fmt.Errorf("private file: multiple hard links")
	}
	return checkHandle(h)
}
