package users

import (
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// @Summary Get all users
// @Description Retrieves a list of all users
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "users: array of user objects"
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
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{} "user: user object"
// @Failure 400 {object} map[string]string "error: Invalid user ID"
// @Failure 404 {object} map[string]string "error: User not found"
// @Failure 500 {object} map[string]string "error: error message"
// @Router /users/{id} [get]
func GetUserByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID d'utilisateur invalide"})
		return
	}

	var user models.User
	result := db.DB.First(&user, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		return
	}

	// Ne pas envoyer le mot de passe
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
	if result := db.DB.First(&user, userID); result.Error != nil {
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
