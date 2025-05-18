// internal/repository/system_prompt_repository_test.go
package repository_test

import (
	"context"
	"testing"

	models "github.com/abeselom-personal/go-ai-service/internal/model"
	"github.com/abeselom-personal/go-ai-service/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)
	err = db.AutoMigrate(&models.SystemPrompt{})
	assert.NoError(t, err)
	return db
}

func TestSystemPromptRepo_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewSystemPromptRepo(db)
	ctx := context.Background()

	// Create
	prompt := &models.SystemPrompt{
		ID:           uuid.New(),
		ModuleName:   "test-module",
		Provider:     "ChatGPT",
		SystemPrompt: "You are a helper.",
		UserPrompt:   "What's the weather?",
		PromptHash:   "abc123",
	}
	err := repo.Create(ctx, prompt)
	assert.NoError(t, err)

	// Get
	fetched, err := repo.GetByHash(ctx, "abc123")
	assert.NoError(t, err)
	assert.Equal(t, prompt.ID, fetched.ID)

	// Update
	fetched.SystemPrompt = "You are a very helpful assistant."
	err = repo.Update(ctx, fetched)
	assert.NoError(t, err)

	updated, err := repo.GetByHash(ctx, "abc123")
	assert.NoError(t, err)
	assert.Equal(t, "You are a very helpful assistant.", updated.SystemPrompt)

	// List
	all, err := repo.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, all, 1)

	// Delete
	err = repo.Delete(ctx, prompt.ID.String())
	assert.NoError(t, err)

	_, err = repo.GetByHash(ctx, "abc123")
	assert.Error(t, err)
}
