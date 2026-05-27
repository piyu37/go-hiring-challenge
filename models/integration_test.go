//go:build integration

package models_test

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/mytheresa/go-hiring-challenge/app/config"
	"github.com/mytheresa/go-hiring-challenge/app/database"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func integrationConfig(t *testing.T) config.Config {
	t.Helper()

	for _, path := range []string{".env", "../.env"} {
		if _, err := os.Stat(path); err == nil {
			require.NoError(t, godotenv.Load(path))
			break
		}
	}

	cfg, err := config.Load()
	require.NoError(t, err)

	return cfg
}

func setupIntegrationDB(t *testing.T) *models.ProductsRepository {
	t.Helper()

	cfg := integrationConfig(t)

	db, closeDB, err := database.Open(context.Background(), cfg.Database)
	require.NoError(t, err)
	t.Cleanup(func() { _ = closeDB() })

	runMigrations(t, db)

	return models.NewProductsRepository(db)
}

func setupIntegrationCategoryRepo(t *testing.T) *models.CategoriesRepository {
	t.Helper()

	cfg := integrationConfig(t)

	db, closeDB, err := database.Open(context.Background(), cfg.Database)
	require.NoError(t, err)
	t.Cleanup(func() { _ = closeDB() })

	runMigrations(t, db)

	return models.NewCategoriesRepository(db)
}

func runMigrations(t *testing.T, db *gorm.DB) {
	t.Helper()

	dir := resolveSQLDir(t)

	files, err := os.ReadDir(dir)
	require.NoError(t, err)

	var sqlFiles []os.DirEntry
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file)
		}
	}

	sort.Slice(sqlFiles, func(i, j int) bool {
		return sqlFiles[i].Name() < sqlFiles[j].Name()
	})

	for _, file := range sqlFiles {
		content, err := os.ReadFile(filepath.Join(dir, file.Name()))
		require.NoError(t, err)
		require.NoError(t, db.Exec(string(content)).Error)
	}
}

func resolveSQLDir(t *testing.T) string {
	t.Helper()

	candidates := []string{}
	if dir := os.Getenv("POSTGRES_SQL_DIR"); dir != "" {
		candidates = append(candidates, dir)
	}
	candidates = append(candidates,
		"sql",
		"../sql",
		filepath.Join(moduleRoot(t), "sql"),
	)

	for _, candidate := range candidates {
		path, err := filepath.Abs(candidate)
		require.NoError(t, err)

		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			return path
		}
	}

	t.Fatal("sql directory not found")
	return ""
}

func moduleRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func TestProductsRepositoryIntegration(t *testing.T) {
	repo := setupIntegrationDB(t)
	ctx := context.Background()

	t.Run("lists products with pagination metadata", func(t *testing.T) {
		products, total, err := repo.List(ctx, models.ProductListFilter{
			Offset: 0,
			Limit:  3,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(8), total)
		assert.Len(t, products, 3)
		assert.NotEmpty(t, products[0].Category.Code)
	})

	t.Run("filters by category", func(t *testing.T) {
		products, total, err := repo.List(ctx, models.ProductListFilter{
			Offset:       0,
			Limit:        10,
			CategoryCode: "clothing",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, products, 3)
		for _, product := range products {
			assert.Equal(t, "clothing", product.Category.Code)
		}
	})

	t.Run("filters by price less than", func(t *testing.T) {
		price := decimal.NewFromFloat(10)
		products, total, err := repo.List(ctx, models.ProductListFilter{
			Offset:        0,
			Limit:         10,
			PriceLessThan: &price,
		})
		require.NoError(t, err)
		assert.Greater(t, total, int64(0))
		for _, product := range products {
			assert.True(t, product.Price.LessThan(price))
		}
	})

	t.Run("loads product with variants and category", func(t *testing.T) {
		product, err := repo.GetByCode(ctx, "PROD001")
		require.NoError(t, err)
		assert.Equal(t, "PROD001", product.Code)
		assert.Equal(t, "clothing", product.Category.Code)
		assert.NotEmpty(t, product.Variants)

		effective := models.EffectiveVariantPrice(product.Variants[1].Price, product.Price)
		assert.True(t, effective.Equal(product.Price))
	})
}

func TestCategoriesRepositoryIntegration(t *testing.T) {
	repo := setupIntegrationCategoryRepo(t)
	ctx := context.Background()

	t.Run("lists seeded categories", func(t *testing.T) {
		categories, err := repo.ListAll(ctx)
		require.NoError(t, err)
		assert.Len(t, categories, 3)
	})

	t.Run("checks category existence", func(t *testing.T) {
		exists, err := repo.ExistsByCode(ctx, "clothing")
		require.NoError(t, err)
		assert.True(t, exists)

		exists, err = repo.ExistsByCode(ctx, "unknown")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}
