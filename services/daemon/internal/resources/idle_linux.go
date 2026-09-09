//go:build linux

package resources

import (
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// idleTime usa xprintidle (X11). Wayland: sem fonte portátil; retorna erro.
func idleTime() (time.Duration, error) {
	out, err := exec.Command("xprintidle").Output()
	if err != nil {
		return 0, err
	}
	ms, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return 0, err
	}
	return time.Duration(ms) * time.Millisecond, nil
}
