package kafka

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"notification-service/internal/core/services"
// 	"testing"

// 	"github.com/IBM/sarama"
// 	"github.com/sirupsen/logrus"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// )

// // --- Mock Notification Service ---

// type MockNotificationService struct {
// 	mock.Mock
// }

// func (m *MockNotificationService) SendNotification(ctx context.Context, req services.SendRequest) {
// 	m.Called(ctx, req)
// }

// // --- Mock Sarama Session ---

// type MockConsumerGroupSession struct {
// 	mock.Mock
// }

// func (m *MockConsumerGroupSession) Claims() map[string][]int32 {
// 	args := m.Called()
// 	return args.Get(0).(map[string][]int32)
// }
// func (m *MockConsumerGroupSession) MemberID() string {
// 	args := m.Called()
// 	return args.String(0)
// }
// func (m *MockConsumerGroupSession) GenerationID() int32 {
// 	args := m.Called()
// 	return int32(args.Int(0))
// }
// func (m *MockConsumerGroupSession) MarkOffset(topic string, partition int32, offset int64, metadata string) {
// 	m.Called(topic, partition, offset, metadata)
// }
// func (m *MockConsumerGroupSession) ResetOffset(topic string, partition int32, offset int64, metadata string) {
// 	m.Called(topic, partition, offset, metadata)
// }
// func (m *MockConsumerGroupSession) MarkMessage(msg *sarama.ConsumerMessage, metadata string) {
// 	m.Called(msg, metadata)
// }
// func (m *MockConsumerGroupSession) Context() context.Context {
// 	return context.Background()
// }

// // --- Test Suite ---

// func TestConsumerGroupHandler_ConsumeClaim(t *testing.T) {
// 	logger := logrus.New()
// 	logger.SetOutput(&bytes.Buffer{})

// 	t.Run("Success Case - Valid Message", func(t *testing.T) {
// 		// Arrange
// 		mockService := new(MockNotificationService)
// 		handler := NewConsumerGroupHandler(mockService, logger)

// 		// Prepare the message payload
// 		payload := map[string]interface{}{"email": "test@example.com", "user_name": "Test"}
// 		payloadBytes, _ := json.Marshal(payload)
// 		message := &sarama.ConsumerMessage{
// 			Topic: "notification.send-welcome",
// 			Value: payloadBytes,
// 		}

// 		// Mock the service call
// 		expectedReq := services.SendRequest{
// 			To:           "test@example.com",
// 			TemplateName: "welcome",
// 			Data:         payload,
// 		}
// 		mockService.On("SendNotification", mock.Anything, expectedReq).Return()

// 		// Mock the session
// 		mockSession := new(MockConsumerGroupSession)
// 		mockSession.On("MarkMessage", message, "").Return()

// 		// Setup channels for the test
// 		messages := make(chan *sarama.ConsumerMessage, 1)
// 		messages <- message
// 		close(messages)
// 		claim := sarama.ConsumerGroupClaim{
// 			Topic:     "notification.send-welcome",
// 			Partition: 0,
// 			Messages:  messages,
// 		}

// 		// Act
// 		err := handler.ConsumeClaim(mockSession, claim)

// 		// Assert
// 		assert.NoError(t, err)
// 		mockService.AssertExpectations(t)
// 		mockSession.AssertExpectations(t)
// 	})

// 	t.Run("Failure Case - Malformed JSON", func(t *testing.T) {
// 		// Arrange
// 		mockService := new(MockNotificationService)
// 		handler := NewConsumerGroupHandler(mockService, logger)
// 		message := &sarama.ConsumerMessage{
// 			Topic: "notification.send-welcome",
// 			Value: []byte("{invalid json"), // Malformed
// 		}
// 		mockSession := new(MockConsumerGroupSession)
// 		mockSession.On("MarkMessage", message, "").Return()

// 		messages := make(chan *sarama.ConsumerMessage, 1)
// 		messages <- message
// 		close(messages)
// 		claim := sarama.ConsumerGroupClaim{Messages: messages}

// 		// Act
// 		err := handler.ConsumeClaim(mockSession, claim)

// 		// Assert
// 		assert.NoError(t, err)
// 		mockService.AssertNotCalled(t, "SendNotification", mock.Anything, mock.Anything) // Service should not be called
// 		mockSession.AssertExpectations(t)
// 	})

// 	t.Run("Failure Case - Missing Email", func(t *testing.T) {
// 		// Arrange
// 		mockService := new(MockNotificationService)
// 		handler := NewConsumerGroupHandler(mockService, logger)
// 		payload := map[string]interface{}{"user_name": "Test"} // Missing "email"
// 		payloadBytes, _ := json.Marshal(payload)
// 		message := &sarama.ConsumerMessage{
// 			Topic: "notification.send-welcome",
// 			Value: payloadBytes,
// 		}
// 		mockSession := new(MockConsumerGroupSession)
// 		mockSession.On("MarkMessage", message, "").Return()

// 		messages := make(chan *sarama.ConsumerMessage, 1)
// 		messages <- message
// 		close(messages)
// 		claim := sarama.ConsumerGroupClaim{Messages: messages}

// 		// Act
// 		err := handler.ConsumeClaim(mockSession, claim)

// 		// Assert
// 		assert.NoError(t, err)
// 		mockService.AssertNotCalled(t, "SendNotification", mock.Anything, mock.Anything)
// 		mockSession.AssertExpectations(t)
// 	})
// }
