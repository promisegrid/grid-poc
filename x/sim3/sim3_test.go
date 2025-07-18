package main

import "testing"

// TestSimulationTrade verifies that a trade occurs during the simulation.
// With arbitrage enabled, intermediaries modify the BID such that a BID of 10
// becomes 9 before reaching the seller, and then 8 for the last hop.
// The seller responds with a CONFIRM using the received bid amount (8).
// Intermediaries then propagate confirmation using the original upstream bid:
// Carol sends CONFIRM with 9 (from Bob's original bid) to Bob, and Bob sends
// CONFIRM with 10 (from Alice's bid) to Alice. Therefore, after the trade,
// we expect the buyer (Alice) to have a liability entry for 10 ALICE and the
// seller (Dave) to have an asset entry for 10 DAVE.
func TestSimulationTrade(t *testing.T) {
	alice, _, _, dave := RunSimulation()
	// With arbitrage, the final trade executes at a confirm price of 10.
	expectedBuyerLiability := 10.0
	expectedSellerAsset := 10.0

	buyerLiability := alice.Liabilities[alice.Currency]
	sellerAsset := dave.Assets[dave.Currency]

	if buyerLiability != expectedBuyerLiability {
		t.Errorf("Expected Alice liability in %s to be %.2f, got %.2f",
			alice.Currency, expectedBuyerLiability, buyerLiability)
	}
	if sellerAsset != expectedSellerAsset {
		t.Errorf("Expected Dave asset in %s to be %.2f, got %.2f",
			dave.Currency, expectedSellerAsset, sellerAsset)
	}
}
