//go:build windows

package player

import (
	"os/exec"
	"syscall"
)

// detach découple le processus enfant pour qu'il survive au serveur.
func detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x00000008} // DETACHED_PROCESS
}
