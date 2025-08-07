package services

import (
	"encoding/json"
	"telegram-store-hub/internal/models"
	"time"

	"gorm.io/gorm"
)

type SessionService struct {
	db *gorm.DB
}

func NewSessionService(db *gorm.DB) *SessionService {
	return &SessionService{db: db}
}

// SetUserState sets user conversation state
func (s *SessionService) SetUserState(telegramID int64, state string, data interface{}) error {
	dataJSON, _ := json.Marshal(data)
	
	session := models.UserSession{
		TelegramID: telegramID,
		State:      state,
		Data:       string(dataJSON),
		UpdatedAt:  time.Now(),
	}
	
	return s.db.Save(&session).Error
}

// GetUserState gets user conversation state
func (s *SessionService) GetUserState(telegramID int64) (*models.UserSession, error) {
	var session models.UserSession
	err := s.db.Where("telegram_id = ?", telegramID).First(&session).Error
	if err != nil {
		return nil, err
	}
	
	return &session, nil
}

// ClearUserState clears user conversation state
func (s *SessionService) ClearUserState(telegramID int64) error {
	return s.db.Where("telegram_id = ?", telegramID).Delete(&models.UserSession{}).Error
}

// GetSessionData gets and unmarshals session data
func (s *SessionService) GetSessionData(telegramID int64, data interface{}) error {
	session, err := s.GetUserState(telegramID)
	if err != nil {
		return err
	}
	
	return json.Unmarshal([]byte(session.Data), data)
}