package main

import "testing"

// TestSimulationTrade verifies that a trade occurs during the simulation.
// With arbitrage enabled, intermediaries modify the BID such that a BID of 50
// becomes 49 before reaching the seller, and then 48 for the last hop.
// The seller responds with a CONFIRM using the received bid amount (48).
// Intermediaries then propagate confirmation using the original upstream bid:
// Carol sends CONFIRM with 49 (from Bob's original bid) to Bob, and Bob sends
// CONFIRM with 50 (from Alice's bid) to Alice. Therefore, we expect the buyer
// (Alice) to pay 50 and the seller (Dave) to receive that amount.
func TestSimulationTrade(t *testing.T) {
	alice, _, _, dave := RunSimulation()
	// With arbitrage, the final trade executes at a CONFIRM price of 50.
	expectedBuyerBalance := 100.0 - 50.0
	expectedSellerBalance := 100.0 + 50.0

	if alice.Balance != expectedBuyerBalance {
		t.Errorf("Expected Alice balance to be %.2f, got %.2f",
			expectedBuyerBalance, alice.Balance)
	}
	if dave.Balance != expectedSellerBalance {
		t.Errorf("Expected Dave balance to be %.2f, got %.2f",
			expectedSellerBalance, dave.Balance)
	}
}
