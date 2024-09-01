package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/slack/events", handleSlackEvents)
	if port == "" {
		port = "3000"
	}
	fmt.Printf("Server listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
