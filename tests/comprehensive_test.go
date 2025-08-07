package tests

import (
	"fmt"
	"log"
	"os"
	"testing"
	"telegram-store-hub/internal/config"
	"telegram-store-hub/internal/database"
	"telegram-store-hub/internal/models"
	"telegram-store-hub/internal/services"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

// TestConfig holds test configuration
type TestConfig struct {
	DB  *gorm.DB
	Bot *tgbotapi.BotAPI
}

// setupTestEnvironment sets up test environment
func setupTestEnvironment(t *testing.T) *TestConfig {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Skipf("Skipping integration tests: %v", err)
	}

	// Connect to test database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	if err := database.Migrate(db); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Create test bot (mock if no token)
	var bot *tgbotapi.BotAPI
	if cfg.MotherBotToken != "" {
		bot, err = tgbotapi.NewBotAPI(cfg.MotherBotToken)
		if err != nil {
			log.Printf("Warning: Could not create bot API: %v", err)
		}
	}

	return &TestConfig{
		DB:  db,
		Bot: bot,
	}
}

// TestDatabaseConnection tests database connectivity
func TestDatabaseConnection(t *testing.T) {
	testConfig := setupTestEnvironment(t)
	
	// Test health check
	if err := database.HealthCheck(testConfig.DB); err != nil {
		t.Fatalf("Database health check failed: %v", err)
	}
	
	t.Log("✅ Database connection test passed")
}

// TestUserService tests user service functionality
func TestUserService(t *testing.T) {
	testConfig := setupTestEnvironment(t)
	
	userService := services.NewUserService(testConfig.DB)
	
	// Test user creation
	testUser := &models.User{
		TelegramID: 123456789,
		Username:   "testuser",
		FirstName:  "Test",
		LastName:   "User",
	}
	
	if err := userService.CreateOrUpdateUser(testUser); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	
	// Test user retrieval
	retrievedUser, err := userService.GetUserByTelegramID(123456789)
	if err != nil {
		t.Fatalf("Failed to retrieve user: %v", err)
	}
	
	if retrievedUser.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", retrievedUser.Username)
	}
	
	// Test user update
	testUser.FirstName = "Updated"
	if err := userService.CreateOrUpdateUser(testUser); err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}
	
	updatedUser, err := userService.GetUserByTelegramID(123456789)
	if err != nil {
		t.Fatalf("Failed to retrieve updated user: %v", err)
	}
	
	if updatedUser.FirstName != "Updated" {
		t.Errorf("Expected first name 'Updated', got '%s'", updatedUser.FirstName)
	}
	
	t.Log("✅ User service tests passed")
}

// TestStoreManager tests store management functionality
func TestStoreManager(t *testing.T) {
	testConfig := setupTestEnvironment(t)
	
	userService := services.NewUserService(testConfig.DB)
	storeManager := services.NewStoreManagerService(testConfig.DB)
	
	// Create test user first
	testUser := &models.User{
		TelegramID: 987654321,
		Username:   "storeowner",
		FirstName:  "Store",
		LastName:   "Owner",
	}
	
	if err := userService.CreateOrUpdateUser(testUser); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	// Test store creation
	testStore := &models.Store{
		OwnerID:        testUser.ID,
		Name:           "Test Store",
		Description:    "A test store",
		PlanType:       models.PlanTypeFree,
		ExpiresAt:      time.Now().AddDate(0, 1, 0),
		IsActive:       true,
		ProductLimit:   10,
		CommissionRate: 5,
	}
	
	if err := storeManager.CreateStore(testStore); err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	
	// Test store retrieval
	retrievedStore, err := storeManager.GetStoreByID(testStore.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve store: %v", err)
	}
	
	if retrievedStore.Name != "Test Store" {
		t.Errorf("Expected store name 'Test Store', got '%s'", retrievedStore.Name)
	}
	
	// Test store update
	retrievedStore.Description = "Updated description"
	if err := storeManager.UpdateStore(retrievedStore); err != nil {
		t.Fatalf("Failed to update store: %v", err)
	}
	
	t.Log("✅ Store manager tests passed")
}

