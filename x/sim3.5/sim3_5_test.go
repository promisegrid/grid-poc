package main

import "testing"

// TestSimulationTrade verifies that a trade is executed in the open market
// simulation. In this test, Alice submits a BID order for 10 TOKEN (with OrderID "BID1"),
// and Dave submits an ASK order for 8 TOKEN (with OrderID "ASK1"). If the matching engine
// functions properly, the trade should execute at the buyer's bid price. The buyer (Alice)
// should record an increase in asset "TOKEN" of 10 and an increase in liability "CASH" of 10,
// while the seller (Dave) should record an increase in asset "CASH" of 10 and an increase in
// liability "TOKEN" of 10.
func TestSimulationTrade(t *testing.T) {
    alice, _, _, dave := RunSimulation()

    // Expected values based on double-entry transaction:
    // For buyer (Alice): Assets["TOKEN"] and Liabilities["CASH"] should be 10.
    // For seller (Dave): Assets["CASH"] and Liabilities["TOKEN"] should be 10.

    expectedValue := 10.0

    // Check buyer's asset "TOKEN"
    buyerTokenAsset := alice.Assets["TOKEN"]
    if buyerTokenAsset != expectedValue {
        t.Errorf("Expected Alice asset TOKEN to be %.2f, got %.2f", expectedValue, buyerTokenAsset)
    }
    // Check buyer's liability "CASH"
    buyerCashLiability := alice.Liabilities["CASH"]
    if buyerCashLiability != expectedValue {
        t.Errorf("Expected Alice liability CASH to be %.2f, got %.2f", expectedValue, buyerCashLiability)
    }
    // Check seller's asset "CASH"
    sellerCashAsset := dave.Assets["CASH"]
    if sellerCashAsset != expectedValue {
        t.Errorf("Expected Dave asset CASH to be %.2f, got %.2f", expectedValue, sellerCashAsset)
    }
    // Check seller's liability "TOKEN"
    sellerTokenLiability := dave.Liabilities["TOKEN"]
    if sellerTokenLiability != expectedValue {
        t.Errorf("Expected Dave liability TOKEN to be %.2f, got %.2f", expectedValue, sellerTokenLiability)
    }
}
