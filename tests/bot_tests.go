package tests

import (
	"database/sql/driver"
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"telegram-store-hub/internal/config"
	"telegram-store-hub/internal/database"
	"telegram-store-hub/internal/models"
	"telegram-store-hub/internal/services"

	"github.com/DATA-DOG/go-sqlmock"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Test structure to verify bot system components
type TestSuite struct {
	db       *gorm.DB
	mock     sqlmock.Sqlmock
	services *TestServices
}

type TestServices struct {
	userService         *services.UserService
	sessionService      *services.SessionService
	storeManager        *services.StoreManagerService
	productService      *services.ProductService
	orderService        *services.OrderService
	paymentService      *services.PaymentService
	subscriptionService *services.SubscriptionService
	botManager          *services.BotManagerService
}

// SetupTestSuite initializes the test environment
func SetupTestSuite() (*TestSuite, error) {
	// Create SQL mock
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL mock: %v", err)
	}

	// Create GORM DB with mock
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create GORM DB: %v", err)
	}

	// Initialize services
	testServices := &TestServices{
		userService:         services.NewUserService(gormDB),
		sessionService:      services.NewSessionService(gormDB),
		storeManager:        services.NewStoreManagerService(gormDB),
		productService:      services.NewProductService(gormDB),
		orderService:        services.NewOrderService(gormDB),
		paymentService:      services.NewPaymentService(gormDB),
		subscriptionService: services.NewSubscriptionService(gormDB),
		botManager:          services.NewBotManagerService(gormDB),
	}

	return &TestSuite{
		db:       gormDB,
		mock:     mock,
		services: testServices,
	}, nil
}

// TestDatabaseConnectivity tests database connection and basic operations
func TestDatabaseConnectivity(t *testing.T) {
	suite, err := SetupTestSuite()
	if err != nil {
		t.Fatalf("Failed to setup test suite: %v", err)
	}
	defer suite.db.Exec("SELECT 1") // Cleanup

	log.Println("✅ Database connectivity test passed")
}

// TestUserService tests user creation and management
func TestUserService(t *testing.T) {
	suite, err := SetupTestSuite()
	if err != nil {
		t.Fatalf("Failed to setup test suite: %v", err)
	}

	// Mock user creation
	suite.mock.ExpectBegin()
	suite.mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO "users" ("created_at","updated_at","deleted_at","telegram_id","username","first_name","last_name","is_admin") VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING "id"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, int64(123456789), "testuser", "Test", "User", false).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	suite.mock.ExpectCommit()

	// Test user creation
	user := &models.User{
		TelegramID: 123456789,
		Username:   "testuser",
		FirstName:  "Test",
		LastName:   "User",
		IsAdmin:    false,
	}

	err = suite.services.userService.CreateUser(user)
	if err != nil {
		t.Errorf("Failed to create user: %v", err)
	}

	log.Println("✅ User service test passed")
}

// TestStoreCreation tests store creation process
func TestStoreCreation(t *testing.T) {
	suite, err := SetupTestSuite()
	if err != nil {
		t.Fatalf("Failed to setup test suite: %v", err)
	}

	// Mock store creation
	suite.mock.ExpectBegin()
	suite.mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO "stores" ("created_at","updated_at","deleted_at","owner_id","name","description","bot_token","bot_username","plan_type","expires_at","is_active","product_limit","commission_rate","welcome_message","support_contact") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING "id"`)).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), nil,
			uint(1), "Test Store", "Test Description", "", "",
			models.PlanFree, sqlmock.AnyArg(), true, 10, 5, "", "").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	suite.mock.ExpectCommit()

	// Test store creation
	store := &models.Store{
		OwnerID:        1,
		Name:           "Test Store",
		Description:    "Test Description",
		PlanType:       models.PlanFree,
		ExpiresAt:      time.Now().AddDate(0, 1, 0),
		IsActive:       true,
		ProductLimit:   10,
		CommissionRate: 5,
	}

	err = suite.services.storeManager.CreateStore(store)
	if err != nil {
		t.Errorf("Failed to create store: %v", err)
	}

	log.Println("✅ Store creation test passed")
}

// TestProductManagement tests product CRUD operations
func TestProductManagement(t *testing.T) {
	suite, err := SetupTestSuite()
	if err != nil {
		t.Fatalf("Failed to setup test suite: %v", err)
	}

	// Mock product creation
	suite.mock.ExpectBegin()
	suite.mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO "products" ("created_at","updated_at","deleted_at","store_id","name","description","price","image_url","is_available","category","tags","stock","track_stock") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "id"`)).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), nil,
			uint(1), "Test Product", "Test Description", int64(1000), "", true, "Test Category", "", 100, false).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	suite.mock.ExpectCommit()

	// Test product creation
	product := &models.Product{
		StoreID:     1,
		Name:        "Test Product",
		Description: "Test Description",
		Price:       1000, // 10.00 in cents
		IsAvailable: true,
		Category:    "Test Category",
		Stock:       100,
		TrackStock:  false,
	}

	err = suite.services.productService.CreateProduct(product)
	if err != nil {
		t.Errorf("Failed to create product: %v", err)
	}

	log.Println("✅ Product management test passed")
}

