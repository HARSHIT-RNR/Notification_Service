package kafka

import (
	"context"
	"encoding/json"
	"notification-service/internal/core/services"
	"strings"
	"sync"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

type ConsumerGroupHandler struct {
	notificationSvc *services.NotificationService
	logger          *logrus.Logger
}

func NewConsumerGroupHandler(svc *services.NotificationService, logger *logrus.Logger) *ConsumerGroupHandler {
	return &ConsumerGroupHandler{
		notificationSvc: svc,
		logger:          logger,
	}
}

func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		log := h.logger.WithFields(logrus.Fields{
			"topic":     message.Topic,
			"partition": message.Partition,
			"offset":    message.Offset,
		})
		log.Info("Kafka message claimed")

		// Derives template name from topic, e.g., "notification.send-welcome" -> "welcome"
		// This is a simple convention; you might want a more robust mapping.
		templateName := strings.TrimPrefix(message.Topic, "notification.send-")

		var data map[string]interface{}
		if err := json.Unmarshal(message.Value, &data); err != nil {
			log.WithError(err).Error("Failed to unmarshal kafka message, skipping")
			session.MarkMessage(message, "") // Acknowledge and move on
			continue
		}

		recipient, ok := data["email"].(string)
		if !ok || recipient == "" {
			log.Error("Recipient 'email' not found or is empty in message, skipping")
			session.MarkMessage(message, "") // Acknowledge and move on
			continue
		}

		req := services.SendRequest{
			To:           recipient,
			TemplateName: templateName,
			Data:         data,
		}

		// The service now handles its own logging, including failures.
		h.notificationSvc.SendNotification(session.Context(), req)

		// Mark the message as processed regardless of success or failure.
		// The failure is logged to the DB for later inspection or retry.
		session.MarkMessage(message, "")
	}
	return nil
}

func StartConsumerGroup(ctx context.Context, wg *sync.WaitGroup, brokers, topics []string, groupID string, handler sarama.ConsumerGroupHandler) {
	defer wg.Done()
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_1_0 // or a version compatible with your kafka cluster
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create Sarama consumer group")
	}
	defer consumerGroup.Close()

	for {
		if err := consumerGroup.Consume(ctx, topics, handler); err != nil {
			if err == sarama.ErrClosedConsumerGroup {
				return // Exit cleanly
			}
			logrus.WithError(err).Error("Error from consumer")
		}
		if ctx.Err() != nil {
			return // Context was cancelled
		}
	}
}
