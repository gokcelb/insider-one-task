package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/insider/event-ingestion/clickhouse"
	"github.com/insider/event-ingestion/clickhouse/repository"
	"github.com/insider/event-ingestion/config"
	"github.com/insider/event-ingestion/events"
	"github.com/insider/event-ingestion/kafka"
	"github.com/insider/event-ingestion/metrics"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	producer, err := kafka.NewProducer(cfg.Kafka)
	if err != nil {
		log.Fatalf("failed to create kafka producer: %v", err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Printf("failed to close kafka producer: %v", err)
		}
	}()

	chClient, err := clickhouse.NewClient(cfg.ClickHouse)
	if err != nil {
		log.Fatalf("failed to connect to clickhouse: %v", err)
	}
	defer chClient.Close()

	if err := clickhouse.RunMigrations(cfg.ClickHouse); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	metricsRepo := repository.NewMetricsRepository(chClient.Conn())

	eventService := events.NewService(producer)
	eventHandler := events.NewHandler(eventService)

	metricsService := metrics.NewService(metricsRepo)
	metricsHandler := metrics.NewHandler(metricsService)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	r.GET("/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := chClient.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "error": "clickhouse unavailable"})
			return
		}

		if err := producer.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "error": "kafka unavailable"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	eventHandler.RegisterRoutes(r)
	metricsHandler.RegisterRoutes(r)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("starting server on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited")
}
