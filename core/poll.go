package core

import (
	"sync"
	"time"
)

// Every `pollTime` seconds, run the `PollingFunc` function.
// Expect a bool on the quit channel to stop gracefully.
func (a *App) poll(pollable Pollable) chan bool {
	ticker := time.NewTicker(pollable.PollTime())
	quit := make(chan bool)
	go func() {
		for {
			select {
			case <-ticker.C:
				if !a.InMaintenanceMode() {
					pollable.PollAction(a.reapLock)
				}
			case <-quit:
				pollable.PollStop()
				return
			}
		}
	}()
	return quit
}

// Pollable is base abstraction for backends and services that support polling
type Pollable interface {
	PollTime() time.Duration
	PollAction(*sync.RWMutex)
	PollStop()
}
