package routes

import (
	"pec2-backend/handlers/user_settings"
	"pec2-backend/middleware"

	"github.com/gin-gonic/gin"
)

func UserSettingsRoutes(r *gin.Engine) {
	settingsRoutes := r.Group("/user-settings")
	settingsRoutes.Use(middleware.JWTAuth())
	{
		settingsRoutes.GET("", user_settings.GetUserSettings)
		settingsRoutes.PUT("", user_settings.UpdateUserSettings)
	}
}
