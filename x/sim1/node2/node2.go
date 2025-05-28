package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"sim1/agent2"
	"sim1/kernel"
)

func main() {
	port := flag.Int("port", 7272, "listen port for node2")
	flag.Parse()

	// Create a kernel instance and start listening on the specified port.
	k := kernel.NewKernel()
	err := k.Start(*port)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Kernel start failed:", err)
		os.Exit(1)
	}

	// Create and register Agent2 with the kernel.
	a := agent2.NewAgent(k)
	k.AddAgent(a)

	fmt.Println("Node2 (hosting Agent2) running. Press Ctrl+C to exit...")
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-ctx.Done()
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	a.Stop()
	cancel()
	k.Stop()
}
