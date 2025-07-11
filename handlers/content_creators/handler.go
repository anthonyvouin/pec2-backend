package content_creators

import (
	"errors"
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"
	"pec2-backend/utils"
	mailsmodels "pec2-backend/utils/mails-models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @Summary Apply to become a content creator
// @Description Submit an application to become a content creator
// @Tags content-creators
// @Accept multipart/form-data
// @Produce json
// @Param companyName formData string true "Company name" default(My Creative Company)
// @Param companyType formData string true "Company type" default(SARL)
// @Param siretNumber formData string true "SIRET number" default(12345678901234)
// @Param vatNumber formData string false "VAT number" default(FR12345678901)
// @Param streetAddress formData string true "Street address" default(123 Business Street)
// @Param postalCode formData string true "Postal code" default(75001)
// @Param city formData string true "City" default(Paris)
// @Param country formData string true "Country" default(France)
// @Param iban formData string true "IBAN" default(FR7630006000011234567890189)
// @Param bic formData string true "BIC" default(BNPAFRPP)
// @Param file formData file true "Document proof (PDF, image)"
// @Success 201 {object} map[string]interface{} "message: Application submitted successfully"
// @Failure 400 {object} map[string]interface{} "error: Invalid input"
// @Failure 409 {object} map[string]interface{} "error: Application already exists"
// @Failure 500 {object} map[string]interface{} "error: Error message"
// @Security BearerAuth
// @Router /content-creators [post]
func Apply(c *gin.Context) {
	var contentCreatorInfoCreate models.ContentCreatorInfoCreate

	if err := c.ShouldBind(&contentCreatorInfoCreate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input: " + err.Error(),
		})
		return
	}

	userID := c.MustGet("user_id").(string)

	var user models.User
	if err := db.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error fetching user information",
		})
		return
	}

	var existingApplication models.ContentCreatorInfo
	if err := db.DB.Where("user_id = ?", userID).First(&existingApplication).Error; err == nil {
		if existingApplication.Status == models.ContentCreatorStatusApproved {
			c.JSON(http.StatusConflict, gin.H{
				"error": "You are already a content creator. Please use the update endpoint if you need to modify your information",
			})
			return
		} else if existingApplication.Status == models.ContentCreatorStatusPending {
			c.JSON(http.StatusConflict, gin.H{
				"error": "You have already applied to become a content creator",
			})
			return
		} else if existingApplication.Status == models.ContentCreatorStatusRejected {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Your application was rejected. Please use the update endpoint to resubmit your application",
			})
			return
		}
	}

	var existingSiret models.ContentCreatorInfo
	if err := db.DB.Where("siret_number = ? AND user_id != ?", contentCreatorInfoCreate.SiretNumber, userID).First(&existingSiret).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "This SIRET number is already registered by another content creator",
		})
		return
	}

	isValid, err := utils.VerifySiret(contentCreatorInfoCreate.SiretNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error verifying SIRET number: " + err.Error(),
		})
		return
	}
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid SIRET number: The provided SIRET number does not exist or is not active",
		})
		return
	}

	file, err := c.FormFile("file")
	if err != nil || file == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Document proof is required",
		})
		return
	}

	documentURL, err := utils.UploadImage(file, "content_creator_documents", "document")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error uploading document: " + err.Error(),
		})
		return
	}

	contentCreatorInfo := models.ContentCreatorInfo{
		UserID:           userID,
		CompanyName:      contentCreatorInfoCreate.CompanyName,
		CompanyType:      contentCreatorInfoCreate.CompanyType,
		SiretNumber:      contentCreatorInfoCreate.SiretNumber,
		VatNumber:        contentCreatorInfoCreate.VatNumber,
		StreetAddress:    contentCreatorInfoCreate.StreetAddress,
		PostalCode:       contentCreatorInfoCreate.PostalCode,
		City:             contentCreatorInfoCreate.City,
		Country:          contentCreatorInfoCreate.Country,
		Iban:             contentCreatorInfoCreate.Iban,
		Bic:              contentCreatorInfoCreate.Bic,
		DocumentProofUrl: documentURL,
		Status:           models.ContentCreatorStatusPending,
	}

	result := db.DB.Create(&contentCreatorInfo)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	mailsmodels.ContentCreatorConfirmation(mailsmodels.ContentCreatorConfirmationData{
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Email:         user.Email,
		CompanyName:   contentCreatorInfo.CompanyName,
		CompanyType:   contentCreatorInfo.CompanyType,
		SiretNumber:   contentCreatorInfo.SiretNumber,
		VatNumber:     contentCreatorInfo.VatNumber,
		StreetAddress: contentCreatorInfo.StreetAddress,
		PostalCode:    contentCreatorInfo.PostalCode,
		City:          contentCreatorInfo.City,
		Country:       contentCreatorInfo.Country,
		Iban:          contentCreatorInfo.Iban,
		Bic:           contentCreatorInfo.Bic,
	})

	utils.LogSuccessWithUser(userID, "Content creator application submitted successfully in Apply")
	c.JSON(http.StatusCreated, gin.H{
		"message": "Content creator application submitted successfully",
	})
}

