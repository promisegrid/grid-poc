package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
)

const (
	TopicName      = "/x/promisegrid/1.0.0"
	CheckInterval  = 5 * time.Second
	IPFSAPIDefault = "localhost:5001"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to local IPFS daemon
	sh := shell.NewShell(IPFSAPIDefault)
	if !sh.IsUp() {
		fmt.Println("IPFS daemon not running. Start it first with 'ipfs daemon'")
		os.Exit(1)
	}

	// Get node identity
	id, err := sh.ID()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Connected to IPFS node %s\n", id.ID)

	// Subscribe to topic
	sub, err := sh.PubSubSubscribe(TopicName)
	if err != nil {
		panic(err)
	}
	defer sub.Cancel()

	// Start message receiver
	go func() {
		for {
			msg, err := sub.Next() // Fixed: removed context parameter
			if err != nil {
				if ctx.Err() == nil {
					fmt.Printf("Subscription error: %v\n", err)
				}
				return
			}
			fmt.Printf("\n[%s] From %s:\n%s\n",
				time.Now().Format("15:04:05"),
				shortID(msg.From.String()), // Fixed: added String() conversion
				string(msg.Data))
		}
	}()

	// Start message sender
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("Enter messages (press enter to send):")
		for scanner.Scan() {
			text := scanner.Text()
			if text == "" {
				continue
			}
			if err := sh.PubSubPublish(TopicName, text); err != nil {
				fmt.Printf("Error sending message: %v\n", err)
			}
		}
	}()

	// Wait for exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	fmt.Println("\nShutting down...")
}

func shortID(id string) string {
	if len(id) < 8 {
		return id
	}
	return id[len(id)-8:]
}
