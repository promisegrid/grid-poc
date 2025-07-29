package main

import (
	"fmt"
	"strings"
)

// simulateArbitrage enables intermediaries to modify the bid.
var simulateArbitrage = true

// tradeExecuted prevents multiple trades from executing.
var tradeExecuted bool = false

// allAgents holds the list of all agents in the simulation for global
// lookup during ledger updates.
var allAgents []*Agent

// Message represents a bid or confirm message in the simulation.
// Every message must be either a BID or a CONFIRM and include a personal
// currency symbol and an amount. OrigBid carries the bid amount received
// from the upstream agent and is used by intermediaries when generating
// confirm messages.
type Message struct {
	Type    string   // "BID" or "CONFIRM"
	Amount  float64  // bid or confirm amount
	Symbol  string   // personal currency symbol (e.g. "ALICE")
	From    string   // sender agent ID
	History []string // list of agent IDs that have handled the message
	OrigBid float64  // original bid amount received from upstream (if any)
}

// Agent represents a simulation participant.
type Agent struct {
	ID          string
	Currency    string // personal currency (e.g. "ALICE")
	Balance     float64
	Assets      map[string]float64 // assets ledger by currency
	Liabilities map[string]float64 // liabilities ledger by currency
	Peers       []*Agent
	IsSeller    bool // Only Dave is the seller.
	IsBuyer     bool // Only Alice is the buyer.
	NextHop     *Agent
	PrevHop     *Agent
	upstreamBid float64 // bid amount received from upstream
}

// PrintBalanceSheet prints the agent's current balance sheet: assets,
// liabilities, and net worth (assets minus liabilities).
func (a *Agent) PrintBalanceSheet() {
	totalAssets := 0.0
	totalLiabilities := 0.0

	assetsStr := ""
	for curr, amt := range a.Assets {
		assetsStr += fmt.Sprintf("%s: %.2f  ", curr, amt)
		totalAssets += amt
	}

	liabStr := ""
	for curr, amt := range a.Liabilities {
		liabStr += fmt.Sprintf("%s: %.2f  ", curr, amt)
		totalLiabilities += amt
	}

	netWorth := totalAssets - totalLiabilities
	fmt.Printf("Balance Sheet for %s -> Assets: [%s] Liabilities: [%s] "+
		"Net Worth: %.2f\n", a.ID, assetsStr, liabStr, netWorth)
}

// SendBidMessage sends a BID message to the next agent in the chain.
func (a *Agent) SendBidMessage(msg Message) {
	if a.NextHop == nil {
		return
	}
	// Append own ID to history.
	msg.History = append(msg.History, a.ID)
	fmt.Printf("%s sends %s message (%.2f %s) to %s\n",
		a.ID, msg.Type, msg.Amount, msg.Symbol, a.NextHop.ID)
	a.PrintBalanceSheet()
	a.NextHop.ReceiveMessage(msg, a)
}

// SendConfirmMessage sends a CONFIRM message to the previous agent in the
// chain; if there is no previous agent (i.e. the buyer), it processes the
// final confirmation.
func (a *Agent) SendConfirmMessage(msg Message) {
	if a.PrevHop == nil {
		a.ReceiveFinalConfirm(msg)
		return
	}
	// Append own ID to history.
	msg.History = append(msg.History, a.ID)
	fmt.Printf("%s sends %s message (%.2f %s) to %s\n",
		a.ID, msg.Type, msg.Amount, msg.Symbol, a.PrevHop.ID)
	a.PrintBalanceSheet()
	a.PrevHop.ReceiveMessage(msg, a)
}

// ReceiveMessage processes an incoming message based on its type and the role of
// the agent. For BID messages, intermediaries arbitrage by subtracting 1 from the
// incoming bid and storing the upstream bid for later use in CONFIRM. The seller
// responds to a BID with a CONFIRM using the exact bid amount. CONFIRM messages are
// forwarded backward along the chain with intermediaries replacing the bid amount with
// the upstream bid value.
func (a *Agent) ReceiveMessage(msg Message, sender *Agent) {
	// Prevent processing the same message more than once.
	if contains(msg.History, a.ID) {
		return
	}
	if msg.Type == "BID" {
		// For a BID message, if this agent is the seller, respond with CONFIRM.
		if a.IsSeller {
			// Seller uses the bid's amount (using the bid currency).
			confirmMsg := Message{
				Type:    "CONFIRM",
				Amount:  msg.Amount,
				Symbol:  msg.Symbol, // Use the bid's currency symbol.
				From:    a.ID,
				History: []string{a.ID},
			}
			fmt.Printf("%s received BID from %s, responds with CONFIRM (%.2f %s)\n",
				a.ID, sender.ID, confirmMsg.Amount, confirmMsg.Symbol)
			a.PrintBalanceSheet()
			a.SendConfirmMessage(confirmMsg)
			return
		} else if !a.IsBuyer && simulateArbitrage {
			// Intermediate agent: store the upstream bid amount.
			a.upstreamBid = msg.Amount
			// Arbitrage: subtract 1 from the incoming bid.
			newBidAmount := msg.Amount - 1
			newBid := Message{
				Type:    "BID",
				Amount:  newBidAmount,
				Symbol:  a.Currency, // Use own currency for new BID.
				From:    a.ID,
				History: append([]string{}, msg.History...),
				OrigBid: msg.Amount, // Preserve the upstream bid.
			}
			fmt.Printf("%s (intermediary) received BID from %s, arbitraging to "+
				"new BID: %.2f %s\n", a.ID, sender.ID, newBid.Amount,
				newBid.Symbol)
			a.PrintBalanceSheet()
			a.SendBidMessage(newBid)
			return
		}
		// If buyer or not eligible, simply forward the BID.
		fmt.Printf("%s received BID message from %s, forwarding...\n",
			a.ID, sender.ID)
		a.PrintBalanceSheet()
		a.SendBidMessage(msg)
	} else if msg.Type == "CONFIRM" {
		if a.IsBuyer {
			// Buyer processes the final CONFIRM.
			a.ReceiveFinalConfirm(msg)
			return
		}
		// Intermediate agent: generate a new CONFIRM using the stored upstream
		// bid amount and the currency of the previous hop.
		newConfirm := Message{
			Type:   "CONFIRM",
			Amount: a.upstreamBid,
			// The confirm uses the upstream agent's currency.
			Symbol: a.PrevHop.Currency,
			From:   msg.From,
			// Start history with current agent for the backward journey.
			History: []string{a.ID},
		}
		fmt.Printf("%s processed CONFIRM message from %s, generating new "+
			"CONFIRM with price %.2f %s\n", a.ID, sender.ID, newConfirm.Amount,
			newConfirm.Symbol)
		a.PrintBalanceSheet()
		a.SendConfirmMessage(newConfirm)
	} else {
		fmt.Printf("%s received unknown message type %s from %s\n",
			a.ID, msg.Type, sender.ID)
	}
}