// @Summary Update a content creator application
// @Description Update an existing content creator application (rejected or approved)
// @Tags content-creators
// @Accept multipart/form-data
// @Produce json
// @Param companyName formData string true "Company name" default(My Creative Company)
// @Param companyType formData string true "Company type" default(SARL)
// @Param siretNumber formData string true "SIRET number" default(12345678901234)
// @Param vatNumber formData string false "VAT number" default(FR12345678901)
// @Param streetAddress formData string true "Street address" default(123 Business Street)
// @Param postalCode formData string true "Postal code" default(75001)
// @Param city formData string true "City" default(Paris)
// @Param country formData string true "Country" default(France)
// @Param iban formData string true "IBAN" default(FR7630006000011234567890189)
// @Param bic formData string true "BIC" default(BNPAFRPP)
// @Success 200 {object} map[string]interface{} "message: Application updated successfully"
// @Failure 400 {object} map[string]interface{} "error: Invalid input"
// @Failure 404 {object} map[string]interface{} "error: No application found"
// @Failure 403 {object} map[string]interface{} "error: Application cannot be updated"
// @Failure 500 {object} map[string]interface{} "error: Error message"
// @Security BearerAuth
// @Router /content-creators [put]
func UpdateContentCreatorInfo(c *gin.Context) {
	var contentCreatorInfoCreate models.ContentCreatorInfoCreate

	if err := c.ShouldBind(&contentCreatorInfoCreate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input: " + err.Error(),
		})
		return
	}

	userID := c.MustGet("user_id").(string)

	var user models.User
	if err := db.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error fetching user information",
		})
		return
	}

	var existingApplication models.ContentCreatorInfo
	if err := db.DB.Where("user_id = ?", userID).First(&existingApplication).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No application found for this user",
		})
		return
	}

	if existingApplication.Status != models.ContentCreatorStatusRejected &&
		existingApplication.Status != models.ContentCreatorStatusApproved {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Application cannot be updated. Your application is currently pending",
		})
		return
	}

	if existingApplication.SiretNumber != contentCreatorInfoCreate.SiretNumber {
		var existingSiret models.ContentCreatorInfo
		if err := db.DB.Where("siret_number = ? AND user_id != ?", contentCreatorInfoCreate.SiretNumber, userID).First(&existingSiret).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{
				"error": "This SIRET number is already registered by another content creator",
			})
			return
		}
	}

	isValid, err := utils.VerifySiret(contentCreatorInfoCreate.SiretNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error verifying SIRET number: " + err.Error(),
		})
		return
	}
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid SIRET number: The provided SIRET number does not exist or is not active",
		})
		return
	}

	// file, err := c.FormFile("file")
	// if err != nil || file == nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{
	// 		"error": "Document proof is required",
	// 	})
	// 	return
	// }

	// oldDocumentURL := existingApplication.DocumentProofUrl

	// documentURL, err := utils.UploadImage(file, "content_creator_documents", "document")
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{
	// 		"error": "Error uploading document: " + err.Error(),
	// 	})
	// 	return
	// }

	// if oldDocumentURL != "" {
	// 	if err := utils.DeleteImage(oldDocumentURL); err != nil {
	// 		fmt.Printf("Error deleting old document: %v\n", err)
	// 	}
	// }

	previousStatus := existingApplication.Status

	existingApplication.CompanyName = contentCreatorInfoCreate.CompanyName
	existingApplication.CompanyType = contentCreatorInfoCreate.CompanyType
	existingApplication.SiretNumber = contentCreatorInfoCreate.SiretNumber
	existingApplication.VatNumber = contentCreatorInfoCreate.VatNumber
	existingApplication.StreetAddress = contentCreatorInfoCreate.StreetAddress
	existingApplication.PostalCode = contentCreatorInfoCreate.PostalCode
	existingApplication.City = contentCreatorInfoCreate.City
	existingApplication.Country = contentCreatorInfoCreate.Country
	existingApplication.Iban = contentCreatorInfoCreate.Iban
	existingApplication.Bic = contentCreatorInfoCreate.Bic
	// existingApplication.DocumentProofUrl = documentURL

	if previousStatus == models.ContentCreatorStatusRejected {
		existingApplication.Status = models.ContentCreatorStatusPending
	}

	if err := db.DB.Save(&existingApplication).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	mailsmodels.ContentCreatorUpdate(mailsmodels.ContentCreatorUpdateData{
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Email:         user.Email,
		CompanyName:   existingApplication.CompanyName,
		CompanyType:   existingApplication.CompanyType,
		SiretNumber:   existingApplication.SiretNumber,
		VatNumber:     existingApplication.VatNumber,
		StreetAddress: existingApplication.StreetAddress,
		PostalCode:    existingApplication.PostalCode,
		City:          existingApplication.City,
		Country:       existingApplication.Country,
		Iban:          existingApplication.Iban,
		Bic:           existingApplication.Bic,
	})

	utils.LogSuccessWithUser(userID, "Content creator information updated successfully in UpdateContentCreatorInfo")
	c.JSON(http.StatusOK, gin.H{
		"message": "Content creator information updated successfully",
	})
}

