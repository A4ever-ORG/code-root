package services

import (
	"telegram-store-hub/internal/models"

	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// GetOrCreateUser gets existing user or creates a new one
func (s *UserService) GetOrCreateUser(telegramID int64, username, firstName, lastName string) (*models.User, error) {
	var user models.User
	
	err := s.db.Where("telegram_id = ?", telegramID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new user
			user = models.User{
				TelegramID: telegramID,
				Username:   username,
				FirstName:  firstName,
				LastName:   lastName,
				IsAdmin:    false,
			}
			
			if err := s.db.Create(&user).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		// Update user info if changed
		updated := false
		if user.Username != username {
			user.Username = username
			updated = true
		}
		if user.FirstName != firstName {
			user.FirstName = firstName
			updated = true
		}
		if user.LastName != lastName {
			user.LastName = lastName
			updated = true
		}
		
		if updated {
			s.db.Save(&user)
		}
	}
	
	return &user, nil
}

// IsAdmin checks if user is admin
func (s *UserService) IsAdmin(telegramID int64) bool {
	var user models.User
	err := s.db.Where("telegram_id = ? AND is_admin = ?", telegramID, true).First(&user).Error
	return err == nil
}

// GetUserStores gets all stores owned by user
func (s *UserService) GetUserStores(telegramID int64) ([]models.Store, error) {
	var user models.User
	err := s.db.Preload("Stores").Where("telegram_id = ?", telegramID).First(&user).Error
	if err != nil {
		return nil, err
	}
	
	return user.Stores, nil
}