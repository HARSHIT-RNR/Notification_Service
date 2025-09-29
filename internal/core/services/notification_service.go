package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"notification-service/internal/adapters/database/db"
	"notification-service/internal/adapters/mailme"
	"notification-service/internal/core/repository"

	"github.com/sirupsen/logrus"
)

type SendRequest struct {
	To           string
	TemplateName string
	Data         map[string]interface{}
}

type NotificationService struct {
	templateRepo repository.TemplateRepository
	logRepo      repository.NotificationLogRepository
	mailer       *mailme.Mailer
	logger       *logrus.Logger
}

func NewNotificationService(
	templateRepo repository.TemplateRepository,
	logRepo repository.NotificationLogRepository,
	mailer *mailme.Mailer,
	logger *logrus.Logger,
) *NotificationService {
	return &NotificationService{
		templateRepo: templateRepo,
		logRepo:      logRepo,
		mailer:       mailer,
		logger:       logger,
	}
}

func (s *NotificationService) SendNotification(ctx context.Context, req SendRequest) {
	log := s.logger.WithFields(logrus.Fields{
		"recipient": req.To,
		"template":  req.TemplateName,
	})

	template, err := s.templateRepo.GetTemplate(ctx, req.TemplateName)
	if err != nil {
		log.WithError(err).Error("Failed to get template")
		s.logAttempt(ctx, req, "failed", "template not found")
		return
	}

	// The mailer's Mail function handles subject/body parsing
	err = s.mailer.Mail(req.To, template.Subject, template.BodyHTML, template.BodyText, req.Data)

	if err != nil {
		log.WithError(err).Error("Failed to send notification")
		s.logAttempt(ctx, req, "failed", err.Error())
		return
	}

	log.Info("Notification sent successfully")
	s.logAttempt(ctx, req, "sent", "Successfully sent")
}

func (s *NotificationService) logAttempt(ctx context.Context, req SendRequest, status, details string) {
	dataJSON, err := json.Marshal(req.Data)
	if err != nil {
		s.logger.WithError(err).Error("Failed to marshal notification data for logging")
		dataJSON = []byte("{}") // Log empty JSON on error
	}

	params := db.CreateNotificationLogParams{
		Recipient:    req.To,
		TemplateName: req.TemplateName,
		Status:       status,
		Details:      sql.NullString{String: details, Valid: true},
		Data:         dataJSON,
		AttemptedAt:  time.Now(),
	}

	if err := s.logRepo.CreateLog(ctx, params); err != nil {
		s.logger.WithError(err).Error("Failed to write notification log to database")
	}
}
