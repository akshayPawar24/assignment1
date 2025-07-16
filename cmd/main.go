package main

import (
	"assignment1/setup"
)

func main() {
	app := setup.Initialize()

	app.Service.StartBackgroundSync()

	app.Router.Run(":" + app.Config.Port)
}
