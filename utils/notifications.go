package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"disko-backend/models"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// NotificationService handles multi-channel notifications
type NotificationService struct {
	emailEnabled    bool
	slackEnabled    bool
	webhookEnabled  bool
	slackWebhookURL string
	webhookURL      string
}

// FeedbackNotification represents a feedback notification
type FeedbackNotification struct {
	BoardID      string    `json:"boardId"`
	BoardName    string    `json:"boardName"`
	IdeaID       string    `json:"ideaId"`
	IdeaTitle    string    `json:"ideaTitle"`
	FeedbackType string    `json:"feedbackType"`
	ClientIP     string    `json:"clientIp"`
	Timestamp    time.Time `json:"timestamp"`
	AdminEmail   string    `json:"adminEmail,omitempty"`
}

// SlackMessage represents a Slack webhook message
type SlackMessage struct {
	Text        string            `json:"text"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

// SlackAttachment represents a Slack message attachment
type SlackAttachment struct {
	Color  string       `json:"color"`
	Fields []SlackField `json:"fields"`
}

// SlackField represents a field in a Slack attachment
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// NewNotificationService creates a new notification service
func NewNotificationService() *NotificationService {
	return &NotificationService{
		emailEnabled:    os.Getenv("EMAIL_ENABLED") == "true",
		slackEnabled:    os.Getenv("SLACK_WEBHOOK_URL") != "",
		webhookEnabled:  os.Getenv("WEBHOOK_URL") != "",
		slackWebhookURL: os.Getenv("SLACK_WEBHOOK_URL"),
		webhookURL:      os.Getenv("WEBHOOK_URL"),
	}
}

// SendFeedbackNotification sends notifications across all configured channels
func (ns *NotificationService) SendFeedbackNotification(boardID, ideaID, feedbackType, clientIP string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get board and idea information
	notification, err := ns.buildNotification(ctx, boardID, ideaID, feedbackType, clientIP)
	if err != nil {
		log.Printf("Failed to build notification: %v", err)
		return
	}

	// Send notifications concurrently
	if ns.emailEnabled {
		go ns.sendEmailNotification(notification)
	}

	if ns.slackEnabled {
		go ns.sendSlackNotification(notification)
	}

	if ns.webhookEnabled {
		go ns.sendWebhookNotification(notification)
	}

	// Trigger real-time feedback animation on admin board
	emoji := ""
	if len(feedbackType) > 6 && feedbackType[:6] == "emoji:" {
		emoji = feedbackType[6:]
		feedbackType = "emoji"
	}
	BroadcastFeedbackAnimation(boardID, ideaID, feedbackType, emoji)

	log.Printf("Feedback notification sent: Board=%s, Idea=%s, Type=%s",
		boardID, ideaID, feedbackType)
}

// buildNotification creates a notification object with board and idea details
func (ns *NotificationService) buildNotification(ctx context.Context, boardID, ideaID, feedbackType, clientIP string) (*FeedbackNotification, error) {
	// Get board information
	boardsCollection := models.GetCollection(models.BoardsCollection)
	var board models.Board
	err := boardsCollection.FindOne(ctx, bson.M{"_id": boardID}).Decode(&board)
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %v", err)
	}

	// Get idea information
	ideasCollection := models.GetCollection(models.IdeasCollection)
	var idea models.Idea
	err = ideasCollection.FindOne(ctx, bson.M{"_id": ideaID}).Decode(&idea)
	if err != nil {
		return nil, fmt.Errorf("failed to get idea: %v", err)
	}

	// TODO: Get admin email from Clerk user info
	// For now, we'll use a placeholder
	adminEmail := "admin@example.com"

	return &FeedbackNotification{
		BoardID:      boardID,
		BoardName:    board.Name,
		IdeaID:       ideaID,
		IdeaTitle:    idea.OneLiner,
		FeedbackType: feedbackType,
		ClientIP:     clientIP,
		Timestamp:    time.Now().UTC(),
		AdminEmail:   adminEmail,
	}, nil
}

// sendEmailNotification sends an email notification
func (ns *NotificationService) sendEmailNotification(notification *FeedbackNotification) {
	// This is a placeholder for email notification
	// In a real implementation, you would integrate with an email service like:
	// - SendGrid
	// - AWS SES
	// - Mailgun
	// - SMTP server

	log.Printf("EMAIL NOTIFICATION: %s received feedback on '%s' in board '%s'",
		notification.FeedbackType, notification.IdeaTitle, notification.BoardName)

	// Example email content
	subject := fmt.Sprintf("New feedback on your idea: %s", notification.IdeaTitle)
	body := fmt.Sprintf(`
Hello,

You've received new feedback on your idea "%s" in board "%s".

Feedback Type: %s
Time: %s
IP Address: %s

View your board: %s

Best regards,
Disko Team
`,
		notification.IdeaTitle,
		notification.BoardName,
		notification.FeedbackType,
		notification.Timestamp.Format("2006-01-02 15:04:05 UTC"),
		notification.ClientIP,
		fmt.Sprintf("https://yourdomain.com/board/%s", notification.BoardID),
	)

	// TODO: Implement actual email sending
	log.Printf("Email would be sent to %s with subject: %s", notification.AdminEmail, subject)
	log.Printf("Email body: %s", body)
}

// sendSlackNotification sends a Slack webhook notification
func (ns *NotificationService) sendSlackNotification(notification *FeedbackNotification) {
	if ns.slackWebhookURL == "" {
		return
	}

	// Create Slack message
	message := SlackMessage{
		Text: "🎉 New feedback received on your Disko board!",
		Attachments: []SlackAttachment{
			{
				Color: "#36a64f", // Green color
				Fields: []SlackField{
					{
						Title: "Board",
						Value: notification.BoardName,
						Short: true,
					},
					{
						Title: "Idea",
						Value: notification.IdeaTitle,
						Short: true,
					},
					{
						Title: "Feedback Type",
						Value: notification.FeedbackType,
						Short: true,
					},
					{
						Title: "Time",
						Value: notification.Timestamp.Format("2006-01-02 15:04:05 UTC"),
						Short: true,
					},
					{
						Title: "Board Link",
						Value: fmt.Sprintf("https://yourdomain.com/board/%s", notification.BoardID),
						Short: false,
					},
				},
			},
		},
	}

	// Send to Slack
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal Slack message: %v", err)
		return
	}

	resp, err := http.Post(ns.slackWebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to send Slack notification: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Slack notification failed with status: %d", resp.StatusCode)
		return
	}

	log.Printf("Slack notification sent successfully")
}

// sendWebhookNotification sends a generic webhook notification
func (ns *NotificationService) sendWebhookNotification(notification *FeedbackNotification) {
	if ns.webhookURL == "" {
		return
	}

	// Send the full notification object as JSON
	jsonData, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Failed to marshal webhook notification: %v", err)
		return
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Post(ns.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to send webhook notification: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("Webhook notification failed with status: %d", resp.StatusCode)
		return
	}

	log.Printf("Webhook notification sent successfully")
}

// Global notification service instance
var notificationService *NotificationService

// InitNotificationService initializes the global notification service
func InitNotificationService() {
	notificationService = NewNotificationService()
}

// SendFeedbackNotification is a convenience function to send notifications
func SendFeedbackNotification(boardID, ideaID, feedbackType, clientIP string) {
	if notificationService == nil {
		InitNotificationService()
	}
	notificationService.SendFeedbackNotification(boardID, ideaID, feedbackType, clientIP)
}
