package models

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a connection pool to the test database
func setupTestDB(t *testing.T) *pgxpool.Pool {
	// Get database URL from environment or use default test database
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/stock_in_order_test?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err, "Failed to connect to test database")

	// Ping to verify connection
	err = pool.Ping(context.Background())
	require.NoError(t, err, "Failed to ping test database")

	return pool
}

// cleanupTestData removes test data from the database
func cleanupTestData(t *testing.T, db *pgxpool.Pool) {
	ctx := context.Background()

	// Delete in reverse order of foreign key dependencies
	queries := []string{
		"DELETE FROM stock_movements WHERE user_id = 999",
		"DELETE FROM order_items WHERE order_id IN (SELECT id FROM sales_orders WHERE user_id = 999)",
		"DELETE FROM sales_orders WHERE user_id = 999",
		"DELETE FROM product_batches WHERE user_id = 999",
		"DELETE FROM products WHERE user_id = 999",
		"DELETE FROM customers WHERE user_id = 999",
		"DELETE FROM users WHERE id = 999",
	}

	for _, query := range queries {
		_, err := db.Exec(ctx, query)
		if err != nil {
			t.Logf("Warning: cleanup query failed: %s - %v", query, err)
		}
	}
}

// createTestUser creates a test user with organization
func createTestUser(t *testing.T, db *pgxpool.Pool) int64 {
	ctx := context.Background()

	// Create test organization first (if not exists)
	const insertOrg = `
		INSERT INTO organizations (id, name, created_at)
		VALUES (999, 'Test Organization', NOW())
		ON CONFLICT (id) DO NOTHING`
	_, err := db.Exec(ctx, insertOrg)
	require.NoError(t, err, "Failed to create test organization")

	// Create test user
	const insertUser = `
		INSERT INTO users (id, name, email, password_hash, role, organization_id, created_at)
		VALUES (999, 'Test User', 'test@example.com', '$2a$10$test', 'admin', 999, NOW())
		ON CONFLICT (id) DO UPDATE SET email = EXCLUDED.email
		RETURNING id`

	var userID int64
	err = db.QueryRow(ctx, insertUser).Scan(&userID)
	require.NoError(t, err, "Failed to create test user")

	return userID
}

// createTestProduct creates a test product
func createTestProduct(t *testing.T, db *pgxpool.Pool, userID int64, name string) int64 {
	ctx := context.Background()

	const insertProduct = `
		INSERT INTO products (name, sku, price, stock_minimo, user_id, created_at)
		VALUES ($1, $2, 100.0, 5, $3, NOW())
		RETURNING id`

	var productID int64
	err := db.QueryRow(ctx, insertProduct, name, "SKU-TEST-"+name, userID).Scan(&productID)
	require.NoError(t, err, "Failed to create test product")

	return productID
}

// createTestBatch creates a test batch for a product
func createTestBatch(t *testing.T, db *pgxpool.Pool, productID, userID int64, quantity int, loteNumber string, expiryDate time.Time) int64 {
	ctx := context.Background()

	const insertBatch = `
		INSERT INTO product_batches (product_id, user_id, lote_number, quantity, expiry_date, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id`

	var batchID int64
	err := db.QueryRow(ctx, insertBatch, productID, userID, loteNumber, quantity, expiryDate).Scan(&batchID)
	require.NoError(t, err, "Failed to create test batch")

	return batchID
}

// getBatchQuantity retrieves the current quantity of a batch
func getBatchQuantity(t *testing.T, db *pgxpool.Pool, batchID int64) int {
	ctx := context.Background()

	const query = `SELECT quantity FROM product_batches WHERE id = $1`

	var quantity int
	err := db.QueryRow(ctx, query, batchID).Scan(&quantity)
	require.NoError(t, err, "Failed to get batch quantity")

	return quantity
}

