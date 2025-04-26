package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

type TaskRequest struct {
	DelayMs int `json:"delay_ms"`
}

type TaskResponse struct {
	ServerID  int           `json:"server_id"`
	Delay     time.Duration `json:"delay"`
	Timestamp string        `json:"timestamp"`
}

func (s *Server) HandleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/process":
		s.handleProcessTask(w, r)
	case "/health":
		s.handleHealthCheck(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleProcessTask(w http.ResponseWriter, r *http.Request) {
	// Получаем значение execTime из заголовка запроса
	execTimeStr := r.Header.Get("Execution-Time")
	if execTimeStr != "" {
		execTime, err := strconv.Atoi(execTimeStr)
		if err != nil {
			log.Printf("Invalid execution time: %s, error: %w", execTimeStr, err)
			http.Error(w, "Invalid execution time", http.StatusBadRequest)
			return
		}
		req := TaskRequest{DelayMs: execTime}
		// Логируем начало обработки
		log.Printf("Server %d: Starting task with delay %dms", s.ID, req.DelayMs)

		// Имитируем обработку
		processingTime := s.ProcessTask(time.Duration(req.DelayMs))
		// Формируем ответ
		resp := TaskResponse{
			ServerID:  s.ID,
			Delay:     processingTime,
			Timestamp: time.Now().Format(time.RFC3339Nano),
		}

		// Логируем завершение
		log.Printf("Server %d: Task completed in %v", s.ID, processingTime)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	} else {
		http.Error(w, "Execution time not specified", http.StatusBadRequest)
	}
}

func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if s.IsHealthy() {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
