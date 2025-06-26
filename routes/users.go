package routes

import (
	"pec2-backend/handlers/users"
	"pec2-backend/middleware"

	"github.com/gin-gonic/gin"
)

func UsersRoutes(r *gin.Engine) {

	// Route accessible sans authentification
	//r.GET("/users/:id", users.GetUserByID)

	userRoutes := r.Group("/users")
	userRoutes.POST("/password/reset/request", users.RequestPasswordReset)
	userRoutes.POST("/password/reset/confirm", users.ConfirmPasswordReset)

	userRoutes.Use(middleware.JWTAuth())
	{
		// Routes accessibles uniquement aux administrateurs
		userRoutes.GET("", users.GetAllUsers)
		userRoutes.GET("/statistics", middleware.AdminAuth(), users.GetUserStatistics)
		userRoutes.GET("/stats/roles", middleware.AdminAuth(), users.GetUserRoleStats)
		userRoutes.GET("/stats/gender", middleware.AdminAuth(), users.GetUserGenderStats)

		// Routes accessibles à tout utilisateur authentifié
		userRoutes.PUT("/password", users.UpdatePassword)
		userRoutes.PUT("/profile", users.UpdateUserProfile)
		userRoutes.GET("/profile", users.GetUserProfile)
		userRoutes.GET("/:username", users.GetUserByUsername)
		userRoutes.POST(":id/follow", users.FollowUser)
		userRoutes.DELETE(":id/follow", users.UnfollowUser)
		userRoutes.GET("/followings", users.GetMyFollowings)
		userRoutes.GET("/followers", users.GetMyFollowers)
		userRoutes.GET("/id/:id/follow-counts", users.GetUserFollowCounts)
		userRoutes.GET("/stats/creator", users.GetCreatorStats)

	}
}
