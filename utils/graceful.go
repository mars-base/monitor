package utils

import (
	"os"
	"os/signal"
	"syscall"
)

type GracefulExitCallback func() error

func AddGracefulExit(fn GracefulExitCallback) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM) // 15 term
	signal.Notify(sig, syscall.SIGINT)  // 2 ctrl+c

	// Block until a signal is received.
	<-sig

	fn() // run self defined function

	os.Exit(0)
}
