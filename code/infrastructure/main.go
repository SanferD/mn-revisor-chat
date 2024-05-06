package main

import (
	"code/infrastructure/watchers"
	"fmt"
	"time"
)

func main() {
	bgInterruptWatcher := watchers.InitializeBackgroundInterruptWatcher()
	bgInterruptWatcher.StartBackgroundWatcher()
	for !bgInterruptWatcher.IsInterrupted() {
		fmt.Println("not-interrupted")
		time.Sleep(time.Second)
	}
	fmt.Println("interrupted")
}
