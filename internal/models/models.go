package models

import (
        "time"
        "gorm.io/gorm"
)

// PlanType represents the subscription plan type
type PlanType string

const (
        PlanFree PlanType = "free"
        PlanPro  PlanType = "pro"
        PlanVIP  PlanType = "vip"
)

// User represents a telegram user
type User struct {
        ID        uint           `gorm:"primarykey" json:"id"`
        CreatedAt time.Time      `json:"created_at"`
        UpdatedAt time.Time      `json:"updated_at"`
        DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
        
        TelegramID   int64  `gorm:"uniqueIndex" json:"telegram_id"`
        Username     string `json:"username"`
        FirstName    string `json:"first_name"`
        LastName     string `json:"last_name"`
        IsAdmin      bool   `gorm:"default:false" json:"is_admin"`
        
        // Bot relationship
        Stores []Store `gorm:"foreignKey:OwnerID" json:"stores,omitempty"`
}

// Store represents a shop/store created by a user
type Store struct {
        ID        uint           `gorm:"primarykey" json:"id"`
        CreatedAt time.Time      `json:"created_at"`
        UpdatedAt time.Time      `json:"updated_at"`
        DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
        
        OwnerID     uint   `json:"owner_id"`
        Owner       User   `gorm:"foreignKey:OwnerID" json:"owner"`
        
        Name        string `json:"name"`
        Description string `json:"description"`
        BotToken    string `json:"bot_token"`
        BotUsername string `json:"bot_username"`
        
        // Subscription details
        PlanType        PlanType  `json:"plan_type"` // "free", "pro", "vip"
        ExpiresAt       time.Time `json:"expires_at"`
        IsActive        bool      `gorm:"default:true" json:"is_active"`
        ProductLimit    int       `json:"product_limit"`
        CommissionRate  int       `json:"commission_rate"`
        
        // Store settings
        WelcomeMessage  string `json:"welcome_message"`
        SupportContact  string `json:"support_contact"`
        
        // Relationships
        Products []Product `gorm:"foreignKey:StoreID" json:"products,omitempty"`
        Orders   []Order   `gorm:"foreignKey:StoreID" json:"orders,omitempty"`
}

// Product represents a product in a store
type Product struct {
        ID        uint           `gorm:"primarykey" json:"id"`
        CreatedAt time.Time      `json:"created_at"`
        UpdatedAt time.Time      `json:"updated_at"`
        DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
        
        StoreID     uint  `json:"store_id"`
        Store       Store `gorm:"foreignKey:StoreID" json:"store"`
        
        Name        string `json:"name"`
        Description string `json:"description"`
        Price       int64  `json:"price"` // in cents/smallest currency unit
        ImageURL    string `json:"image_url"`
        IsAvailable bool   `gorm:"default:true" json:"is_available"`
        
        // Product organization
        Category string `json:"category"`
        Tags     string `json:"tags"` // JSON array as string
        
        // Inventory
        Stock       int  `json:"stock"`
        TrackStock  bool `gorm:"default:false" json:"track_stock"`
        
        // Relationships
        OrderItems []OrderItem `gorm:"foreignKey:ProductID" json:"order_items,omitempty"`
}

// Order represents a customer order
type Order struct {
        ID        uint           `gorm:"primarykey" json:"id"`
        CreatedAt time.Time      `json:"created_at"`
        UpdatedAt time.Time      `json:"updated_at"`
        DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
        
        StoreID     uint  `json:"store_id"`
        Store       Store `gorm:"foreignKey:StoreID" json:"store"`
        
        // Customer info
        CustomerTelegramID int64  `json:"customer_telegram_id"`
        CustomerName       string `json:"customer_name"`
        CustomerUsername   string `json:"customer_username"`
        
        // Order details
        TotalAmount    int64  `json:"total_amount"`
        Status         string `json:"status"` // "pending", "confirmed", "shipped", "delivered", "cancelled"
        PaymentMethod  string `json:"payment_method"`
        PaymentStatus  string `json:"payment_status"` // "pending", "paid", "failed"
        
        // Delivery info
        DeliveryAddress string `json:"delivery_address"`
        DeliveryPhone   string `json:"delivery_phone"`
        DeliveryNotes   string `json:"delivery_notes"`
        
        // Commission
        CommissionAmount int64 `json:"commission_amount"`
        
        // Relationships
        OrderItems []OrderItem `gorm:"foreignKey:OrderID" json:"order_items,omitempty"`
}

// OrderItem represents individual items in an order
type OrderItem struct {
        ID        uint           `gorm:"primarykey" json:"id"`
        CreatedAt time.Time      `json:"created_at"`
        UpdatedAt time.Time      `json:"updated_at"`
        DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
        
        OrderID   uint    `json:"order_id"`
        Order     Order   `gorm:"foreignKey:OrderID" json:"order"`
        ProductID uint    `json:"product_id"`
        Product   Product `gorm:"foreignKey:ProductID" json:"product"`
        
        Quantity  int   `json:"quantity"`
        UnitPrice int64 `json:"unit_price"`
        SubTotal  int64 `json:"sub_total"`
}

// Payment represents payment records
type Payment struct {
        ID        uint           `gorm:"primarykey" json:"id"`
        CreatedAt time.Time      `json:"created_at"`
        UpdatedAt time.Time      `json:"updated_at"`
        DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
        
        StoreID uint  `json:"store_id"`
        Store   Store `gorm:"foreignKey:StoreID" json:"store"`
        
        Amount      int64  `json:"amount"`
        PaymentType string `json:"payment_type"` // "subscription", "commission"
        Status      string `json:"status"`       // "pending", "confirmed", "failed"
        
        // Payment proof
        ProofImageURL string `json:"proof_image_url"`
        Notes         string `json:"notes"`
        
        // Admin verification
        VerifiedBy   *uint      `json:"verified_by,omitempty"`
        VerifiedAt   *time.Time `json:"verified_at,omitempty"`
}

// UserSession represents user conversation state
type UserSession struct {
        TelegramID int64             `gorm:"primarykey" json:"telegram_id"`
        State      string            `json:"state"`
        Data       string            `json:"data"` // JSON data for current operation
        UpdatedAt  time.Time         `json:"updated_at"`
}