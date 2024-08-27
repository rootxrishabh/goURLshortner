package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type shortenRequest struct {
	LongURL     string `json:"long_url"`
	CustomAlias string `json:"custom_alias,omitempty"`
	TTLSeconds  int    `json:"ttl_seconds,omitempty"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
}

type analyticsResponse struct {
	Alias       string   `json:"alias"`
	LongUrl     string   `json:"long_url"`
	AccessCount int      `json:"access_count"`
	AccessTimes []string `json:"access_times"`
}

type updateRequest struct {
	CustomAlias string `json:"custom_alias,omitempty"`
	TTLSeconds  int    `json:"ttl_seconds,omitempty"`
}

type urlInfo struct {
	LongURL      string
	CreationTime time.Time
	TTL          time.Duration
	AccessCount  int
	AccessTimes  []time.Time
}

var (
	urlStore = make(map[string]urlInfo)
	baseUrl  = "http://localhost:8080/"
)

func generateAlias() string {
	const aliasLength = 6
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, aliasLength)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}

	return string(result)
}

// POST /shorten
func shortenURLHandler(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.LongURL == "" {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	alias := req.CustomAlias
	if alias == "" {
		alias = generateAlias()
	}

	ttl := time.Duration(req.TTLSeconds) * time.Second
	if ttl == 0 {
		ttl = 120 * time.Second
	}

	urlStore[alias] = urlInfo{
		LongURL:      req.LongURL,
		CreationTime: time.Now(),
		TTL:          ttl,
		AccessCount:  0,
		AccessTimes:  []time.Time{},
	}

	resp := shortenResponse{
		ShortURL: baseUrl + alias,
	}

	json.NewEncoder(w).Encode(resp)
}

// GET "/"
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	alias := r.URL.Path[len("/"):]
	url, exists := urlStore[alias]
	if !exists {
		http.Error(w, "Alias not found or expired", http.StatusNotFound)
		return
	}

	url.AccessCount++
	url.AccessTimes = append(url.AccessTimes, time.Now())
	urlStore[alias] = url

	http.Redirect(w, r, url.LongURL, http.StatusFound)
}

// GET /analytics/:alias
func analyticsHandler(w http.ResponseWriter, r *http.Request) {
	alias := r.URL.Path[len("/analytics/"):]

	url, exists := urlStore[alias]
	if !exists {
		http.Error(w, "Alias not found or expired", http.StatusNotFound)
		return
	}

	resp := analyticsResponse{
		Alias:       alias,
		LongUrl:     url.LongURL,
		AccessCount: url.AccessCount,
		AccessTimes: []string{},
	}

	for i := len(url.AccessTimes) - 1; i >= 0; i-- {
		resp.AccessTimes = append(resp.AccessTimes, url.AccessTimes[i].Format(time.RFC3339))
	}

	for  {
		if len(resp.AccessTimes) > 10 {
			resp.AccessTimes = resp.AccessTimes[:len(resp.AccessTimes)-1]
		} else {
			break
		}
	}

	json.NewEncoder(w).Encode(&resp)
}

// PUT /update/:alias
func updateHandler(w http.ResponseWriter, r *http.Request) {
	alias := r.URL.Path[len("/update/"):]
	var req updateRequest
	err := json.NewDecoder(r.Body).Decode(&req) // Use &req to decode into the pointer
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	url, exists := urlStore[alias]
	if !exists || time.Since(url.CreationTime) > url.TTL {
		http.Error(w, "Alias not found or expired", http.StatusNotFound)
		return
	}

	if req.CustomAlias != "" {
		delete(urlStore, alias)
		alias = req.CustomAlias
	}

	if req.TTLSeconds > 0 {
		url.TTL = time.Duration(req.TTLSeconds) * time.Second
	}

	urlStore[alias] = url
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "URL updated")
}

// DELETE /delete/:alias
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	alias := r.URL.Path[len("/delete/"):]
	_, exists := urlStore[alias]
	if !exists {
		http.Error(w, "Alias not found or expired", http.StatusNotFound)
		return
	}

	delete(urlStore, alias)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "URL deleted")
}

func main() {

	go func() {
		for {
			time.Sleep(10 * time.Second) // Check for expired URLs every minute
			now := time.Now()
			for alias, url := range urlStore {
				if now.Sub(url.CreationTime) > url.TTL {
					delete(urlStore, alias)
				}
			}
		}
	}()

	http.HandleFunc("/shorten", shortenURLHandler)
	http.HandleFunc("/", redirectHandler)
	http.HandleFunc("/analytics/", analyticsHandler)
	http.HandleFunc("/update/", updateHandler)
	http.HandleFunc("/delete/", deleteHandler)

	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}
