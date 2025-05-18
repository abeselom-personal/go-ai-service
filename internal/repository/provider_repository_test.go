package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	models "github.com/abeselom-personal/go-ai-service/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupProviderTest(t *testing.T) (ProviderRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	repo := NewProviderRepository(gormDB)
	return repo, mock, func() { db.Close() }
}

func TestProviderRepository_Create(t *testing.T) {
	repo, mock, teardown := setupProviderTest(t)
	defer teardown()

	providerID := uuid.New()
	provider := &models.Provider{
		ID:        providerID,
		Name:      "test-provider",
		BaseURL:   "https://api.example.com",
		APIKey:    "test-key",
		Default:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "providers"`).
			WithArgs(
				provider.Name,    // 1
				provider.BaseURL, // 2
				provider.APIKey,  // 3
				provider.Default, // 4
				sqlmock.AnyArg(), // 5 CreatedAt
				sqlmock.AnyArg(), // 6 UpdatedAt
				provider.ID,      // 7 ID (GORM places UUID last)
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(providerID))
		mock.ExpectCommit()

		err := repo.Create(context.Background(), provider)
		assert.NoError(t, err)
	})
}

func TestProviderRepository_GetByID(t *testing.T) {
	repo, mock, teardown := setupProviderTest(t)
	defer teardown()

	providerID := uuid.New()
	expectedProvider := &models.Provider{
		ID:   providerID,
		Name: "test-provider",
	}

	t.Run("success", func(t *testing.T) {
		// Mock provider query
		providerRows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(providerID, "test-provider")

		// Mock related models query
		modelRows := sqlmock.NewRows([]string{"id", "name"})

		mock.ExpectQuery(`SELECT \* FROM "providers" WHERE id = \$1 ORDER BY "providers"\."id" LIMIT \$2`).
			WithArgs(providerID, 1).
			WillReturnRows(providerRows)

		mock.ExpectQuery(`SELECT \* FROM "models" WHERE "models"\."provider_id" = \$1`).
			WithArgs(providerID).
			WillReturnRows(modelRows)

		result, err := repo.GetByID(context.Background(), providerID.String())
		assert.NoError(t, err)
		assert.Equal(t, expectedProvider.Name, result.Name)
	})
}

func TestProviderRepository_Create_Invalid(t *testing.T) {
	repo, _, teardown := setupProviderTest(t)
	defer teardown()

	t.Run("empty name", func(t *testing.T) {
		err := repo.Create(context.Background(), &models.Provider{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider name cannot be empty")
	})
}

func TestProviderRepository_GetByName(t *testing.T) {
	repo, mock, teardown := setupProviderTest(t)
	defer teardown()

	t.Run("found", func(t *testing.T) {
		name := "existing-provider"
		mock.ExpectQuery(`SELECT \* FROM "providers" WHERE name = .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(uuid.New(), name))
		mock.ExpectQuery(`SELECT \* FROM "models" WHERE .*`).
			WillReturnRows(sqlmock.NewRows(nil))

		_, err := repo.GetByName(context.Background(), name)
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "providers" WHERE name = .*`).
			WillReturnError(gorm.ErrRecordNotFound)

		_, err := repo.GetByName(context.Background(), "nonexistent")
		assert.Equal(t, ErrProviderNotFound, err)
	})
}

func TestProviderRepository_Update(t *testing.T) {
	repo, mock, teardown := setupProviderTest(t)
	defer teardown()

	provider := &models.Provider{
		ID:      uuid.New(),
		Name:    "updated-provider",
		BaseURL: "",
		APIKey:  "",
		Default: false,
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "providers" SET .*`).
			WithArgs(
				sqlmock.AnyArg(),   // id (updated id, dynamic)
				"updated-provider", // name
				sqlmock.AnyArg(),   // updated_at
				sqlmock.AnyArg(),   // original id for WHERE clause
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(context.Background(), provider)
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "providers" SET .*`).
			WithArgs(
				sqlmock.AnyArg(),   // id
				"updated-provider", // name
				sqlmock.AnyArg(),   // updated_at
				sqlmock.AnyArg(),   // original id
			).
			WillReturnResult(sqlmock.NewResult(0, 0)) // no rows affected
		mock.ExpectRollback() // expect rollback, NOT commit

		err := repo.Update(context.Background(), provider)
		assert.Equal(t, ErrProviderNotFound, err)
	})
}

func TestProviderRepository_Delete(t *testing.T) {
	repo, mock, teardown := setupProviderTest(t)
	defer teardown()

	providerID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "providers" WHERE id = .*`).
			WithArgs(providerID.String()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Delete(context.Background(), providerID.String())
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "providers" WHERE id = .*`).
			WithArgs("nonexistent").
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.Delete(context.Background(), "nonexistent")
		assert.Equal(t, ErrProviderNotFound, err)
	})
}
