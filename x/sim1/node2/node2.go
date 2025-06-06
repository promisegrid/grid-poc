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
	port := flag.Int("port", 7272, "listen port for node2")
	name := flag.String("name", "agent2", "agent name")
	flag.Parse()

	// Create a kernel instance and start listening on the specified port.
	k := kernel.NewKernel()
	err := k.Start(*port)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Kernel start failed:", err)
		os.Exit(1)
	}

	// Create and register a hello1 agent with the kernel using the agent name.
	a := hello1.NewAgent(k, *name)
	k.AddAgent(a)

	fmt.Println("Node2 (hosting hello1 agent with name", *name,
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
