package services

import (
	"fmt"
	"log"
	"telegram-store-hub/internal/messages"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ChannelVerificationService handles channel membership verification
type ChannelVerificationService struct {
	bot                *tgbotapi.BotAPI
	requiredChannelID  string
	channelUsername    string
	isRequired         bool
}

// NewChannelVerificationService creates a new channel verification service
func NewChannelVerificationService(bot *tgbotapi.BotAPI, channelID string) *ChannelVerificationService {
	return &ChannelVerificationService{
		bot:               bot,
		requiredChannelID: channelID,
		isRequired:        channelID != "",
	}
}

// SetChannelUsername sets the channel username for display purposes
func (c *ChannelVerificationService) SetChannelUsername(username string) {
	c.channelUsername = username
}

// CheckAndHandleMembership checks if user is member of required channel
// Returns true if user is member or no channel verification required
func (c *ChannelVerificationService) CheckAndHandleMembership(chatID int64) (bool, error) {
	// If no channel verification required, allow access
	if !c.isRequired {
		return true, nil
	}

	// Check membership
	isMember, err := c.IsUserMember(chatID)
	if err != nil {
		log.Printf("Error checking channel membership for user %d: %v", chatID, err)
		return false, err
	}

	if !isMember {
		// Send join channel message
		c.SendJoinChannelMessage(chatID)
		return false, nil
	}

	return true, nil
}

// IsUserMember checks if user is member of the required channel
func (c *ChannelVerificationService) IsUserMember(userID int64) (bool, error) {
	if !c.isRequired {
		return true, nil
	}

	// Get chat member info
	memberConfig := tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: c.requiredChannelID,
			UserID: userID,
		},
	}

	chatMember, err := c.bot.GetChatMember(memberConfig)
	if err != nil {
		// If user not found in channel, they're not a member
		if err.Error() == "Bad Request: user not found" {
			return false, nil
		}
		return false, fmt.Errorf("failed to get chat member: %w", err)
	}

	// Check membership status
	switch chatMember.Status {
	case "creator", "administrator", "member":
		return true, nil
	case "left", "kicked":
		return false, nil
	default:
		return false, nil
	}
}

// SendJoinChannelMessage sends a message asking user to join the channel
func (c *ChannelVerificationService) SendJoinChannelMessage(chatID int64) {
	var keyboard tgbotapi.InlineKeyboardMarkup

	// Create keyboard with channel link and verification button
	if c.channelUsername != "" {
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL(
					messages.ButtonJoinChannel,
					fmt.Sprintf("https://t.me/%s", c.channelUsername),
				),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					messages.ButtonCheckMembership,
					"check_membership",
				),
			),
		)
	} else {
		// If no username provided, just show verification button
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					messages.ButtonCheckMembership,
					"check_membership",
				),
			),
		)
	}

	// Prepare message text
	messageText := messages.ChannelJoinRequired
	if c.channelUsername != "" {
		messageText = fmt.Sprintf("%s\n\n🔗 کانال: @%s", messageText, c.channelUsername)
	}

	msg := tgbotapi.NewMessage(chatID, messageText)
	msg.ReplyMarkup = keyboard

	_, err := c.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending join channel message: %v", err)
	}
}

// SendMembershipConfirmation sends confirmation message when user joins
func (c *ChannelVerificationService) SendMembershipConfirmation(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, messages.ChannelJoinSuccess)
	
	// Add main menu keyboard
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.ButtonRegisterStore, "register_store"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.ButtonManageStore, "manage_store"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(messages.ButtonViewPlans, "view_plans"),
			tgbotapi.NewInlineKeyboardButtonData(messages.ButtonSupport, "support"),
		),
	)
	
	msg.ReplyMarkup = keyboard

	_, err := c.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending membership confirmation: %v", err)
	}
}

// GetChannelInfo returns channel information
func (c *ChannelVerificationService) GetChannelInfo() (string, string, bool) {
	return c.requiredChannelID, c.channelUsername, c.isRequired
}

// GetChannelMemberCount returns the number of members in the channel
func (c *ChannelVerificationService) GetChannelMemberCount() (int, error) {
	if !c.isRequired {
		return 0, nil
	}

	chatConfig := tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: c.requiredChannelID,
		},
	}

	chat, err := c.bot.GetChat(chatConfig)
	if err != nil {
		return 0, fmt.Errorf("failed to get chat info: %w", err)
	}

	return chat.MembersCount, nil
}

// ValidateChannelAccess verifies that the bot has access to the channel
func (c *ChannelVerificationService) ValidateChannelAccess() error {
	if !c.isRequired {
		return nil
	}

	// Try to get bot's own membership in the channel
	botUser := c.bot.Self
	memberConfig := tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: c.requiredChannelID,
			UserID: int64(botUser.ID),
		},
	}

	chatMember, err := c.bot.GetChatMember(memberConfig)
	if err != nil {
		return fmt.Errorf("bot cannot access channel %s: %w", c.requiredChannelID, err)
	}

	// Bot should be admin or member to check other users
	switch chatMember.Status {
	case "creator", "administrator":
		log.Printf("Bot has admin access to channel %s", c.requiredChannelID)
		return nil
	case "member":
		log.Printf("Bot has member access to channel %s", c.requiredChannelID)
		return nil
	default:
		return fmt.Errorf("bot has insufficient permissions in channel %s (status: %s)", c.requiredChannelID, chatMember.Status)
	}
}