package main

import (
	_ "time/tzdata"

	_ "github.com/vucongthanh92/courier/chat-service/docs"
	"github.com/vucongthanh92/courier/chat-service/startup"
)

// @title Courier Chat Service API
// @version 1.0
// @description Conversation and messaging APIs for Courier.
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	startup.Execute()
}
