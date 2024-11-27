package main

import (
	"fmt"
	"net/http"
)

func main() {
	r := SetupRouter()
	fmt.Println("Listening on port 8080")

	http.ListenAndServe(":8080", r)
}
