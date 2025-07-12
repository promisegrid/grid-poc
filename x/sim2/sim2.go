package main

import (
	"fmt"
	"strings"
)

var simulateArbitrage = true
var tradeExecuted bool = false

// Message represents a bid or ask message in the simulation.
// Every message must be either a bid or an ask and include a personal
// currency symbol and an amount.
type Message struct {
	Type    string   // "BID" or "ASK"
	Amount  float64  // bid or ask amount
	Symbol  string   // personal currency symbol (e.g. "ALICE")
	From    string   // sender agent ID
	History []string // list of agent IDs that have handled the message
}

// Agent represents a simulation participant.
type Agent struct {
	ID       string
	Currency string // personal currency (e.g. "ALICE")
	Balance  float64
	Peers    []*Agent
	IsSeller bool // Only Dave is the seller.
	IsBuyer  bool // Only Alice is the buyer.
}

// SendMessage sends a message to all peers, skipping those already in the
// message history to avoid loops. If the message's symbol is empty, it uses
// the sender's personal currency.
func (a *Agent) SendMessage(msg Message) {
	if msg.Symbol == "" {
		msg.Symbol = a.Currency
	}
	// Append self to history.
	msg.History = append(msg.History, a.ID)
	for _, peer := range a.Peers {
		// Avoid sending to peers already in the history.
		if contains(msg.History, peer.ID) {
			continue
		}
		fmt.Printf("%s sends %s message (%.2f %s) to %s\n", a.ID, msg.Type,
			msg.Amount, msg.Symbol, peer.ID)
		peer.ReceiveMessage(msg, a)
	}
}

// ReceiveMessage processes an incoming message based on type and role.
// Sellers respond to BID messages by issuing an ASK using their own currency;
// buyers accept acceptable ASK messages. Intermediate agents, which in a real
// network would perform arbitrage (i.e. they would not merely forward messages
// but create their own orders to profit from price differences), modify the BID
// by reducing its amount by 1 and changing the currency to their own before
// forwarding.
func (a *Agent) ReceiveMessage(msg Message, sender *Agent) {
	if contains(msg.History, a.ID) {
		return
	}
	msg.History = append(msg.History, a.ID)
	if msg.Type == "BID" {
		if a.IsSeller {
			askPrice := msg.Amount - 5.0
			if askPrice < 0 {
				askPrice = 0
			}
			askMsg := Message{
				Type:    "ASK",
				Amount:  askPrice,
				Symbol:  a.Currency,
				From:    a.ID,
				History: []string{},
			}
			fmt.Printf("%s received BID from %s, responds with ASK (%.2f%s)\n",
				a.ID, msg.From, askMsg.Amount, askMsg.Symbol)
			a.SendMessage(askMsg)
			return
		} else if !a.IsBuyer && simulateArbitrage {
			newBid := Message{
				Type:    "BID",
				Amount:  msg.Amount - 1,
				Symbol:  a.Currency,
				From:    a.ID,
				History: append([]string{}, msg.History...),
			}
			fmt.Printf("%s (intermediary) received BID from %s, arbitraging to "+
				"new BID: %.2f%s\n", a.ID, msg.From, newBid.Amount,
				newBid.Symbol)
			a.SendMessage(newBid)
			return
		}
		fmt.Printf("%s received BID message from %s, forwarding...\n",
			a.ID, sender.ID)
		a.SendMessage(msg)
	} else if msg.Type == "ASK" {
		if a.IsBuyer {
			if !tradeExecuted && msg.Amount <= 50.0 {
				fmt.Printf("%s (buyer) received ASK from %s with price %.2f%s, "+
					"trade executed!\n", a.ID, msg.From, msg.Amount,
					msg.Symbol)
				a.Balance -= msg.Amount
				seller := findAgentByID(a.Peers, msg.From)
				if seller != nil {
					seller.Balance += msg.Amount
					fmt.Printf("Trade settled: %s pays %.2f%s to %s\n",
						a.ID, msg.Amount, msg.Symbol, seller.ID)
				} else {
					fmt.Printf("Seller %s not found among direct peers.\n",
						msg.From)
				}
				tradeExecuted = true
			} else if !tradeExecuted && msg.Amount > 50.0 {
				fmt.Printf("%s (buyer) received high ASK price %.2f%s from %s, "+
					"rejecting trade\n", a.ID, msg.Amount, msg.Symbol, msg.From)
			}
			return
		}
		fmt.Printf("%s received ASK message from %s, forwarding...\n",
			a.ID, sender.ID)
		a.SendMessage(msg)
	} else {
		fmt.Printf("%s received unknown message type %s from %s\n",
			a.ID, msg.Type, sender.ID)
	}
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
// Scenario: Alice, Bob, and Carol are peers; Bob, Carol, and Dave are peers.
// Alice and Dave cannot communicate directly.
func RunSimulation() (alice, bob, carol, dave *Agent) {
	tradeExecuted = false
	alice = &Agent{ID: "Alice", Balance: 100.0, IsBuyer: true,
		Currency: "ALICE"}
	bob = &Agent{ID: "Bob", Balance: 100.0, Currency: "BOB"}
	carol = &Agent{ID: "Carol", Balance: 100.0, Currency: "CAROL"}
	dave = &Agent{ID: "Dave", Balance: 100.0, IsSeller: true,
		Currency: "DAVE"}

	// Set up peer connections:
	// Alice, Bob, and Carol are mutually connected.
	// Bob, Carol, and Dave are mutually connected.
	alice.Peers = []*Agent{bob, carol}
	bob.Peers = []*Agent{alice, carol, dave}
	carol.Peers = []*Agent{alice, bob, dave}
	dave.Peers = []*Agent{bob, carol}

	// Alice initiates the auction by sending a BID message with her currency.
	bidMsg := Message{
		Type:    "BID",
		Amount:  50.0,
		Symbol:  alice.Currency,
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
	// Display each agent's balance along with their personal currency.
	fmt.Printf("Alice: %.2f %s\n", alice.Balance, alice.Currency)
	fmt.Printf("Bob: %.2f %s\n", bob.Balance, bob.Currency)
	fmt.Printf("Carol: %.2f %s\n", carol.Balance, carol.Currency)
	fmt.Printf("Dave: %.2f %s\n", dave.Balance, dave.Currency)
}

func main() {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("Starting Auction Simulation")
	fmt.Println(strings.Repeat("=", 70))
	simulateAuction()
}
