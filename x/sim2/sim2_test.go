package main

import "testing"

// TestSimulationTrade verifies that a trade occurs during the simulation.
// We expect the buyer (Alice) to pay the ask price (assumed to be 5 less than
// the bid of 50) and the seller (Dave) to receive that amount.
func TestSimulationTrade(t *testing.T) {
	alice, _, _, dave := RunSimulation()
	// In our simulation, Dave (seller) responds with an ASK price of 45.
	expectedBuyerBalance := 100.0 - 45.0
	expectedSellerBalance := 100.0 + 45.0

	if alice.Balance != expectedBuyerBalance {
		t.Errorf("Expected Alice balance to be %.2f, got %.2f",
			expectedBuyerBalance, alice.Balance)
	}
	if dave.Balance != expectedSellerBalance {
		t.Errorf("Expected Dave balance to be %.2f, got %.2f",
			expectedSellerBalance, dave.Balance)
	}
}
