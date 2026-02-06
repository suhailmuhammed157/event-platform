package http

import (
	"encoding/json"
	"event-platform/ingest-service/internal/batching"
	"log"
	"net/http"
	"time"
)

type Server struct {
	batcher *batching.Batcher
}

func NewServer(b *batching.Batcher) *Server {
	return &Server{batcher: b}
}

func (s *Server) Run(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ingest", s.ingestHandler)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	log.Printf("Ingest service listening on %s", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

type EventRequest struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"timestamp"`
	Type      string `json:"type"`
	Payload   string `json:"payload"`
}

func (s *Server) ingestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event EventRequest
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	s.batcher.Add(batching.Event{
		Key:   event.ID,
		Value: []byte(event.Payload),
		Topic: "events",
	})

	w.WriteHeader(http.StatusAccepted)
}