// @Summary Get all content creator applications (Admin)
// @Description Retrieves a list of all content creator applications (Admin access only)
// @Tags content-creators
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.ContentCreatorInfo
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 403 {object} map[string]string "error: Forbidden - Admin access required"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /content-creators/all [get]
func GetAllContentCreators(c *gin.Context) {
	var contentCreators []models.ContentCreatorInfo
	result := db.DB.Order("created_at DESC").Find(&contentCreators)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, contentCreators)
}

// @Summary Update content creator application status (Admin)
// @Description Update the status of a content creator application (Admin access only)
// @Tags content-creators
// @Accept json
// @Produce json
// @Param id path string true "Content Creator Application ID"
// @Param status body models.ContentCreatorStatusUpdate true "New status"
// @Security BearerAuth
// @Success 200 {object} map[string]string "message: Status updated successfully"
// @Failure 400 {object} map[string]string "error: Invalid input"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 403 {object} map[string]string "error: Forbidden - Admin access required"
// @Failure 404 {object} map[string]string "error: Content creator application not found"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /content-creators/{id}/status [put]
func UpdateContentCreatorStatus(c *gin.Context) {
	id := c.Param("id")
	var statusUpdate models.ContentCreatorStatusUpdate

	if err := c.ShouldBindJSON(&statusUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input data: " + err.Error(),
		})
		return
	}

	validStatus := false
	for _, status := range []models.ContentCreatorStatusType{
		models.ContentCreatorStatusPending,
		models.ContentCreatorStatusApproved,
		models.ContentCreatorStatusRejected,
	} {
		if statusUpdate.Status == status {
			validStatus = true
			break
		}
	}

	if !validStatus {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status",
		})
		return
	}

	var contentCreator models.ContentCreatorInfo
	if result := db.DB.First(&contentCreator, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Content creator application not found",
		})
		return
	}

	var user models.User
	if err := db.DB.First(&user, "id = ?", contentCreator.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error fetching user information",
		})
		return
	}

	if result := db.DB.Model(&contentCreator).Update("status", statusUpdate.Status); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	var newRole models.Role
	if statusUpdate.Status == models.ContentCreatorStatusApproved {
		newRole = models.ContentCreator
	} else {
		newRole = models.UserRole
	}

	if result := db.DB.Model(&user).Update("role", newRole); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error updating user role: " + result.Error.Error(),
		})
		return
	}

	mailsmodels.ContentCreatorStatusUpdate(mailsmodels.ContentCreatorStatusUpdateData{
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		CompanyName: contentCreator.CompanyName,
		Status:      statusUpdate.Status,
	})

	utils.LogSuccessWithUser(user.ID, "Content creator status updated successfully in UpdateContentCreatorStatus")
	c.JSON(http.StatusOK, gin.H{
		"message": "Status updated successfully",
	})
}

