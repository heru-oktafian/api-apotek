package main

import (
	os "os"

	utils "github.com/heru-oktafian/scafold/utils"
)

func main() {
	// Get port from environment
	serverPort := os.Getenv("SERVER_PORT")

	app := utils.App()
	app.Listen(":" + serverPort)
}
