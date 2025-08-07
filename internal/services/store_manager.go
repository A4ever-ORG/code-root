package services

import (
	"telegram-store-hub/internal/models"
	"time"

	"gorm.io/gorm"
)

type StoreManagerService struct {
	db *gorm.DB
}

func NewStoreManagerService(db *gorm.DB) *StoreManagerService {
	return &StoreManagerService{db: db}
}

// CreateStore creates a new store
func (s *StoreManagerService) CreateStore(ownerID uint, name, description string, planType models.PlanType) (*models.Store, error) {
	// Set plan limits
	var productLimit, commissionRate int
	var expiresAt time.Time
	
	switch planType {
	case models.PlanFree:
		productLimit = 10
		commissionRate = 5
		expiresAt = time.Now().AddDate(1, 0, 0) // 1 year
	case models.PlanPro:
		productLimit = 200
		commissionRate = 5
		expiresAt = time.Now().AddDate(0, 1, 0) // 1 month
	case models.PlanVIP:
		productLimit = -1 // unlimited
		commissionRate = 0
		expiresAt = time.Now().AddDate(0, 1, 0) // 1 month
	}
	
	store := models.Store{
		OwnerID:        ownerID,
		Name:           name,
		Description:    description,
		PlanType:       planType,
		ExpiresAt:      expiresAt,
		IsActive:       true,
		ProductLimit:   productLimit,
		CommissionRate: commissionRate,
	}
	
	if err := s.db.Create(&store).Error; err != nil {
		return nil, err
	}
	
	return &store, nil
}

// GetStoreByID gets store by ID
func (s *StoreManagerService) GetStoreByID(storeID uint) (*models.Store, error) {
	var store models.Store
	err := s.db.Preload("Owner").Preload("Products").Preload("Orders").First(&store, storeID).Error
	return &store, err
}

// GetUserStores gets all stores for a user
func (s *StoreManagerService) GetUserStores(ownerID uint) ([]models.Store, error) {
	var stores []models.Store
	err := s.db.Where("owner_id = ?", ownerID).Preload("Products").Find(&stores).Error
	return stores, err
}

// UpdateStore updates store information
func (s *StoreManagerService) UpdateStore(store *models.Store) error {
	return s.db.Save(store).Error
}

// DeactivateStore deactivates a store
func (s *StoreManagerService) DeactivateStore(storeID uint) error {
	return s.db.Model(&models.Store{}).Where("id = ?", storeID).Update("is_active", false).Error
}

// ActivateStore activates a store
func (s *StoreManagerService) ActivateStore(storeID uint) error {
	return s.db.Model(&models.Store{}).Where("id = ?", storeID).Update("is_active", true).Error
}

// ExtendStorePlan extends store plan expiration
func (s *StoreManagerService) ExtendStorePlan(storeID uint, months int) error {
	var store models.Store
	if err := s.db.First(&store, storeID).Error; err != nil {
		return err
	}
	
	// Extend from current expiry or now (whichever is later)
	baseTime := store.ExpiresAt
	if time.Now().After(store.ExpiresAt) {
		baseTime = time.Now()
	}
	
	store.ExpiresAt = baseTime.AddDate(0, months, 0)
	store.IsActive = true
	
	return s.db.Save(&store).Error
}

// GetExpiredStores gets stores that have expired
func (s *StoreManagerService) GetExpiredStores() ([]models.Store, error) {
	var stores []models.Store
	err := s.db.Where("expires_at < ? AND is_active = ?", time.Now(), true).Find(&stores).Error
	return stores, err
}

// GetExpiringSoonStores gets stores expiring within days
func (s *StoreManagerService) GetExpiringSoonStores(days int) ([]models.Store, error) {
	var stores []models.Store
	futureDate := time.Now().AddDate(0, 0, days)
	err := s.db.Where("expires_at BETWEEN ? AND ? AND is_active = ?", time.Now(), futureDate, true).
		Preload("Owner").Find(&stores).Error
	return stores, err
}

// CheckStoreProductLimit checks if store can add more products
func (s *StoreManagerService) CheckStoreProductLimit(storeID uint) (bool, error) {
	var store models.Store
	if err := s.db.Preload("Products").First(&store, storeID).Error; err != nil {
		return false, err
	}
	
	if store.ProductLimit == -1 { // unlimited
		return true, nil
	}
	
	return len(store.Products) < store.ProductLimit, nil
}

// GetAllStores gets all stores for admin
func (s *StoreManagerService) GetAllStores(limit, offset int) ([]models.Store, error) {
	var stores []models.Store
	err := s.db.Preload("Owner").Limit(limit).Offset(offset).Find(&stores).Error
	return stores, err
}

// GetStoreStats gets store statistics
func (s *StoreManagerService) GetStoreStats(storeID uint) (map[string]interface{}, error) {
	var store models.Store
	if err := s.db.Preload("Products").Preload("Orders").First(&store, storeID).Error; err != nil {
		return nil, err
	}
	
	stats := map[string]interface{}{
		"total_products":   len(store.Products),
		"active_products":  0,
		"total_orders":     len(store.Orders),
		"pending_orders":   0,
		"completed_orders": 0,
		"total_revenue":    int64(0),
	}
	
	// Calculate active products
	for _, product := range store.Products {
		if product.IsAvailable {
			stats["active_products"] = stats["active_products"].(int) + 1
		}
	}
	
	// Calculate order stats
	for _, order := range store.Orders {
		switch order.Status {
		case "pending":
			stats["pending_orders"] = stats["pending_orders"].(int) + 1
		case "completed":
			stats["completed_orders"] = stats["completed_orders"].(int) + 1
			stats["total_revenue"] = stats["total_revenue"].(int64) + order.TotalAmount
		}
	}
	
	return stats, nil
}