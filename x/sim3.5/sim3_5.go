package main

import (
	"fmt"
	"strings"
	"sync"
)

// Global list of agents and the exchange kernel.
var allAgents []*Agent
var Trent *Kernel

// Kernel represents the exchange that matches orders from buyers and sellers.
// It maintains order books for bids and asks. In this open market model, agents
// submit BID and ASK orders identified by unique order IDs. When orders are matched,
// the kernel acts as the exchange, executing a balanced double‐entry transaction.
// In each trade, one party debits an asset while crediting a liability, ensuring
// that the fundamental accounting equation Assets = Liabilities + Equity is satisfied.
type Kernel struct {
	agents map[string]*Agent
	bids   []Message
	asks   []Message
	mu     sync.Mutex
}

// NewKernel creates a new Kernel (exchange) instance.
func NewKernel() *Kernel {
	return &Kernel{
		agents: make(map[string]*Agent),
		bids:   []Message{},
		asks:   []Message{},
	}
}

// RegisterAgent registers an agent with the exchange.
func (k *Kernel) RegisterAgent(agent *Agent) {
	k.agents[agent.ID] = agent
}

// SubmitOrder processes an order (BID or ASK) submitted by an agent. It attempts to match
// the order with an opposing order in the order book. When a match is found, the trade
// is executed at the buyer's bid price and a balanced double-entry transaction is recorded.
// The buyer's transaction debits the asset account "TOKEN" and credits the liability account "CASH".
// Conversely, the seller's transaction debits the asset account "CASH" and credits the liability
// account "TOKEN". A trade confirmation message is then sent to both participants.
func (k *Kernel) SubmitOrder(order Message) {
	k.mu.Lock()
	defer k.mu.Unlock()

	switch order.Type {
	case "BID":
		// Add the bid order to the order book.
		k.bids = append(k.bids, order)
		// Attempt to match with an existing ask order.
		for _, ask := range k.asks {
			if order.Amount >= ask.Amount && order.Symbol == ask.Symbol {
				// A match is found; execute trade at the buyer's bid price.
				tradePrice := order.Amount
				buyer := k.agents[order.From]
				seller := k.agents[ask.From]
				// For double-entry accounting on the buyer:
				// Debit asset "TOKEN" and credit liability "CASH".
				buyer.Assets["TOKEN"] += tradePrice
				buyer.Liabilities["CASH"] += tradePrice
				// For the seller, debit asset "CASH" and credit liability "TOKEN".
				seller.Assets["CASH"] += tradePrice
				seller.Liabilities["TOKEN"] += tradePrice
				// Create a confirmation message from the exchange.
				confirmMsg := Message{
					Type:   "CONFIRM",
					Amount: tradePrice,
					Symbol: order.Symbol,
					From:   "Exchange",
				}
				buyer.ReceiveConfirm(confirmMsg)
				seller.ReceiveConfirm(confirmMsg)
				// Remove the matched ask and bid orders using their order IDs.
				k.removeAskOrder(ask.OrderID)
				k.removeBidOrder(order.OrderID)
				return
			}
		}
	case "ASK":
		// Add the ask order to the order book.
		k.asks = append(k.asks, order)
		// Attempt to match with an existing bid order.
		for _, bid := range k.bids {
			if bid.Amount >= order.Amount && bid.Symbol == order.Symbol {
				tradePrice := bid.Amount
				buyer := k.agents[bid.From]
				seller := k.agents[order.From]
				// For double-entry accounting on the buyer:
				// Debit asset "TOKEN" and credit liability "CASH".
				buyer.Assets["TOKEN"] += tradePrice
				buyer.Liabilities["CASH"] += tradePrice
				// For the seller, debit asset "CASH" and credit liability "TOKEN".
				seller.Assets["CASH"] += tradePrice
				seller.Liabilities["TOKEN"] += tradePrice
				confirmMsg := Message{
					Type:   "CONFIRM",
					Amount: tradePrice,
					Symbol: order.Symbol,
					From:   "Exchange",
				}
				buyer.ReceiveConfirm(confirmMsg)
				seller.ReceiveConfirm(confirmMsg)
				// Remove the matched bid and ask orders using their order IDs.
				k.removeBidOrder(bid.OrderID)
				k.removeAskOrder(order.OrderID)
				return
			}
		}
	}
}

// removeBidOrder removes a bid order from the order book identified by the given order ID.
func (k *Kernel) removeBidOrder(orderID string) {
	for i, bid := range k.bids {
		if bid.OrderID == orderID {
			k.bids = append(k.bids[:i], k.bids[i+1:]...)
			return
		}
	}
}

// removeAskOrder removes an ask order from the order book identified by the given order ID.
func (k *Kernel) removeAskOrder(orderID string) {
	for i, ask := range k.asks {
		if ask.OrderID == orderID {
			k.asks = append(k.asks[:i], k.asks[i+1:]...)
			return
		}
	}
}