// TestProductService tests product management functionality
func TestProductService(t *testing.T) {
	testConfig := setupTestEnvironment(t)
	
	productService := services.NewProductService(testConfig.DB)
	
	// First create a store for testing
	testStore := &models.Store{
		Name:           "Product Test Store",
		Description:    "Store for product testing",
		PlanType:       models.PlanTypeFree,
		ExpiresAt:      time.Now().AddDate(0, 1, 0),
		IsActive:       true,
		ProductLimit:   10,
		CommissionRate: 5,
	}
	
	if err := testConfig.DB.Create(testStore).Error; err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	
	// Test product creation
	testProduct := &models.Product{
		StoreID:     testStore.ID,
		Name:        "Test Product",
		Description: "A test product",
		Price:       10000,
		IsAvailable: true,
		Category:    "Electronics",
	}
	
	if err := productService.CreateProduct(testProduct); err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	
	// Test product retrieval
	products, err := productService.GetStoreProducts(testStore.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve products: %v", err)
	}
	
	if len(products) != 1 {
		t.Errorf("Expected 1 product, got %d", len(products))
	}
	
	if products[0].Name != "Test Product" {
		t.Errorf("Expected product name 'Test Product', got '%s'", products[0].Name)
	}
	
	// Test product count
	count, err := productService.GetStoreProductCount(testStore.ID)
	if err != nil {
		t.Fatalf("Failed to get product count: %v", err)
	}
	
	if count != 1 {
		t.Errorf("Expected product count 1, got %d", count)
	}
	
	t.Log("✅ Product service tests passed")
}

// TestSubscriptionService tests subscription management
func TestSubscriptionService(t *testing.T) {
	testConfig := setupTestEnvironment(t)
	
	subscriptionSrv := services.NewSubscriptionService(
		testConfig.Bot,
		testConfig.DB,
		"1234-5678-9012-3456",
		"Test Card Holder",
	)
	
	// Test plan retrieval
	plans := subscriptionSrv.GetAvailablePlans()
	if len(plans) != 3 {
		t.Errorf("Expected 3 plans, got %d", len(plans))
	}
	
	// Test plan by type
	freePlan := subscriptionSrv.GetPlanByType("free")
	if freePlan == nil {
		t.Error("Free plan should not be nil")
	}
	
	if freePlan.Price != 0 {
		t.Errorf("Expected free plan price 0, got %d", freePlan.Price)
	}
	
	if freePlan.ProductLimit != 10 {
		t.Errorf("Expected free plan product limit 10, got %d", freePlan.ProductLimit)
	}
	
	proPlan := subscriptionSrv.GetPlanByType("pro")
	if proPlan == nil {
		t.Error("Pro plan should not be nil")
	}
	
	if proPlan.Price != 50000 {
		t.Errorf("Expected pro plan price 50000, got %d", proPlan.Price)
	}
	
	vipPlan := subscriptionSrv.GetPlanByType("vip")
	if vipPlan == nil {
		t.Error("VIP plan should not be nil")
	}
	
	if vipPlan.ProductLimit != -1 {
		t.Errorf("Expected VIP plan unlimited products (-1), got %d", vipPlan.ProductLimit)
	}
	
	t.Log("✅ Subscription service tests passed")
}

