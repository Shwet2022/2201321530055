package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)



type LogEntry struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Resource  string    `json:"resource"`
	Timestamp time.Time `json:"timestamp"`
	TraceID   string    `json:"traceId"`
}

type LoggerService struct {
}

func NewLoggerService() *LoggerService {
	return &LoggerService{}
}

func (ls *LoggerService) Log(level, message, resource, traceID string) {
	entry := LogEntry{
		Level:     level,
		Message:   message,
		Resource:  resource,
		Timestamp: time.Now().UTC(),
		TraceID:   traceID,
	}
	logJSON, _ := json.Marshal(entry)
	fmt.Println(string(logJSON)) 
}

func (ls *LoggerService) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := fmt.Sprintf("%d", time.Now().UnixNano())
		ls.Log("info", fmt.Sprintf("Request received: %s %s", r.Method, r.URL.Path), r.URL.Path, traceID)
		next.ServeHTTP(w, r)
		ls.Log("info", fmt.Sprintf("Response sent for: %s %s", r.Method, r.URL.Path), r.URL.Path, traceID)
	})
}


type ShortURL struct {
	ID        string    `json:"id"`
	OriginalURL string    `json:"originalUrl"`
	ShortCode string    `json:"shortcode"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	Hits      int       `json:"hits"`
}

// CreateURLRequest is the request body for creating a short URL.
type CreateURLRequest struct {
	URL       string `json:"url"`
	Validity  int    `json:"validity,omitempty"` // in minutes
	ShortCode string `json:"shortcode,omitempty"`
}

// CreateURLResponse is the response for a successful creation.
type CreateURLResponse struct {
	ShortLink string    `json:"shortLink"`
	Expiry    time.Time `json:"expiry"`
}

// URLStatsResponse provides statistics for a short URL.
type URLStatsResponse struct {
	OriginalURL string    `json:"originalUrl"`
	CreatedAt   time.Time `json:"createdAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
	Hits        int       `json:"hits"`
}

// ErrorResponse is a generic error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Datastore is our in-memory database.
type Datastore struct {
	mu   sync.RWMutex
	urls map[string]*ShortURL // maps shortcode to ShortURL object
}

var db *Datastore
var logger *LoggerService
var hostname string

func init() {
	db = &Datastore{
		urls: make(map[string]*ShortURL),
	}
	logger = NewLoggerService()
	hostname = "http://localhost:8080" // Change if needed
}

