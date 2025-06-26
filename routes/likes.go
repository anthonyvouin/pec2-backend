package routes

import (
	"pec2-backend/handlers/posts/likes"
	"pec2-backend/middleware"

	"github.com/gin-gonic/gin"
)

func LikesRoutes(r *gin.Engine) {
	likesRoutes := r.Group("/likes")
	likesRoutes.Use(middleware.JWTAuth(), middleware.AdminAuth())
	{
		likesRoutes.GET("/statistics", likes.GetLikesStatistics)
	}
}
