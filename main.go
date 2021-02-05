package main

import (
	"myNotes/core/http"
	"myNotes/core/mongo"
)

func main() {
	db, err := mongo.NDB("default", "myNotes")
	if err != nil {
		panic(err)
	}

	bot := *http.NEmailSender(http.BotAccount.Email, http.BotAccount.Password, 587)

	ws := http.NWS("127.0.0.1", "./web", 5504, db, bot)

	ws.RegisterHandlers()

	ws.Run()
}