// TestReminderService tests reminder functionality
func TestReminderService(t *testing.T) {
	testConfig := setupTestEnvironment(t)
	
	subscriptionSrv := services.NewSubscriptionService(
		testConfig.Bot,
		testConfig.DB,
		"1234-5678-9012-3456",
		"Test Card Holder",
	)
	
	reminderService := services.NewReminderService(
		testConfig.Bot,
		testConfig.DB,
		subscriptionSrv,
		[]int{7, 3, 1},
	)
	
	// Create test store with expiration
	testStore := &models.Store{
		Name:           "Reminder Test Store",
		Description:    "Store for reminder testing",
		PlanType:       models.PlanTypeFree,
		ExpiresAt:      time.Now().AddDate(0, 0, 7), // Expires in 7 days
		IsActive:       true,
		ProductLimit:   10,
		CommissionRate: 5,
	}
	
	if err := testConfig.DB.Create(testStore).Error; err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	
	// Test upcoming expirations
	upcomingStores, err := reminderService.GetUpcomingExpirations(10)
	if err != nil {
		t.Fatalf("Failed to get upcoming expirations: %v", err)
	}
	
	found := false
	for _, store := range upcomingStores {
		if store.ID == testStore.ID {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Test store should be in upcoming expirations")
	}
	
	t.Log("✅ Reminder service tests passed")
}

// TestBotManagerService tests bot management functionality
func TestBotManagerService(t *testing.T) {
	testConfig := setupTestEnvironment(t)
	
	botManager := services.NewBotManagerService(testConfig.Bot, testConfig.DB)
	
	// Create test store
	testStore := &models.Store{
		Name:           "Bot Test Store",
		Description:    "Store for bot testing",
		PlanType:       models.PlanTypePro,
		ExpiresAt:      time.Now().AddDate(0, 1, 0),
		IsActive:       true,
		ProductLimit:   200,
		CommissionRate: 5,
	}
	
	if err := testConfig.DB.Create(testStore).Error; err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	
	// Test bot creation (this will be simulated)
	if err := botManager.CreateSubBot(testStore.ID); err != nil {
		t.Fatalf("Failed to create sub-bot: %v", err)
	}
	
	// Wait a moment for async processing
	time.Sleep(3 * time.Second)
	
	// Check bot status
	status, err := botManager.GetBotStatus(testStore.ID)
	if err != nil {
		t.Fatalf("Failed to get bot status: %v", err)
	}
	
	if status != services.BotStatusActive && status != services.BotStatusCreating {
		t.Errorf("Expected bot status to be active or creating, got %s", status)
	}
	
	t.Log("✅ Bot manager service tests passed")
}

// TestCompleteWorkflow tests the complete workflow
func TestCompleteWorkflow(t *testing.T) {
	testConfig := setupTestEnvironment(t)
	
	// Create all services
	userService := services.NewUserService(testConfig.DB)
	storeManager := services.NewStoreManagerService(testConfig.DB)
	productService := services.NewProductService(testConfig.DB)
	sessionService := services.NewSessionService(testConfig.DB)
	subscriptionSrv := services.NewSubscriptionService(
		testConfig.Bot,
		testConfig.DB,
		"1234-5678-9012-3456",
		"Test Workflow",
	)
	botManager := services.NewBotManagerService(testConfig.Bot, testConfig.DB)
	
	// Step 1: Create user
	testUser := &models.User{
		TelegramID: 111222333,
		Username:   "workflowuser",
		FirstName:  "Workflow",
		LastName:   "Test",
	}
	
	if err := userService.CreateOrUpdateUser(testUser); err != nil {
		t.Fatalf("Step 1 failed - user creation: %v", err)
	}
	
	// Step 2: Create store
	testStore := &models.Store{
		OwnerID:        testUser.ID,
		Name:           "Complete Workflow Store",
		Description:    "Testing complete workflow",
		PlanType:       models.PlanTypePro,
		ExpiresAt:      time.Now().AddDate(0, 1, 0),
		IsActive:       true,
		ProductLimit:   200,
		CommissionRate: 5,
	}
	
	if err := storeManager.CreateStore(testStore); err != nil {
		t.Fatalf("Step 2 failed - store creation: %v", err)
	}
	
	// Step 3: Add products
	for i := 1; i <= 3; i++ {
		product := &models.Product{
			StoreID:     testStore.ID,
			Name:        fmt.Sprintf("Product %d", i),
			Description: fmt.Sprintf("Description for product %d", i),
			Price:       int64(i * 10000),
			IsAvailable: true,
			Category:    "Test Category",
		}
		
		if err := productService.CreateProduct(product); err != nil {
			t.Fatalf("Step 3 failed - product creation: %v", err)
		}
	}
	
	// Step 4: Verify product count
	productCount, err := productService.GetStoreProductCount(testStore.ID)
	if err != nil {
		t.Fatalf("Step 4 failed - product count: %v", err)
	}
	
	if productCount != 3 {
		t.Errorf("Expected 3 products, got %d", productCount)
	}
	
	// Step 5: Create sub-bot
	if err := botManager.CreateSubBot(testStore.ID); err != nil {
		t.Fatalf("Step 5 failed - bot creation: %v", err)
	}
	
	// Step 6: Test plan renewal
	if err := subscriptionSrv.ProcessPlanRenewal(testStore.ID, "vip"); err != nil {
		t.Fatalf("Step 6 failed - plan renewal: %v", err)
	}
	
	// Step 7: Verify plan update
	updatedStore, err := storeManager.GetStoreByID(testStore.ID)
	if err != nil {
		t.Fatalf("Step 7 failed - store retrieval: %v", err)
	}
	
	if updatedStore.PlanType != models.PlanTypeVIP {
		t.Errorf("Expected plan type VIP, got %s", updatedStore.PlanType)
	}
	
	if updatedStore.CommissionRate != 0 {
		t.Errorf("Expected commission rate 0 for VIP, got %d", updatedStore.CommissionRate)
	}
	
	t.Log("✅ Complete workflow test passed")
}

// TestSystemStats tests system statistics
func TestSystemStats(t *testing.T) {
	testConfig := setupTestEnvironment(t)
	
	// Get database statistics
	stats, err := database.GetStats(testConfig.DB)
	if err != nil {
		t.Fatalf("Failed to get database stats: %v", err)
	}
	
	// Stats should be non-negative
	for key, value := range stats {
		if value < 0 {
			t.Errorf("Stat %s should not be negative, got %d", key, value)
		}
	}
	
	t.Logf("✅ System stats test passed - Users: %d, Stores: %d, Products: %d", 
		stats["users"], stats["stores"], stats["products"])
}

// TestCleanupOperations tests cleanup operations
func TestCleanupOperations(t *testing.T) {
	testConfig := setupTestEnvironment(t)
	
	// Test session cleanup
	if err := database.CleanupExpiredSessions(testConfig.DB); err != nil {
		t.Fatalf("Failed to cleanup expired sessions: %v", err)
	}
	
	// Test reminder log cleanup (if reminder service is available)
	reminderService := services.NewReminderService(
		testConfig.Bot,
		testConfig.DB,
		nil,
		[]int{7, 3, 1},
	)
	
	if err := reminderService.CleanupOldReminderLogs(); err != nil {
		t.Fatalf("Failed to cleanup old reminder logs: %v", err)
	}
	
	t.Log("✅ Cleanup operations test passed")
}

// BenchmarkUserCreation benchmarks user creation performance
func BenchmarkUserCreation(b *testing.B) {
	// Skip if no database configuration
	if os.Getenv("DATABASE_URL") == "" {
		b.Skip("No DATABASE_URL configured")
	}
	
	testConfig := setupTestEnvironment(&testing.T{})
	userService := services.NewUserService(testConfig.DB)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		user := &models.User{
			TelegramID: int64(1000000 + i),
			Username:   fmt.Sprintf("benchuser%d", i),
			FirstName:  "Bench",
			LastName:   "User",
		}
		
		userService.CreateOrUpdateUser(user)
	}
}

// TestMain runs setup and teardown for tests
func TestMain(m *testing.M) {
	log.Println("🧪 Starting comprehensive tests...")
	
	// Run tests
	code := m.Run()
	
	log.Println("🏁 Comprehensive tests completed")
	
	// Exit with the code returned from tests
	os.Exit(code)
}