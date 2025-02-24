package api

import (
	"net/http"

	"github.com/Emmanuella-codes/burnished-microservice/internal/config"
)

type Server struct {
	cfg 		*config.Config
	router 	*http.ServeMux
}

func NewServer(cfg *config.Config) *Server {
	s := &Server{
		cfg: cfg,
		router: http.NewServeMux(),
	}
	
}

func (s *Server) setupRoutes() {
	s.router.HandleFunc("/health", s.healthHandler)
	s.router.HandleFunc("/process-cv", s.processCVHandler)
	s.router.HandleFunc("/format-cv", s.formatCVHandler)
	s.router.HandleFunc("/roast-cv", s.roastCVHandler)
	s.router.HandleFunc("/generate-cover-letter", s.generateCoverLetterHandler)
}

func (s *Server) Start() error {
	server := &http.Server{
		
	}
}
