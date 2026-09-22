package metrics

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

type Server struct {
	server   *http.Server
	listener net.Listener
	result   result
}

func NewServer() *Server {
	return &Server{}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) Start() (string, error) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/metrics", s.handleMetrics)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("failed to start metrics server: %w", err)
	}

	s.listener = listener
	s.server = &http.Server{
		Handler: mux,
	}

	go func() {
		_ = s.server.Serve(listener)
	}()

	url := fmt.Sprintf(
		"ws://%s/api/v1/metrics",
		listener.Addr().String(),
	)

	return url, nil
}

func (s *Server) Stop() error {
	if s.server == nil {
		return nil
	}

	return s.server.Close()
}

func (s *Server) Result() result {
	return s.result
}
