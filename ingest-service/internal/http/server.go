package http

import (
	"context"
	"encoding/json"
	"event-platform/ingest-service/internal/producer"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Server struct {
	producer *producer.KafkaProducer
}

func NewServer(p *producer.KafkaProducer) *Server {
	return &Server{producer: p}
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

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := s.producer.Publish(ctx, event.ID, []byte(event.Payload)); err != nil {
		fmt.Println(err)
		http.Error(w, "failed to enqueue event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
