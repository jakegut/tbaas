package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jakegut/tbaas"
)

func main() {
	// Bucket with 5 tokens per minute
	b := tbaas.MakeBucket(5, time.Minute)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		amt, err := b.Take(r.Context(), r.RemoteAddr, 1)
		if err != nil {
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintf(w, "Too many requests")
			return
		}
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "Hello there! You have %d requests remaining", amt)
	})

	http.ListenAndServe(":8080", nil)
}
