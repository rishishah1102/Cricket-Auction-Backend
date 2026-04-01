package utils

import (
	"fmt"
	"os"
	"strconv"

	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

func buildOTPEmailHTML(subject string, otp int) string {
	otpStr := fmt.Sprintf("%d", otp)

	purposeText := "log in to"
	if subject == "Registration OTP" {
		purposeText = "complete your registration on"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"></head>
<body style="margin:0;padding:0;background-color:#f0f2f5;font-family:'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif;">
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f0f2f5;padding:32px 16px;">
<tr><td align="center">
<table role="presentation" width="520" cellpadding="0" cellspacing="0" style="max-width:520px;width:100%%;background-color:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,0.08);">

  <!-- Header -->
  <tr>
    <td style="background:linear-gradient(135deg,#0d2249 0%%,#143676 100%%);padding:36px 40px 28px;text-align:center;">
      <table role="presentation" cellpadding="0" cellspacing="0" style="margin:0 auto;">
        <tr>
          <td style="width:44px;height:44px;background:linear-gradient(135deg,#3b82f6,#6366f1);border-radius:12px;text-align:center;vertical-align:middle;">
            <span style="font-size:24px;color:#ffffff;">&#9928;</span>
          </td>
          <td style="width:12px;"></td>
          <td style="font-size:24px;font-weight:700;color:#ffffff;letter-spacing:0.5px;font-family:'Segoe UI',Roboto,Arial,sans-serif;">
            Cricket Auction
          </td>
        </tr>
      </table>
      <p style="color:rgba(255,255,255,0.7);font-size:14px;margin:14px 0 0;font-weight:400;">
        Secure Verification Code
      </p>
    </td>
  </tr>

  <!-- Body -->
  <tr>
    <td style="padding:36px 40px 12px;">
      <h1 style="margin:0 0 8px;font-size:22px;font-weight:700;color:#0d2249;font-family:'Segoe UI',Roboto,Arial,sans-serif;">
        Your One-Time Password
      </h1>
      <p style="margin:0 0 28px;font-size:15px;color:#64748b;line-height:1.6;">
        Use the code below to %s <strong style="color:#0d2249;">Cricket Auction</strong>.
        Do not share this code with anyone.
      </p>

      <!-- OTP Code -->
      <table role="presentation" cellpadding="0" cellspacing="0" style="margin:0 auto 28px;">
        <tr>
          <td style="background-color:#f0f4ff;border:2px solid #3b82f6;border-radius:12px;padding:16px 36px;text-align:center;">
            <span style="font-size:34px;font-weight:800;color:#0d2249;letter-spacing:14px;font-family:'Segoe UI',Roboto,monospace;">%s</span>
          </td>
        </tr>
      </table>

      <!-- Timer badge -->
      <table role="presentation" cellpadding="0" cellspacing="0" style="margin:0 auto 28px;">
        <tr>
          <td style="background-color:#fef3c7;border:1px solid #fbbf24;border-radius:8px;padding:10px 20px;text-align:center;">
            <span style="font-size:13px;color:#92400e;font-weight:600;">&#9200; This code expires in 5 minutes</span>
          </td>
        </tr>
      </table>

      <!-- Divider -->
      <table role="presentation" width="100%%" cellpadding="0" cellspacing="0">
        <tr><td style="border-top:1px solid #e2e8f0;padding-top:20px;"></td></tr>
      </table>

      <!-- Security note -->
      <table role="presentation" cellpadding="0" cellspacing="0" style="width:100%%;">
        <tr>
          <td style="width:32px;vertical-align:top;padding-top:2px;">
            <span style="font-size:18px;">&#128274;</span>
          </td>
          <td style="font-size:13px;color:#94a3b8;line-height:1.5;">
            If you didn't request this code, you can safely ignore this email.
            Someone may have entered your email address by mistake.
          </td>
        </tr>
      </table>
    </td>
  </tr>

  <!-- Footer -->
  <tr>
    <td style="background-color:#f8fafc;padding:24px 40px;text-align:center;border-top:1px solid #e2e8f0;">
      <p style="margin:0 0 4px;font-size:13px;color:#94a3b8;">
        &copy; 2026 Cricket Auction. All rights reserved.
      </p>
      <p style="margin:0;font-size:12px;color:#cbd5e1;">
        This is an automated message — please do not reply.
      </p>
    </td>
  </tr>

</table>
</td></tr>
</table>
</body>
</html>`, purposeText, otpStr)
}

// SendEmail sends an OTP email to the recipient
func SendEmail(recipient string, subject string, otp int, logger *zap.Logger) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	body := buildOTPEmailHTML(subject, otp)

	message := gomail.NewMessage()
	message.SetHeader("From", os.Getenv("SMTP_MAIL"))
	message.SetHeader("To", recipient)
	message.SetHeader("Subject", subject+" — Cricket Auction")
	message.SetBody("text/html", body)

	port, err := strconv.Atoi(smtpPort)
	if err != nil {
		logger.Error("unable to convert port to integer", zap.Error(err))
		return err
	}

	dialer := gomail.NewDialer(smtpHost, port, smtpUsername, smtpPassword)

	err = dialer.DialAndSend(message)
	if err != nil {
		logger.Error("unable to send email", zap.Error(err))
		return err
	}

	return nil
}