// @Summary Get general statistiques
// @Description statistic for followers and subscribers for one creator
// @Tags content-creators
// @Accept json
// @Produce json
// @Param isSubscriberSearch query bool true "subscribers stats?"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Statistics successfully retrieved"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /content-creators/stats/creator [get]
func GetCreatorStats(c *gin.Context) {
	isSubscriberSearchStr := c.Query("isSubscriberSearch")
	isSubscriberSearch, err := strconv.ParseBool(isSubscriberSearchStr)
	if err != nil {
		isSubscriberSearch = false
	}
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

	subscribersOrFollowers := getSubscribersOrFollowers(userID, isSubscriberSearch)
	agePercents := getSubscriberAge(subscribersOrFollowers)
	genderPercents := getSubscriberGender(subscribersOrFollowers)
	mostLikedPost := getMostLikedPost(userID, isSubscriberSearch)
	mostCommentedPost := getMostCommentsPost(userID, isSubscriberSearch)
	threeLastPosts := getThreeLastPost(userID, isSubscriberSearch)

	c.JSON(http.StatusOK, gin.H{
		"subscribersOrFollowers": subscribersOrFollowers,
		"subscriberLength":       len(subscribersOrFollowers),
		"gender":                 genderPercents,
		"subscriberAge":          agePercents,
		"mostLikedPost":          mostLikedPost,
		"mostCommentedPost":      mostCommentedPost,
		"threeLastPost":          threeLastPosts,
	})

}

func getSubscriberGender(listUsers []models.LiteUser) map[string]float64 {
	genderCounts := map[string]int{
		"femme": 0,
		"homme": 0,
		"autre": 0,
	}

	for _, user := range listUsers {
		switch user.Sexe {
		case "WOMAN":
			genderCounts["femme"]++
		case "MAN":
			genderCounts["homme"]++
		default:
			genderCounts["autre"]++
		}
	}

	totalCountGender := 0
	for _, count := range genderCounts {
		totalCountGender += count
	}

	genderPercents := map[string]float64{}
	for key, count := range genderCounts {
		if totalCountGender == 0 {
			genderPercents[key] = 0.0
		} else {
			genderPercents[key] = (float64(count) / float64(totalCountGender)) * 100
		}
	}
	return genderPercents
}

func getSubscriberAge(listUsers []models.LiteUser) map[string]float64 {
	subscriberAge := map[string]int{
		"under18":        0,
		"between18And25": 0,
		"between26And40": 0,
		"over40":         0,
	}

	for _, user := range listUsers {
		age := calculateAge(user.BirthDayDate)

		switch {
		case age < 18:
			subscriberAge["under18"]++
		case age <= 25:
			subscriberAge["between18And25"]++
		case age <= 40:
			subscriberAge["between26And40"]++
		default:
			subscriberAge["over40"]++
		}
	}
	totalCountAge := 0
	for _, count := range subscriberAge {
		totalCountAge += count
	}

	agePercents := map[string]float64{}
	for key, count := range subscriberAge {
		if totalCountAge == 0 {
			agePercents[key] = 0.0
		} else {
			agePercents[key] = (float64(count) / float64(totalCountAge)) * 100
		}
	}
	return agePercents
}

func calculateAge(birthDate time.Time) int {
	now := time.Now()
	years := now.Year() - birthDate.Year()
	if now.YearDay() < birthDate.YearDay() {
		years--
	}
	return years
}

