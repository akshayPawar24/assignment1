package main

import (
	"assignment1/setup"
)

func main() {
	app := setup.Initialize()

	app.Service.StartBackgroundSync()

	err := app.Router.Run(":" + app.Config.Port)
	if err != nil {
		return
	}
}
