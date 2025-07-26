package main

import (
	"fmt"
	"strings"
	"sync"
)

// Global list of agents and the exchange kernel.
var allAgents []*Agent
var Exchange *Kernel

// Kernel represents the exchange that matches orders from buyers and
// sellers. It maintains order books for bids and asks. In this open market
// model, agents submit BID and ASK orders identified by unique order IDs.
// When orders are matched, the kernel acts as the exchange, executing a
// balanced double‐entry transaction. In each trade, one party debits an asset
// while crediting a liability, ensuring that the fundamental accounting
// equation Assets = Liabilities + Equity is satisfied. In addition, each agent
// issues its own personal currency. The kernel maintains a ledger of these
// currencies and verifies that each agent's personal currency is unique.
type Kernel struct {
	agents             map[string]*Agent
	bids               []Message
	asks               []Message
	personalCurrencies map[string]bool
	mu                 sync.Mutex
}

// NewKernel creates a new Kernel (exchange) instance.
func NewKernel() *Kernel {
	return &Kernel{
		agents:             make(map[string]*Agent),
		bids:               []Message{},
		asks:               []Message{},
		personalCurrencies: make(map[string]bool),
	}
}

// RegisterAgent registers an agent with the exchange. It enforces that each
// agent's personal currency is unique.
func (k *Kernel) RegisterAgent(agent *Agent) {
	k.mu.Lock()
	defer k.mu.Unlock()
	// Ensure the agent's personal currency is unique.
	if _, exists := k.personalCurrencies[agent.PersonalCurrency]; exists {
		fmt.Printf("Error: personal currency %s already registered\n",
			agent.PersonalCurrency)
		return
	}
	k.personalCurrencies[agent.PersonalCurrency] = true
	k.agents[agent.ID] = agent
}