// TestConsumeStockFEFO tests the First Expired First Out logic
func TestConsumeStockFEFO(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Clean up before and after test
	cleanupTestData(t, db)
	defer cleanupTestData(t, db)

	// Setup test data
	userID := createTestUser(t, db)
	productID := createTestProduct(t, db, userID, "Leche")

	// Create two batches:
	// Lote A: 10 units, expires TOMORROW (should be consumed first)
	tomorrow := time.Now().AddDate(0, 0, 1)
	batchA := createTestBatch(t, db, productID, userID, 10, "LOTE-A", tomorrow)

	// Lote B: 10 units, expires IN A MONTH (should be consumed second)
	nextMonth := time.Now().AddDate(0, 1, 0)
	batchB := createTestBatch(t, db, productID, userID, 10, "LOTE-B", nextMonth)

	t.Run("Consume 15 units - FEFO Logic", func(t *testing.T) {
		// Start a transaction to test ConsumeStockFEFO
		ctx := context.Background()
		tx, err := db.Begin(ctx)
		require.NoError(t, err, "Failed to begin transaction")
		defer tx.Rollback(ctx)

		// Consume 15 units (10 from Lote A + 5 from Lote B)
		err = ConsumeStockFEFO(ctx, tx, productID, userID, 15)
		assert.NoError(t, err, "ConsumeStockFEFO should not return error")

		// Commit the transaction
		err = tx.Commit(ctx)
		require.NoError(t, err, "Failed to commit transaction")

		// Verify Lote A: Should be 0 (fully consumed)
		quantityA := getBatchQuantity(t, db, batchA)
		assert.Equal(t, 0, quantityA, "Lote A should be fully consumed (0 units)")

		// Verify Lote B: Should be 5 (10 - 5 = 5 remaining)
		quantityB := getBatchQuantity(t, db, batchB)
		assert.Equal(t, 5, quantityB, "Lote B should have 5 units remaining")

		t.Logf("✅ FEFO Logic verified: Lote A=%d, Lote B=%d", quantityA, quantityB)
	})
}

// TestConsumeStockFEFO_InsufficientStock tests the rollback when there's not enough stock
func TestConsumeStockFEFO_InsufficientStock(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Clean up before and after test
	cleanupTestData(t, db)
	defer cleanupTestData(t, db)

	// Setup test data
	userID := createTestUser(t, db)
	productID := createTestProduct(t, db, userID, "Leche_Insufficient")

	// Create two batches with total of 20 units
	tomorrow := time.Now().AddDate(0, 0, 1)
	batchA := createTestBatch(t, db, productID, userID, 10, "LOTE-A", tomorrow)

	nextMonth := time.Now().AddDate(0, 1, 0)
	batchB := createTestBatch(t, db, productID, userID, 10, "LOTE-B", nextMonth)

	// Record initial quantities
	initialQuantityA := getBatchQuantity(t, db, batchA)
	initialQuantityB := getBatchQuantity(t, db, batchB)

	t.Run("Consume 25 units - Should fail and rollback", func(t *testing.T) {
		// Start a transaction
		ctx := context.Background()
		tx, err := db.Begin(ctx)
		require.NoError(t, err, "Failed to begin transaction")
		defer tx.Rollback(ctx)

		// Try to consume 25 units (we only have 20)
		err = ConsumeStockFEFO(ctx, tx, productID, userID, 25)
		assert.Error(t, err, "ConsumeStockFEFO should return error for insufficient stock")
		assert.ErrorIs(t, err, ErrInsufficientStock, "Error should be ErrInsufficientStock")

		// Explicitly rollback the transaction
		err = tx.Rollback(ctx)
		require.NoError(t, err, "Failed to rollback transaction")

		// Verify stock was NOT modified (rollback successful)
		quantityA := getBatchQuantity(t, db, batchA)
		quantityB := getBatchQuantity(t, db, batchB)

		assert.Equal(t, initialQuantityA, quantityA, "Lote A should remain unchanged after rollback")
		assert.Equal(t, initialQuantityB, quantityB, "Lote B should remain unchanged after rollback")

		t.Logf("✅ Rollback verified: Lote A=%d (expected %d), Lote B=%d (expected %d)",
			quantityA, initialQuantityA, quantityB, initialQuantityB)
	})
}

