//go:build !darwin && !linux && !windows

package resources

import (
	"errors"
	"time"
)

func idleTime() (time.Duration, error) {
	return 0, errors.New("idle time não suportado nesta plataforma")
}