// SubmitOrder processes an order (BID or ASK) submitted by an agent.
// It attempts to match the order with an opposing order in the order book.
// For a trade to match, the Message.Symbol field must indicate the target
// personal currency, and if the GoodSymbol field is provided then both orders
// must match on GoodSymbol and GoodQty as well. When a match is found, the trade
// is executed as a bilateral swap: the buyer receives the seller's personal
// currency as an asset while incurring a liability in his own currency, and vice-versa
// for the seller. A trade confirmation message is then sent to both participants.
func (k *Kernel) SubmitOrder(order Message) {
	k.mu.Lock()
	defer k.mu.Unlock()

	switch order.Type {
	case "BID":
		// Add the bid order to the order book.
		k.bids = append(k.bids, order)
		// Attempt to match with an existing ask order.
		for _, ask := range k.asks {
			// Check if the good or service is involved; if either order specifies a
			// GoodSymbol then both must match in GoodSymbol and GoodQty.
			if order.GoodSymbol != "" || ask.GoodSymbol != "" {
				if order.GoodSymbol != ask.GoodSymbol ||
					order.GoodQty != ask.GoodQty {
					continue
				}
			}
			if order.Amount >= ask.Amount &&
				order.Symbol == ask.Symbol {
				// A match is found; execute trade.
				tradePrice := order.Amount
				buyer := k.agents[order.From]
				seller := k.agents[ask.From]
				// The trade is executed as a swap between the buyer's and
				// seller's personal currencies. The buyer receives the asset
				// of the target currency (seller's currency) and incurs a liability
				// in his own personal currency.
				// For the buyer:
				//   Debit asset: seller's personal currency.
				//   Credit liability: buyer's personal currency.
				buyer.Assets[order.Symbol] += tradePrice
				buyer.Liabilities[buyer.PersonalCurrency] += tradePrice
				// For the seller:
				//   Debit asset: buyer's personal currency.
				//   Credit liability: seller's personal currency.
				seller.Assets[buyer.PersonalCurrency] += tradePrice
				seller.Liabilities[seller.PersonalCurrency] += tradePrice
				// Create a confirmation message.
				confirmMsg := Message{
					Type:       "CONFIRM",
					Amount:     tradePrice,
					Symbol:     order.Symbol,
					From:       "Exchange",
					GoodSymbol: order.GoodSymbol,
					GoodQty:    order.GoodQty,
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
			// Check for matching goods if specified.
			if order.GoodSymbol != "" || bid.GoodSymbol != "" {
				if order.GoodSymbol != bid.GoodSymbol ||
					order.GoodQty != bid.GoodQty {
					continue
				}
			}
			if bid.Amount >= order.Amount &&
				bid.Symbol == order.Symbol {
				tradePrice := bid.Amount
				buyer := k.agents[bid.From]
				seller := k.agents[order.From]
				// The trade uses the target personal currency (order.Symbol).
				// For the buyer:
				//   Debit asset: seller's personal currency.
				//   Credit liability: buyer's personal currency.
				buyer.Assets[order.Symbol] += tradePrice
				buyer.Liabilities[buyer.PersonalCurrency] += tradePrice
				// For the seller:
				//   Debit asset: buyer's personal currency.
				//   Credit liability: seller's personal currency.
				seller.Assets[buyer.PersonalCurrency] += tradePrice
				seller.Liabilities[seller.PersonalCurrency] += tradePrice
				confirmMsg := Message{
					Type:       "CONFIRM",
					Amount:     tradePrice,
					Symbol:     order.Symbol,
					From:       "Exchange",
					GoodSymbol: order.GoodSymbol,
					GoodQty:    order.GoodQty,
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

// removeBidOrder removes a bid order from the order book identified by the
// given order ID.
func (k *Kernel) removeBidOrder(orderID string) {
	for i, bid := range k.bids {
		if bid.OrderID == orderID {
			k.bids = append(k.bids[:i], k.bids[i+1:]...)
			return
		}
	}
}

// removeAskOrder removes an ask order from the order book identified by the
// given order ID.
func (k *Kernel) removeAskOrder(orderID string) {
	for i, ask := range k.asks {
		if ask.OrderID == orderID {
			k.asks = append(k.asks[:i], k.asks[i+1:]...)
			return
		}
	}
}

// Message represents an order or trade confirmation in the exchange.
// The message type can be "BID", "ASK", or "CONFIRM". Each order message is
// identified by a unique OrderID. The Symbol field indicates the personal
// currency being traded. In addition, the GoodSymbol field indicates the
// good or service involved in the transaction and the GoodQty field specifies
// the quantity of the good.
type Message struct {
	OrderID    string  // Unique identifier for the order
	Type       string  // "BID", "ASK", or "CONFIRM"
	Amount     float64 // Order amount or confirmed trade price
	Symbol     string  // Target personal currency (e.g., seller's currency)
	From       string  // Agent ID that submitted the order (or "Exchange")
	GoodSymbol string  // Indicates the specific good or service being exchanged
	GoodQty    float64 // Quantity of the good or service being exchanged
}

// Agent represents a market participant. Agents hold their own balance
// sheets that display assets and liabilities. The balance sheet follows the
// double-entry accounting model where Assets = Liabilities + Equity. In addition,
// each agent issues its own personal currency used to transact on the exchange.
type Agent struct {
	ID               string
	PersonalCurrency string
	Assets           map[string]float64 // Ledger of assets by account name.
	Liabilities      map[string]float64 // Ledger of liabilities by account name.
}

// PrintBalanceSheet prints the agent's current balance sheet, showing their
// assets, liabilities, and computed equity (Assets - Liabilities).
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
// The order must include a unique OrderID. The Symbol field should indicate
// the target personal currency for the transaction. Optionally, GoodSymbol and
// GoodQty can specify the good or service being exchanged.
func (a *Agent) SubmitOrder(order Message) {
	fmt.Printf("%s submits %s order (OrderID: %s, %.2f %s)\n",
		a.ID, order.Type, order.OrderID, order.Amount, order.Symbol)
	Exchange.SubmitOrder(order)
}

// ReceiveConfirm processes a trade confirmation message from the exchange.
func (a *Agent) ReceiveConfirm(msg Message) {
	fmt.Printf("%s receives CONFIRM: Trade executed at price %.2f %s by %s\n",
		a.ID, msg.Amount, msg.Symbol, msg.From)
	a.PrintBalanceSheet()
}

// RunSimulation initializes four agents and simulates a basic open market trade.
// Each agent issues its own personal currency. In this simulation, the buyer
// (Alice) submits a BID order to acquire Dave's personal currency, and the seller
// (Dave) submits an ASK order offering his own currency. When the BID and ASK match,
// the exchange performs a balanced double-entry transaction: the buyer receives
// Dave's currency (asset) and accrues a liability in his own currency, while the seller
// receives Alice's currency (asset) and accrues a liability in his own currency.
func RunSimulation() (alice, bob, carol, dave *Agent) {
	// Initialize agents with empty ledger maps and assign unique personal
	// currencies. Here we set the personal currency to be the same as the ID.
	alice = &Agent{
		ID:               "Alice",
		PersonalCurrency: "Alice",
		Assets:           make(map[string]float64),
		Liabilities:      make(map[string]float64),
	}
	bob = &Agent{
		ID:               "Bob",
		PersonalCurrency: "Bob",
		Assets:           make(map[string]float64),
		Liabilities:      make(map[string]float64),
	}
	carol = &Agent{
		ID:               "Carol",
		PersonalCurrency: "Carol",
		Assets:           make(map[string]float64),
		Liabilities:      make(map[string]float64),
	}
	dave = &Agent{
		ID:               "Dave",
		PersonalCurrency: "Dave",
		Assets:           make(map[string]float64),
		Liabilities:      make(map[string]float64),
	}

	// Initialize global agent list.
	allAgents = []*Agent{alice, bob, carol, dave}

	// Initialize the exchange kernel and register all agents.
	Exchange = NewKernel()
	for _, agent := range allAgents {
		Exchange.RegisterAgent(agent)
	}

	// Simulation: Alice (buyer) submits a BID order. She indicates that she
	// wishes to acquire Dave's personal currency, so the Symbol is set to "Dave".
	// In this example, no specific good is traded so GoodSymbol is left empty.
	bidMsg := Message{
		OrderID: "BID1",
		Type:    "BID",
		Amount:  10.0,
		Symbol:  "Dave",
		From:    alice.ID,
	}
	alice.SubmitOrder(bidMsg)

	// Simulation: Dave (seller) submits an ASK order offering his own currency.
	askMsg := Message{
		OrderID: "ASK1",
		Type:    "ASK",
		Amount:  10.0,
		Symbol:  "Dave",
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
