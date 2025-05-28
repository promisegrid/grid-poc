package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"sim1/agent3"
)

func main() {
	peer := flag.String("peer", "localhost:7272", "peer address for dialing")
	flag.Parse()

	stop, err := agent3.Start(*peer)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting Agent3:", err)
		os.Exit(1)
	}

	fmt.Println("Node3 (hosting Agent3) running. Press Ctrl+C to exit...")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	stop()
}
