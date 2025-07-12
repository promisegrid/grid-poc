package main

import "testing"

// TestSimulationTrade verifies that a trade occurs during the simulation.
// With arbitrage enabled, intermediaries modify the BID such that a BID of 50
// becomes 49 before reaching the seller. The seller then responds with an ASK
// price of 44 (i.e. 49 - 5). Therefore, we expect the buyer (Alice) to pay 44
// and the seller (Dave) to receive that amount.
func TestSimulationTrade(t *testing.T) {
	alice, _, _, dave := RunSimulation()
	// With arbitrage, the trade executes at an ASK price of 44.
	expectedBuyerBalance := 100.0 - 44.0
	expectedSellerBalance := 100.0 + 44.0

	if alice.Balance != expectedBuyerBalance {
		t.Errorf("Expected Alice balance to be %.2f, got %.2f",
			expectedBuyerBalance, alice.Balance)
	}
	if dave.Balance != expectedSellerBalance {
		t.Errorf("Expected Dave balance to be %.2f, got %.2f",
			expectedSellerBalance, dave.Balance)
	}
}
