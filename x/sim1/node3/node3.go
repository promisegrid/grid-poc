package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"sim1/agent3"
	"sim1/kernel"
)

func main() {
	peer := flag.String("peer", "localhost:7272",
		"peer address for dialing")
	flag.Parse()

	// Create a kernel instance and configure it for dialing.
	k := kernel.NewKernel()
	k.SetPeer(*peer)
	err := k.Start(0)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Kernel start failed:", err)
		os.Exit(1)
	}

	// Create and register Agent3 with the kernel.
	a := agent3.NewAgent(k)
	k.AddAgent(a)

	fmt.Println("Node3 (hosting Agent3) running. Press Ctrl+C to exit...")
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
