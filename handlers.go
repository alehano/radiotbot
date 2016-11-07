package main

import (
	"fmt"
	"net/http"
)

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Path[len("/test/"):]

	Search(searchIndex, q)

	fmt.Fprintf(w, "ok")
}