// TestConsumeStockFEFO_ExactAmount tests consuming the exact available amount
func TestConsumeStockFEFO_ExactAmount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cleanupTestData(t, db)
	defer cleanupTestData(t, db)

	userID := createTestUser(t, db)
	productID := createTestProduct(t, db, userID, "Leche_Exact")

	tomorrow := time.Now().AddDate(0, 0, 1)
	batchA := createTestBatch(t, db, productID, userID, 10, "LOTE-A", tomorrow)

	nextMonth := time.Now().AddDate(0, 1, 0)
	batchB := createTestBatch(t, db, productID, userID, 10, "LOTE-B", nextMonth)

	t.Run("Consume exact 20 units - Should consume all", func(t *testing.T) {
		ctx := context.Background()
		tx, err := db.Begin(ctx)
		require.NoError(t, err, "Failed to begin transaction")
		defer tx.Rollback(ctx)

		err = ConsumeStockFEFO(ctx, tx, productID, userID, 20)
		assert.NoError(t, err, "ConsumeStockFEFO should not return error")

		err = tx.Commit(ctx)
		require.NoError(t, err, "Failed to commit transaction")

		quantityA := getBatchQuantity(t, db, batchA)
		quantityB := getBatchQuantity(t, db, batchB)

		assert.Equal(t, 0, quantityA, "Lote A should be fully consumed")
		assert.Equal(t, 0, quantityB, "Lote B should be fully consumed")

		t.Logf("✅ Exact consumption verified: Lote A=%d, Lote B=%d", quantityA, quantityB)
	})
}

// TestConsumeStockFEFO_SingleBatch tests consuming from a single batch
func TestConsumeStockFEFO_SingleBatch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cleanupTestData(t, db)
	defer cleanupTestData(t, db)

	userID := createTestUser(t, db)
	productID := createTestProduct(t, db, userID, "Leche_Single")

	tomorrow := time.Now().AddDate(0, 0, 1)
	batchA := createTestBatch(t, db, productID, userID, 10, "LOTE-A", tomorrow)

	nextMonth := time.Now().AddDate(0, 1, 0)
	batchB := createTestBatch(t, db, productID, userID, 10, "LOTE-B", nextMonth)

	t.Run("Consume 5 units - Should only affect first batch", func(t *testing.T) {
		ctx := context.Background()
		tx, err := db.Begin(ctx)
		require.NoError(t, err, "Failed to begin transaction")
		defer tx.Rollback(ctx)

		err = ConsumeStockFEFO(ctx, tx, productID, userID, 5)
		assert.NoError(t, err, "ConsumeStockFEFO should not return error")

		err = tx.Commit(ctx)
		require.NoError(t, err, "Failed to commit transaction")

		quantityA := getBatchQuantity(t, db, batchA)
		quantityB := getBatchQuantity(t, db, batchB)

		assert.Equal(t, 5, quantityA, "Lote A should have 5 units remaining")
		assert.Equal(t, 10, quantityB, "Lote B should remain untouched")

		t.Logf("✅ Single batch consumption verified: Lote A=%d, Lote B=%d", quantityA, quantityB)
	})
}

