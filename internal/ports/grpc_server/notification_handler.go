package grpc_server

import (
	"context"
	"fmt"
	"notification-service/api/proto/pb"
	"notification-service/internal/adapters/kafka"
	"strings" // Import the strings package

	"github.com/sirupsen/logrus"
)

// Server implements the gRPC server functionality.
type Server struct {
	pb.UnimplementedNotificationServiceServer
	kafkaProducer *kafka.Producer
	logger        *logrus.Logger
}

// NewGrpcServer creates a new gRPC server.
func NewGrpcServer(producer *kafka.Producer, logger *logrus.Logger) *Server {
	return &Server{
		kafkaProducer: producer,
		logger:        logger,
	}
}

// SendNotification is the gRPC method handler. It now publishes an event to Kafka.
func (s *Server) SendNotification(ctx context.Context, req *pb.SendNotificationRequest) (*pb.SendNotificationResponse, error) {
	log := s.logger.WithFields(logrus.Fields{
		"recipient": req.To,
		"template":  req.TemplateName,
		"source":    "grpc",
	})
	log.Info("Received gRPC request, preparing to publish to Kafka")

	// --- START: ADDED SANITIZATION ---
	// Sanitize the template name to remove trailing characters that are invalid in Kafka topics.
	sanitizedTemplateName := strings.TrimRight(req.TemplateName, "!?. ")
	// --- END: ADDED SANITIZATION ---

	// The topic name is derived from the sanitized template name.
	topic := fmt.Sprintf("notification.send-%s", sanitizedTemplateName)

	// Convert the protobuf struct to a standard map[string]interface{}
	data := req.Data.AsMap()

	// The recipient's email must be included in the message payload
	// for the consumer to use it.
	data["email"] = req.To

	// Publish the event to Kafka.
	if err := s.kafkaProducer.Publish(ctx, topic, data); err != nil {
		log.WithError(err).Error("Failed to publish notification event to Kafka")
		// Return a generic error to the client, as the failure is internal.
		return &pb.SendNotificationResponse{
			Success: false,
			Message: "Failed to queue notification for sending.",
		}, nil
	}

	log.Info("Successfully published notification event to Kafka")
	return &pb.SendNotificationResponse{
		Success: true,
		Message: "Notification has been successfully queued for sending.",
	}, nil
}