// ReceiveFinalConfirm is called by the buyer when no previous agent exists.
// It finalizes the trade by updating the agents' balance sheets using double-entry
// accounting. The buyer records a liability in their own currency, while the seller
// records an asset in their own currency.
func (a *Agent) ReceiveFinalConfirm(msg Message) {
	if !tradeExecuted {
		fmt.Printf("%s (buyer) received final CONFIRM with price %.2f %s, trade "+
			"executed!\n", a.ID, msg.Amount, msg.Symbol)
		// Find the seller in the simulation.
		seller := findSeller(allAgents)
		if seller != nil {
			// Buyer creates a liability in their own currency.
			a.Liabilities[a.Currency] += msg.Amount
			// Seller recognizes an asset in their own currency.
			seller.Assets[seller.Currency] += msg.Amount
			fmt.Printf("Trade ledger updated: %s records liability of %.2f %s, "+
				"%s records asset of %.2f %s\n", a.ID, msg.Amount,
				a.Currency, seller.ID, msg.Amount, seller.Currency)
			a.PrintBalanceSheet()
			seller.PrintBalanceSheet()
		} else {
			fmt.Printf("Seller not found in the simulation.\n")
		}
		tradeExecuted = true
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

// findSeller returns the first agent in agents who is marked as the seller.
func findSeller(agents []*Agent) *Agent {
	for _, agent := range agents {
		if agent.IsSeller {
			return agent
		}
	}
	return nil
}

// RunSimulation initializes the four agents and simulates a double-auction
// message pass using personal currencies across a multi-hop network.
// Scenario: Alice, Bob, and Carol form a BID chain and Bob, Carol, and Dave form
// the corresponding chain for forwarding. Alice and Dave cannot communicate
// directly.
// This simulation reflects the design example where Alice initiates a BID of
// 10 ALICE, Bob arbitrages to 9 BOB, Carol arbitrages to 8 CAROL, and Dave, the
// seller, accepts the bid.
func RunSimulation() (alice, bob, carol, dave *Agent) {
	tradeExecuted = false
	alice = &Agent{
		ID:          "Alice",
		Balance:     0.0,
		IsBuyer:     true,
		Currency:    "ALICE",
		Assets:      make(map[string]float64),
		Liabilities: make(map[string]float64),
	}
	bob = &Agent{
		ID:          "Bob",
		Balance:     0.0,
		Currency:    "BOB",
		Assets:      make(map[string]float64),
		Liabilities: make(map[string]float64),
	}
	carol = &Agent{
		ID:          "Carol",
		Balance:     0.0,
		Currency:    "CAROL",
		Assets:      make(map[string]float64),
		Liabilities: make(map[string]float64),
	}
	dave = &Agent{
		ID:          "Dave",
		Balance:     0.0,
		IsSeller:    true,
		Currency:    "DAVE",
		Assets:      make(map[string]float64),
		Liabilities: make(map[string]float64),
	}

	// Set up peer connections (full mesh for potential lookups).
	alice.Peers = []*Agent{bob, carol}
	bob.Peers = []*Agent{alice, carol, dave}
	carol.Peers = []*Agent{alice, bob, dave}
	dave.Peers = []*Agent{bob, carol}

	// Define the chain order using NextHop and PrevHop.
	// Chain: Alice -> Bob -> Carol -> Dave.
	alice.NextHop = bob
	alice.PrevHop = nil

	bob.NextHop = carol
	bob.PrevHop = alice

	carol.NextHop = dave
	carol.PrevHop = bob

	dave.NextHop = nil
	dave.PrevHop = carol

	// Initialize global agent list.
	allAgents = []*Agent{alice, bob, carol, dave}

	// Alice initiates the auction by sending a BID message with her currency.
	bidMsg := Message{
		Type:    "BID",
		Amount:  10.0,
		Symbol:  alice.Currency,
		From:    alice.ID,
		History: []string{},
	}
	alice.SendBidMessage(bidMsg)
	return alice, bob, carol, dave
}

// simulateAuction runs the simulation and prints final agent ledger
// balance sheets.
func simulateAuction() {
	alice, bob, carol, dave := RunSimulation()
	fmt.Println("\nFinal Ledger Balance Sheets:")
	alice.PrintBalanceSheet()
	bob.PrintBalanceSheet()
	carol.PrintBalanceSheet()
	dave.PrintBalanceSheet()
}

func main() {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("Starting Auction Simulation")
	fmt.Println(strings.Repeat("=", 70))
	simulateAuction()
}
