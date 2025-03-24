package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Emmanuella-codes/burnished-microservice/internal/config"
	"github.com/Emmanuella-codes/burnished-microservice/internal/documents"

	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg 		*config.Config
	router 	*gin.Engine
	server 	*http.Server
	docProc *documents.Processor
}

func NewServer(cfg *config.Config) *Server {
	router := gin.Default()
	processor := documents.NewProcessor(cfg)
	s := &Server{
		cfg: cfg,
		router: router,
		docProc: processor,
		server: &http.Server{
			Addr: 	 ":" + cfg.Port,
			Handler: router,
		},
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	api := s.router.Group("/api/v1")
	api.Use(loggingMiddleware())
	api.Use(validateRequestMiddleware())
	api.Use(rateLimitMiddleware())
	{
		api.GET("/health", s.healthHandler)
		api.POST("/process-cv", s.processCVHandler)
		api.POST("/format-cv", s.formatCVHandler)
		api.POST("/roast-cv", s.roastCVHandler)
		api.POST("/generate-cover-letter", s.generateCoverLetterHandler)
	}
}

func (s *Server) Start() error {
	// setup shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.server.Shutdown(ctx); err != nil {
			fmt.Printf("Server forced to shutdown: %v\n", err)
		}
	}()
	
	fmt.Printf("Server starting on port %s\n", s.cfg.Port)
	return s.server.ListenAndServe()
}
