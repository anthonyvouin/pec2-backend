package users

import (
	"errors"
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"
	"pec2-backend/utils"
	mailsmodels "pec2-backend/utils/mails-models"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Struct pour Swagger : demande de code de réinitialisation
// @Description Email pour demander un code de réinitialisation
// @name PasswordResetRequest
// @Param email body string true "Email de l'utilisateur"
type PasswordResetRequest struct {
	Email string `json:"email" example:"utilisateur@exemple.com"`
}

// Struct pour Swagger : confirmation de réinitialisation
// @Description Email, code et nouveau mot de passe pour confirmer la réinitialisation
// @name PasswordResetConfirm
// @Param email body string true "Email de l'utilisateur"
// @Param code body string true "Code reçu par email"
// @Param newPassword body string true "Nouveau mot de passe"
type PasswordResetConfirm struct {
	Email       string `json:"email" example:"utilisateur@exemple.com"`
	Code        string `json:"code" example:"123456"`
	NewPassword string `json:"newPassword" example:"NouveauMotdepasse123"`
}

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
		utils.LogError(result.Error, "Error when retrieving all users in GetAllUsers")
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	for i := range users {
		users[i].Password = ""
	}

	userID, exists := c.Get("user_id")
	if !exists {
		userID = "0"
	}
	utils.LogSuccessWithUser(userID, "List of users retrieved successfully in GetAllUsers")
	c.JSON(http.StatusOK, users)
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
		utils.LogError(errors.New("user_id manquant"), "User not found dans UpdatePassword")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var passwordUpdate models.PasswordUpdate
	if err := c.ShouldBindJSON(&passwordUpdate); err != nil {
		utils.LogError(err, "Error when binding JSON in UpdatePassword")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data: " + err.Error()})
		return
	}

	if len(passwordUpdate.NewPassword) < 6 {
		utils.LogError(errors.New("new password is too short"), "New password is too short in UpdatePassword")
		c.JSON(http.StatusBadRequest, gin.H{"error": "The new password must contain at least 6 characters"})
		return
	}

	if passwordUpdate.OldPassword == passwordUpdate.NewPassword {
		utils.LogError(errors.New("new password is the same as the old password"), "New password is the same as the old password in UpdatePassword")
		c.JSON(http.StatusBadRequest, gin.H{"error": "The new password must be different from the old password"})
		return
	}

	var user models.User
	if result := db.DB.Where("id = ?", userID).First(&user); result.Error != nil {
		utils.LogError(result.Error, "User not found in UpdatePassword")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(passwordUpdate.OldPassword)); err != nil {
		utils.LogError(err, "Incorrect old password in UpdatePassword")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect old password"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordUpdate.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.LogError(err, "Error when hashing the new password in UpdatePassword")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}

	if result := db.DB.Model(&user).Update("password", string(hashedPassword)); result.Error != nil {
		utils.LogError(result.Error, "Error when updating password in UpdatePassword")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating password"})
		return
	}

	utils.LogSuccessWithUser(userID, "Password updated successfully in UpdatePassword")
	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// @Summary Update user profile
// @Description Update the current authenticated user's profile information with optional profile picture
// @Tags users
// @Accept multipart/form-data
// @Produce json
// @Param userName formData string false "UserName"
// @Param firstName formData string false "First name"
// @Param lastName formData string false "Last name"
// @Param bio formData string false "Biography"
// @Param email formData string false "Email address"
// @Param sexe formData string false "Sexe"
// @Param birthDayDate formData string false "BirthDayDate"
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
		utils.LogError(errors.New("user_id manquant"), "User not found in token dans UpdateUserProfile")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in token"})
		return
	}

	var user models.User
	if result := db.DB.Where("id = ?", userID).First(&user); result.Error != nil {
		utils.LogError(result.Error, "User not found in UpdateUserProfile")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var formData models.UserUpdateFormData
	if err := c.ShouldBind(&formData); err != nil {
		utils.LogError(err, "Error when binding form data in UpdateUserProfile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data: " + err.Error()})
		return
	}

	if formData.UserName != "" {
		var existingUser models.User
		if err := db.DB.Where("user_name = ? AND id != ?", formData.UserName, userID).First(&existingUser).Error; err == nil {
			utils.LogError(errors.New("username déjà utilisé"), "Username already taken in UpdateUserProfile")
			c.JSON(http.StatusConflict, gin.H{
				"error": "This username is already taken",
			})
			return
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			utils.LogError(err, "Error when checking the username existence in UpdateUserProfile")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error when checking the username existence",
			})
			return
		}
		user.UserName = formData.UserName
	}
	if formData.Bio != "" {
		user.Bio = formData.Bio
	}
	if formData.Email != "" {
		if !utils.ValidateEmail(formData.Email) {
			utils.LogError(errors.New("format email invalide"), "Invalid email format in UpdateUserProfile")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
			return
		}
		user.Email = formData.Email
	}

	if formData.FirstName != "" {
		user.FirstName = formData.FirstName
	}
	if formData.LastName != "" {
		user.LastName = formData.LastName
	}

	file, err := c.FormFile("profilePicture")
	if err == nil && file != nil {
		oldImageURL := user.ProfilePicture

		imageURL, err := utils.UploadImage(file, "profile_pictures", "profile")
		if err != nil {
			utils.LogError(err, "Error when uploading profile picture in UpdateUserProfile")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error uploading profile picture: " + err.Error()})
			return
		}

		user.ProfilePicture = imageURL

		if oldImageURL != "" {
			_ = utils.DeleteImage(oldImageURL)
		}
	}

	if result := db.DB.Save(&user); result.Error != nil {
		utils.LogError(result.Error, "Error when saving user profile in UpdateUserProfile")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating profile: " + result.Error.Error()})
		return
	}

	user.Password = ""

	utils.LogSuccessWithUser(userID, "User profile updated successfully in UpdateUserProfile")
	c.JSON(http.StatusOK, user)
}

