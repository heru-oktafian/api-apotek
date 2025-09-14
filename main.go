package main

import (
	utils "github.com/heru-oktafian/scafold/utils"
)

func main() {
	app := utils.App()
	app.Listen(":8080")
}
