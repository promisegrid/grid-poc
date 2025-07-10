package main

import (
	"fmt"
	"strings"
)

// Message represents a bid or ask message in the simulation.
// Every message must be either a bid or an ask and include a personal
// currency symbol and an amount.
type Message struct {
	Type    string   // "BID" or "ASK"
	Amount  float64  // bid or ask amount
	Symbol  string   // personal currency symbol (e.g. "$")
	From    string   // sender agent ID
	History []string // list of agent IDs that have handled the message
}

// Agent represents a simulation participant.
type Agent struct {
	ID       string
	Balance  float64
	Peers    []*Agent
	IsSeller bool // In our simulation, only one agent (Dave) is considered a
	// resource provider (seller).
	IsBuyer  bool // Only one agent (Alice) is the buyer.
}

// SendMessage sends a message to all peers, skipping those already in the
// message history to avoid loops.
func (a *Agent) SendMessage(msg Message) {
	// Append self to history.
	msg.History = append(msg.History, a.ID)
	for _, peer := range a.Peers {
		// Avoid sending to peers already in the history.
		if contains(msg.History, peer.ID) {
			continue
		}
		fmt.Printf("%s sends %s message (%.2f%s) to %s\n",
			a.ID, msg.Type, msg.Amount, msg.Symbol, peer.ID)
		peer.ReceiveMessage(msg, a)
	}
}

// ReceiveMessage processes an incoming message based on type and role.
// Sellers respond to BID messages by issuing an ASK; buyers accept acceptable
// ASK messages. Intermediate agents simply forward messages.
func (a *Agent) ReceiveMessage(msg Message, sender *Agent) {
	// Avoid reprocessing if already in history.
	if contains(msg.History, a.ID) {
		return
	}
	msg.History = append(msg.History, a.ID)

	// Seller behavior: on receiving a BID, respond with an ASK.
	if msg.Type == "BID" && a.IsSeller {
		// For this simulation, seller sets ask price lower than bid by 5.
		askPrice := msg.Amount - 5.0
		if askPrice < 0 {
			askPrice = 0
		}
		askMsg := Message{
			Type:    "ASK",
			Amount:  askPrice,
			Symbol:  msg.Symbol,
			From:    a.ID,
			History: []string{},
		}
		fmt.Printf("%s received BID from %s, responds with ASK (%.2f%s)\n",
			a.ID, msg.From, askMsg.Amount, askMsg.Symbol)
		a.SendMessage(askMsg)
		return
	}

	// Buyer behavior: on receiving an ASK, if the price is acceptable, execute
	// the trade.
	if msg.Type == "ASK" && a.IsBuyer {
		// For simplicity, buyer accepts the ASK if the price is less than or equal
		// to 50.
		if msg.Amount <= 50.0 {
			fmt.Printf("%s (buyer) received ASK from %s with price %.2f%s, trade executed!\n",
				a.ID, msg.From, msg.Amount, msg.Symbol)
			a.Balance -= msg.Amount
			// Find the seller in the network; for this simulation, we assume that
			// the seller is Dave. If a direct peer is found, credit his balance.
			seller := findAgentByID(a.Peers, msg.From)
			if seller != nil {
				seller.Balance += msg.Amount
				fmt.Printf("Trade settled: %s pays %.2f%s to %s\n",
					a.ID, msg.Amount, msg.Symbol, seller.ID)
			} else {
				fmt.Printf("Seller %s not found among direct peers.\n", msg.From)
			}
			return
		}
		fmt.Printf("%s (buyer) received high ASK price %.2f%s from %s, rejecting trade\n",
			a.ID, msg.Amount, msg.Symbol, msg.From)
		return
	}

	// Intermediate agent: forward the message along the network.
	fmt.Printf("%s received %s message from %s, forwarding...\n",
		a.ID, msg.Type, sender.ID)
	a.SendMessage(msg)
}

// contains returns true if s is in the slice.
func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// findAgentByID returns the first agent in agents with a matching ID.
func findAgentByID(agents []*Agent, id string) *Agent {
	for _, agent := range agents {
		if agent.ID == id {
			return agent
		}
	}
	return nil
}

// RunSimulation initializes the four agents and simulates a double-auction
// message pass using personal currencies across a multi-hop network.
// The scenario: Alice, Bob, and Carol are peers; Bob, Carol, and Dave are peers.
// Alice and Dave cannot communicate directly. Each message is relayed locally.
func RunSimulation() (alice, bob, carol, dave *Agent) {
	// Create agents with an initial balance of 100.
	alice = &Agent{ID: "Alice", Balance: 100.0, IsBuyer: true}
	bob = &Agent{ID: "Bob", Balance: 100.0}
	carol = &Agent{ID: "Carol", Balance: 100.0}
	dave = &Agent{ID: "Dave", Balance: 100.0, IsSeller: true}

	// Set up peer connections:
	// Alice, Bob, and Carol are mutually connected.
	// Bob, Carol, and Dave are mutually connected.
	alice.Peers = []*Agent{bob, carol}
	bob.Peers = []*Agent{alice, carol, dave}
	carol.Peers = []*Agent{alice, bob, dave}
	dave.Peers = []*Agent{bob, carol}

	// Alice initiates the auction by sending a BID message.
	bidMsg := Message{
		Type:    "BID",
		Amount:  50.0,
		Symbol:  "$",
		From:    alice.ID,
		History: []string{},
	}
	alice.SendMessage(bidMsg)
	return alice, bob, carol, dave
}

// simulateAuction runs the simulation and prints final agent balances.
func simulateAuction() {
	alice, bob, carol, dave := RunSimulation()
	fmt.Println("\nFinal Balances:")
	// Trim any trailing zeros for nicer display.
	fmt.Printf("Alice: %s\n", fmt.Sprintf("%.2f$", alice.Balance))
	fmt.Printf("Bob: %s\n", fmt.Sprintf("%.2f$", bob.Balance))
	fmt.Printf("Carol: %s\n", fmt.Sprintf("%.2f$", carol.Balance))
	fmt.Printf("Dave: %s\n", fmt.Sprintf("%.2f$", dave.Balance))
}

func main() {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("Starting Auction Simulation")
	fmt.Println(strings.Repeat("=", 70))
	simulateAuction()
}
