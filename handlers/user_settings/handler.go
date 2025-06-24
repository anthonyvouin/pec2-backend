package user_settings

import (
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"
	"pec2-backend/utils"

	"github.com/gin-gonic/gin"
)

// Struct pour les mises à jour des préférences
type UpdateSettingsRequest struct {
	CommentEnabled *bool `json:"commentEnabled"`
	MessageEnabled *bool `json:"messageEnabled"`
	SubscriptionEnabled *bool `json:"subscriptionEnabled"`
}

// Structure pour uniformiser la réponse
type UserSettingsResponse struct {
	CommentEnabled bool `json:"commentEnabled"`
	MessageEnabled bool `json:"messageEnabled"`
	SubscriptionEnabled bool `json:"subscriptionEnabled"`
}

// @Summary Get user settings
// @Description Retrieves the settings for the authenticated user
// @Tags user-settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserSettingsResponse
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 404 {object} map[string]string "error: User not found"
// @Failure 500 {object} map[string]string "error: error message"
// @Router /user-settings [get]
func GetUserSettings(c *gin.Context) {
	// Récupérer l'ID de l'utilisateur authentifié depuis le contexte
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	result := db.DB.Where("id = ?", userID).First(&user)

	if result.Error != nil {
		utils.LogError(result.Error, "Erreur lors de la récupération de l'utilisateur")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération de l'utilisateur"})
		return
	}

	// Créer la réponse avec les paramètres de l'utilisateur
	settings := UserSettingsResponse{
		CommentEnabled: user.CommentsEnable,
		MessageEnabled: user.MessageEnable,
		SubscriptionEnabled: user.SubscriptionEnable,
	}

	c.JSON(http.StatusOK, settings)
}

// @Summary Update user settings
// @Description Updates the settings for the authenticated user
// @Tags user-settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param settings body UpdateSettingsRequest true "Settings to update"
// @Success 200 {object} UserSettingsResponse
// @Failure 400 {object} map[string]string "error: Bad request"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: error message"
// @Router /user-settings [put]
func UpdateUserSettings(c *gin.Context) {
	// Récupérer l'ID de l'utilisateur authentifié depuis le contexte
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}

	// Valider les données de la requête
	var request UpdateSettingsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.LogError(err, "Données de mise à jour des paramètres invalides")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides"})
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	result := db.DB.Where("id = ?", userID).First(&user)

	if result.Error != nil {
		utils.LogError(result.Error, "Erreur lors de la récupération de l'utilisateur")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération de l'utilisateur"})
		return
	}

	// Mettre à jour les champs si fournis dans la requête
	if request.CommentEnabled != nil {
		user.CommentsEnable = *request.CommentEnabled
	}
	if request.MessageEnabled != nil {
		user.MessageEnable = *request.MessageEnabled
	}
	if request.SubscriptionEnabled != nil {
		user.SubscriptionEnable = *request.SubscriptionEnabled
	}

	// Enregistrer les modifications
	if err := db.DB.Save(&user).Error; err != nil {
		utils.LogError(err, "Erreur lors de la sauvegarde des paramètres utilisateur")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la sauvegarde des paramètres"})
		return
	}

	// Préparer la réponse
	response := UserSettingsResponse{
		CommentEnabled:      user.CommentsEnable,
		MessageEnabled:      user.MessageEnable,
		SubscriptionEnabled: user.SubscriptionEnable,
	}

	c.JSON(http.StatusOK, response)
}