// TestSalesOrderModel_Create_IntegrationFEFO tests the full sales order creation with FEFO
func TestSalesOrderModel_Create_IntegrationFEFO(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cleanupTestData(t, db)
	defer cleanupTestData(t, db)

	userID := createTestUser(t, db)

	// Create customer
	ctx := context.Background()
	const insertCustomer = `
		INSERT INTO customers (name, email, user_id, created_at)
		VALUES ('Test Customer', 'customer@test.com', $1, NOW())
		RETURNING id`
	var customerID int64
	err := db.QueryRow(ctx, insertCustomer, userID).Scan(&customerID)
	require.NoError(t, err, "Failed to create test customer")

	productID := createTestProduct(t, db, userID, "Leche_Integration")

	tomorrow := time.Now().AddDate(0, 0, 1)
	batchA := createTestBatch(t, db, productID, userID, 10, "LOTE-A", tomorrow)

	nextMonth := time.Now().AddDate(0, 1, 0)
	batchB := createTestBatch(t, db, productID, userID, 10, "LOTE-B", nextMonth)

	t.Run("Create sales order with FEFO consumption", func(t *testing.T) {
		model := &SalesOrderModel{DB: db}

		order := &SalesOrder{
			CustomerID:  sql.NullInt64{Int64: customerID, Valid: true},
			OrderDate:   time.Now(),
			Status:      "pending",
			TotalAmount: sql.NullFloat64{Float64: 1500.0, Valid: true},
			UserID:      userID,
		}

		items := []OrderItem{
			{
				ProductID: productID,
				Quantity:  15,
				UnitPrice: 100.0,
			},
		}

		err := model.Create(order, items)
		assert.NoError(t, err, "Create sales order should not return error")

		// Verify FEFO logic was applied
		quantityA := getBatchQuantity(t, db, batchA)
		quantityB := getBatchQuantity(t, db, batchB)

		assert.Equal(t, 0, quantityA, "Lote A should be fully consumed")
		assert.Equal(t, 5, quantityB, "Lote B should have 5 units remaining")

		t.Logf("✅ Sales order integration test passed: Order ID=%d, Lote A=%d, Lote B=%d",
			order.ID, quantityA, quantityB)
	})
}

// TestSalesOrderModel_Create_InsufficientStock tests sales order creation failure
func TestSalesOrderModel_Create_InsufficientStock(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	cleanupTestData(t, db)
	defer cleanupTestData(t, db)

	userID := createTestUser(t, db)

	ctx := context.Background()
	const insertCustomer = `
		INSERT INTO customers (name, email, user_id, created_at)
		VALUES ('Test Customer 2', 'customer2@test.com', $1, NOW())
		RETURNING id`
	var customerID int64
	err := db.QueryRow(ctx, insertCustomer, userID).Scan(&customerID)
	require.NoError(t, err, "Failed to create test customer")

	productID := createTestProduct(t, db, userID, "Leche_Insufficient_Order")

	tomorrow := time.Now().AddDate(0, 0, 1)
	batchA := createTestBatch(t, db, productID, userID, 10, "LOTE-A", tomorrow)

	nextMonth := time.Now().AddDate(0, 1, 0)
	batchB := createTestBatch(t, db, productID, userID, 10, "LOTE-B", nextMonth)

	initialQuantityA := getBatchQuantity(t, db, batchA)
	initialQuantityB := getBatchQuantity(t, db, batchB)

	t.Run("Create sales order with insufficient stock - Should fail", func(t *testing.T) {
		model := &SalesOrderModel{DB: db}

		order := &SalesOrder{
			CustomerID:  sql.NullInt64{Int64: customerID, Valid: true},
			OrderDate:   time.Now(),
			Status:      "pending",
			TotalAmount: sql.NullFloat64{Float64: 2500.0, Valid: true},
			UserID:      userID,
		}

		items := []OrderItem{
			{
				ProductID: productID,
				Quantity:  25, // More than available
				UnitPrice: 100.0,
			},
		}

		err := model.Create(order, items)
		assert.Error(t, err, "Create should return error for insufficient stock")

		// Verify error is InsufficientStockError
		var insufficientStockErr *InsufficientStockError
		assert.ErrorAs(t, err, &insufficientStockErr, "Error should be InsufficientStockError")

		// Verify stock was NOT modified
		quantityA := getBatchQuantity(t, db, batchA)
		quantityB := getBatchQuantity(t, db, batchB)

		assert.Equal(t, initialQuantityA, quantityA, "Lote A should remain unchanged")
		assert.Equal(t, initialQuantityB, quantityB, "Lote B should remain unchanged")

		t.Logf("✅ Sales order rejection verified: Stock unchanged (A=%d, B=%d)", quantityA, quantityB)
	})
}
