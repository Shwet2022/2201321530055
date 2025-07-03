package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	URL string
}

// NewLoggerService creates a new instance of the LoggerService.
func NewLoggerService(url string) *LoggerService {
	return &LoggerService{URL: url}
}


func (ls *LoggerService) Log(level, message, resource, traceID string) {
	entry := LogEntry{
		Level:     level,
		Message:   message,
		Resource:  resource,
		Timestamp: time.Now().UTC(),
		TraceID:   traceID,
	}

	
	logJSON, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("Error marshalling log entry: %v\n", err)
		return
	}
	fmt.Println(string(logJSON))
}

func (ls *LoggerService) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := fmt.Sprintf("%d", time.Now().UnixNano())

		// Log the incoming request
		ls.Log("info", "Request received", r.URL.Path, traceID)

		// Call the next handler in the chain
		next.ServeHTTP(w, r)

		// Log the outgoing response (simplified)
		ls.Log("info", "Response sent", r.URL.Path, traceID)
	})
}

func main() {
	
	logger := NewLoggerService("http://test-log-server/log") // Dummy URL

	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Log("info", "Executing helloHandler", r.URL.Path, "trace-123")
		w.Write([]byte("Hello, World!"))
	})

	http.Handle("/", logger.Middleware(helloHandler))

	fmt.Println("Test server with logging middleware starting on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		fmt.Printf("Server failed: %s\n", err)
	}
}
