package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/abeselom-personal/go-ai-service/internal/config"
	models "github.com/abeselom-personal/go-ai-service/internal/model"
	"github.com/abeselom-personal/go-ai-service/internal/repository"
	"gorm.io/gorm"
)

type SystemPromptService struct {
	repo *repository.SystemPromptRepo
	db   *gorm.DB
	cfg  *config.Config
}

func NewSystemPromptService(db *gorm.DB, repo *repository.SystemPromptRepo, cfg *config.Config) *SystemPromptService {
	return &SystemPromptService{repo: repo, db: db, cfg: cfg}
}

func hashPrompt(systemPrompt, userPrompt, moduleName string) string {
	raw := systemPrompt + userPrompt + moduleName
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (s *SystemPromptService) Create(ctx context.Context, module, provider, sys, modelname string) (*models.SystemPrompt, error) {

	sp := &models.SystemPrompt{
		ModuleName:   module,
		ModelName:    modelname,
		Provider:     provider,
		SystemPrompt: sys,
	}
	err := s.repo.Create(ctx, sp)
	return sp, err
}

func (s *SystemPromptService) Get(ctx context.Context) ([]models.SystemPrompt, error) {
	return s.repo.List(ctx)
}

func (s *SystemPromptService) GetHash(ctx context.Context, hash string) (*models.SystemPrompt, error) {
	return s.repo.GetByHash(ctx, hash)
}
func (s *SystemPromptService) Update(ctx context.Context, id string, sys, user string) error {
	var sp models.SystemPrompt
	if err := s.db.WithContext(ctx).First(&sp, "id = ?", id).Error; err != nil {
		return err
	}
	sp.SystemPrompt = sys
	return s.repo.Update(ctx, &sp)
}

func (s *SystemPromptService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *SystemPromptService) getActiveProviderAndModel() (*config.ProviderConfig, *config.ModelConfig, error) {
	activeModel := s.cfg.Defaults.Model
	for i := range s.cfg.Defaults.Providers {
		provider := &s.cfg.Defaults.Providers[i]
		for j := range provider.Models {
			model := &provider.Models[j]
			if model.Name == activeModel {
				return provider, model, nil
			}
		}
	}
	return nil, nil, errors.New("active model not found in any provider")
}

func (s *SystemPromptService) SendPrompt(
	ctx context.Context,
	module,
	sys,
	user string,
	bypassCache bool,
) (*models.AIUsageLog, error) {
	hash := hashPrompt(sys, user, module)

	// Check cache first unless bypass is requested
	if !bypassCache {
		cached, err := s.getCachedResponse(ctx, hash)
		if err == nil {
			return cached, nil
		}
	}

	// Proceed with API call
	provider, model, err := s.getActiveProviderAndModel()
	if err != nil {
		return nil, err
	}
	// // Rate limit check
	// if err := s.checkRateLimit(ctx, module, provider.Name); err != nil {
	// 	return nil, err
	// }

	// Make API call
	response, err := s.callAIAPI(ctx, provider, model, sys, user)
	if err != nil {
		return nil, err
	}

	// Store in database
	logEntry := &models.AIUsageLog{
		ModuleName: module,
		Provider:   provider.Name,
		PromptHash: hash,
		Request:    sys + "\n" + user, // Store combined request
		Response:   response,
	}

	if err := s.db.Create(logEntry).Error; err != nil {
		return nil, fmt.Errorf("failed to store response: %v", err)
	}

	return logEntry, nil
}

func (s *SystemPromptService) getCachedResponse(ctx context.Context, hash string) (*models.AIUsageLog, error) {
	var logEntry models.AIUsageLog
	err := s.db.WithContext(ctx).
		Where("prompt_hash = ?", hash).
		Order("used_at DESC").
		First(&logEntry).
		Error

	if err != nil {
		return nil, fmt.Errorf("cache miss: %v", err)
	}
	return &logEntry, nil
}

func (s *SystemPromptService) checkRateLimit(ctx context.Context, module, provider string) error {
	var limit models.RateLimit
	result := s.db.WithContext(ctx).
		Where("module_name = ? AND provider = ?", module, provider).
		First(&limit)

	// Only enforce if rate limit exists
	if result.Error == nil {
		var count int64
		start := time.Now().Add(-time.Duration(limit.PerSeconds) * time.Second)
		err := s.db.Model(&models.AIUsageLog{}).
			Where("module_name = ? AND provider = ? AND used_at >= ?", module, provider, start).
			Count(&count).
			Error

		if err != nil {
			return fmt.Errorf("failed to check usage: %w", err)
		}

		if count >= int64(limit.MaxRequests) {
			return fmt.Errorf("rate limit exceeded for %s/%s (%d requests per %d seconds)",
				module, provider, limit.MaxRequests, limit.PerSeconds)
		}
	}

	return nil
}
func (s *SystemPromptService) callAIAPI(
	ctx context.Context,
	provider *config.ProviderConfig,
	model *config.ModelConfig,
	sys, user string,
) (string, error) {
	// Construct request body using template
	tmpl, err := template.New("request").Parse(model.Config)
	if err != nil {
		return "", fmt.Errorf("invalid request template: %w", err)
	}

	var bodyBuf bytes.Buffer
	err = tmpl.Execute(&bodyBuf, struct {
		SystemPrompt string
		UserPrompt   string
	}{sys, user})
	if err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s%s:generateContent", provider.BaseURL, model.Name)
	if provider.AuthMethod == "query_param" {
		url += fmt.Sprintf("?key=%s", provider.APIKey)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, &bodyBuf)
	if err != nil {
		return "", fmt.Errorf("request creation failed: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if provider.AuthMethod == "header" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", provider.APIKey))
	}

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle errors
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return s.extractResponse(responseBody, model.ResponsePath)
}

func (s *SystemPromptService) extractResponse(body []byte, path string) (string, error) {
	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("invalid JSON response: %w", err)
	}

	// Simple JSON path implementation
	parts := strings.Split(path, ".")
	var current interface{} = result

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		case []interface{}:
			index, err := strconv.Atoi(part)
			if err != nil || index >= len(v) {
				return "", fmt.Errorf("invalid array index in path")
			}
			current = v[index]
		default:
			return "", fmt.Errorf("invalid response structure")
		}
	}

	if str, ok := current.(string); ok {
		return str, nil
	}
	return "", fmt.Errorf("response text not found at path")
}
