//go:build windows

package resources

import (
	"syscall"
	"time"
	"unsafe"
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	getLastInputInfo = user32.NewProc("GetLastInputInfo")
	getTickCount     = kernel32.NewProc("GetTickCount")
)

type lastInputInfo struct {
	cbSize uint32
	dwTime uint32
}

func idleTime() (time.Duration, error) {
	var lii lastInputInfo
	lii.cbSize = uint32(unsafe.Sizeof(lii))
	r, _, err := getLastInputInfo.Call(uintptr(unsafe.Pointer(&lii)))
	if r == 0 {
		return 0, err
	}
	tick, _, _ := getTickCount.Call()
	return time.Duration(uint32(tick)-lii.dwTime) * time.Millisecond, nil
}