func getSubscribersOrFollowers(userID any, isSubscriberSearch bool) []models.LiteUser {

	if isSubscriberSearch {
		var subscriptions []models.Subscription
		errSub := db.DB.
			Where("content_creator_id = ?", userID).
			Preload("User").
			Find(&subscriptions).Error
		if errSub != nil {
			utils.LogError(errSub, "Error when getList subscribers id")
		}

		return getSearchSubscribersUser(subscriptions)

	} else {
		var followers []models.UserFollow
		errSub := db.DB.
			Where("followed_id = ?", userID).
			Preload("Follower").
			Find(&followers).Error
		if errSub != nil {
			utils.LogError(errSub, "Error when getList subscribers id")
		}

		return getSearchFollowersUser(followers)
	}

}

func getSearchSubscribersUser(usersList []models.Subscription) []models.LiteUser {
	var subscribers []models.LiteUser
	seen2 := make(map[string]bool)

	for _, sub := range usersList {
		user := sub.User
		if !seen2[user.ID] {
			seen2[user.ID] = true
			subscribers = append(subscribers, models.LiteUser{
				ID:             user.ID,
				UserName:       user.UserName,
				ProfilePicture: user.ProfilePicture,
				BirthDayDate:   user.BirthDayDate,
				Sexe:           user.Sexe,
			})
		}
	}
	return subscribers
}

func getSearchFollowersUser(usersList []models.UserFollow) []models.LiteUser {
	var followers []models.LiteUser
	seen2 := make(map[string]bool)

	for _, sub := range usersList {
		user := sub.Follower
		if !seen2[user.ID] {
			seen2[user.ID] = true
			followers = append(followers, models.LiteUser{
				ID:             user.ID,
				UserName:       user.UserName,
				ProfilePicture: user.ProfilePicture,
				BirthDayDate:   user.BirthDayDate,
				Sexe:           user.Sexe,
			})
		}
	}
	return followers
}

func getMostLikedPost(userID any, isSubscriberSearch bool) models.MostLikedPost {

	var mostLikedPost models.MostLikedPost

	errPost := db.DB.
		Table("posts").
		Select("posts.name, posts.picture_url, posts.description, COUNT(likes.id) AS like_count").
		Joins("LEFT JOIN likes ON likes.post_id = posts.id").
		Where("posts.user_id = ? AND posts.is_free = ?", userID, !isSubscriberSearch).
		Group("posts.id").
		Order("like_count DESC").
		Limit(1).
		Scan(&mostLikedPost).Error

	if errPost != nil {
		utils.LogError(errPost, "Error when get most liked post")
	}

	return mostLikedPost
}

func getMostCommentsPost(userID any, isSubscriberSearch bool) models.MostCommentPost {

	var mostCommentedPost models.MostCommentPost

	errPost := db.DB.
		Table("posts").
		Select("posts.name, posts.picture_url, posts.description, COUNT(comments.id) AS comment_count").
		Joins("LEFT JOIN comments ON comments.post_id = posts.id::text").
		Where("posts.user_id::text = ? AND posts.is_free = ?", userID, !isSubscriberSearch).
		Group("posts.id").
		Order("comment_count DESC").
		Limit(1).
		Scan(&mostCommentedPost).Error

	if errPost != nil {
		utils.LogError(errPost, "Error when getting most commented post")
	}

	return mostCommentedPost
}

func getThreeLastPost(userID any, isSubscriberSearch bool) []models.LastPost {
	var lastPosts []models.LastPost

	err := db.DB.
		Table("posts").
		Select("name, picture_url").
		Where("user_id = ? AND is_free = ?", userID, !isSubscriberSearch).
		Order("created_at DESC").
		Limit(3).
		Scan(&lastPosts).Error

	if err != nil {
		utils.LogError(err, "Erreur lors de la récupération des derniers posts")
	}

	return lastPosts
}

