// routes/routes.go
package routes

import (
	"html/template"
	"net/http"

	"github.com/abeselom-personal/go-ai-service/internal/config"
	"github.com/abeselom-personal/go-ai-service/internal/controller"
	"github.com/abeselom-personal/go-ai-service/internal/repository"
	"github.com/abeselom-personal/go-ai-service/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	repo := repository.NewSystemPromptRepo(db)
	svc := service.NewSystemPromptService(db, repo, cfg)
	ctrl := controller.NewSystemPromptController(svc)

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	r.SetHTMLTemplate(tmpl)

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Title": "AI System Prompts Manager",
		})
	})

	api := r.Group("/api/system-prompts")
	{
		api.POST("/", ctrl.Create)
		api.GET("/", ctrl.Get)
		api.PUT("/:id", ctrl.Update)
		api.DELETE("/:id", ctrl.Delete)
		api.POST("/send", ctrl.Send)
	}

}
