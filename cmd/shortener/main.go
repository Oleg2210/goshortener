package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

var store = map[string]string{}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func randomID() string {
	n := rnd.Intn(6) + 5 // 5–10 символов
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rnd.Intn(len(letters))]
	}
	return string(b)
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	url := string(body)
	id := randomID()

	store[id] = url
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "http://localhost:8080/%s", id)
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[1:]
	url, ok := store[id]
	if !ok {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/" {
			handlePost(w, r)
			return
		}
		if r.Method == http.MethodGet {
			handleGet(w, r)
			return
		}
		http.Error(w, "bad request", http.StatusBadRequest)
	})

	http.ListenAndServe(":8080", nil)
}
