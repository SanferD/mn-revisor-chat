package watchers

import (
	"os"
	"os/signal"
	"sync"
)

type BackgroundWatcher struct {
	signalChan         chan os.Signal
	isInterrupted      bool
	isInterruptedMutex sync.Mutex
}

func InitializeBackgroundInterruptWatcher() *BackgroundWatcher {
	signalChan := make(chan os.Signal, 1)
	return &BackgroundWatcher{signalChan: signalChan, isInterrupted: false, isInterruptedMutex: sync.Mutex{}}
}

func (bw *BackgroundWatcher) StartBackgroundWatcher() {
	signal.Notify(bw.signalChan, os.Interrupt)
	go func() {
		// wait until someone hits cntl+C
		<-bw.signalChan

		// update the isInterrupted value
		bw.isInterruptedMutex.Lock()
		bw.isInterrupted = true
		bw.isInterruptedMutex.Unlock()
	}()
}

func (bw *BackgroundWatcher) ForceSIGTERM() {
	close(bw.signalChan)
}

func (bw *BackgroundWatcher) IsInterrupted() bool {
	bw.isInterruptedMutex.Lock()
	defer bw.isInterruptedMutex.Unlock()
	return bw.isInterrupted
}
