package main

import (
	"bookings/pkg/handlers"
	"fmt"
	"net/http"
)

const portNumber = ":8080"

func main() {
	http.HandleFunc("/", handlers.Home)
	http.HandleFunc("/about", handlers.About)

	fmt.Println("Startig application on port", portNumber)
	_ = http.ListenAndServe(portNumber, nil)
}
