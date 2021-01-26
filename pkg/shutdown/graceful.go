package shutdown

import (
	"os"
	"os/signal"
)

func Gracefully() <-chan struct{} {
	ch := make(chan os.Signal)
	notify := make(chan struct{})

	signal.Notify(ch, os.Interrupt)

	go func() {
		defer close(notify)
		defer close(ch)

		<-ch
	}()

	return notify
}