// @Summary Get user profile
// @Description Get the current authenticated user's profile information
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "user: user object"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 404 {object} map[string]string "error: User not found"
// @Failure 500 {object} map[string]string "error: Error retrieving profile"
// @Router /users/profile [get]
func GetUserProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.LogError(errors.New("user_id manquant"), "User not found in token dans GetUserProfile")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in token"})
		return
	}

	var user models.User
	result := db.DB.First(&user, "id = ?", userID)
	if result.Error != nil {
		utils.LogError(result.Error, "User not found in GetUserProfile")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Password = ""

	utils.LogSuccessWithUser(userID, "User profile retrieved successfully in GetUserProfile")
	c.JSON(http.StatusOK, user)
}

func GetUserByUsername(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		utils.LogError(errors.New("username manquant"), "Username parameter is missing in GetUserByUsername")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username parameter is required"})
		return
	}

	var searchUser models.User
	result := db.DB.Where("user_name = ?", username).First(&searchUser)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			utils.LogError(result.Error, "User not found in GetUserByUsername")
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		utils.LogError(result.Error, "Error when retrieving user by username in GetUserByUsername")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user: " + result.Error.Error()})
		return
	}

	searchUser.Password = ""

	userID, exists := c.Get("user_id")
	if !exists {
		userID = "0"
	}

	var subs []models.Subscription
	today := time.Now()

	resultSubscription := db.DB.
		Where("user_id = ? AND content_creator_id = ? AND end_date > ?", userID, searchUser.ID, today).
		Find(&subs)

	if resultSubscription.Error != nil {

	}

	var lastActive *models.Subscription
	var lastCanceled *models.Subscription

	for i := range subs {
		sub := &subs[i]

		if sub.Status == "ACTIVE" {
			if lastActive == nil || sub.EndDate.After(*lastActive.EndDate) {
				lastActive = sub
			}
		} else if sub.Status == "CANCELED" {
			if lastCanceled == nil || sub.EndDate.After(*lastCanceled.EndDate) {
				lastCanceled = sub
			}
		}
	}

	var lastSubscription *models.Subscription
	if lastActive != nil {
		lastSubscription = lastActive
	} else if lastCanceled != nil {
		lastSubscription = lastCanceled
	} else {
		lastSubscription = nil
	}

	isSubscriber := len(subs) > 0

	utils.LogSuccessWithUser(userID, "User retrieved successfully by username in GetUserByUsername")
	c.JSON(http.StatusOK, gin.H{
		"user":                     searchUser,
		"isSubscriberToSearchUser": isSubscriber,
		"canceledSubscription": func() *bool {
			if lastSubscription == nil {
				return nil
			}
			b := lastSubscription.Status == models.SubscriptionCanceled
			return &b
		}(),
		"subscriberUntil": func() *time.Time {
			if lastSubscription == nil {
				return nil
			}
			return lastSubscription.EndDate
		}(),
	})
}

