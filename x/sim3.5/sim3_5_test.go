package main

import "testing"

// TestSimulationTrade verifies that a trade is executed in the open market
// simulation. In this test, Alice submits a BID order for 10 units, where she
// wishes to acquire Dave's personal currency. Dave submits an ASK order offering
// his personal currency. When the matching engine functions properly, the trade
// should execute at the buyer's bid price.
// For the buyer (Alice):
//   - Assets: the asset corresponding to "Dave" should increase by 10.
//   - Liabilities: the liability corresponding to "Alice" should increase by 10.
// For the seller (Dave):
//   - Assets: the asset corresponding to "Alice" should increase by 10.
//   - Liabilities: the liability corresponding to "Dave" should increase by 10.
func TestSimulationTrade(t *testing.T) {
	alice, _, _, dave := RunSimulation()

	expectedValue := 10.0

	// Check buyer's asset for target currency "Dave".
	buyerAsset := alice.Assets["Dave"]
	if buyerAsset != expectedValue {
		t.Errorf("Expected Alice asset for Dave to be %.2f, got %.2f",
			expectedValue, buyerAsset)
	}
	// Check buyer's liability for her own currency "Alice".
	buyerLiability := alice.Liabilities["Alice"]
	if buyerLiability != expectedValue {
		t.Errorf("Expected Alice liability for Alice to be %.2f, got %.2f",
			expectedValue, buyerLiability)
	}
	// Check seller's asset for "Alice" currency.
	sellerAsset := dave.Assets["Alice"]
	if sellerAsset != expectedValue {
		t.Errorf("Expected Dave asset for Alice to be %.2f, got %.2f",
			expectedValue, sellerAsset)
	}
	// Check seller's liability for his own currency "Dave".
	sellerLiability := dave.Liabilities["Dave"]
	if sellerLiability != expectedValue {
		t.Errorf("Expected Dave liability for Dave to be %.2f, got %.2f",
			expectedValue, sellerLiability)
	}
}

