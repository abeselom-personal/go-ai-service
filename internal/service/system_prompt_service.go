package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type SystemPromptService struct {
	repo  *repository.SystemPromptRepo
	db    *gorm.DB
	cfg   *config.Config
	redis *redis.Client
}

func NewSystemPromptService(db *gorm.DB, repo *repository.SystemPromptRepo, cfg *config.Config, redis *redis.Client) *SystemPromptService {
	return &SystemPromptService{repo: repo, db: db, cfg: cfg, redis: redis}
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
	sp.UserPrompt = user
	sp.PromptHash = hashPrompt(sys, user, sp.ModuleName)
	return s.repo.Update(ctx, &sp)
}

func (s *SystemPromptService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *SystemPromptService) getActiveProviderAndModel() (*config.ProviderConfig, *config.ModelConfig, error) {
	activeProvider := s.cfg.Defaults.Provider
	activeModel := s.cfg.Defaults.Model
	for i := range s.cfg.Defaults.Providers {
		provider := &s.cfg.Defaults.Providers[i]
		if provider.Name != activeProvider {
			continue
		}
		for j := range provider.Models {
			model := &provider.Models[j]
			if model.Name == activeModel {
				return provider, model, nil
			}
		}
		return nil, nil, fmt.Errorf("model %q not found in provider %q", activeModel, activeProvider)
	}
	return nil, nil, fmt.Errorf("provider %q not found in config", activeProvider)
}

func (s *SystemPromptService) SendPrompt(
	ctx context.Context,
	module,
	sys,
	user string,
	bypassCache bool,
	clientIP string,
) (*models.AIUsageLog, error) {
	hash := hashPrompt(sys, user, module)

	if err := s.checkGlobalRateLimit(ctx, clientIP); err != nil {
		return nil, err
	}
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

	// Make API call
	response, err := s.callAIAPI(ctx, provider, model, sys, user)
	if err != nil {
		return nil, err
	}

	logEntry := &models.AIUsageLog{
		ModuleName: module,
		Provider:   provider.Name,
		PromptHash: hash,
		Request:    sys + "\n" + user,
		Response:   response,
	}

	if err := s.db.WithContext(ctx).Create(logEntry).Error; err != nil {
		return nil, fmt.Errorf("failed to store response: %w", err)
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
		return nil, fmt.Errorf("cache miss: %w", err)
	}
	return &logEntry, nil
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

	// Build URL based on provider type
	url := provider.BaseURL
	if strings.Contains(provider.BaseURL, "%s") {
		url = fmt.Sprintf(provider.BaseURL, model.Name)
	} else {
		url = strings.TrimSuffix(provider.BaseURL, "/") + "/" + model.Name
	}
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
	var result any
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("invalid JSON response: %w", err)
	}

	// Simple JSON path implementation
	parts := strings.Split(path, ".")
	var current any = result

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			current = v[part]
		case []any:
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
func (s *SystemPromptService) checkGlobalRateLimit(ctx context.Context, clientIP string) error {
	if !s.cfg.RateLimit.Enabled {
		return nil
	}

	for _, ip := range s.cfg.RateLimit.IPWhitelist {
		if ip == clientIP {
			return nil
		}
	}

	window, err := time.ParseDuration(s.cfg.RateLimit.Window)
	if err != nil {
		return fmt.Errorf("invalid rate limit window: %w", err)
	}

	key := fmt.Sprintf("rl:global:%s", clientIP)
	count, err := s.redis.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("rate limiter unavailable: %w", err)
	}
	if count == 1 {
		s.redis.Expire(ctx, key, window)
	}

	if count > int64(s.cfg.RateLimit.Requests) {
		return fmt.Errorf("global rate limit exceeded")
	}
	return nil
}
