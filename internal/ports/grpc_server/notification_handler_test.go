package grpc_server

import (
	"context"
	"notification-service/api/proto/pb"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
)

type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) Publish(ctx context.Context, topic string, message interface{}) error {
	args := m.Called(ctx, topic, message)
	return args.Error(0)
}

func TestNotificationHandler_SendNotification_Success(t *testing.T) {
	mockProducer := new(MockProducer)
	handler := NewNotificationHandler(mockProducer)

	data := map[string]interface{}{"user_name": "Bhavana"}
	pbData, _ := structpb.NewStruct(data)

	req := &pb.SendRequest{
		To:           "test@example.com",
		TemplateName: "welcome",
		Data:         pbData,
	}

	mockProducer.On("Publish", mock.Anything, "notification.send-welcome",
		mock.MatchedBy(func(m map[string]interface{}) bool {
			return m["user_name"] == "Bhavana" && m["email"] == "test@example.com"
		}),
	).Return(nil)

	_, err := handler.SendNotification(context.Background(), req)
	assert.NoError(t, err)

	mockProducer.AssertExpectations(t)
}

func TestNotificationHandler_SendNotification_Failure(t *testing.T) {
	mockProducer := new(MockProducer)
	handler := NewNotificationHandler(mockProducer)

	data := map[string]interface{}{"user_name": "Bhavana"}
	pbData, _ := structpb.NewStruct(data)

	req := &pb.SendRequest{
		To:           "test@example.com",
		TemplateName: "welcome",
		Data:         pbData,
	}

	mockProducer.On("Publish", mock.Anything, "notification.send-welcome", mock.Anything).
		Return(assert.AnError)

	_, err := handler.SendNotification(context.Background(), req)
	assert.NoError(t, err) // handler swallows the error

	mockProducer.AssertExpectations(t)
}

func TestNotificationHandler_SendNotification_Sanitization(t *testing.T) {
	mockProducer := new(MockProducer)
	handler := NewNotificationHandler(mockProducer)

	req := &pb.SendRequest{
		To:           "test@example.com",
		TemplateName: "welcome!?. ",
		Data:         nil,
	}

	mockProducer.On("Publish", mock.Anything, "notification.send-welcome", mock.Anything).
		Return(nil)

	_, err := handler.SendNotification(context.Background(), req)
	assert.NoError(t, err)

	mockProducer.AssertExpectations(t)
}

func TestNotificationHandler_SendNotification_NilData(t *testing.T) {
	mockProducer := new(MockProducer)
	handler := NewNotificationHandler(mockProducer)

	// Data is nil
	req := &pb.SendRequest{
		To:           "nil@example.com",
		TemplateName: "welcome",
		Data:         nil,
	}

	mockProducer.On("Publish", mock.Anything, "notification.send-welcome",
		mock.MatchedBy(func(m map[string]interface{}) bool {
			return m["email"] == "nil@example.com"
		}),
	).Return(nil)

	_, err := handler.SendNotification(context.Background(), req)
	assert.NoError(t, err)

	mockProducer.AssertExpectations(t)
}
