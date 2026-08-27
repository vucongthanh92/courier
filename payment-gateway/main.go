package main

import (
	_ "time/tzdata"

	"github.com/vucongthanh92/courier/payment-gateway/startup"
)

// @title Courier Payment Gateway API
// @version 1.0
// @description Wallet top-up and payment ledger APIs.
// @BasePath /api/v1
func main() {
	startup.Execute()
}
