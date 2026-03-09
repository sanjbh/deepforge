package tools

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type FileEmailSender struct {
	outputDir string
}

func NewFileEmailSender(outputDir string) *FileEmailSender {
	return &FileEmailSender{outputDir: outputDir}
}

func (e *FileEmailSender) Send(subject string, htmlBody string) error {
	htmlTemplate := `
		<!DOCTYPE html>
		<html>
			<head><title>%s</title>
		</head>
		<body>%s</body>
		</html>
	`
	html := fmt.Sprintf(htmlTemplate, subject, htmlBody)

	fileName := filepath.Join(e.outputDir, time.Now().Format("report_20060102_150405.html"))

	if err := os.MkdirAll(e.outputDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	err := os.WriteFile(fileName, []byte(html), 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	log.Printf("Email written to %s successfully", fileName)
	return nil
}
