package main

import (
    "fmt"
    "strings"
)

// Global list of agents and the exchange kernel.
var allAgents []*Agent
var Trent *Kernel

// Kernel represents the exchange that matches orders from buyers and sellers.
// It maintains order books for bids and asks.
type Kernel struct {
    agents map[string]*Agent
    bids   []Message
    asks   []Message
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

// SubmitOrder processes an order (BID or ASK) submitted by an agent and
// attempts to match it with an opposing order in the order book. When a match
// is found, the trade is executed at the buyer’s bid price, and both the buyer’s
// liability and the seller’s asset are updated. A trade confirmation message is
// sent to both participants.
func (k *Kernel) SubmitOrder(order Message) {
    switch order.Type {
    case "BID":
        // Add the bid order to the order book.
        k.bids = append(k.bids, order)
        // Attempt to match with an existing ask order.
        for i, ask := range k.asks {
            if order.Amount >= ask.Amount && order.Symbol == ask.Symbol {
                // A match is found; the trade price is determined by the buyer's bid.
                tradePrice := order.Amount
                buyer := k.agents[order.From]
                seller := k.agents[ask.From]
                // Update the ledgers of the buyer and seller.
                buyer.Liabilities[buyer.Currency] += tradePrice
                seller.Assets[seller.Currency] += tradePrice
                // Create a confirmation message from the exchange.
                confirmMsg := Message{
                    Type:   "CONFIRM",
                    Amount: tradePrice,
                    Symbol: order.Symbol,
                    From:   "Exchange",
                }
                buyer.ReceiveConfirm(confirmMsg)
                seller.ReceiveConfirm(confirmMsg)
                // Remove the matched ask from the ask order book.
                k.asks = append(k.asks[:i], k.asks[i+1:]...)
                // Remove the bid order since it has been matched.
                k.removeBidOrder(order.From)
                return
            }
        }
    case "ASK":
        // Add the ask order to the order book.
        k.asks = append(k.asks, order)
        // Attempt to match with an existing bid order.
        for i, bid := range k.bids {
            if bid.Amount >= order.Amount && bid.Symbol == order.Symbol {
                tradePrice := bid.Amount
                buyer := k.agents[bid.From]
                seller := k.agents[order.From]
                buyer.Liabilities[buyer.Currency] += tradePrice
                seller.Assets[seller.Currency] += tradePrice
                confirmMsg := Message{
                    Type:   "CONFIRM",
                    Amount: tradePrice,
                    Symbol: order.Symbol,
                    From:   "Exchange",
                }
                buyer.ReceiveConfirm(confirmMsg)
                seller.ReceiveConfirm(confirmMsg)
                // Remove the matched bid from the bid order book.
                k.bids = append(k.bids[:i], k.bids[i+1:]...)
                // Remove the ask order from the ask order book.
                k.removeAskOrder(order.From)
                return
            }
        }
    }
}

// removeBidOrder removes a bid order from the order book for the given agent ID.
func (k *Kernel) removeBidOrder(agentID string) {
    for i, bid := range k.bids {
        if bid.From == agentID {
            k.bids = append(k.bids[:i], k.bids[i+1:]...)
            return
        }
    }
}

// removeAskOrder removes an ask order from the order book for the given agent ID.
func (k *Kernel) removeAskOrder(agentID string) {
    for i, ask := range k.asks {
        if ask.From == agentID {
            k.asks = append(k.asks[:i], k.asks[i+1:]...)
            return
        }
    }
}

// Message represents an order or trade confirmation in the exchange.
// The message type can be "BID", "ASK", or "CONFIRM".
type Message struct {
    Type   string  // "BID", "ASK", or "CONFIRM"
    Amount float64 // Order amount or confirmed trade price
    Symbol string  // Order symbol (e.g., "TOKEN")
    From   string  // Agent ID that submitted the order (or "Exchange" for CONFIRM)
}

// Agent represents a market participant. Agents hold their own currency,
// asset balances, and liability records. In this open market model, agents
// interact solely with the exchange kernel.
type Agent struct {
    ID          string
    Currency    string             // Personal currency identifier (e.g., "ALICE")
    Assets      map[string]float64 // Ledger of assets by currency
    Liabilities map[string]float64 // Ledger of liabilities by currency
}

// PrintBalanceSheet prints the agent's current balance sheet, showing their
// assets, liabilities, and net worth.
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

// SubmitOrder allows an agent to submit an order (BID or ASK) to the exchange.
func (a *Agent) SubmitOrder(order Message) {
    fmt.Printf("%s submits %s order (%.2f %s)\n", a.ID, order.Type,
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
func RunSimulation() (alice, bob, carol, dave *Agent) {
    alice = &Agent{
        ID:          "Alice",
        Currency:    "ALICE",
        Assets:      make(map[string]float64),
        Liabilities: make(map[string]float64),
    }
    bob = &Agent{
        ID:          "Bob",
        Currency:    "BOB",
        Assets:      make(map[string]float64),
        Liabilities: make(map[string]float64),
    }
    carol = &Agent{
        ID:          "Carol",
        Currency:    "CAROL",
        Assets:      make(map[string]float64),
        Liabilities: make(map[string]float64),
    }
    dave = &Agent{
        ID:          "Dave",
        Currency:    "DAVE",
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
        Type:   "BID",
        Amount: 10.0,
        Symbol: "TOKEN",
        From:   alice.ID,
    }
    alice.SubmitOrder(bidMsg)

    // Simulation: Dave (seller) submits an ASK order.
    askMsg := Message{
        Type:   "ASK",
        Amount: 8.0,
        Symbol: "TOKEN",
        From:   dave.ID,
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
