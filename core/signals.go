package core

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// HandleSignals listens for and captures signals used for orchestration
func (a *App) handleSignals() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGUSR1, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for signal := range sig {
			switch signal {
			case syscall.SIGUSR1:
				a.ToggleMaintenanceMode()
			case syscall.SIGTERM:
				a.Terminate()
			case syscall.SIGHUP:
				a.Reload()
			}
		}
	}()
}

// reapChildren cleans up zombies
// this section is borrows heavily from hashicorp/go-reap but wraps itself
// in a goroutine and doesn't bother with the various back-communication chans
func reapChildren(reapLock *sync.RWMutex) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGCHLD)
	go func() {
		// listen for signals on the channel until it closes
		for _ = range sig {
			func() {
				if reapLock != nil {
					reapLock.Lock()
					defer func() {
						reapLock.Unlock()
					}()
				}
			POLL:
				// only 1 SIGCHLD can be handled at a time from the channel,
				// so we need to allow for the possibility that multiple child
				// processes have terminated while one is already being reaped.
				var wstatus syscall.WaitStatus
				pid, err := syscall.Wait4(-1, &wstatus,
					syscall.WNOHANG|syscall.WUNTRACED|syscall.WCONTINUED,
					nil)
				switch err {
				case nil:
					if pid > 0 {
						goto POLL
					}
					return
				case syscall.ECHILD:
					// return to the outer loop and wait for another signal
					return
				case syscall.EINTR:
					goto POLL
				default:
					// return to the outer loop and wait for another signal
					return
				}
			}()
		}
	}()
}