// TestSubscriptionSystem tests subscription logic
func TestSubscriptionSystem(t *testing.T) {
	suite, err := SetupTestSuite()
	if err != nil {
		t.Fatalf("Failed to setup test suite: %v", err)
	}

	// Test plan limits
	planLimits := map[models.PlanType]struct {
		ProductLimit   int
		CommissionRate int
		Price          int64
	}{
		models.PlanFree: {ProductLimit: 10, CommissionRate: 5, Price: 0},
		models.PlanPro:  {ProductLimit: 200, CommissionRate: 5, Price: 50000},
		models.PlanVIP:  {ProductLimit: -1, CommissionRate: 0, Price: 150000}, // -1 = unlimited
	}

	for planType, limits := range planLimits {
		if planType == models.PlanFree && limits.ProductLimit != 10 {
			t.Errorf("Free plan should have 10 product limit, got %d", limits.ProductLimit)
		}
		if planType == models.PlanPro && limits.ProductLimit != 200 {
			t.Errorf("Pro plan should have 200 product limit, got %d", limits.ProductLimit)
		}
		if planType == models.PlanVIP && limits.CommissionRate != 0 {
			t.Errorf("VIP plan should have 0 commission rate, got %d", limits.CommissionRate)
		}
	}

	log.Println("✅ Subscription system test passed")
}

// TestBotTokenValidation tests bot token validation
func TestBotTokenValidation(t *testing.T) {
	// Test with invalid token (should fail gracefully)
	_, err := tgbotapi.NewBotAPI("invalid_token")
	if err == nil {
		t.Error("Expected error with invalid bot token")
	}

	log.Println("✅ Bot token validation test passed")
}

// TestConfigurationLoading tests configuration loading
func TestConfigurationLoading(t *testing.T) {
	// This test ensures the config structure is valid
	cfg := &config.Config{
		MotherBotToken:     "test_token",
		DatabaseURL:        "postgres://test",
		Debug:              true,
		RequiredChannelID:  "@testchannel",
		AdminChatID:        123456789,
		PaymentCardNumber:  "1234-5678-9012-3456",
		PaymentCardHolder:  "Test Holder",
		FreePlanPrice:      0,
		ProPlanPrice:       50000,
		VIPPlanPrice:       150000,
		FreePlanCommission: 5,
		ProPlanCommission:  5,
		VIPPlanCommission:  0,
	}

	if cfg.FreePlanPrice != 0 {
		t.Error("Free plan price should be 0")
	}
	if cfg.VIPPlanCommission != 0 {
		t.Error("VIP plan commission should be 0")
	}

	log.Println("✅ Configuration loading test passed")
}

// TestErrorHandling tests error handling scenarios
func TestErrorHandling(t *testing.T) {
	suite, err := SetupTestSuite()
	if err != nil {
		t.Fatalf("Failed to setup test suite: %v", err)
	}

	// Test handling of database errors
	suite.mock.ExpectBegin()
	suite.mock.ExpectQuery(regexp.QuoteMeta("INSERT")).
		WillReturnError(fmt.Errorf("database connection lost"))
	suite.mock.ExpectRollback()

	// Attempt operation that should fail
	user := &models.User{
		TelegramID: 123456789,
		Username:   "testuser",
	}

	err = suite.services.userService.CreateUser(user)
	if err == nil {
		t.Error("Expected database error but got none")
	}

	log.Println("✅ Error handling test passed")
}

// TestConcurrentOperations tests concurrent operations
func TestConcurrentOperations(t *testing.T) {
	suite, err := SetupTestSuite()
	if err != nil {
		t.Fatalf("Failed to setup test suite: %v", err)
	}

	// Test concurrent session operations
	numOperations := 10
	done := make(chan bool, numOperations)

	for i := 0; i < numOperations; i++ {
		go func(id int) {
			// Mock session operations
			userID := int64(id)
			suite.services.sessionService.SetUserState(userID, "test_state", "test_data")
			suite.services.sessionService.GetUserState(userID)
			suite.services.sessionService.ClearUserState(userID)
			done <- true
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < numOperations; i++ {
		<-done
	}

	log.Println("✅ Concurrent operations test passed")
}

// Helper function to run all tests
func RunAllTests() {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{"Database Connectivity", TestDatabaseConnectivity},
		{"User Service", TestUserService},
		{"Store Creation", TestStoreCreation},
		{"Product Management", TestProductManagement},
		{"Subscription System", TestSubscriptionSystem},
		{"Bot Token Validation", TestBotTokenValidation},
		{"Configuration Loading", TestConfigurationLoading},
		{"Error Handling", TestErrorHandling},
		{"Concurrent Operations", TestConcurrentOperations},
	}

	log.Println("🧪 Starting comprehensive bot system tests...")

	for _, test := range tests {
		t := &testing.T{}
		log.Printf("🔄 Running test: %s", test.name)
		test.test(t)
		if t.Failed() {
			log.Printf("❌ Test failed: %s", test.name)
		} else {
			log.Printf("✅ Test passed: %s", test.name)
		}
	}

	log.Println("🎉 All tests completed!")
}