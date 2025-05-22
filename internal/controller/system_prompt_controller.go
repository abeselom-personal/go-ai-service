// controller/system_prompt_controller.go
package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/abeselom-personal/go-ai-service/internal/service"
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
		ModuleName   string `json:"module_name" binding:"required"`
		SystemPrompt string `json:"system_prompt" binding:"required"`
		UserPrompt   string `json:"user_prompt" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get cache control parameter
	bypassCache, _ := strconv.ParseBool(ctx.Query("cache"))

	response, err := c.svc.SendPrompt(
		ctx,
		req.ModuleName,
		req.SystemPrompt,
		req.UserPrompt,
		bypassCache,
	)

	if err != nil {
		ctx.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"response":  response.Response,
		"cached":    !bypassCache && time.Since(response.UsedAt) > time.Second,
		"timestamp": response.UsedAt,
	})
}
