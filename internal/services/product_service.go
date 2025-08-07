package services

import (
	"telegram-store-hub/internal/models"

	"gorm.io/gorm"
)

type ProductService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{db: db}
}

// CreateProduct creates a new product
func (s *ProductService) CreateProduct(storeID uint, name, description string, price int64, imageURL string) (*models.Product, error) {
	product := models.Product{
		StoreID:     storeID,
		Name:        name,
		Description: description,
		Price:       price,
		ImageURL:    imageURL,
		IsAvailable: true,
		Stock:       0,
		TrackStock:  false,
	}
	
	if err := s.db.Create(&product).Error; err != nil {
		return nil, err
	}
	
	return &product, nil
}

// GetProductByID gets product by ID
func (s *ProductService) GetProductByID(productID uint) (*models.Product, error) {
	var product models.Product
	err := s.db.Preload("Store").First(&product, productID).Error
	return &product, err
}

// GetStoreProducts gets all products for a store
func (s *ProductService) GetStoreProducts(storeID uint) ([]models.Product, error) {
	var products []models.Product
	err := s.db.Where("store_id = ?", storeID).Order("created_at DESC").Find(&products).Error
	return products, err
}

// GetActiveStoreProducts gets only active products for a store
func (s *ProductService) GetActiveStoreProducts(storeID uint) ([]models.Product, error) {
	var products []models.Product
	err := s.db.Where("store_id = ? AND is_available = ?", storeID, true).Order("created_at DESC").Find(&products).Error
	return products, err
}

// UpdateProduct updates product information
func (s *ProductService) UpdateProduct(product *models.Product) error {
	return s.db.Save(product).Error
}

// DeleteProduct deletes a product
func (s *ProductService) DeleteProduct(productID uint) error {
	return s.db.Delete(&models.Product{}, productID).Error
}

// ToggleProductAvailability toggles product availability
func (s *ProductService) ToggleProductAvailability(productID uint) error {
	var product models.Product
	if err := s.db.First(&product, productID).Error; err != nil {
		return err
	}
	
	product.IsAvailable = !product.IsAvailable
	return s.db.Save(&product).Error
}

// UpdateProductStock updates product stock
func (s *ProductService) UpdateProductStock(productID uint, stock int) error {
	return s.db.Model(&models.Product{}).Where("id = ?", productID).Updates(map[string]interface{}{
		"stock":       stock,
		"track_stock": true,
	}).Error
}

// CheckProductStock checks if product has enough stock
func (s *ProductService) CheckProductStock(productID uint, quantity int) (bool, error) {
	var product models.Product
	if err := s.db.First(&product, productID).Error; err != nil {
		return false, err
	}
	
	if !product.TrackStock {
		return true, nil // No stock tracking
	}
	
	return product.Stock >= quantity, nil
}

// GetProductsByCategory gets products by category
func (s *ProductService) GetProductsByCategory(storeID uint, category string) ([]models.Product, error) {
	var products []models.Product
	err := s.db.Where("store_id = ? AND category = ? AND is_available = ?", storeID, category, true).
		Order("created_at DESC").Find(&products).Error
	return products, err
}

// SearchProducts searches products by name or description
func (s *ProductService) SearchProducts(storeID uint, query string) ([]models.Product, error) {
	var products []models.Product
	searchQuery := "%" + query + "%"
	err := s.db.Where("store_id = ? AND (name ILIKE ? OR description ILIKE ?) AND is_available = ?", 
		storeID, searchQuery, searchQuery, true).
		Order("created_at DESC").Find(&products).Error
	return products, err
}

// GetTopSellingProducts gets top selling products for a store
func (s *ProductService) GetTopSellingProducts(storeID uint, limit int) ([]models.Product, error) {
	var products []models.Product
	err := s.db.Table("products").
		Select("products.*, COALESCE(SUM(order_items.quantity), 0) as total_sold").
		Joins("LEFT JOIN order_items ON products.id = order_items.product_id").
		Joins("LEFT JOIN orders ON order_items.order_id = orders.id").
		Where("products.store_id = ? AND (orders.status = 'completed' OR orders.status IS NULL)", storeID).
		Group("products.id").
		Order("total_sold DESC").
		Limit(limit).
		Find(&products).Error
	return products, err
}

// GetProductSalesStats gets sales statistics for a product
func (s *ProductService) GetProductSalesStats(productID uint) (map[string]interface{}, error) {
	var stats struct {
		TotalSold    int   `json:"total_sold"`
		TotalRevenue int64 `json:"total_revenue"`
		OrderCount   int   `json:"order_count"`
	}
	
	err := s.db.Table("order_items").
		Select("COALESCE(SUM(quantity), 0) as total_sold, COALESCE(SUM(sub_total), 0) as total_revenue, COUNT(DISTINCT order_id) as order_count").
		Joins("JOIN orders ON order_items.order_id = orders.id").
		Where("order_items.product_id = ? AND orders.status = 'completed'", productID).
		Scan(&stats).Error
	
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"total_sold":    stats.TotalSold,
		"total_revenue": stats.TotalRevenue,
		"order_count":   stats.OrderCount,
	}, nil
}