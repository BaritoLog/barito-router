package srvkit

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func AsyncGracefulShutdown(shutdownFunc func()) {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	go func() {
		sig := <-gracefulStop
		fmt.Printf("Caught signal: %+v\n", sig)
		shutdownFunc()
		os.Exit(0)
	}()
}

func GracefullShutdown(shutdownFunc func()) {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	sig := <-gracefulStop
	fmt.Printf("Caught signal: %+v\n", sig)
	shutdownFunc()
	os.Exit(0)
}
