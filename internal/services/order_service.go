package services

import (
	"telegram-store-hub/internal/models"
	"time"

	"gorm.io/gorm"
)

type OrderService struct {
	db *gorm.DB
}

func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{db: db}
}

// CreateOrder creates a new order
func (s *OrderService) CreateOrder(storeID uint, customerTelegramID int64, customerName, customerUsername string) (*models.Order, error) {
	order := models.Order{
		StoreID:            storeID,
		CustomerTelegramID: customerTelegramID,
		CustomerName:       customerName,
		CustomerUsername:   customerUsername,
		TotalAmount:        0,
		Status:             "pending",
		PaymentStatus:      "pending",
	}
	
	if err := s.db.Create(&order).Error; err != nil {
		return nil, err
	}
	
	return &order, nil
}

// AddOrderItem adds an item to an order
func (s *OrderService) AddOrderItem(orderID, productID uint, quantity int, unitPrice int64) error {
	orderItem := models.OrderItem{
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  quantity,
		UnitPrice: unitPrice,
		SubTotal:  unitPrice * int64(quantity),
	}
	
	if err := s.db.Create(&orderItem).Error; err != nil {
		return err
	}
	
	// Update order total
	return s.UpdateOrderTotal(orderID)
}

// UpdateOrderTotal recalculates and updates order total
func (s *OrderService) UpdateOrderTotal(orderID uint) error {
	var total int64
	s.db.Table("order_items").
		Where("order_id = ?", orderID).
		Select("COALESCE(SUM(sub_total), 0)").
		Row().Scan(&total)
	
	return s.db.Model(&models.Order{}).Where("id = ?", orderID).Update("total_amount", total).Error
}

// GetOrderByID gets order by ID with all items
func (s *OrderService) GetOrderByID(orderID uint) (*models.Order, error) {
	var order models.Order
	err := s.db.Preload("OrderItems").Preload("OrderItems.Product").Preload("Store").First(&order, orderID).Error
	return &order, err
}

// GetStoreOrders gets all orders for a store
func (s *OrderService) GetStoreOrders(storeID uint, limit, offset int) ([]models.Order, error) {
	var orders []models.Order
	err := s.db.Where("store_id = ?", storeID).
		Preload("OrderItems").
		Preload("OrderItems.Product").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error
	return orders, err
}

// GetOrdersByStatus gets orders by status
func (s *OrderService) GetOrdersByStatus(storeID uint, status string) ([]models.Order, error) {
	var orders []models.Order
	err := s.db.Where("store_id = ? AND status = ?", storeID, status).
		Preload("OrderItems").
		Preload("OrderItems.Product").
		Order("created_at DESC").
		Find(&orders).Error
	return orders, err
}

// UpdateOrderStatus updates order status
func (s *OrderService) UpdateOrderStatus(orderID uint, status string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	
	// If completed, mark payment as paid
	if status == "completed" {
		updates["payment_status"] = "paid"
	}
	
	return s.db.Model(&models.Order{}).Where("id = ?", orderID).Updates(updates).Error
}

// UpdateOrderPaymentStatus updates payment status
func (s *OrderService) UpdateOrderPaymentStatus(orderID uint, paymentStatus string) error {
	return s.db.Model(&models.Order{}).Where("id = ?", orderID).Update("payment_status", paymentStatus).Error
}

// CancelOrder cancels an order
func (s *OrderService) CancelOrder(orderID uint, reason string) error {
	updates := map[string]interface{}{
		"status":         "cancelled",
		"delivery_notes": reason,
		"updated_at":     time.Now(),
	}
	
	return s.db.Model(&models.Order{}).Where("id = ?", orderID).Updates(updates).Error
}

// GetCustomerOrders gets orders for a specific customer
func (s *OrderService) GetCustomerOrders(customerTelegramID int64, limit int) ([]models.Order, error) {
	var orders []models.Order
	err := s.db.Where("customer_telegram_id = ?", customerTelegramID).
		Preload("Store").
		Preload("OrderItems").
		Preload("OrderItems.Product").
		Order("created_at DESC").
		Limit(limit).
		Find(&orders).Error
	return orders, err
}

