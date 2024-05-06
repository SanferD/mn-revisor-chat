package watchers

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestWatchers(t *testing.T) {
	t.Run("test interrupt watcher is not interrupted when not started", func(t *testing.T) {
		bgInterruptWatcher := InitializeBackgroundInterruptWatcher()
		assert.False(t, bgInterruptWatcher.IsInterrupted(), "expected to not be interrupted, but was interrupted")
	})

	t.Run("test interrupt watcher when started and interrupted is interrupted", func(t *testing.T) {
		bgInterruptWatcher := InitializeBackgroundInterruptWatcher()
		bgInterruptWatcher.StartBackgroundWatcher()
		time.Sleep(time.Millisecond * 40)
		assert.False(t, bgInterruptWatcher.IsInterrupted(), "expected that ")
		bgInterruptWatcher.ForceSIGTERM()
		time.Sleep(time.Millisecond * 40)
		assert.True(t, bgInterruptWatcher.IsInterrupted(), "expected that ")
	})
}
