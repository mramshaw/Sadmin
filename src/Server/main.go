package main

import (
	"os"

	// local import
	"admin-server/application"
)

func main() {
	app := application.App{}
	app.Initialize(
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_DB"),
		os.Getenv("AUTH_USER"),
		os.Getenv("AUTH_PASSWORD"))
	app.Run(os.Getenv("PORT"))
}