// GetOrderStats gets order statistics for a store
func (s *OrderService) GetOrderStats(storeID uint, days int) (map[string]interface{}, error) {
	startDate := time.Now().AddDate(0, 0, -days)
	
	var stats struct {
		TotalOrders     int64 `json:"total_orders"`
		CompletedOrders int64 `json:"completed_orders"`
		PendingOrders   int64 `json:"pending_orders"`
		CancelledOrders int64 `json:"cancelled_orders"`
		TotalRevenue    int64 `json:"total_revenue"`
	}
	
	// Get order counts
	s.db.Model(&models.Order{}).Where("store_id = ? AND created_at >= ?", storeID, startDate).Count(&stats.TotalOrders)
	s.db.Model(&models.Order{}).Where("store_id = ? AND status = 'completed' AND created_at >= ?", storeID, startDate).Count(&stats.CompletedOrders)
	s.db.Model(&models.Order{}).Where("store_id = ? AND status = 'pending' AND created_at >= ?", storeID, startDate).Count(&stats.PendingOrders)
	s.db.Model(&models.Order{}).Where("store_id = ? AND status = 'cancelled' AND created_at >= ?", storeID, startDate).Count(&stats.CancelledOrders)
	
	// Get total revenue
	s.db.Table("orders").
		Where("store_id = ? AND status = 'completed' AND created_at >= ?", storeID, startDate).
		Select("COALESCE(SUM(total_amount), 0)").
		Row().Scan(&stats.TotalRevenue)
	
	return map[string]interface{}{
		"total_orders":     stats.TotalOrders,
		"completed_orders": stats.CompletedOrders,
		"pending_orders":   stats.PendingOrders,
		"cancelled_orders": stats.CancelledOrders,
		"total_revenue":    stats.TotalRevenue,
	}, nil
}

// GetDailySales gets daily sales data for charts
func (s *OrderService) GetDailySales(storeID uint, days int) ([]map[string]interface{}, error) {
	startDate := time.Now().AddDate(0, 0, -days)
	
	var results []struct {
		Date         time.Time `json:"date"`
		OrderCount   int64     `json:"order_count"`
		TotalRevenue int64     `json:"total_revenue"`
	}
	
	err := s.db.Table("orders").
		Select("DATE(created_at) as date, COUNT(*) as order_count, COALESCE(SUM(total_amount), 0) as total_revenue").
		Where("store_id = ? AND status = 'completed' AND created_at >= ?", storeID, startDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error
	
	if err != nil {
		return nil, err
	}
	
	// Convert to map slice
	data := make([]map[string]interface{}, len(results))
	for i, result := range results {
		data[i] = map[string]interface{}{
			"date":          result.Date.Format("2006-01-02"),
			"order_count":   result.OrderCount,
			"total_revenue": result.TotalRevenue,
		}
	}
	
	return data, nil
}

// UpdateOrderDeliveryInfo updates delivery information
func (s *OrderService) UpdateOrderDeliveryInfo(orderID uint, address, phone, notes string) error {
	updates := map[string]interface{}{
		"delivery_address": address,
		"delivery_phone":   phone,
		"delivery_notes":   notes,
		"updated_at":       time.Now(),
	}
	
	return s.db.Model(&models.Order{}).Where("id = ?", orderID).Updates(updates).Error
}

// CalculateCommission calculates commission for an order
func (s *OrderService) CalculateCommission(orderID uint) error {
	var order models.Order
	if err := s.db.Preload("Store").First(&order, orderID).Error; err != nil {
		return err
	}
	
	commissionAmount := order.TotalAmount * int64(order.Store.CommissionRate) / 100
	
	return s.db.Model(&models.Order{}).Where("id = ?", orderID).Update("commission_amount", commissionAmount).Error
}