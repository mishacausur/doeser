package main

import "net/http"

func main() {
	error := http.ListenAndServe(":7540", nil)

	if error != nil {
		panic(error)
	}
}
