package services

import (
	"fmt"
	"time"
	"telegram-store-hub/internal/models"
	"gorm.io/gorm"
)

type StoreService struct {
	db *gorm.DB
}

func NewStoreService(db *gorm.DB) *StoreService {
	return &StoreService{db: db}
}

func (s *StoreService) CreateStore(ownerID uint, name, description, planType string) (*models.Store, error) {
	// Calculate plan details
	productLimit, commissionRate := s.getPlanDetails(planType)
	expiresAt := time.Now().AddDate(0, 1, 0) // 1 month from now

	store := models.Store{
		OwnerID:        ownerID,
		Name:           name,
		Description:    description,
		PlanType:       planType,
		ExpiresAt:      expiresAt,
		IsActive:       false, // Will be activated after payment confirmation
		ProductLimit:   productLimit,
		CommissionRate: commissionRate,
		BotToken:       "", // Will be set when sub-bot is created
		BotUsername:    "", // Will be set when sub-bot is created
	}

	err := s.db.Create(&store).Error
	return &store, err
}

func (s *StoreService) GetUserStores(ownerID uint) ([]models.Store, error) {
	var stores []models.Store
	err := s.db.Where("owner_id = ?", ownerID).Find(&stores).Error
	return stores, err
}

func (s *StoreService) GetStoreByID(id uint) (*models.Store, error) {
	var store models.Store
	err := s.db.Preload("Owner").First(&store, id).Error
	return &store, err
}

func (s *StoreService) ActivateStore(storeID uint, botToken, botUsername string) error {
	return s.db.Model(&models.Store{}).Where("id = ?", storeID).Updates(map[string]interface{}{
		"is_active":    true,
		"bot_token":    botToken,
		"bot_username": botUsername,
	}).Error
}

func (s *StoreService) DeactivateStore(storeID uint) error {
	return s.db.Model(&models.Store{}).Where("id = ?", storeID).Update("is_active", false).Error
}

func (s *StoreService) UpdateStore(storeID uint, updates map[string]interface{}) error {
	return s.db.Model(&models.Store{}).Where("id = ?", storeID).Updates(updates).Error
}

func (s *StoreService) DeleteStore(storeID uint) error {
	return s.db.Delete(&models.Store{}, storeID).Error
}

func (s *StoreService) GetActiveStores() ([]models.Store, error) {
	var stores []models.Store
	err := s.db.Where("is_active = ?", true).Preload("Owner").Find(&stores).Error
	return stores, err
}

func (s *StoreService) GetPendingStores() ([]models.Store, error) {
	var stores []models.Store
	err := s.db.Where("is_active = ?", false).Preload("Owner").Find(&stores).Error
	return stores, err
}

func (s *StoreService) RenewStore(storeID uint, months int) error {
	var store models.Store
	if err := s.db.First(&store, storeID).Error; err != nil {
		return err
	}

	newExpiryDate := store.ExpiresAt.AddDate(0, months, 0)
	if newExpiryDate.Before(time.Now()) {
		newExpiryDate = time.Now().AddDate(0, months, 0)
	}

	return s.db.Model(&store).Update("expires_at", newExpiryDate).Error
}

func (s *StoreService) GetExpiringStores(days int) ([]models.Store, error) {
	var stores []models.Store
	expiryDate := time.Now().AddDate(0, 0, days)
	
	err := s.db.Where("expires_at <= ? AND is_active = ?", expiryDate, true).
		Preload("Owner").Find(&stores).Error
	return stores, err
}

func (s *StoreService) GetStoresCount() (int64, error) {
	var count int64
	err := s.db.Model(&models.Store{}).Count(&count).Error
	return count, err
}

func (s *StoreService) GetActiveStoresCount() (int64, error) {
	var count int64
	err := s.db.Model(&models.Store{}).Where("is_active = ?", true).Count(&count).Error
	return count, err
}

func (s *StoreService) getPlanDetails(planType string) (productLimit, commissionRate int) {
	switch planType {
	case "free":
		return 10, 5
	case "pro":
		return 200, 5
	case "vip":
		return -1, 0 // -1 means unlimited
	default:
		return 10, 5
	}
}

func (s *StoreService) GenerateBotUsername(storeName string, storeID uint) string {
	// Clean store name and create bot username
	cleanName := s.cleanStringForUsername(storeName)
	return fmt.Sprintf("%s_%d_bot", cleanName, storeID)
}

func (s *StoreService) cleanStringForUsername(str string) string {
	// Convert Persian/Arabic characters and remove special chars
	// This is a simple implementation - you might want to use a more sophisticated approach
	result := ""
	for _, char := range str {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			result += string(char)
		}
	}
	if len(result) == 0 {
		result = "store"
	}
	if len(result) > 10 {
		result = result[:10]
	}
	return result
}