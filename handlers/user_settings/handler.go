package user_settings

import (
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"
	"pec2-backend/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// Struct pour les mises à jour des préférences
type UpdateSettingsRequest struct {
	CommentEnabled *bool `json:"commentEnabled"`
	MessageEnabled *bool `json:"messageEnabled"`
}

// @Summary Get user settings
// @Description Retrieves the settings for the authenticated user
// @Tags user-settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.UserSettings
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 404 {object} map[string]string "error: Settings not found"
// @Failure 500 {object} map[string]string "error: error message"
// @Router /user-settings [get]
func GetUserSettings(c *gin.Context) {
	// Récupérer l'ID de l'utilisateur authentifié depuis le contexte
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}

	// Vérifier si des paramètres existent déjà pour cet utilisateur
	var settings models.UserSettings
	result := db.DB.Where("user_id = ?", userID).First(&settings)

	if result.Error != nil {
		// Si les paramètres n'existent pas, créer des paramètres par défaut
		if result.Error.Error() == "record not found" {
			settings = models.UserSettings{
				UserID:         userID.(string),
				CommentEnabled: true,
				MessageEnabled: true,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			if err := db.DB.Create(&settings).Error; err != nil {
				utils.LogError(err, "Erreur lors de la création des paramètres utilisateur")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création des paramètres"})
				return
			}
		} else {
			utils.LogError(result.Error, "Erreur lors de la récupération des paramètres utilisateur")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des paramètres"})
			return
		}
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
// @Success 200 {object} models.UserSettings
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

	// Chercher les paramètres existants ou créer de nouveaux paramètres
	var settings models.UserSettings
	result := db.DB.Where("user_id = ?", userID).First(&settings)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			// Créer de nouveaux paramètres si aucun n'existe
			settings = models.UserSettings{
				UserID:         userID.(string),
				CommentEnabled: true,
				MessageEnabled: true,
				CreatedAt:      time.Now(),
			}
		} else {
			utils.LogError(result.Error, "Erreur lors de la récupération des paramètres utilisateur")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des paramètres"})
			return
		}
	}
	// Mettre à jour les champs si fournis dans la requête
	if request.CommentEnabled != nil {
		settings.CommentEnabled = *request.CommentEnabled
	}
	if request.MessageEnabled != nil {
		settings.MessageEnabled = *request.MessageEnabled
	}
	settings.UpdatedAt = time.Now()

	// Enregistrer les modifications
	var saveErr error
	if result.Error != nil && result.Error.Error() == "record not found" {
		saveErr = db.DB.Create(&settings).Error
	} else {
		saveErr = db.DB.Save(&settings).Error
	}

	if saveErr != nil {
		utils.LogError(saveErr, "Erreur lors de la sauvegarde des paramètres utilisateur")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la sauvegarde des paramètres"})
		return
	}

	// Synchroniser les paramètres avec le modèle User
	var user models.User
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err == nil {
		// Mettre à jour les champs correspondants dans User
		if request.CommentEnabled != nil {
			user.CommentsEnable = *request.CommentEnabled
		}
		if request.MessageEnabled != nil {
			user.MessageEnable = *request.MessageEnabled
		}
		
		// Sauvegarder les modifications de l'utilisateur
		if err := db.DB.Save(&user).Error; err != nil {
			utils.LogError(err, "Erreur lors de la synchronisation des paramètres avec l'utilisateur")
			// Continuer malgré l'erreur car les paramètres ont été sauvegardés
		}
	}

	c.JSON(http.StatusOK, settings)
}
