package services

import (
	"telegram-store-hub/internal/models"
	"time"

	"gorm.io/gorm"
)

type PaymentService struct {
	db *gorm.DB
}

func NewPaymentService(db *gorm.DB) *PaymentService {
	return &PaymentService{db: db}
}

// CreatePayment creates a new payment record
func (s *PaymentService) CreatePayment(storeID uint, amount int64, paymentType string, proofImageURL, notes string) (*models.Payment, error) {
	payment := models.Payment{
		StoreID:       storeID,
		Amount:        amount,
		PaymentType:   paymentType,
		Status:        "pending",
		ProofImageURL: proofImageURL,
		Notes:         notes,
	}
	
	if err := s.db.Create(&payment).Error; err != nil {
		return nil, err
	}
	
	return &payment, nil
}

// GetPaymentByID gets payment by ID
func (s *PaymentService) GetPaymentByID(paymentID uint) (*models.Payment, error) {
	var payment models.Payment
	err := s.db.Preload("Store").Preload("Store.Owner").First(&payment, paymentID).Error
	return &payment, err
}

// GetPendingPayments gets all pending payments
func (s *PaymentService) GetPendingPayments(limit, offset int) ([]models.Payment, error) {
	var payments []models.Payment
	err := s.db.Where("status = ?", "pending").
		Preload("Store").
		Preload("Store.Owner").
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&payments).Error
	return payments, err
}

// GetStorePayments gets all payments for a store
func (s *PaymentService) GetStorePayments(storeID uint) ([]models.Payment, error) {
	var payments []models.Payment
	err := s.db.Where("store_id = ?", storeID).
		Order("created_at DESC").
		Find(&payments).Error
	return payments, err
}

// ApprovePayment approves a payment
func (s *PaymentService) ApprovePayment(paymentID uint, adminUserID uint) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":      "confirmed",
		"verified_by": adminUserID,
		"verified_at": &now,
	}
	
	return s.db.Model(&models.Payment{}).Where("id = ?", paymentID).Updates(updates).Error
}

// RejectPayment rejects a payment
func (s *PaymentService) RejectPayment(paymentID uint, adminUserID uint, reason string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":      "failed",
		"verified_by": adminUserID,
		"verified_at": &now,
		"notes":       reason,
	}
	
	return s.db.Model(&models.Payment{}).Where("id = ?", paymentID).Updates(updates).Error
}

// GetPaymentStats gets payment statistics
func (s *PaymentService) GetPaymentStats(days int) (map[string]interface{}, error) {
	startDate := time.Now().AddDate(0, 0, -days)
	
	var stats struct {
		TotalPayments     int64 `json:"total_payments"`
		ConfirmedPayments int64 `json:"confirmed_payments"`
		PendingPayments   int64 `json:"pending_payments"`
		FailedPayments    int64 `json:"failed_payments"`
		TotalAmount       int64 `json:"total_amount"`
		ConfirmedAmount   int64 `json:"confirmed_amount"`
	}
	
	// Get payment counts
	s.db.Model(&models.Payment{}).Where("created_at >= ?", startDate).Count(&stats.TotalPayments)
	s.db.Model(&models.Payment{}).Where("status = 'confirmed' AND created_at >= ?", startDate).Count(&stats.ConfirmedPayments)
	s.db.Model(&models.Payment{}).Where("status = 'pending' AND created_at >= ?", startDate).Count(&stats.PendingPayments)
	s.db.Model(&models.Payment{}).Where("status = 'failed' AND created_at >= ?", startDate).Count(&stats.FailedPayments)
	
	// Get amounts
	s.db.Table("payments").
		Where("created_at >= ?", startDate).
		Select("COALESCE(SUM(amount), 0)").
		Row().Scan(&stats.TotalAmount)
	
	s.db.Table("payments").
		Where("status = 'confirmed' AND created_at >= ?", startDate).
		Select("COALESCE(SUM(amount), 0)").
		Row().Scan(&stats.ConfirmedAmount)
	
	return map[string]interface{}{
		"total_payments":     stats.TotalPayments,
		"confirmed_payments": stats.ConfirmedPayments,
		"pending_payments":   stats.PendingPayments,
		"failed_payments":    stats.FailedPayments,
		"total_amount":       stats.TotalAmount,
		"confirmed_amount":   stats.ConfirmedAmount,
	}, nil
}

// GetMonthlyRevenue gets monthly revenue data
func (s *PaymentService) GetMonthlyRevenue(months int) ([]map[string]interface{}, error) {
	startDate := time.Now().AddDate(0, -months, 0)
	
	var results []struct {
		Month       string `json:"month"`
		TotalAmount int64  `json:"total_amount"`
		PaymentCount int64  `json:"payment_count"`
	}
	
	err := s.db.Table("payments").
		Select("TO_CHAR(created_at, 'YYYY-MM') as month, COALESCE(SUM(amount), 0) as total_amount, COUNT(*) as payment_count").
		Where("status = 'confirmed' AND created_at >= ?", startDate).
		Group("TO_CHAR(created_at, 'YYYY-MM')").
		Order("month ASC").
		Scan(&results).Error
	
	if err != nil {
		return nil, err
	}
	
	// Convert to map slice
	data := make([]map[string]interface{}, len(results))
	for i, result := range results {
		data[i] = map[string]interface{}{
			"month":         result.Month,
			"total_amount":  result.TotalAmount,
			"payment_count": result.PaymentCount,
		}
	}
	
	return data, nil
}

// CreateSubscriptionPayment creates a subscription payment
func (s *PaymentService) CreateSubscriptionPayment(storeID uint, planType models.PlanType, proofImageURL string) (*models.Payment, error) {
	var amount int64
	
	switch planType {
	case models.PlanFree:
		amount = 0
	case models.PlanPro:
		amount = 50000
	case models.PlanVIP:
		amount = 150000
	}
	
	return s.CreatePayment(storeID, amount, "subscription", proofImageURL, string(planType)+" plan subscription")
}

// CreateRenewalPayment creates a plan renewal payment
func (s *PaymentService) CreateRenewalPayment(storeID uint, months int, proofImageURL string) (*models.Payment, error) {
	var store models.Store
	if err := s.db.First(&store, storeID).Error; err != nil {
		return nil, err
	}
	
	var monthlyAmount int64
	switch store.PlanType {
	case models.PlanPro:
		monthlyAmount = 50000
	case models.PlanVIP:
		monthlyAmount = 150000
	default:
		monthlyAmount = 0
	}
	
	totalAmount := monthlyAmount * int64(months)
	notes := "Plan renewal for " + string(store.PlanType) + " plan, " + string(rune(months)) + " months"
	
	return s.CreatePayment(storeID, totalAmount, "renewal", proofImageURL, notes)
}