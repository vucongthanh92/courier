package main

import (
	_ "time/tzdata"

	"github.com/vucongthanh92/courier/user-service/startup"
)

func main() {
	startup.Execute()
}
