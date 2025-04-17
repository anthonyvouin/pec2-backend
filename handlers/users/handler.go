package users

import (
	"log"
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"
	"pec2-backend/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// @Summary Get all users (Admin)
// @Description Retrieves a list of all users (Admin access only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "users: array of user objects"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 403 {object} map[string]string "error: Forbidden - Admin access required"
// @Failure 500 {object} map[string]string "error: error message"
// @Router /users [get]
func GetAllUsers(c *gin.Context) {
	var users []models.User
	result := db.DB.Order("created_at DESC").Find(&users)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	for i := range users {
		users[i].Password = ""
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// @Summary Get user by ID
// @Description Retrieves a user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} map[string]interface{} "user: user object"
// @Failure 400 {object} map[string]string "error: Invalid user ID"
// @Failure 404 {object} map[string]string "error: User not found"
// @Failure 500 {object} map[string]string "error: error message"
// @Router /users/{id} [get]
func GetUserByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID d'utilisateur invalide"})
		return
	}

	var user models.User
	result := db.DB.Where("id = ?", id).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		return
	}

	user.Password = ""

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// @Summary Update user password
// @Description Update user's password by verifying the old password and setting a new one
// @Tags users
// @Accept json
// @Produce json
// @Param password body models.PasswordUpdate true "Password update information"
// @Security BearerAuth
// @Success 200 {object} map[string]string "message: Password updated successfully"
// @Failure 400 {object} map[string]string "error: Invalid request"
// @Failure 401 {object} map[string]string "error: Invalid old password"
// @Failure 404 {object} map[string]string "error: User not found"
// @Failure 500 {object} map[string]string "error: Error updating password"
// @Router /users/password [put]
func UpdatePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var passwordUpdate models.PasswordUpdate
	if err := c.ShouldBindJSON(&passwordUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data: " + err.Error()})
		return
	}

	if len(passwordUpdate.NewPassword) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The new password must contain at least 6 characters"})
		return
	}

	var user models.User
	if result := db.DB.Where("id = ?", userID).First(&user); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(passwordUpdate.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect old password"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordUpdate.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}

	if result := db.DB.Model(&user).Update("password", string(hashedPassword)); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// @Summary Update user profile
// @Description Update the current authenticated user's profile information with optional profile picture
// @Tags users
// @Accept multipart/form-data
// @Produce json
// @Param username formData string false "Username"
// @Param bio formData string false "Biography"
// @Param email formData string false "Email address"
// @Param profilePicture formData file false "Profile picture image file"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "message: Profile updated successfully, user: updated user object"
// @Failure 400 {object} map[string]string "error: Invalid request data"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 404 {object} map[string]string "error: User not found"
// @Failure 500 {object} map[string]string "error: Error updating profile"
// @Router /users/profile [put]
func UpdateUserProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in token"})
		return
	}
	log.Printf("Mise à jour du profil pour l'utilisateur ID: %v", userID)

	var user models.User
	if result := db.DB.Where("id = ?", userID).First(&user); result.Error != nil {
		log.Printf("Utilisateur non trouvé: %v", result.Error)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	log.Printf("Utilisateur trouvé: %s (ID: %s)", user.Email, user.ID)
	log.Printf("Valeur initiale de ProfilePicture: %s", user.ProfilePicture)

	var formData models.UserUpdateFormData
	if err := c.ShouldBind(&formData); err != nil {
		log.Printf("Erreur de binding du formulaire: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data: " + err.Error()})
		return
	}
	log.Printf("Données du formulaire reçues: UserName=%s, Bio=%s, Email=%s",
		formData.UserName, formData.Bio, formData.Email)

	if formData.UserName != "" {
		user.UserName = formData.UserName
	}
	if formData.Bio != "" {
		user.Bio = formData.Bio
	}
	if formData.Email != "" {
		if !utils.ValidateEmail(formData.Email) {
			log.Printf("Format d'email invalide: %s", formData.Email)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
			return
		}
		user.Email = formData.Email
	}

	// Vérifier si un fichier image a été envoyé
	file, err := c.FormFile("profilePicture")
	if err != nil {
		log.Printf("Aucun fichier image trouvé ou erreur: %v", err)
	} else if file != nil {
		log.Printf("Fichier image reçu: %s, taille: %d", file.Filename, file.Size)

		// Si une image a été téléchargée
		imageURL, err := utils.UploadProfilePicture(file)
		if err != nil {
			log.Printf("Erreur lors de l'upload sur Cloudinary: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error uploading profile picture: " + err.Error()})
			return
		}
		log.Printf("Image téléchargée sur Cloudinary avec succès, URL: %s", imageURL)
		user.ProfilePicture = imageURL
		log.Printf("URL d'image mise à jour dans l'objet utilisateur: %s", user.ProfilePicture)
	}

	// Sauvegarder les modifications
	log.Printf("Avant sauvegarde - ProfilePicture: %s", user.ProfilePicture)

	if result := db.DB.Save(&user); result.Error != nil {
		log.Printf("Erreur lors de la sauvegarde en BDD: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating profile: " + result.Error.Error()})
		return
	}
	log.Printf("Profil utilisateur mis à jour avec succès en BDD")

	// Vérifier après sauvegarde
	var updatedUser models.User
	if result := db.DB.Where("id = ?", userID).First(&updatedUser); result.Error != nil {
		log.Printf("Erreur lors de la récupération de l'utilisateur après mise à jour: %v", result.Error)
	} else {
		log.Printf("Après sauvegarde - ProfilePicture: %s", updatedUser.ProfilePicture)
	}

	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    user,
	})
}
