package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/gomail.v2"
)

// ContactRequest represents the contact form data
type ContactRequest struct {
	Subject string `json:"subject" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
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

// Simple in-memory rate limiting for contact form
var contactRateLimitStore = make(map[string]time.Time)

// isContactRateLimited checks if an IP is rate limited for contact form
func isContactRateLimited(ip string) bool {
	if lastRequest, exists := contactRateLimitStore[ip]; exists {
		// Rate limit: 1 contact form submission per hour per IP
		if time.Since(lastRequest) < time.Hour {
			return true
		}
	}
	return false
}

// setContactRateLimit sets the rate limit for an IP
func setContactRateLimit(ip string) {
	contactRateLimitStore[ip] = time.Now()

	// Clean up old entries after 2 hours
	go func() {
		time.Sleep(2 * time.Hour)
		delete(contactRateLimitStore, ip)
	}()
}

// HandleContactSubmit handles contact form submissions
func HandleContactSubmit(c *gin.Context) {
	clientIP := c.ClientIP()

	// Check rate limiting
	if isContactRateLimited(clientIP) {
		log.Printf("[Contact] Rate limited contact form submission from IP: %s", clientIP)
		c.JSON(http.StatusTooManyRequests, ContactResponse{
			Success: false,
			Message: "Too many contact form submissions. Please wait at least 1 hour before submitting another message.",
		})
		return
	}

	var req ContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Contact] Invalid request data from IP %s: %v", clientIP, err)
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

	// Set rate limit before processing
	setContactRateLimit(clientIP)

	// Send email notification
	if err := sendContactEmail(req); err != nil {
		log.Printf("[Contact] Failed to send contact email from IP %s: %v", clientIP, err)
		c.JSON(http.StatusInternalServerError, ContactResponse{
			Success: false,
			Message: "Failed to send message. Please try again later.",
		})
		return
	}

	log.Printf("[Contact] Contact form submitted successfully from IP %s, Email: %s", clientIP, req.Email)
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
	smtpPortInt, _ := strconv.Atoi(smtpPort)
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")

	//
	fromName := os.Getenv("FROM_NAME")
	fromEmail := os.Getenv("FROM_EMAIL")

	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" || fromEmail == "" {
		log.Printf("[Contact] Email configuration missing, skipping email send")
		return nil // Don't fail the request if email is not configured
	}

	// Create email message
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", fromName, fromEmail))
	m.SetHeader("To", fromEmail)
	m.SetHeader("Subject", fmt.Sprintf("[Disko][Contact] %s - %s", req.Subject, req.Email))

	// Set email body
	body := generateContactEmailBody(req)
	m.SetBody("text/html", body)

	// Send email
	d := gomail.NewDialer(smtpHost, smtpPortInt, smtpUser, smtpPass)
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
                <div class="label">Message:</div>
                <div class="value">%s</div>
            </div>
            <div class="footer">
                <p>This message was sent from the Disko App.</p>
            </div>
        </div>
    </div>
</body>
</html>`, now, req.Subject, req.Email, req.Message)

	return html
}
