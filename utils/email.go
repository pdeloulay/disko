package utils

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"disko-backend/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"gopkg.in/gomail.v2"
)

// SendBoardInviteEmail sends an HTML invitation email for a board
func SendBoardInviteEmail(email, subject, message string, board models.Board, userID string) error {
	// Get email configuration from environment variables
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	fromEmail := os.Getenv("FROM_EMAIL")

	log.Printf("[Email] Configuration check - SMTP_HOST: %s, SMTP_PORT: %s, SMTP_USER: %s, FROM_EMAIL: %s, APP_URL: %s",
		smtpHost, smtpPortStr, smtpUser, fromEmail, os.Getenv("APP_URL"))

	log.Printf("[Email] Requested email - To: %s, Subject: %s", email, subject)

	if smtpHost == "" || smtpPortStr == "" || smtpUser == "" || smtpPass == "" || fromEmail == "" {
		log.Printf("[Email] Configuration incomplete - missing required environment variables")
		return fmt.Errorf("email configuration incomplete - check SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS, FROM_EMAIL environment variables")
	}

	smtpPort, _ := strconv.Atoi(smtpPortStr)

	// Get user email from Clerk if userID is provided
	// fromEmailWithName := fromEmail
	// if userID != "" {
	// 	_, err := getUserEmailFromClerk(userID)
	// 	if err != nil {
	// 		log.Printf("[Email] Failed to get user email from Clerk: %v, using default email", err)
	// 	} else {
	// 		fromEmailWithName = fmt.Sprintf("Disko <noreply@%s>", extractDomain(fromEmail))
	// 	}
	// }

	// Create email message - send to the email address provided in the form
	m := gomail.NewMessage()
	m.SetHeader("From", fromEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", generateInviteEmailHTML(board, message))

	// Create dialer
	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)

	// Send email
	if err := d.DialAndSend(m); err != nil {
		log.Printf("[Email] Failed to send invite email - Error: %v, To: %s, BoardID: %s", err, email, board.ID)
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.Printf("[Email] Invite email sent successfully - To: %s, BoardID: %s, BoardName: %s", email, board.ID, board.Name)
	return nil
}

// getUserEmailFromClerk retrieves user email from Clerk
func getUserEmailFromClerk(userID string) (string, error) {
	// Initialize Clerk client
	clerkSecretKey := os.Getenv("CLERK_SECRET_KEY")
	if clerkSecretKey == "" {
		return "", fmt.Errorf("CLERK_SECRET_KEY not set")
	}

	// For now, we'll use a placeholder since the Clerk SDK might not be available
	// In a real implementation, you would use the Clerk SDK to get user information
	log.Printf("[Email] Getting user email from Clerk for userID: %s", userID)

	return "", fmt.Errorf("Clerk SDK integration not yet implemented")
}

// generateInviteEmailHTML creates a compelling HTML email template with Disko branding
func generateInviteEmailHTML(board models.Board, message string) string {
	publicURL := fmt.Sprintf("%s/public/%s", os.Getenv("APP_URL"), board.PublicLink)

	// Get board statistics
	ideasCount := getBoardIdeasCount(board.ID)
	reactionsCount := getBoardReactionsCount(board.ID)
	recentIdeas := getRecentIdeas(board.ID, 5)

	// Build the HTML template with proper escaping
	htmlTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.BoardName}} - Board Invitation</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            margin: 0;
            padding: 0;
            background-color: #f9fafb;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            background-color: #ffffff;
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
        }
        .header {
            background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
            color: white;
            padding: 40px 30px;
            text-align: center;
        }
        .logo {
            font-size: 32px;
            font-weight: 700;
            margin-bottom: 16px;
        }
        .header h1 {
            margin: 0;
            font-size: 28px;
            font-weight: 700;
        }
        .header p {
            margin: 10px 0 0 0;
            opacity: 0.9;
            font-size: 16px;
        }
        .content {
            padding: 40px 30px;
        }
        .board-info {
            background-color: #f8fafc;
            border-radius: 8px;
            padding: 24px;
            margin-bottom: 30px;
            border-left: 4px solid #3b82f6;
        }
        .board-name {
            font-size: 24px;
            font-weight: 700;
            color: #1e293b;
            margin: 0 0 8px 0;
        }
        .board-description {
            color: #64748b;
            margin: 0 0 16px 0;
        }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
            gap: 16px;
            margin-bottom: 24px;
        }
        .stat-item {
            text-align: center;
            padding: 16px;
            background-color: #ffffff;
            border-radius: 8px;
            border: 1px solid #e2e8f0;
        }
        .stat-number {
            font-size: 24px;
            font-weight: 700;
            color: #3b82f6;
            display: block;
        }
        .stat-label {
            font-size: 12px;
            color: #64748b;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        .emoji-recaps {
            margin-top: 20px;
            padding: 16px;
            background-color: #f8fafc;
            border-radius: 8px;
            text-align: center;
            border: 1px solid #e2e8f0;
        }
        .recaps-label {
            display: block;
            font-size: 12px;
            color: #64748b;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            margin-bottom: 8px;
        }
        .recaps-emojis {
            font-size: 24px;
            letter-spacing: 8px;
        }
        .recent-ideas {
            margin-bottom: 30px;
        }
        .recent-ideas h3 {
            font-size: 18px;
            font-weight: 600;
            color: #1e293b;
            margin: 0 0 16px 0;
        }
        .idea-item {
            padding: 12px 16px;
            background-color: #f8fafc;
            border-radius: 6px;
            margin-bottom: 8px;
            border-left: 3px solid #10b981;
        }
        .idea-title {
            font-weight: 600;
            color: #1e293b;
            margin: 0 0 4px 0;
        }
        .idea-meta {
            font-size: 12px;
            color: #64748b;
        }
        .idea-feedback-summary {
            margin-top: 8px;
            padding-top: 8px;
            border-top: 1px solid #e2e8f0;
        }
        .feedback-label {
            font-size: 11px;
            color: #64748b;
            font-weight: 600;
            margin-right: 8px;
        }
        .feedback-items {
            font-size: 12px;
            color: #3b82f6;
        }
        .cta-section {
            text-align: center;
            padding: 30px;
            background-color: #f8fafc;
            border-radius: 8px;
        }
        .cta-button {
            display: inline-block;
            background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
            color: white;
            text-decoration: none;
            padding: 16px 32px;
            border-radius: 8px;
            font-weight: 600;
            font-size: 16px;
            transition: transform 0.2s ease;
        }
        .cta-button:hover {
            transform: translateY(-2px);
        }
        .footer {
            background-color: #f1f5f9;
            padding: 24px 30px;
            text-align: center;
            color: #64748b;
            font-size: 14px;
        }
        .footer-logo {
            margin-bottom: 16px;
            text-align: center;
        }
        .footer-logo img {
            max-width: 120px;
            height: auto;
        }
        .footer p {
            margin: 0 0 8px 0;
        }
        .footer a {
            color: #3b82f6;
            text-decoration: none;
        }
        .footer-links {
            margin-top: 16px;
            padding-top: 16px;
            border-top: 1px solid #e2e8f0;
        }
        .footer-links a {
            margin: 0 8px;
            color: #64748b;
            text-decoration: none;
        }
        .footer-links a:hover {
            color: #3b82f6;
        }
        .footer-cta {
            margin: 16px 0;
            padding: 12px;
            background-color: #f8fafc;
            border-radius: 6px;
            border-left: 3px solid #3b82f6;
        }
        .footer-cta p {
            margin: 0;
            color: #1e293b;
            font-weight: 500;
        }
        .footer-cta a {
            color: #3b82f6;
            text-decoration: none;
            font-weight: 600;
        }
        .footer-cta a:hover {
            text-decoration: underline;
        }
        @media (max-width: 600px) {
            .container {
                margin: 0;
                border-radius: 0;
            }
            .header, .content, .footer {
                padding: 20px;
            }
            .stats-grid {
                grid-template-columns: repeat(2, 1fr);
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">🚀</div>
            <h1>You're Invited!</h1>
            <p>Someone has invited you to view their Disko board</p>
        </div>
        
        <div class="content">
            <div class="board-info">
                <h2 class="board-name">{{.BoardName}}</h2>
                <p class="board-description">{{.BoardDescription}}</p>
                
                <div class="stats-grid">
                    <div class="stat-item">
                        <span class="stat-number">{{.IdeasCount}}</span>
                        <span class="stat-label">Ideas</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-number">{{.ReactionsCount}}</span>
                        <span class="stat-label">Reactions</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-number">{{.UpdatedAgo}}</span>
                        <span class="stat-label">Updated</span>
                    </div>
                </div>
                
                <div class="emoji-recaps">
                    <span class="recaps-label">Board Highlights:</span>
                    <span class="recaps-emojis">{{.EmojiRecaps}}</span>
                </div>
            </div>
            
            {{if .Message}}
            <div class="personal-message">
                <h3>💬 Personal Message</h3>
                <div class="message-content">
                    {{.Message}}
                </div>
            </div>
            {{end}}
            
            <div class="recent-ideas">
                <h3>💡 Recent Ideas</h3>
                {{.RecentIdeasHTML}}
            </div>
            
            <div class="cta-section">
                <h3 style="margin: 0 0 16px 0; color: #1e293b;">Ready to explore?</h3>
                <p style="margin: 0 0 24px 0; color: #64748b;">Click the button below to view the board and provide feedback on ideas.</p>
                <a href="{{.PublicURL}}" class="cta-button">View Board</a>
            </div>
        </div>
        
        <div class="footer">
            <div class="footer-logo">
                <img src="{{.AppURL}}/static/images/logo-sm.png" alt="Disko" width="120" height="30" style="border: 0; display: block;">
            </div>
            <p>This invitation was sent from <a href="{{.AppURL}}">Disko</a>, a Nomadis service.</p>
            <p>If you didn't expect this invitation, you can safely ignore this email.</p>
            <div class="footer-cta">
                <p>Want to start your own board? <a href="{{.AppURL}}">Sign up for Disko</a></p>
            </div>
            <div class="footer-links">
                <a href="{{.AboutURL}}">About Disko</a>
                <a href="{{.PrivacyURL}}">Privacy Policy</a>
                <a href="{{.TermsURL}}">Terms of Service</a>
					 <a href="{{.ContactURL}}">Contact Us</a>
            </div>
        </div>
    </div>
</body>
</html>`

	// Create template data
	templateData := struct {
		BoardName        string
		BoardDescription string
		IdeasCount       int
		ReactionsCount   int
		UpdatedAgo       string
		EmojiRecaps      string
		RecentIdeasHTML  string
		PublicURL        string
		AppURL           string
		AboutURL         string
		PrivacyURL       string
		TermsURL         string
		ContactURL       string
		Message          string // Added Message field
	}{
		BoardName:        board.Name,
		BoardDescription: board.Description,
		IdeasCount:       ideasCount,
		ReactionsCount:   reactionsCount,
		UpdatedAgo:       formatTimeAgo(board.UpdatedAt),
		EmojiRecaps:      generateEmojiRecaps(board),
		RecentIdeasHTML:  generateRecentIdeasHTML(recentIdeas),
		PublicURL:        publicURL,
		AppURL:           os.Getenv("APP_URL"),
		AboutURL:         fmt.Sprintf("%s/about", os.Getenv("APP_URL")),
		PrivacyURL:       fmt.Sprintf("%s/privacy", os.Getenv("APP_URL")),
		TermsURL:         fmt.Sprintf("%s/terms", os.Getenv("APP_URL")),
		ContactURL:       fmt.Sprintf("%s/contact", os.Getenv("APP_URL")),
		Message:          message, // Pass the message to the template
	}

	// Use Go's text/template to properly handle the template
	tmpl, err := template.New("email").Parse(htmlTemplate)
	if err != nil {
		log.Printf("[Email] Failed to parse email template: %v", err)
		return ""
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, templateData)
	if err != nil {
		log.Printf("[Email] Failed to execute email template: %v", err)
		return ""
	}

	html := buf.String()

	return html
}

// Helper functions for email generation
func getBoardIdeasCount(boardID string) int {
	// Query the database for actual ideas count
	ideasCollection := models.GetCollection(models.IdeasCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"board_id": boardID}
	count, err := ideasCollection.CountDocuments(ctx, filter)
	if err != nil {
		log.Printf("[Email] Failed to count ideas for board %s: %v", boardID, err)
		return 0
	}

	return int(count)
}

// getBoardReactionsCount gets the total reactions count for a board
func getBoardReactionsCount(boardID string) int {
	ideasCollection := models.GetCollection(models.IdeasCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{"$match": bson.M{"board_id": boardID}},
		{"$project": bson.M{
			"totalReactions": bson.M{
				"$add": []interface{}{
					"$thumbs_up",
					bson.M{"$reduce": bson.M{
						"input":        "$emoji_reactions",
						"initialValue": 0,
						"in":           bson.M{"$add": []string{"$$value", "$$this.count"}},
					}},
				},
			},
		}},
		{"$group": bson.M{
			"_id":            nil,
			"totalReactions": bson.M{"$sum": "$totalReactions"},
		}},
	}

	cursor, err := ideasCollection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf("[Email] Failed to get reactions count for board %s: %v", boardID, err)
		return 0
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err := cursor.All(ctx, &result); err != nil || len(result) == 0 {
		return 0
	}

	if total, ok := result[0]["totalReactions"].(int32); ok {
		return int(total)
	}

	return 0
}

// generateEmojiRecaps creates emoji recaps for the board
func generateEmojiRecaps(board models.Board) string {
	// Query the database for real board statistics
	ideasCollection := models.GetCollection(models.IdeasCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	recaps := []string{}

	// Get all emoji reactions aggregated across the board
	emojiPipeline := []bson.M{
		{"$match": bson.M{"board_id": board.ID}},
		{"$unwind": "$emoji_reactions"},
		{"$group": bson.M{
			"_id":   "$emoji_reactions.emoji",
			"count": bson.M{"$sum": "$emoji_reactions.count"},
		}},
		{"$sort": bson.M{"count": -1}},
		{"$limit": 5}, // Show top 5 most popular emojis
	}

	emojiCursor, err := ideasCollection.Aggregate(ctx, emojiPipeline)
	if err == nil {
		defer emojiCursor.Close(ctx)
		var emojiResults []bson.M
		if err := emojiCursor.All(ctx, &emojiResults); err == nil {
			for _, emoji := range emojiResults {
				if emojiStr, ok := emoji["_id"].(string); ok {
					if count, ok := emoji["count"].(int32); ok && count > 0 {
						// Add emoji with count if it has reactions
						recaps = append(recaps, emojiStr)
					}
				}
			}
		}
	}

	// Get total thumbs up count
	thumbsUpPipeline := []bson.M{
		{"$match": bson.M{"board_id": board.ID}},
		{"$group": bson.M{
			"_id":         nil,
			"totalThumbs": bson.M{"$sum": "$thumbs_up"},
		}},
	}

	thumbsCursor, err := ideasCollection.Aggregate(ctx, thumbsUpPipeline)
	if err == nil {
		defer thumbsCursor.Close(ctx)
		var thumbsResults []bson.M
		if err := thumbsCursor.All(ctx, &thumbsResults); err == nil && len(thumbsResults) > 0 {
			if totalThumbs, ok := thumbsResults[0]["totalThumbs"].(int32); ok && totalThumbs > 0 {
				recaps = append(recaps, "👍") // Add thumbs up emoji if there are any
			}
		}
	}

	// Add contextual emojis based on board activity
	if len(board.VisibleColumns) > 0 {
		recaps = append(recaps, "📊") // Board structure
	}

	// Add emoji based on recent activity
	if time.Since(board.UpdatedAt) < 24*time.Hour {
		recaps = append(recaps, "🔥") // Recently updated
	}

	// Add emoji based on board type/description
	if board.Description != "" {
		recaps = append(recaps, "💡") // Has description
	}

	// Add default emoji if no specific ones
	if len(recaps) == 0 {
		recaps = append(recaps, "🚀") // Default Disko emoji
	}

	// Join emojis with spaces
	return strings.Join(recaps, " ")
}

func getRecentIdeas(boardID string, limit int) []models.Idea {
	// Query the database for actual recent ideas
	ideasCollection := models.GetCollection(models.IdeasCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"board_id": boardID}
	opts := options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(int64(limit))

	cursor, err := ideasCollection.Find(ctx, filter, opts)
	if err != nil {
		log.Printf("[Email] Failed to get recent ideas for board %s: %v", boardID, err)
		return []models.Idea{}
	}
	defer cursor.Close(ctx)

	var ideas []models.Idea
	if err := cursor.All(ctx, &ideas); err != nil {
		log.Printf("[Email] Failed to decode recent ideas for board %s: %v", boardID, err)
		return []models.Idea{}
	}

	return ideas
}

func generateRecentIdeasHTML(ideas []models.Idea) string {
	if len(ideas) == 0 {
		return `<p style="color: #64748b; font-style: italic;">No recent ideas to display</p>`
	}

	html := ""
	for _, idea := range ideas {
		// Generate feedback summary
		feedbackSummary := generateFeedbackSummary(idea)

		html += fmt.Sprintf(`
            <div class="idea-item">
                <div class="idea-title">%s</div>
                <div class="idea-meta">%s • %s</div>
                %s
            </div>
        `,
			idea.OneLiner,
			formatColumn(idea.Column),
			formatTimeAgo(idea.CreatedAt),
			feedbackSummary,
		)
	}

	return html
}

func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

func generateFeedbackSummary(idea models.Idea) string {
	var feedbackParts []string

	// Add thumbs up count if greater than 0
	if idea.ThumbsUp > 0 {
		feedbackParts = append(feedbackParts, fmt.Sprintf("👍 %d", idea.ThumbsUp))
	}

	// Add emoji reactions if any exist
	if len(idea.EmojiReactions) > 0 {
		for _, reaction := range idea.EmojiReactions {
			if reaction.Count > 0 {
				feedbackParts = append(feedbackParts, fmt.Sprintf("%s %d", reaction.Emoji, reaction.Count))
			}
		}
	}

	// If no feedback, return empty string
	if len(feedbackParts) == 0 {
		return ""
	}

	// Return feedback summary with styling
	return fmt.Sprintf(`
		<div class="idea-feedback-summary">
			<span class="feedback-label">Feedback:</span>
			<span class="feedback-items">%s</span>
		</div>
	`, strings.Join(feedbackParts, " "))
}

func formatColumn(column string) string {
	switch column {
	case "now":
		return "Now"
	case "next":
		return "Next"
	case "later":
		return "Later"
	case "wont-do":
		return "Won't Do"
	default:
		return column
	}
}
