package main

import "testing"

// TestSimulationTrade verifies that a trade is executed in the open market
// simulation. In this test, Alice submits a BID order for 10 TOKEN, and Dave
// submits an ASK order for 8 TOKEN. If the matching engine functions properly,
// the trade should execute at the buyer's bid price, resulting in Alice having a
// liability of 10 in her currency and Dave having an asset of 10 in his currency.
func TestSimulationTrade(t *testing.T) {
    alice, _, _, dave := RunSimulation()

    expectedBuyerLiability := 10.0
    expectedSellerAsset := 10.0

    // Check the buyer's liability in their assigned currency.
    buyerLiability := alice.Liabilities[alice.Currency]
    if buyerLiability != expectedBuyerLiability {
        t.Errorf("Expected Alice liability in %s to be %.2f, got %.2f",
            alice.Currency, expectedBuyerLiability, buyerLiability)
    }

    // Check the seller's asset in their assigned currency.
    sellerAsset := dave.Assets[dave.Currency]
    if sellerAsset != expectedSellerAsset {
        t.Errorf("Expected Dave asset in %s to be %.2f, got %.2f",
            dave.Currency, expectedSellerAsset, sellerAsset)
    }
}
