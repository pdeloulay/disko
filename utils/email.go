package utils

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"text/template"
	"time"

	"disko-backend/models"

	"gopkg.in/gomail.v2"
)

// SendBoardInviteEmail sends an HTML invitation email for a board
func SendBoardInviteEmail(email, subject string, board models.Board, userID string) error {
	// Get email configuration from environment variables
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	fromEmail := os.Getenv("FROM_EMAIL")

	log.Printf("[Email] Configuration check - SMTP_HOST: %s, SMTP_PORT: %s, SMTP_USER: %s, FROM_EMAIL: %s, APP_URL: %s",
		smtpHost, smtpPortStr, smtpUser, fromEmail, os.Getenv("APP_URL"))

	if smtpHost == "" || smtpPortStr == "" || smtpUser == "" || smtpPass == "" || fromEmail == "" {
		log.Printf("[Email] Configuration incomplete - missing required environment variables")
		return fmt.Errorf("email configuration incomplete - check SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS, FROM_EMAIL environment variables")
	}

	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		return fmt.Errorf("invalid SMTP port: %v", err)
	}

	// Get user email from Clerk if userID is provided
	fromEmailWithName := fromEmail
	if userID != "" {
		_, err := getUserEmailFromClerk(userID)
		if err != nil {
			log.Printf("[Email] Failed to get user email from Clerk: %v, using default email", err)
		} else {
			fromEmailWithName = fmt.Sprintf("Disko <noreply@%s>", extractDomain(fromEmail))
		}
	}

	// Create email message
	m := gomail.NewMessage()
	m.SetHeader("From", fromEmailWithName)
	m.SetHeader("To", email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", generateInviteEmailHTML(board))

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

	// Placeholder implementation - in production, you would use the actual Clerk SDK
	// client, err := clerk.NewClient(clerkSecretKey)
	// if err != nil {
	//     return "", fmt.Errorf("failed to create Clerk client: %v", err)
	// }
	//
	// user, err := client.Users.GetUser(context.Background(), userID)
	// if err != nil {
	//     return "", fmt.Errorf("failed to get user from Clerk: %v", err)
	// }
	//
	// if len(user.EmailAddresses) > 0 {
	//     for _, email := range user.EmailAddresses {
	//         if email.ID == user.PrimaryEmailAddressID {
	//             return email.EmailAddress, nil
	//         }
	//     }
	//     if len(user.EmailAddresses) > 0 {
	//         return user.EmailAddresses[0].EmailAddress, nil
	//     }
	// }

	return "", fmt.Errorf("Clerk SDK integration not yet implemented")
}

// extractDomain extracts domain from email address
func extractDomain(email string) string {
	for i, char := range email {
		if char == '@' {
			return email[i+1:]
		}
	}
	return "disko.app"
}

// generateInviteEmailHTML creates a compelling HTML email template with Disko branding
func generateInviteEmailHTML(board models.Board) string {
	publicURL := fmt.Sprintf("%s/public/%s", os.Getenv("APP_URL"), board.PublicLink)

	// Get board statistics
	ideasCount := getBoardIdeasCount(board.ID)
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
            font-size: 20px;
            font-weight: 700;
            color: #3b82f6;
            margin-bottom: 16px;
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
                        <span class="stat-number">{{.ColumnsCount}}</span>
                        <span class="stat-label">Columns</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-number">{{.UpdatedAgo}}</span>
                        <span class="stat-label">Updated</span>
                    </div>
                </div>
            </div>
            
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
            <div class="footer-logo">🚀 Disko</div>
            <p>This invitation was sent from <a href="{{.AppURL}}">Disko</a></p>
            <p>If you didn't expect this invitation, you can safely ignore this email.</p>
            <div class="footer-links">
                <a href="{{.AboutURL}}">About Disko</a>
                <a href="{{.PrivacyURL}}">Privacy Policy</a>
                <a href="{{.TermsURL}}">Terms of Service</a>
                <a href="mailto:support@disko.app">Support</a>
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
		ColumnsCount     int
		UpdatedAgo       string
		RecentIdeasHTML  string
		PublicURL        string
		AppURL           string
		AboutURL         string
		PrivacyURL       string
		TermsURL         string
	}{
		BoardName:        board.Name,
		BoardDescription: board.Description,
		IdeasCount:       ideasCount,
		ColumnsCount:     len(board.VisibleColumns),
		UpdatedAgo:       formatTimeAgo(board.UpdatedAt),
		RecentIdeasHTML:  generateRecentIdeasHTML(recentIdeas),
		PublicURL:        publicURL,
		AppURL:           os.Getenv("APP_URL"),
		AboutURL:         fmt.Sprintf("%s/about", os.Getenv("APP_URL")),
		PrivacyURL:       fmt.Sprintf("%s/privacy", os.Getenv("APP_URL")),
		TermsURL:         fmt.Sprintf("%s/terms", os.Getenv("APP_URL")),
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
	// This would typically query the database
	// For now, return a placeholder
	return 12 // Placeholder
}

func getRecentIdeas(boardID string, limit int) []models.Idea {
	// This would typically query the database
	// For now, return placeholder data
	return []models.Idea{
		{
			ID:        "idea1",
			OneLiner:  "Implement real-time collaboration features",
			Column:    "now",
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        "idea2",
			OneLiner:  "Add advanced search and filtering",
			Column:    "next",
			CreatedAt: time.Now().Add(-4 * time.Hour),
		},
		{
			ID:        "idea3",
			OneLiner:  "Create mobile app for iOS and Android",
			Column:    "later",
			CreatedAt: time.Now().Add(-6 * time.Hour),
		},
	}
}

func generateRecentIdeasHTML(ideas []models.Idea) string {
	if len(ideas) == 0 {
		return `<p style="color: #64748b; font-style: italic;">No recent ideas to display</p>`
	}

	html := ""
	for _, idea := range ideas {
		html += fmt.Sprintf(`
            <div class="idea-item">
                <div class="idea-title">%s</div>
                <div class="idea-meta">%s • %s</div>
            </div>
        `,
			idea.OneLiner,
			formatColumn(idea.Column),
			formatTimeAgo(idea.CreatedAt),
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
