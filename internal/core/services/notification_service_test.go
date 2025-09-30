package services

import (
	"bytes"
	"context"
	"errors"
	"notification-service/internal/adapters/database/db"
	"notification-service/internal/core/repository"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

// --- Mock Implementations ---

// MockTemplateRepo is a mock for the TemplateRepository.
type MockTemplateRepo struct {
	mock.Mock
}

func (m *MockTemplateRepo) GetTemplate(ctx context.Context, name string) (*repository.Template, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Template), args.Error(1)
}

// MockLogRepo is a mock for the NotificationLogRepository.
type MockLogRepo struct {
	mock.Mock
}

func (m *MockLogRepo) CreateLog(ctx context.Context, params db.CreateNotificationLogParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

// MockMailer is a mock for the mailme.Mailer.
type MockMailer struct {
	mock.Mock
}

func (m *MockMailer) Mail(to, subject, html, text string, data map[string]interface{}) error {
	args := m.Called(to, subject, html, text, data)
	return args.Error(0)
}

// --- Test Suite ---

func TestNotificationService_SendNotification(t *testing.T) {
	// Setup a logger that discards output for clean test runs
	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})

	// Test cases
	testCases := []struct {
		name              string
		setupMocks        func(*MockTemplateRepo, *MockLogRepo, *MockMailer)
		request           SendRequest
		expectMailerCall  bool
		expectedLogStatus string
	}{
		{
			name: "Success Case",
			setupMocks: func(templateRepo *MockTemplateRepo, logRepo *MockLogRepo, mailer *MockMailer) {
				// Mock GetTemplate to return a valid template
				templateRepo.On("GetTemplate", mock.Anything, "welcome").Return(&repository.Template{
					Subject:  "Welcome!",
					BodyHTML: "<h1>Hello {{.user_name}}</h1>",
					BodyText: "Hello {{.user_name}}",
				}, nil)

				// Mock Mailer to succeed
				mailer.On("Mail", "test@example.com", "Welcome!", "<h1>Hello {{.user_name}}</h1>", "Hello {{.user_name}}", mock.Anything).Return(nil)

				// Mock CreateLog to succeed
				logRepo.On("CreateLog", mock.Anything, mock.MatchedBy(func(params db.CreateNotificationLogParams) bool {
					return params.Status == "sent"
				})).Return(nil)
			},
			request: SendRequest{
				To:           "test@example.com",
				TemplateName: "welcome",
				Data:         map[string]interface{}{"user_name": "Test User"},
			},
			expectMailerCall:  true,
			expectedLogStatus: "sent",
		},
		{
			name: "Template Not Found Failure",
			setupMocks: func(templateRepo *MockTemplateRepo, logRepo *MockLogRepo, mailer *MockMailer) {
				// Mock GetTemplate to return an error
				templateRepo.On("GetTemplate", mock.Anything, "non-existent").Return(nil, errors.New("template not found"))

				// Mock CreateLog to record the failure
				logRepo.On("CreateLog", mock.Anything, mock.MatchedBy(func(params db.CreateNotificationLogParams) bool {
					return params.Status == "failed" && params.Details.String == "template not found"
				})).Return(nil)
			},
			request: SendRequest{
				To:           "test@example.com",
				TemplateName: "non-existent",
				Data:         map[string]interface{}{},
			},
			expectMailerCall:  false, // Mailer should not be called if template fails
			expectedLogStatus: "failed",
		},
		{
			name: "Email Sending Failure",
			setupMocks: func(templateRepo *MockTemplateRepo, logRepo *MockLogRepo, mailer *MockMailer) {
				// Mock GetTemplate to succeed
				templateRepo.On("GetTemplate", mock.Anything, "welcome").Return(&repository.Template{
					Subject:  "Welcome!",
					BodyHTML: "<h1>Hello</h1>",
				}, nil)

				// Mock Mailer to return an error
				mailer.On("Mail", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("SMTP connection failed"))

				// Mock CreateLog to record the failure
				logRepo.On("CreateLog", mock.Anything, mock.MatchedBy(func(params db.CreateNotificationLogParams) bool {
					return params.Status == "failed" && params.Details.String == "failed to send email via SMTP: SMTP connection failed"
				})).Return(nil)
			},
			request: SendRequest{
				To:           "test@example.com",
				TemplateName: "welcome",
				Data:         map[string]interface{}{},
			},
			expectMailerCall:  true,
			expectedLogStatus: "failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create instances of our mocks
			mockTemplateRepo := new(MockTemplateRepo)
			mockLogRepo := new(MockLogRepo)
			mockMailer := new(MockMailer)

			// Setup the expectations for this test case
			tc.setupMocks(mockTemplateRepo, mockLogRepo, mockMailer)

			// Create the service with the mocks
			service := NewNotificationService(mockTemplateRepo, mockLogRepo, mockMailer, logger)

			// Execute the method we want to test
			service.SendNotification(context.Background(), tc.request)

			// Assert that our mocks were called as expected
			mockTemplateRepo.AssertExpectations(t)
			mockLogRepo.AssertExpectations(t)
			mockMailer.AssertExpectations(t)

			// A simple way to check if mailer was called or not
			if tc.expectMailerCall {
				mockMailer.AssertNumberOfCalls(t, "Mail", 1)
			} else {
				mockMailer.AssertNumberOfCalls(t, "Mail", 0)
			}
		})
	}
}
