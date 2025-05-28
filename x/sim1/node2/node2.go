package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"sim1/agent2"
)

func main() {
	port := flag.Int("port", 7272, "listen port for node2")
	flag.Parse()

	stop, err := agent2.Start(*port)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting Agent2:", err)
		os.Exit(1)
	}

	fmt.Println("Node2 (hosting Agent2) running. Press Ctrl+C to exit...")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	stop()
}
