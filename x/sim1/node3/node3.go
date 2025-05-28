package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"sim1/hello1"
	"sim1/kernel"
)

func main() {
	peer := flag.String("peer", "localhost:7272", "peer address for dialing")
	name := flag.String("name", "agent3", "agent name")
	flag.Parse()

	// Create a kernel instance and configure it for dialing.
	k := kernel.NewKernel()
	k.SetPeer(*peer)
	err := k.Start(0)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Kernel start failed:", err)
		os.Exit(1)
	}

	// Create and register a hello1 agent with the kernel using the agent name.
	a := hello1.NewAgent(k, *name)
	k.AddAgent(a)

	fmt.Println("Node3 (hosting hello1 agent with name", *name,
		") running. Press Ctrl+C to exit...")
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