// @Summary Get user statistics (Admin)
// @Description Get count of new users by day between start and end dates
// @Tags users
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "total: total number of users, daily_data: array of daily user registration data"
// @Failure 400 {object} map[string]string "error: Invalid date parameters"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: Error retrieving statistics"
// @Router /users/statistics [get]
func GetUserStatistics(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		utils.LogError(nil, "start_date or end_date missing in GetUserStatistics")
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required (format YYYY-MM-DD)"})
		return
	}

	// Correction du format de date pour supporter le format YYYY-MM-DD
	var startDate, endDate time.Time
	var err error

	// Essayer différents formats de date
	formats := []string{"2006-01-02", "2006-01-02T15:04:05Z07:00", "2006-01-02T15:04:05Z"}
	startDateParsed := false

	for _, format := range formats {
		startDate, err = time.Parse(format, startDateStr)
		if err == nil {
			startDateParsed = true
			break
		}
	}

	if !startDateParsed {
		utils.LogError(err, "Invalid start_date format in GetUserStatistics: "+startDateStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format (YYYY-MM-DD)"})
		return
	}

	endDateParsed := false
	for _, format := range formats {
		endDate, err = time.Parse(format, endDateStr)
		if err == nil {
			endDateParsed = true
			break
		}
	}

	if !endDateParsed {
		utils.LogError(err, "Invalid end_date format in GetUserStatistics: "+endDateStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format (YYYY-MM-DD)"})
		return
	}
	if endDate.Before(startDate) {
		utils.LogError(nil, "end_date before start_date in GetUserStatistics")
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_date must be after start_date"})
		return
	}

	var total int64
	err = db.DB.Model(&models.User{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate.Add(24*time.Hour)).
		Count(&total).Error
	if err != nil {
		utils.LogError(err, "Error calculating total user count in GetUserStatistics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error calculating total user count"})
		return
	}

	type DailyData struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}

	var dailyData []DailyData
	err = db.DB.Model(&models.User{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ? AND created_at <= ?", startDate, endDate.Add(24*time.Hour)).
		Group("DATE(created_at)").
		Order("date").
		Scan(&dailyData).Error
	if err != nil {
		utils.LogError(err, "Error fetching daily user data in GetUserStatistics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching daily user data"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		userID = "0"
	}
	utils.LogSuccessWithUser(userID, "User statistics retrieved successfully in GetUserStatistics")
	c.JSON(http.StatusOK, gin.H{
		"total":      total,
		"daily_data": dailyData,
	})
}

// @Summary Get user role statistics (Admin)
// @Description Get the count of users for each role (Admin access only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]int "Role counts"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: Internal server error"
// @Router /users/stats/roles [get]
func GetUserRoleStats(c *gin.Context) {
	var roleCounts = make(map[string]int)

	roleCounts["ADMIN"] = 0
	roleCounts["CONTENT_CREATOR"] = 0
	roleCounts["USER"] = 0

	for _, role := range []models.Role{models.AdminRole, models.ContentCreator, models.UserRole} {
		var count int64
		if err := db.DB.Model(&models.User{}).Where("role = ?", role).Count(&count).Error; err != nil {
			utils.LogError(err, "Error when counting users by role in GetUserRoleStats")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting users by role"})
			return
		}
		roleCounts[string(role)] = int(count)
	}

	userID, exists := c.Get("user_id")
	if !exists {
		userID = "0"
	}
	utils.LogSuccessWithUser(userID, "User role statistics retrieved successfully in GetUserRoleStats")
	c.JSON(http.StatusOK, roleCounts)
}

// @Summary Get user gender statistics (Admin)
// @Description Get the count of users for each gender (Admin access only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]int "Gender counts"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: Internal server error"
// @Router /users/stats/gender [get]
func GetUserGenderStats(c *gin.Context) {
	var genderCounts = make(map[string]int)

	genderCounts["MAN"] = 0
	genderCounts["WOMAN"] = 0
	genderCounts["OTHER"] = 0

	for _, sexe := range []models.Sexe{models.Male, models.Female, models.Other} {
		var count int64
		if err := db.DB.Model(&models.User{}).Where("sexe = ?", sexe).Count(&count).Error; err != nil {
			utils.LogError(err, "Error when counting users by gender in GetUserGenderStats")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting users by gender"})
			return
		}
		genderCounts[string(sexe)] = int(count)
	}

	userID, exists := c.Get("user_id")
	if !exists {
		userID = "0"
	}
	utils.LogSuccessWithUser(userID, "User gender statistics retrieved successfully in GetUserGenderStats")
	c.JSON(http.StatusOK, genderCounts)
}

// @Summary send a reset password code
// @Description send a reset password code to the email if the user exists
// @Tags users
// @Accept json
// @Produce json
// @Param data body PasswordResetRequest true "Email of the user"
// @Success 200 {object} map[string]string "message: Code sent"
// @Failure 404 {object} map[string]string "error: User not found"
// @Router /users/password/reset/request [post]
func RequestPasswordReset(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogError(err, "Error when binding JSON in RequestPasswordReset")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email"})
		return
	}

	if !utils.ValidateEmail(req.Email) {
		utils.LogError(errors.New("format email invalide"), "Invalid email format in RequestPasswordReset")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	var user models.User
	if err := db.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		utils.LogError(err, "User not found in RequestPasswordReset")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	code := utils.GenerateCode()
	end := time.Now().Add(15 * time.Minute)

	user.ResetPasswordCode = code
	user.ResetPasswordCodeEnd = end

	if err := db.DB.Save(&user).Error; err != nil {
		utils.LogError(err, "Error when saving the reset code in RequestPasswordReset")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving the user"})
		return
	}

	mailsmodels.SendResetPasswordCode(user.Email, code)
	userID, exists := c.Get("user_id")
	if !exists {
		userID = "0"
	}
	utils.LogSuccessWithUser(userID, "Reset password code sent successfully in RequestPasswordReset")
	c.JSON(http.StatusOK, gin.H{"message": "Code sent to the email if it exists in our database."})
}

// @Summary Reset password with a code
// @Description Change the password if the code is correct and not expired
// @Tags users
// @Accept json
// @Produce json
// @Param data body PasswordResetConfirm true "Email, code, new password"
// @Success 200 {object} map[string]string "message: Password reset"
// @Failure 400 {object} map[string]string "error: Invalid data or code incorrect/expired"
// @Router /users/password/reset/confirm [post]
func ConfirmPasswordReset(c *gin.Context) {
	var req struct {
		Email       string `json:"email" binding:"required,email"`
		Code        string `json:"code" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogError(err, "Error when binding JSON in ConfirmPasswordReset")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if !utils.ValidateEmail(req.Email) {
		utils.LogError(errors.New("format email invalide"), "Invalid email format in ConfirmPasswordReset")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	var user models.User
	if err := db.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		utils.LogError(err, "User not found in ConfirmPasswordReset")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if user.ResetPasswordCode != req.Code || time.Now().After(user.ResetPasswordCodeEnd) {
		utils.LogError(errors.New("code invalide ou expiré"), "Invalid or expired reset code in ConfirmPasswordReset")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid code or expired"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.LogError(err, "Error when hashing the new password in ConfirmPasswordReset")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing the password"})
		return
	}

	user.Password = string(hashedPassword)
	user.ResetPasswordCode = ""
	user.ResetPasswordCodeEnd = time.Time{}

	if err := db.DB.Save(&user).Error; err != nil {
		utils.LogError(err, "Error when saving the new password in ConfirmPasswordReset")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving the user"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		userID = "0"
	}
	utils.LogSuccessWithUser(userID, "Password reset successfully in ConfirmPasswordReset")
	c.JSON(http.StatusOK, gin.H{"message": "Password reset"})
}

// @Summary Follow a user
// @Description Allows an authenticated user to follow another user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "ID of the user to follow"
// @Security BearerAuth
// @Success 200 {object} map[string]string "message: Follow successful"
// @Failure 400 {object} map[string]string "error: Bad request"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 404 {object} map[string]string "error: User not found"
// @Failure 409 {object} map[string]string "error: Already followed"
// @Failure 500 {object} map[string]string "error: Server error"
// @Router /users/{id}/follow [post]
func FollowUser(c *gin.Context) {
	followerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	followedID := c.Param("id")
	if followedID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User to follow ID is missing"})
		return
	}
	if followerID == followedID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot follow yourself"})
		return
	}

	var user models.User
	if err := db.DB.First(&user, "id = ?", followedID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User to follow not found"})
		return
	}

	var existingFollow models.UserFollow
	err := db.DB.Where("follower_id = ? AND followed_id = ?", followerID, followedID).First(&existingFollow).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "You already follow this user"})
		return
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error when checking the follow"})
		return
	}

	follow := models.UserFollow{
		FollowerID: followerID.(string),
		FollowedID: followedID,
	}
	if err := db.DB.Create(&follow).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error when following the user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User followed successfully"})
}

// @Summary Unfollow a user
// @Description Allows an authenticated user to unfollow another user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "ID of the user to unfollow"
// @Security BearerAuth
// @Success 200 {object} map[string]string "message: Unfollow successful"
// @Failure 400 {object} map[string]string "error: Bad request"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 404 {object} map[string]string "error: User not found or follow does not exist"
// @Failure 500 {object} map[string]string "error: Server error"
// @Router /users/{id}/follow [delete]
func UnfollowUser(c *gin.Context) {
	followerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	followedID := c.Param("id")
	if followedID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User to unfollow ID is missing"})
		return
	}
	if followerID == followedID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot unfollow yourself"})
		return
	}

	var follow models.UserFollow
	err := db.DB.Where("follower_id = ? AND followed_id = ?", followerID, followedID).First(&follow).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Follow relationship does not exist"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error when checking the follow"})
		return
	}

	if err := db.DB.Delete(&follow).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error when unfollowing the user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User unfollowed successfully"})
}

// @Summary List of users followed
// @Description List all users that the authenticated user follows
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.User "List of users followed"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: Server error"
// @Router /users/followings [get]
func GetMyFollowings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var follows []models.UserFollow
	if err := db.DB.Where("follower_id = ?", userID).Find(&follows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching followings"})
		return
	}

	var users []models.User
	for _, follow := range follows {
		var user models.User
		if err := db.DB.Where("id = ?", follow.FollowedID).First(&user).Error; err == nil {
			user.Password = ""
			users = append(users, user)
		}
	}

	c.JSON(http.StatusOK, users)
}

// @Summary Liste des followers
// @Description Liste tous les utilisateurs qui suivent l'utilisateur authentifié
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.User "List of followers"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: Server error"
// @Router /users/followers [get]
func GetMyFollowers(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var follows []models.UserFollow
	if err := db.DB.Where("followed_id = ?", userID).Find(&follows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching followers"})
		return
	}

	var users []models.User
	for _, follow := range follows {
		var user models.User
		if err := db.DB.Where("id = ?", follow.FollowerID).First(&user).Error; err == nil {
			user.Password = ""
			users = append(users, user)
		}
	}

	c.JSON(http.StatusOK, users)
}

// @Summary Number of followers and followings
// @Description Return the number of followers and followings for a given user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID of the user"
// @Success 200 {object} map[string]interface{} "userId, followers, followings"
// @Failure 404 {object} map[string]string "error: User not found"
// @Failure 500 {object} map[string]string "error: Server error"
// @Router /users/id/{id}/follow-counts [get]
func GetUserFollowCounts(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	var user models.User
	if err := db.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var followersCount int64
	var followingsCount int64
	db.DB.Model(&models.UserFollow{}).Where("followed_id = ?", userID).Count(&followersCount)
	db.DB.Model(&models.UserFollow{}).Where("follower_id = ?", userID).Count(&followingsCount)

	c.JSON(http.StatusOK, gin.H{
		"userId":     userID,
		"followers":  followersCount,
		"followings": followingsCount,
	})
}

func GetCreatorStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var user models.User
	if err := db.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var subscriptions []models.Subscription
	errSub := db.DB.
		Where("content_creator_id = ?", userID).
		Preload("User").
		Find(&subscriptions).Error
	if errSub != nil {
		utils.LogError(errSub, "Error when getList subscribers id")
	}

	var userSubscribers []models.User
	seen := make(map[string]bool)

	for _, sub := range subscriptions {
		user := sub.User
		if !seen[user.ID] {
			userSubscribers = append(userSubscribers, user)
			seen[user.ID] = true
		}
	}

	var subscribers []models.LiteUser
	seen2 := make(map[string]bool)

	for _, sub := range subscriptions {
		user := sub.User
		if !seen2[user.ID] {
			seen2[user.ID] = true
			subscribers = append(subscribers, models.LiteUser{
				ID:             user.ID,
				UserName:       user.UserName,
				ProfilePicture: user.ProfilePicture,
			})
		}
	}

	genderCounts := map[string]int{
		"femme": 0,
		"homme": 0,
		"autre": 0,
	}

	for _, user := range userSubscribers {
		switch user.Sexe {
		case "WOMAN":
			genderCounts["femme"]++
		case "MAN":
			genderCounts["homme"]++
		default:
			genderCounts["autre"]++
		}
	}

	total := 0
	for _, count := range genderCounts {
		total += count
	}

	genderPercents := map[string]float64{}
	for key, count := range genderCounts {
		if total == 0 {
			genderPercents[key] = 0.0
		} else {
			genderPercents[key] = (float64(count) / float64(total)) * 100
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"subscribers":      subscribers,
		"subscriberLength": len(subscribers),
		"gender":           genderPercents,
	})

}
