package main

import (
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func main() {
	http.HandleFunc("/kanye", kanyeHandler)
	http.ListenAndServe(":8001", nil)
}

func kanyeHandler(w http.ResponseWriter, r *http.Request) {
	delayParam := r.URL.Query().Get("delay")
	applyDelay := false

	if delayParam != "" {
		applyDelay, _ = strconv.ParseBool(delayParam)
	}

	if applyDelay {
		// add a random delay
		delay := rand.Intn(3)
		time.Sleep(time.Duration(delay) * time.Second)
	}

	// get a kanye quote
	resp, err := http.Get("https://api.kanye.rest")
	if err != nil {
		w.WriteHeader(500)
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}
