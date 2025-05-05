package utils

import (
	"fmt"
	"os"
	"strconv"

	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

// This function is used to send an email
func SendEmail(recipient string, subject string, otp int) error {
	// Reading logger
	logger, err := ConfigLogger()
	if err != nil {
		zap.Must(zap.NewProduction()).Error(err.Error())
		return err
	}

	// SMTP server config
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	body := fmt.Sprintf("<p>Your %s is <span style=\"font-weight:bold;color:blue;\">%d</span>", subject, otp)

	// Creating a new message
	message := gomail.NewMessage()
	message.SetHeader("From", os.Getenv("SMTP_MAIL"))
	message.SetHeader("To", recipient)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", body)

	// int to string conversion of port
	port, err := strconv.Atoi(smtpPort)
	if err != nil {
		logger.Error("unable to convert port to integer", zap.Error(err))
		return err
	}

	// New SMTP Client
	dialer := gomail.NewDialer(smtpHost, port, smtpUsername, smtpPassword)

	// Sending email
	err = dialer.DialAndSend(message)
	if err != nil {
		logger.Error("unable to send email", zap.Error(err))
		return err
	}

	return nil
}
