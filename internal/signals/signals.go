package signals

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func CatchSigterm(isDone chan struct{}) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		fmt.Println("Waiting for signal")
		sig := <-signals
		fmt.Printf("GOT SIGNAL: %s\n", sig.String())
		isDone <- struct{}{}
	}()
}
