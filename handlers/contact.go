package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/gomail.v2"
)

// ContactRequest represents the contact form data
type ContactRequest struct {
	Subject string `json:"subject" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
	Name    string `json:"name"`
	Message string `json:"message" binding:"required"`
}

// ContactResponse represents the response from the contact API
type ContactResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// HandleContactPage renders the contact page
func HandleContactPage(c *gin.Context) {
	// Read version from file
	versionBytes, err := os.ReadFile("static/.version")
	version := "0.0.0"
	if err == nil {
		version = strings.TrimSpace(string(versionBytes))
	}

	c.HTML(http.StatusOK, "contact.html", gin.H{
		"title":   "Contact Us",
		"version": version,
	})
}

// HandleContactSubmit handles contact form submissions
func HandleContactSubmit(c *gin.Context) {
	var req ContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Contact] Invalid request data: %v", err)
		c.JSON(http.StatusBadRequest, ContactResponse{
			Success: false,
			Message: "Invalid request data",
		})
		return
	}

	// Validate required fields
	if req.Subject == "" || req.Email == "" || req.Message == "" {
		c.JSON(http.StatusBadRequest, ContactResponse{
			Success: false,
			Message: "Please fill in all required fields",
		})
		return
	}

	// Send email notification
	if err := sendContactEmail(req); err != nil {
		log.Printf("[Contact] Failed to send contact email: %v", err)
		c.JSON(http.StatusInternalServerError, ContactResponse{
			Success: false,
			Message: "Failed to send message. Please try again later.",
		})
		return
	}

	log.Printf("[Contact] Contact form submitted successfully from %s", req.Email)
	c.JSON(http.StatusOK, ContactResponse{
		Success: true,
		Message: "Thank you for your message! We'll get back to you soon.",
	})
}

// sendContactEmail sends a contact form email notification
func sendContactEmail(req ContactRequest) error {
	// Get email configuration from environment
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	adminEmail := os.Getenv("ADMIN_EMAIL")

	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" || adminEmail == "" {
		log.Printf("[Contact] Email configuration missing, skipping email send")
		return nil // Don't fail the request if email is not configured
	}

	// Create email message
	m := gomail.NewMessage()
	m.SetHeader("From", smtpUser)
	m.SetHeader("To", adminEmail)
	m.SetHeader("Subject", fmt.Sprintf("[Disko Contact] %s - %s", req.Subject, req.Email))

	// Set email body
	body := generateContactEmailBody(req)
	m.SetBody("text/html", body)

	// Send email
	d := gomail.NewDialer(smtpHost, 587, smtpUser, smtpPass)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send contact email: %w", err)
	}

	return nil
}

// generateContactEmailBody generates the HTML body for contact emails
func generateContactEmailBody(req ContactRequest) string {
	now := time.Now().Format("January 2, 2006 at 3:04 PM MST")

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #3b82f6; color: white; padding: 20px; border-radius: 8px 8px 0 0; }
        .content { background: #f9fafb; padding: 20px; border-radius: 0 0 8px 8px; }
        .field { margin-bottom: 15px; }
        .label { font-weight: bold; color: #374151; }
        .value { background: white; padding: 10px; border-radius: 4px; border: 1px solid #d1d5db; }
        .footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #e5e7eb; font-size: 14px; color: #6b7280; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>New Contact Form Submission</h1>
            <p>Received on %s</p>
        </div>
        <div class="content">
            <div class="field">
                <div class="label">Subject:</div>
                <div class="value">%s</div>
            </div>
            <div class="field">
                <div class="label">From:</div>
                <div class="value">%s</div>
            </div>
            <div class="field">
                <div class="label">Name:</div>
                <div class="value">%s</div>
            </div>
            <div class="field">
                <div class="label">Message:</div>
                <div class="value">%s</div>
            </div>
            <div class="footer">
                <p>This message was sent from the Disko contact form.</p>
            </div>
        </div>
    </div>
</body>
</html>`, now, req.Subject, req.Email, req.Name, req.Message)

	return html
}
