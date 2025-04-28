package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	URL string `json:"url"`
	Key string `json:"key"`
}

type URLStore struct {
	mu       sync.RWMutex
	keyToURL map[string]string
	urlToKey map[string]string
}

func NewURLStore() *URLStore {
	return &URLStore{
		keyToURL: make(map[string]string),
		urlToKey: make(map[string]string),
	}
}

func (s *URLStore) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, ok := s.keyToURL[key]
	return url, ok
}

const keyLength = 10
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

var src = rand.NewSource(time.Now().UnixNano())

func generateKey() string {
	sb := strings.Builder{}
	sb.Grow(keyLength)
	for i, cache, remain := keyLength-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return sb.String()
}

type Server struct {
	store *URLStore
}

func NewServer(store *URLStore) *Server {
	return &Server{store: store}
}

func (s *Server) shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&req)
	if err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.URL == "" {
		http.Error(w, "invalid request: url cannot be empty", http.StatusBadRequest)
		return
	}

	s.store.mu.RLock()
	existingKey, found := s.store.urlToKey[req.URL]
	s.store.mu.RUnlock()

	if found {
		log.Printf("URL '%s' already exists with key '%s'", req.URL, existingKey)
		resp := ShortenResponse{URL: req.URL, Key: existingKey}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Error encoding JSON response for existing key: %v", err)
		}
		return
	}

	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	existingKey, found = s.store.urlToKey[req.URL]
	if found {
		log.Printf("URL '%s' already exists with key '%s' (checked after lock)", req.URL, existingKey)
		resp := ShortenResponse{URL: req.URL, Key: existingKey}
		w.Header().Set("Content-Type", "application/json")
		// Status already set by previous WriteHeader, no need to set again if headers written.
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Error encoding JSON response for existing key (after lock): %v", err)
		}
		return
	}

	var key string
	for {
		key = generateKey()
		if _, keyExists := s.store.keyToURL[key]; !keyExists {
			break
		}
		log.Printf("Key collision for generated key %s, retrying...", key)
	}

	s.store.keyToURL[key] = req.URL
	s.store.urlToKey[req.URL] = key

	log.Printf("Shortened '%s' to new key '%s'", req.URL, key)
	resp := ShortenResponse{URL: req.URL, Key: key}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding JSON response for new key: %v", err)
	}
}

func (s *Server) redirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	prefix := "/go/"
	if !strings.HasPrefix(path, prefix) {
		http.NotFound(w, r)
		return
	}
	key := path[len(prefix):]

	if key == "" {
		http.NotFound(w, r)
		return
	}

	originalURL, found := s.store.Get(key)
	if !found {
		log.Printf("Key not found: %s", key)
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}

	log.Printf("Redirecting key '%s' to '%s'", key, originalURL)
	http.Redirect(w, r, originalURL, http.StatusFound)
}

func main() {
	port := flag.Int("port", 6029, "Port for the server to listen on")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	store := NewURLStore()
	server := NewServer(store)

	mux := http.NewServeMux()

	mux.HandleFunc("/shorten", server.shortenHandler)
	mux.HandleFunc("/go/", server.redirectHandler)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintln(w, "URL Shortener is running. Use /shorten (POST) or /go/<key> (GET).")
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting URL shortener server on %s", addr)

	err := http.ListenAndServe(addr, mux)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
