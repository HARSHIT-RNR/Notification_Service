package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"notification-service/internal/core/repository"

	"github.com/sirupsen/logrus"
)

type FileTemplateRepo struct {
	basePath string
	logger   *logrus.Logger
}

func NewFileTemplateRepo(basePath string, logger *logrus.Logger) *FileTemplateRepo {
	return &FileTemplateRepo{
		basePath: basePath,
		logger:   logger,
	}
}

func (r *FileTemplateRepo) GetTemplate(ctx context.Context, name string) (*repository.Template, error) {
	log := r.logger.WithField("template_name", name)
	templateDir := filepath.Join(r.basePath, name)

	subject, err := r.readFile(filepath.Join(templateDir, "subject.txt"))
	if err != nil {
		log.WithError(err).Error("Failed to read subject.txt")
		return nil, err
	}

	bodyHTML, err := r.readFile(filepath.Join(templateDir, "body.html"))
	if err != nil {
		log.WithError(err).Error("Failed to read body.html")
		return nil, err
	}

	bodyText, err := r.readFile(filepath.Join(templateDir, "body.txt"))
	if err != nil {
		// This one can be optional, maybe just a warning
		log.WithError(err).Warn("Could not read body.txt, proceeding without it")
		bodyText = ""
	}

	// Meta.json is useful for validation but not strictly required by the mailer
	// You could extend this to parse and use the metadata
	_, err = r.readFile(filepath.Join(templateDir, "meta.json"))
	if err != nil {
		log.WithError(err).Warn("Could not read meta.json")
	}

	return &repository.Template{
		Name:     name,
		Subject:  subject,
		BodyHTML: bodyHTML,
		BodyText: bodyText,
	}, nil
}

func (r *FileTemplateRepo) readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("could not read file %s: %w", path, err)
	}
	return string(data), nil
}
