package main

import (
	"net/http"
	"os"
)

func main() {

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	indexPage := http.FileServer(http.Dir("./web"))

	http.Handle("/", indexPage)

	error := http.ListenAndServe(":"+port, nil)

	if error != nil {
		panic(error)
	}
}