// @Summary Get advenced statistiques
// @Description statistic for turnover and count subscribers per date
// @Tags content-creators
// @Accept json
// @Produce json
// @Param start query string true "start period"
// @Param end query string true "end period"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Statistics successfully retrieved"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /content-creators/stats-advenced/creator [get]
func GetAdvencedStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	startStr := c.Query("start")
	endStr := c.Query("end")

	start, errStart := time.Parse("2006-01-02", startStr)
	end, errEnd := time.Parse("2006-01-02", endStr)

	if errStart != nil || errEnd != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	dateFormat := "YYYY-MM-DD"
	isGroupedByMonth := false

	if end.Sub(start).Hours() > 31*24 {
		dateFormat = "YYYY-MM"
		isGroupedByMonth = true
	}

	paymentResult := getPaymentsStats(userID, start, end, dateFormat, isGroupedByMonth)
	subscriptionResult := getSubscriptionCounts(userID, start, end, dateFormat, isGroupedByMonth)

	c.JSON(http.StatusOK, gin.H{
		"monthlyRevenue": paymentResult,
		"subscriptions":  subscriptionResult,
	})
}

func getPaymentsStats(userID any, start time.Time, end time.Time, dateFormat string, isGroupedByMonth bool) []models.MonthlyRevenue {
	var rawResults []struct {
		Period string
		Total  int64
	}

	err := db.DB.
		Table("subscription_payments").
		Select("TO_CHAR(subscription_payments.created_at, ?) AS period, SUM(subscription_payments.amount) AS total", dateFormat).
		Joins("JOIN subscriptions ON subscriptions.id = subscription_payments.subscription_id").
		Where("subscriptions.content_creator_id = ? AND subscription_payments.created_at BETWEEN ? AND ?", userID, start, end).
		Group("period").
		Order("period").
		Scan(&rawResults).Error

	if err != nil {
		utils.LogError(err, "Error while fetching advanced stats")
		return nil
	}

	// Convert rawResults to a map for easier lookup
	revenueMap := make(map[string]float64)
	for _, r := range rawResults {
		revenueMap[r.Period] = float64(r.Total) / 100.0
	}

	var results []models.MonthlyRevenue
	current := start

	for !current.After(end) {
		var label string
		if isGroupedByMonth {
			label = current.Format("2006-01")
			current = current.AddDate(0, 1, 0)
		} else {
			label = current.Format("2006-01-02")
			current = current.AddDate(0, 0, 1)
		}

		total := revenueMap[label]
		results = append(results, models.MonthlyRevenue{
			Month: label,
			Total: total,
		})
	}

	return results
}

func getSubscriptionCounts(userID any, start time.Time, end time.Time, dateFormat string, isGroupedByMonth bool) []models.MonthlyRevenue {
	var rawResults []struct {
		Period string
		Count  int64
	}

	err := db.DB.
		Table("subscriptions").
		Select("TO_CHAR(created_at, ?) AS period, COUNT(*) AS count", dateFormat).
		Where("content_creator_id = ? AND created_at BETWEEN ? AND ?", userID, start, end).
		Group("period").
		Order("period").
		Scan(&rawResults).Error

	if err != nil {
		utils.LogError(err, "Error while fetching subscription counts")
		return nil
	}

	subscriptionMap := make(map[string]float64)
	for _, r := range rawResults {
		subscriptionMap[r.Period] = float64(r.Count)
	}

	var results []models.MonthlyRevenue
	current := start

	for !current.After(end) {
		var label string
		if isGroupedByMonth {
			label = current.Format("2006-01")
			current = current.AddDate(0, 1, 0)
		} else {
			label = current.Format("2006-01-02")
			current = current.AddDate(0, 0, 1)
		}

		count := subscriptionMap[label]
		results = append(results, models.MonthlyRevenue{
			Month: label,
			Total: count,
		})
	}

	return results
}

// @Summary Get creator inscription
// @Description Get creator inscription
// @Tags content-creators
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "creator inscription or null"
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /content-creators [get]
func GetCreatorInscription(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.LogError(errors.New("user not auth"), "User not auth")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not auth"})
		return
	}

	var creatorInfo models.ContentCreatorInfo
	err := db.DB.Where("user_id = ?", userID).First(&creatorInfo).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, nil)
			return
		}
		// Erreur DB autre
		utils.LogError(err, "Error when getting inscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error when getting inscription"})
		return
	}

	c.JSON(http.StatusOK, creatorInfo)
}
