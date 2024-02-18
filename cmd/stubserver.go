package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/pixel", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		random := rand.Int31n(10)
		if random < 5 {
			time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if random < 8 {
			time.Sleep(time.Duration(rand.Int31n(500)) * time.Millisecond)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		time.Sleep(time.Duration(rand.Int31n(2000)) * time.Millisecond)

		w.WriteHeader(http.StatusInternalServerError)
		return

	}))
	log.Printf("Running server on port 8081")
	log.Fatal(http.ListenAndServe("localhost:8081", mux))
}
