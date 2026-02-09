package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

func LLM(r *gin.Engine) {
	llmGroup := r.Group("/api/v1/llm")
	llmGroup.Use(middlewares.AuthToken)
	{
		llmGroup.POST("/organize-nodes", middlewares.DemoModeRestrict, api.LLMOrganizeNodes)
		llmGroup.POST("/generate-rules", middlewares.DemoModeRestrict, api.LLMGenerateRules)
	}
}
