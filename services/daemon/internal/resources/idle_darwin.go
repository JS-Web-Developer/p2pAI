//go:build darwin

package resources

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

var hidIdleRe = regexp.MustCompile(`"HIDIdleTime"\s*=\s*(\d+)`)

// idleTime lê HIDIdleTime (nanossegundos) do IOKit via ioreg.
func idleTime() (time.Duration, error) {
	out, err := exec.Command("ioreg", "-c", "IOHIDSystem", "-d", "4").Output()
	if err != nil {
		return 0, err
	}
	m := hidIdleRe.FindSubmatch(out)
	if m == nil {
		return 0, fmt.Errorf("HIDIdleTime não encontrado")
	}
	ns, err := strconv.ParseInt(string(m[1]), 10, 64)
	if err != nil {
		return 0, err
	}
	return time.Duration(ns), nil
}