// Message represents an order or trade confirmation in the exchange.
// The message type can be "BID", "ASK", or "CONFIRM". Each order message is identified
// by a unique OrderID.
type Message struct {
	OrderID string  // Unique identifier for the order
	Type    string  // "BID", "ASK", or "CONFIRM"
	Amount  float64 // Order amount or confirmed trade price
	Symbol  string  // Order symbol (e.g., "TOKEN")
	From    string  // Agent ID that submitted the order (or "Exchange" for CONFIRM)
}

// Agent represents a market participant. Agents hold their own balance sheets which
// display assets and liabilities. The balance sheet follows the double-entry accounting
// model where Assets = Liabilities + Equity. Equity is computed on demand as the difference
// between total assets and total liabilities.
type Agent struct {
	ID          string
	Assets      map[string]float64 // Ledger of assets by account name (e.g., "TOKEN", "CASH")
	Liabilities map[string]float64 // Ledger of liabilities by account name
}

// PrintBalanceSheet prints the agent's current balance sheet, showing their assets,
// liabilities, and computed equity (Assets - Liabilities).
func (a *Agent) PrintBalanceSheet() {
	totalAssets := 0.0
	totalLiabilities := 0.0

	assetsStr := ""
	for acct, amt := range a.Assets {
		assetsStr += fmt.Sprintf("%s: %.2f  ", acct, amt)
		totalAssets += amt
	}

	liabStr := ""
	for acct, amt := range a.Liabilities {
		liabStr += fmt.Sprintf("%s: %.2f  ", acct, amt)
		totalLiabilities += amt
	}

	equity := totalAssets - totalLiabilities
	fmt.Printf("Balance Sheet for %s -> Assets: [%s] Liabilities: [%s] Equity: %.2f\n",
		a.ID, assetsStr, liabStr, equity)
}

// SubmitOrder allows an agent to submit an order (BID or ASK) to the exchange.
// The order must include a unique OrderID.
func (a *Agent) SubmitOrder(order Message) {
	fmt.Printf("%s submits %s order (OrderID: %s, %.2f %s)\n", a.ID, order.Type, order.OrderID,
		order.Amount, order.Symbol)
	Trent.SubmitOrder(order)
}

// ReceiveConfirm processes a trade confirmation message from the exchange.
func (a *Agent) ReceiveConfirm(msg Message) {
	fmt.Printf("%s receives CONFIRM: Trade executed at price %.2f %s by %s\n",
		a.ID, msg.Amount, msg.Symbol, msg.From)
	a.PrintBalanceSheet()
}

// RunSimulation initializes four agents and simulates a basic open market trade.
// In this simulation, the buyer (Alice) submits a BID order and the seller (Dave)
// submits an ASK order. If the BID price is equal to or exceeds the ASK price and
// the order symbols match, the exchange matches the orders and executes a trade.
// The trade is executed as a balanced double-entry transaction following the rules:
// the buyer debits asset "TOKEN" and credits liability "CASH", while the seller debits
// asset "CASH" and credits liability "TOKEN".
func RunSimulation() (alice, bob, carol, dave *Agent) {
	// Initialize agents with empty ledger maps.
	alice = &Agent{
		ID:          "Alice",
		Assets:      make(map[string]float64),
		Liabilities: make(map[string]float64),
	}
	bob = &Agent{
		ID:          "Bob",
		Assets:      make(map[string]float64),
		Liabilities: make(map[string]float64),
	}
	carol = &Agent{
		ID:          "Carol",
		Assets:      make(map[string]float64),
		Liabilities: make(map[string]float64),
	}
	dave = &Agent{
		ID:          "Dave",
		Assets:      make(map[string]float64),
		Liabilities: make(map[string]float64),
	}

	// Initialize global agent list.
	allAgents = []*Agent{alice, bob, carol, dave}

	// Initialize the exchange kernel and register all agents.
	Trent = NewKernel()
	for _, agent := range allAgents {
		Trent.RegisterAgent(agent)
	}

	// Simulation: Alice (buyer) submits a BID order.
	bidMsg := Message{
		OrderID: "BID1",
		Type:    "BID",
		Amount:  10.0,
		Symbol:  "TOKEN",
		From:    alice.ID,
	}
	alice.SubmitOrder(bidMsg)

	// Simulation: Dave (seller) submits an ASK order.
	askMsg := Message{
		OrderID: "ASK1",
		Type:    "ASK",
		Amount:  8.0,
		Symbol:  "TOKEN",
		From:    dave.ID,
	}
	dave.SubmitOrder(askMsg)

	return alice, bob, carol, dave
}

func main() {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("Starting Open Market Simulation")
	fmt.Println(strings.Repeat("=", 70))
	RunSimulation()
}
