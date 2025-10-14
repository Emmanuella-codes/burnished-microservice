package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	cfg 					*config.Config
	router 				*gin.Engine
	server 				*http.Server
	docProc 			*documents.Processor
	docFormatter 	*documents.Formatter
	webhookClient *http.Client
}

func NewServer(cfg *config.Config) *Server {
	router := gin.Default()
	pdfProcessor := documents.NewPDFProcessor()
	processor := documents.NewProcessor(cfg)
	formatter := documents.NewFormatter(cfg, pdfProcessor)
	webhookClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	s := &Server{
		cfg: cfg,
		router: router,
		docProc: processor,
		docFormatter: formatter,
		webhookClient: webhookClient,
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
	api.Use(loggingMiddleware(), rateLimitMiddleware(), dailyRateLimitMiddleware())
	{
		api.GET("/health", s.healthHandler)
		api.POST("/process", s.processCVHandler)

	}
	api.Use(authMiddleware())
	{
		api.GET("/files/:filename", s.serveFileHandler)
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

func (s *Server) sendWebhook(payload ProcessResponse) error {
	webhookURL := os.Getenv("BURNISHED_WEB_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("WEBHOOK_URL not configured")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	webhookSecret := os.Getenv("WEBHOOK_SECRET")
	req.Header.Set("Authorization", "Bearer "+webhookSecret)
	req.Header.Set("Content-Type", "application/json")

	log.Printf("Sending webhook to: %s", webhookURL)
	res, err := s.webhookClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	log.Printf("Webhook response status: %d", res.StatusCode)
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("webhook returned non-200 status: %d", res.StatusCode)
	}
	return nil
}
