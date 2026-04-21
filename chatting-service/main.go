package main

import (
	_ "time/tzdata"

	"github.com/vucongthanh92/courier/chat-service/startup"
)

func main() {
	startup.Execute()
}
