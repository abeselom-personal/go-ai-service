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

func setupModelTest(t *testing.T) (ModelRepository, sqlmock.Sqlmock, func()) {
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

	repo := NewModelRepository(gormDB)
	return repo, mock, func() { db.Close() }
}

func TestModelRepository_Create(t *testing.T) {
	repo, mock, teardown := setupModelTest(t)
	defer teardown()

	modelID := uuid.New()
	providerID := uuid.New()
	model := &models.Model{
		ID:         modelID,
		ProviderID: providerID,
		Name:       "test-model",
		Parameters: "{}",
		Config:     "{}",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "models"`).
			WithArgs(
				model.ProviderID.String(), // 1
				model.Name,                // 2
				model.Parameters,          // 3
				model.Config,              // 4
				sqlmock.AnyArg(),          // 5 CreatedAt
				sqlmock.AnyArg(),          // 6 UpdatedAt
				model.ID,                  // 7 ID
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(modelID))
		mock.ExpectCommit()

		err := repo.Create(context.Background(), model)
		assert.NoError(t, err)
	})
}

func TestModelRepository_Create_InvalidProviderID(t *testing.T) {
	repo, _, teardown := setupModelTest(t)
	defer teardown()

	model := &models.Model{
		ID:   uuid.New(),
		Name: "invalid-provider",
	}

	err := repo.Create(context.Background(), model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid provider ID")
}

func TestModelRepository_Create_Duplicate(t *testing.T) {
	repo, mock, teardown := setupModelTest(t)
	defer teardown()

	model := &models.Model{
		ID:         uuid.New(),
		ProviderID: uuid.New(),
		Name:       "duplicate-model",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "models"`).WillReturnError(gorm.ErrDuplicatedKey)
	mock.ExpectRollback()

	err := repo.Create(context.Background(), model)
	assert.Equal(t, ErrDuplicateModel, err)
}

func TestModelRepository_GetByID(t *testing.T) {
	repo, mock, teardown := setupModelTest(t)
	defer teardown()

	modelID := uuid.New()

	t.Run("found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "models" WHERE id = .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(modelID))

		result, err := repo.GetByID(context.Background(), modelID.String())
		assert.NoError(t, err)
		assert.Equal(t, modelID, result.ID)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "models" WHERE id = .*`).
			WillReturnError(gorm.ErrRecordNotFound)

		_, err := repo.GetByID(context.Background(), "nonexistent")
		assert.Equal(t, ErrModelNotFound, err)
	})
}

func TestModelRepository_GetByProviderAndName(t *testing.T) {
	repo, mock, teardown := setupModelTest(t)
	defer teardown()

	providerID := uuid.New()
	modelName := "test-model"

	t.Run("found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "models" WHERE provider_id = .* AND name = .*`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

		_, err := repo.GetByProviderAndName(context.Background(), providerID.String(), modelName)
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "models" WHERE provider_id = .* AND name = .*`).
			WillReturnError(gorm.ErrRecordNotFound)

		_, err := repo.GetByProviderAndName(context.Background(), providerID.String(), "nonexistent")
		assert.Equal(t, ErrModelNotFound, err)
	})
}

func TestModelRepository_ListByProvider(t *testing.T) {
	repo, mock, teardown := setupModelTest(t)
	defer teardown()

	providerID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id"}).AddRow(uuid.New())
		mock.ExpectQuery(`SELECT \* FROM "models" WHERE provider_id = .*`).WillReturnRows(rows)

		result, err := repo.ListByProvider(context.Background(), providerID)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("empty", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM "models" WHERE provider_id = .*`).WillReturnRows(sqlmock.NewRows(nil))
		result, err := repo.ListByProvider(context.Background(), providerID)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestModelRepository_Update(t *testing.T) {
	repo, mock, teardown := setupModelTest(t)
	defer teardown()

	model := &models.Model{
		ID:         uuid.New(),
		ProviderID: uuid.New(),
		Name:       "updated-model",
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "models" SET .*`).
			WithArgs(
				model.ID,
				model.ProviderID,
				model.Name,
				sqlmock.AnyArg(), // updated_at or other nullable fields
				model.ID,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(context.Background(), model)
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "models" SET .*`).
			WithArgs(
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(0, 0)) // zero rows affected

		mock.ExpectRollback()

		err := repo.Update(context.Background(), &models.Model{
			ID: uuid.New(),
			// add all required fields here to match args count if needed
			ProviderID: uuid.New(),
			Name:       "nonexistent",
		})
		assert.Equal(t, ErrModelNotFound, err)
	})
}

func TestModelRepository_Delete(t *testing.T) {
	repo, mock, teardown := setupModelTest(t)
	defer teardown()

	modelID := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "models" WHERE id = .*`).
			WithArgs(modelID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Delete(context.Background(), modelID)
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "models" WHERE id = .*`).
			WithArgs("nonexistent").
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.Delete(context.Background(), "nonexistent")
		assert.Equal(t, ErrModelNotFound, err)
	})
}
