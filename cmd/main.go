package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/nikhilpratapgit/TodoApp/database"
	"github.com/nikhilpratapgit/TodoApp/server"
)

func main() {

	srv := server.SetupRoutes()
	// make envs
	if err := database.ConnectandMigrate(
		"localhost",
		"5432",
		"postgres",
		"local",
		"local",
		database.SSLModeDisable); err != nil {
		fmt.Printf("Failed while initialize and migrate database: %v", err)
	}
	fmt.Println("server is running")
	ServerErr := http.ListenAndServe(":8080", srv)
	if ServerErr != nil {
		log.Fatal("")
		return
	}

	fmt.Println("server started at:8080")

}
