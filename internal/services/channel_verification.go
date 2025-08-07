package services

import (
        "fmt"
        "telegram-store-hub/internal/messages"

        tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ChannelVerificationService struct {
        bot       *tgbotapi.BotAPI
        channelID string
}

func NewChannelVerificationService(bot *tgbotapi.BotAPI, channelID string) *ChannelVerificationService {
        return &ChannelVerificationService{
                bot:       bot,
                channelID: channelID,
        }
}

// CheckAndHandleMembership checks if user is member and handles join flow
func (s *ChannelVerificationService) CheckAndHandleMembership(chatID int64) (bool, error) {
        // If no channel is configured, allow all users
        if s.channelID == "" {
                return true, nil
        }
        
        // Check membership
        isMember, err := s.IsUserMember(chatID)
        if err != nil {
                return false, err
        }
        
        if !isMember {
                // Send join message
                s.sendJoinMessage(chatID)
                return false, nil
        }
        
        return true, nil
}

// IsUserMember checks if user is member of the channel
func (s *ChannelVerificationService) IsUserMember(chatID int64) (bool, error) {
        if s.channelID == "" {
                return true, nil
        }
        
        member, err := s.bot.GetChatMember(tgbotapi.GetChatMemberConfig{
                ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
                        ChatID: s.channelID,
                        UserID: chatID,
                },
        })
        
        if err != nil {
                return false, err
        }
        
        // Check if user is member, administrator, or creator
        status := member.Status
        return status == "member" || status == "administrator" || status == "creator", nil
}

// sendJoinMessage sends force join message
func (s *ChannelVerificationService) sendJoinMessage(chatID int64) {
        text := fmt.Sprintf(messages.ForceJoinMessage, s.channelID)
        
        keyboard := tgbotapi.NewInlineKeyboardMarkup(
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonURL("📢 عضویت در کانال", fmt.Sprintf("https://t.me/%s", s.channelID)),
                ),
                tgbotapi.NewInlineKeyboardRow(
                        tgbotapi.NewInlineKeyboardButtonData(messages.CheckMembershipBtn, "check_membership"),
                ),
        )
        
        msg := tgbotapi.NewMessage(chatID, text)
        msg.ReplyMarkup = keyboard
        msg.ParseMode = "HTML"
        
        s.bot.Send(msg)
}

// UpdateChannelID updates the channel ID
func (s *ChannelVerificationService) UpdateChannelID(channelID string) {
        s.channelID = channelID
}