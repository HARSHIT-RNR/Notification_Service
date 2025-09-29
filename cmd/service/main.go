package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"notification-service/api/proto/pb"
	"notification-service/internal/adapters/kafka"
	"notification-service/internal/adapters/logger"
	"notification-service/internal/adapters/mailme"
	repo "notification-service/internal/adapters/repository"
	"notification-service/internal/config"
	"notification-service/internal/core/services"
	"notification-service/internal/ports/grpc_server"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.LoadConfig()
	log := logger.NewLogger()
	log.Info("Starting ERP CP Notification Service")

	// --- Database Connection ---
	db, err := sql.Open(cfg.DBDriver, cfg.DBSource)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		log.WithError(err).Fatal("Cannot ping database")
	}

	// --- Dependency Injection ---
	// Dependencies for the CONSUMER side (email sending)
	logRepo := repo.NewNotificationLogRepo(db)
	templateRepo := repo.NewFileTemplateRepo(cfg.TemplatePath, log)
	mailer := mailme.NewMailer(cfg.SmtpHost, cfg.SmtpPort, cfg.SmtpUser, cfg.SmtpPass, cfg.SmtpFrom, log)
	notificationSvc := services.NewNotificationService(templateRepo, logRepo, mailer, log)

	// Dependency for the PRODUCER side (gRPC server)
	kafkaProducer, err := kafka.NewProducer(cfg.KafkaBrokers, log)
	if err != nil {
		log.WithError(err).Fatal("Failed to create Kafka producer")
	}
	defer kafkaProducer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// --- Start Kafka Consumer ---
	wg.Add(1)
	go func() {
		log.WithFields(logrus.Fields{"brokers": cfg.KafkaBrokers, "group": cfg.KafkaGroupID}).Info("Starting Kafka consumer group")
		consumerHandler := kafka.NewConsumerGroupHandler(notificationSvc, log)
		kafka.StartConsumerGroup(ctx, &wg, cfg.KafkaBrokers, cfg.KafkaTopics, cfg.KafkaGroupID, consumerHandler)
		log.Info("Kafka consumer group stopped")
	}()

	// --- Start gRPC Server ---
	wg.Add(1)
	go func() {
		defer wg.Done()
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GrpcPort))
		if err != nil {
			log.WithError(err).Fatalf("Failed to listen on port %s", cfg.GrpcPort)
		}
		grpcServer := grpc.NewServer()
		// CORRECTED: The server now takes the producer, not the notification service.
		server := grpc_server.NewGrpcServer(kafkaProducer, log)
		pb.RegisterNotificationServiceServer(grpcServer, server)
		reflection.Register(grpcServer)

		log.Infof("gRPC server listening on port %s", cfg.GrpcPort)
		go func() {
			if err := grpcServer.Serve(lis); err != nil {
				log.WithError(err).Error("gRPC server failed")
			}
		}()

		<-ctx.Done()
		log.Info("Gracefully stopping gRPC server...")
		grpcServer.GracefulStop()
		log.Info("gRPC server stopped")
	}()

	// --- Graceful Shutdown ---
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm

	log.Warn("Termination signal received. Shutting down...")
	cancel()
	wg.Wait()
	log.Info("Service shut down gracefully")
}
