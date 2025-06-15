// controller/system_prompt_controller.go
package controller

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/abeselom-personal/go-ai-service/internal/service"
	"github.com/abeselom-personal/go-ai-service/internal/utils"
	"github.com/gin-gonic/gin"
)

type SystemPromptController struct {
	svc *service.SystemPromptService
}

func NewSystemPromptController(svc *service.SystemPromptService) *SystemPromptController {
	return &SystemPromptController{svc}
}

func (c *SystemPromptController) Create(ctx *gin.Context) {
	var req struct {
		ModuleName   string `json:"module_name" binding:"required"`
		ModelName    string `json:"model_name" binding:"required"`
		Provider     string `json:"provider" binding:"required"`
		SystemPrompt string `json:"system_prompt" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	prompt, err := c.svc.Create(ctx, req.ModuleName, req.Provider, req.SystemPrompt, req.ModelName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, prompt)
}

func (c *SystemPromptController) Get(ctx *gin.Context) {
	prompt, err := c.svc.Get(ctx)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Prompt not found"})
		return
	}
	ctx.JSON(http.StatusOK, prompt)
}

func (c *SystemPromptController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var req struct {
		SystemPrompt string `json:"system_prompt" binding:"required"`
		UserPrompt   string `json:"user_prompt" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.svc.Update(ctx, id, req.SystemPrompt, req.UserPrompt); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusOK)
}

func (c *SystemPromptController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.svc.Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *SystemPromptController) Send(ctx *gin.Context) {
	var req struct {
		ModuleName   string `json:"module_name"`
		SystemPrompt string `json:"system_prompt" binding:"required"`
		UserPrompt   string `json:"user_prompt" binding:"required"`
		Type         string `json:"type"`
	}

	log.Println("[DEBUG] Binding JSON request body")
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("[ERROR] Failed to bind JSON: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	log.Printf("[DEBUG] Request received: %+v", req)

	if req.Type == "json" {
		log.Println("[DEBUG] JSON type detected, appending formatting rules")
		expectedStructure := getExpectedStructure(req.SystemPrompt)
		jsonRules := fmt.Sprintf(`
		STRICT JSON FORMATTING RULES:
		1. Return ONLY valid JSON with double quotes
		2. No extra text outside the JSON structure
		3. Escape special characters in strings
		4. Validate structure matches: %s
		5. Ensure all brackets/braces are properly closed

		BAD EXAMPLE:
		'''json
		{ options: ['Option 1] }
		'''

		GOOD EXAMPLE:
		%s`, expectedStructure, formatExample(expectedStructure))

		req.SystemPrompt = fmt.Sprintf("%s\n%s", req.SystemPrompt, jsonRules)
		log.Printf("[DEBUG] Updated system prompt with rules: %s", req.SystemPrompt)
	}

	bypassCache, _ := strconv.ParseBool(ctx.Query("cache"))
	clientIP := ctx.ClientIP()
	log.Printf("[DEBUG] Bypass cache flag: %v", bypassCache)
	log.Printf("[DEBUG] UserPrompt: %s", req.UserPrompt)
	log.Printf("[DEBUG] SystemPrompt: %s", req.SystemPrompt)

	log.Println("[DEBUG] Sending prompt to prompt service")
	response, err := c.svc.SendPrompt(
		ctx,
		req.ModuleName,
		req.SystemPrompt,
		req.UserPrompt,
		bypassCache,
		clientIP,
	)

	if err != nil {
		log.Printf("[ERROR] Prompt service error: %v", err)
		ctx.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[DEBUG] AI raw response: %s", response.Response)

	if req.Type == "json" {
		var result map[string]any
		response.Response = strings.ReplaceAll(response.Response, "\n", "")
		log.Println("[DEBUG] Attempting to extract JSON from AI response")
		if err := utils.ExtractJSON(response.Response, &result); err != nil {
			log.Printf("[ERROR] JSON extraction failed: %v", err)
			log.Printf("[DEBUG] Raw AI response: %s", response.Response)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":            "AI response format validation failed",
				"details":          "The AI returned malformed JSON. Please check the prompt instructions.",
				"validation_guide": "JSON must be properly formatted with matching braces and quotes",
				"raw_response":     response.Response,
			})
			return
		}

		if len(result) == 0 {
			log.Println("[WARN] JSON extraction succeeded but result is empty")
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":    "Empty JSON response from AI",
				"solution": "Check if the AI prompt specifies required JSON fields",
			})
			return
		}

		log.Printf("[DEBUG] Valid JSON extracted: %+v", result)
		ctx.JSON(http.StatusOK, result)
		return
	}

	cached := !bypassCache && time.Since(response.UsedAt) > time.Second
	log.Printf("[DEBUG] Final response ready | Cached: %v | TimeUsed: %s", cached, response.UsedAt.Format(time.RFC3339))

	ctx.JSON(http.StatusOK, gin.H{
		"response":  response.Response,
		"cached":    cached,
		"timestamp": response.UsedAt,
	})
}

func getExpectedStructure(prompt string) string {
	log.Printf("[DEBUG] Determining expected structure from prompt: %s", prompt)
	if strings.Contains(prompt, "options") {
		return `{ "options": [], "correct_index": 0 }`
	}
	if strings.Contains(prompt, "question") {
		return `{ "question": "" }`
	}
	return "{}"
}

func formatExample(structure string) string {
	log.Printf("[DEBUG] Formatting example for structure: %s", structure)
	switch {
	case strings.Contains(structure, "options"):
		return `{
  "options": [
    "Mitosis", 
    "Meiosis", 
    "Fertilization", 
    "Osmosis"
  ],
  "correct_index": 1
}`
	case strings.Contains(structure, "question"):
		return `{
  "question": "Which process results in four genetically unique haploid cells?"
}`
	default:
		return `{
  "key": "value"
}`
	}
}
