package server

import (
	"job-tracker/internal/handler"
	"job-tracker/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	r *gin.Engine,
	authHandler *handler.AuthHandler,
	appHandler *handler.ApplicationHandler,
	jobHandler *handler.JobHandler,
	resumeHandler *handler.ResumeHandler,
	statusHandler *handler.StatusHandler,
	notionHandler *handler.NotionHandler,
	jwtSecret string,
) {
	r.Static("/uploads", "./uploads")

	api := r.Group("/api")

	// Public auth routes
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/logout", authHandler.Logout)
	}

	// Public Notion OAuth callback
	api.GET("/notion/callback", notionHandler.Callback)

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.Auth(jwtSecret))
	{
		protected.GET("/user/profile", authHandler.Profile)

		protected.GET("/statuses", statusHandler.GetAll)

		applications := protected.Group("/applications")
		{
			applications.POST("", appHandler.Create)
			applications.GET("", appHandler.GetAll)
			applications.GET("/:id", appHandler.GetByID)
			applications.PUT("/:id", appHandler.Update)
			applications.DELETE("/:id", appHandler.Delete)
		}

		jobs := protected.Group("/jobs")
		{
			jobs.POST("", jobHandler.Create)
			jobs.GET("", jobHandler.GetAll)
			jobs.GET("/:id", jobHandler.GetByID)
			jobs.PUT("/:id", jobHandler.Update)
			jobs.DELETE("/:id", jobHandler.Delete)
		}

		resumes := protected.Group("/resumes")
		{
			resumes.POST("", resumeHandler.Create)
			resumes.GET("", resumeHandler.GetAll)
			resumes.GET("/:id", resumeHandler.GetByID)
			resumes.PUT("/:id", resumeHandler.Update)
			resumes.DELETE("/:id", resumeHandler.Delete)
		}

		notion := protected.Group("/notion")
		{
			notion.GET("/connect", notionHandler.Connect)
			notion.GET("/status", notionHandler.Status)
			notion.POST("/configure", notionHandler.Configure)
			notion.GET("/databases", notionHandler.ListDatabases)
			notion.POST("/sync", notionHandler.Sync)
			notion.DELETE("/disconnect", notionHandler.Disconnect)
			notion.GET("/debug/pages", notionHandler.DebugPages)
		}
	}
}
