package server

import (
	"encoding/json"
	"errors"
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

// Переменные для уменьшения количества логов о состоянии сервера
var (
	logCounter  = 0
	logInterval = 3
)

var (
	ErrNotANumber     = errors.New("execTime is not a number")
	ErrNegativeNumber = errors.New("execTime is negative")
	ErrTooBigNumber   = errors.New("execTime is too big")
)

func (s *Server) handleProcessTask(w http.ResponseWriter, r *http.Request) {
	// Получаем значение execTime из заголовка запроса
	s.mu.Lock()         // Блокируем другие запросы
	defer s.mu.Unlock() // Освобождаем после завершения

	execTimeStr := r.Header.Get("Execution-Time")
	if execTimeStr != "" {
		execTime, err := s.processErrors(w, execTimeStr)
		if err != nil {
			return
		}
		req := TaskRequest{DelayMs: execTime}
		// Логируем начало обработки
		s.Logger.Printf("Server %d: Starting task with delay %dms", s.ID, req.DelayMs)

		// Имитируем обработку
		processingTime := s.ProcessTask(time.Duration(req.DelayMs))
		// Формируем ответ
		resp := TaskResponse{
			ServerID:  s.ID,
			Delay:     processingTime,
			Timestamp: time.Now().Format(time.Millisecond.String()),
		}

		// Логируем завершение
		s.Logger.Printf("Server %d: Task completed in %dms", s.ID, req.DelayMs)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	} else {
		http.Error(w, "Execution time not specified", http.StatusBadRequest)
	}
}

func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	// log.Printf("Received health check request from %s", r.RemoteAddr)
	if s.IsHealthy() {
		w.WriteHeader(http.StatusOK)
		if logCounter%logInterval == 0 {
			s.Logger.Printf("Server is healthy, returning 200 OK")
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		s.Logger.Printf("Server is not healthy, returning 500 Internal Server Error")
	}
	logCounter++
}

func (s *Server) processErrors(w http.ResponseWriter, execTimeStr string) (int, error) {
	execTime, err := strconv.Atoi(execTimeStr)
	if err != nil {
		s.Logger.Printf("Invalid execution time: %s, error: %v", execTimeStr, ErrNotANumber)
		http.Error(w, "Invalid execution time", http.StatusBadRequest)
		return 0, ErrNotANumber
	} else if execTime < 0 {
		s.Logger.Printf("Invalid execution time: %s, error: %v", execTimeStr, ErrNegativeNumber)
		http.Error(w, "Invalid execution time "+strconv.Itoa(execTime), http.StatusBadRequest)
		return 0, ErrNegativeNumber
	} else if execTime >= 10000 {
		s.Logger.Printf("Invalid execution time: %s, error: %v", execTimeStr, ErrTooBigNumber)
		http.Error(w, "Invalid execution time "+strconv.Itoa(execTime), http.StatusBadRequest)
		return 0, ErrTooBigNumber
	}

	return execTime, nil
}