// generateShortCode creates a random, unique shortcode.
func generateShortCode() (string, error) {
	const length = 6
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for i := 0; i < 10; i++ { // Try 10 times to find a unique code
		bytes := make([]byte, length)
		if _, err := rand.Read(bytes); err != nil {
			return "", err
		}
		code := make([]byte, length)
		for i, b := range bytes {
			code[i] = chars[int(b)%len(chars)]
		}
		shortCode := string(code)
		db.mu.RLock()
		_, exists := db.urls[shortCode]
		db.mu.RUnlock()
		if !exists {
			return shortCode, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique shortcode")
}

// createShortURLHandler handles the POST /shorturls request.
func createShortURLHandler(w http.ResponseWriter, r *http.Request) {
	traceID := fmt.Sprintf("%d", time.Now().UnixNano())
	var req CreateURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log("error", "Invalid request body", r.URL.Path, traceID)
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.URL == "" {
		logger.Log("error", "URL is required", r.URL.Path, traceID)
		respondWithError(w, http.StatusBadRequest, "URL is required")
		return
	}

	// Determine shortcode
	shortCode := req.ShortCode
	if shortCode == "" {
		var err error
		shortCode, err = generateShortCode()
		if err != nil {
			logger.Log("error", "Failed to generate shortcode", r.URL.Path, traceID)
			respondWithError(w, http.StatusInternalServerError, "Could not create short URL")
			return
		}
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.urls[shortCode]; exists {
		logger.Log("error", "Custom shortcode already exists", r.URL.Path, traceID)
		respondWithError(w, http.StatusConflict, "Custom shortcode already exists")
		return
	}

	// Determine expiry
	validity := req.Validity
	if validity == 0 {
		validity = 30 // Default to 30 minutes
	}
	expiresAt := time.Now().UTC().Add(time.Duration(validity) * time.Minute)

	newURL := &ShortURL{
		ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
		OriginalURL: req.URL,
		ShortCode:   shortCode,
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   expiresAt,
		Hits:        0,
	}

	db.urls[shortCode] = newURL

	logger.Log("info", fmt.Sprintf("URL shortened: %s -> %s", req.URL, shortCode), r.URL.Path, traceID)
	respondWithJSON(w, http.StatusCreated, CreateURLResponse{
		ShortLink: fmt.Sprintf("%s/%s", hostname, shortCode),
		Expiry:    expiresAt,
	})
}

// redirectHandler handles redirection and stats.
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	traceID := fmt.Sprintf("%d", time.Now().UnixNano())
	shortCode := r.URL.Path[1:] // Remove leading '/'

	// Check if it's a stats request
	isStatsRequest := len(shortCode) > 5 && shortCode[len(shortCode)-5:] == "/stats"
	if isStatsRequest {
		shortCode = shortCode[:len(shortCode)-5]
		getStatsHandler(w, r, shortCode, traceID)
		return
	}

	db.mu.Lock() // Use Lock because we are modifying Hits
	defer db.mu.Unlock()

	url, exists := db.urls[shortCode]
	if !exists {
		logger.Log("warn", "Shortcode not found", r.URL.Path, traceID)
		respondWithError(w, http.StatusNotFound, "Short URL not found")
		return
	}

	if time.Now().UTC().After(url.ExpiresAt) {
		logger.Log("warn", "Short URL has expired", r.URL.Path, traceID)
		// Optionally delete the expired URL
		delete(db.urls, shortCode)
		respondWithError(w, http.StatusNotFound, "Short URL has expired")
		return
	}

	url.Hits++
	logger.Log("info", fmt.Sprintf("Redirecting %s to %s", shortCode, url.OriginalURL), r.URL.Path, traceID)
	http.Redirect(w, r, url.OriginalURL, http.StatusFound)
}

// getStatsHandler provides stats for a short URL.
func getStatsHandler(w http.ResponseWriter, r *http.Request, shortCode, traceID string) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	url, exists := db.urls[shortCode]
	if !exists {
		logger.Log("warn", "Shortcode for stats not found", r.URL.Path, traceID)
		respondWithError(w, http.StatusNotFound, "Statistics not found for this URL")
		return
	}
	
	logger.Log("info", fmt.Sprintf("Stats requested for %s", shortCode), r.URL.Path, traceID)
	respondWithJSON(w, http.StatusOK, URLStatsResponse{
		OriginalURL: url.OriginalURL,
		CreatedAt:   url.CreatedAt,
		ExpiresAt:   url.ExpiresAt,
		Hits:        url.Hits,
	})
}


// --- Helper Functions ---

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, ErrorResponse{Error: message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // For development
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.WriteHeader(code)
	w.Write(response)
}

func corsHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "OPTIONS" {
            w.Header().Set("Access-Control-Allow-Origin", "*")
            w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
            w.WriteHeader(http.StatusOK)
            return
        }
        next.ServeHTTP(w, r)
    })
}


func main() {
	// Using a router to handle specific paths
	mux := http.NewServeMux()
	
	// Endpoint for creating short URLs
	mux.HandleFunc("/shorturls", createShortURLHandler)
	
	// Handler for redirection and stats
	mux.HandleFunc("/", redirectHandler)

	// Wrap the main router with the CORS and logging middleware
	handler := corsHandler(logger.Middleware(mux))

	fmt.Println("URL Shortener Microservice starting on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		fmt.Printf("Server failed: %s\n", err)
	}
}
