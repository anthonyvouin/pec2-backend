package posts

import (
	"encoding/json"
	"fmt"
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"
	"pec2-backend/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// @Summary Create a new post
// @Description Create a new post with the provided information
// @Tags posts
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "Post name"
// @Param description formData string false "Post description"
// @Param isFree formData boolean false "Is the post free"
// @Param enable formData boolean false "Is the post enabled"
// @Param categories formData []string false "Category IDs"
// @Param file formData file false "Post picture"
// @Security BearerAuth
// @Success 201 {object} models.Post
// @Failure 400 {object} map[string]string "error: Invalid input"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /posts [post]
func CreatePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.LogError(nil, "User not found in token in CreatePost")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in token"})
		return
	}

	name := c.Request.FormValue("name")
	if name == "" {
		utils.LogError(nil, "Name is required in CreatePost")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	isFreeStr := c.Request.FormValue("isFree")
	var isFree bool
	switch isFreeStr {
	case "true":
		isFree = true
	case "false":
		isFree = false
	default:
		isFree = false
	}

	categoriesStr := c.Request.FormValue("categories")
	fmt.Println("Categories received in CreatePost:", categoriesStr)
	var categoryIDs []string
	if categoriesStr != "" {
		if err := json.Unmarshal([]byte(categoriesStr), &categoryIDs); err != nil {
			categoryIDs = []string{}
			cleanStr := strings.Trim(categoriesStr, "[]")
			for _, id := range strings.Split(cleanStr, ",") {
				trimmedID := strings.TrimSpace(id)
				if trimmedID != "" {
					categoryIDs = append(categoryIDs, trimmedID)
				}
			}
		}
	}
	description := c.Request.FormValue("description")

	post := models.Post{
		UserID:      userID.(string),
		Name:        name,
		Description: description,
		IsFree:      isFree,
		Enable:      true,
	}

	file, err := c.FormFile("postPicture")
	if err == nil && file != nil {
		imageURL, err := utils.UploadImage(file, "post_pictures", "post")
		if err != nil {
			utils.LogError(err, "Error uploading picture in CreatePost")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error uploading picture: " + err.Error()})
			return
		}
		post.PictureURL = imageURL
	} else {
		utils.LogError(nil, "Picture is required in CreatePost")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Picture is required"})
		return
	}

	if len(categoryIDs) > 0 {
		var categories []models.Category
		if err := db.DB.Where("id IN ?", categoryIDs).Find(&categories).Error; err != nil {
			utils.LogError(err, "Error finding categories in CreatePost")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding categories: " + err.Error()})
			return
		}

		if len(categories) == 0 {
			utils.LogError(nil, "No valid categories found in CreatePost")
			c.JSON(http.StatusBadRequest, gin.H{"error": "No valid categories found"})
			return
		}

		post.Categories = categories
	}

	if err := db.DB.Create(&post).Error; err != nil {
		utils.LogError(err, "Error creating post in CreatePost")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating post: " + err.Error()})
		return
	}

	//! C'est à moitié useless, mais c'est pour renvoyer les catégories sinon je les voient pas dans la réponse
	if err := db.DB.Preload("Categories").Where("id = ?", post.ID).First(&post).Error; err != nil {
		utils.LogError(err, "Error retrieving created post in CreatePost")
		fmt.Println("Error retrieving created post:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving created post: " + err.Error()})
		return
	}

	utils.LogSuccessWithUser(userID, "Post created successfully in CreatePost")
	c.JSON(http.StatusCreated, post)
}

// @Summary Get all posts
// @Description Retrieve all posts with optional filtering and pagination
// @Tags posts
// @Produce json
// @Param isFree query boolean false "Filter by free posts"
// @Param userIs query boolean false "Filter by user"
// @Param homeFeed query boolean false "Filter by current user following"
// @Param subscriptionFeed query boolean false "Filter by current user active subscriptions"
// @Param categories query []string false "Filter by category IDs (can provide multiple)"
// @Param limit query integer false "Number of items per page (default: 10)"
// @Param page query integer false "Page number (default: 1)"
// @Success 200 {object} map[string]interface{} "posts and pagination info"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /posts [get]
func GetAllPosts(c *gin.Context) {
	userID, exists := c.Get("user_id")
	var posts []models.Post
	query := db.DB.Preload("Categories").Order("created_at DESC")

	if isFree := c.Query("isFree"); isFree != "" {
		query = query.Where("is_free = ?", isFree == "true")
	}

	if userIs := c.Query("userIs"); userIs != "" {
		query = query.Where("user_id = ?", userIs)
	}

	if c.Query("homeFeed") == "true" {
		var userFollow []models.UserFollow
		errUserFollow := db.DB.
			Where("follower_id = ?", userID).Find(&userFollow).Error
		if errUserFollow != nil {
			utils.LogError(errUserFollow, "Error when getList userFollow id")
		}

		if len(userFollow) > 0 {
			var followedIDs []string
			for _, follow := range userFollow {
				if follow.FollowedID != "" {
					followedIDs = append(followedIDs, follow.FollowedID)
				}
			}

			if len(followedIDs) > 0 {
				query = query.Where("user_id IN ?", followedIDs)
			} else {
				query = query.Where("1 = 0")
			}
		} else {
			query = query.Where("1 = 0")
		}
	}
	// Filtrer par abonnements actifs de l'utilisateur
	if c.Query("subscriptionFeed") == "true" && exists {
		utils.LogSuccess("Filtering by subscriptions for user: " + userID.(string))
		var subscriptions []models.Subscription
		if err := db.DB.Where("user_id = ? AND (status = ? OR end_date > ?)",
			userID, models.SubscriptionActive, time.Now()).Find(&subscriptions).Error; err != nil {
			utils.LogError(err, "Error finding user subscriptions in GetAllPosts")
		}

		utils.LogSuccess(fmt.Sprintf("Found %d active subscriptions for user %s", len(subscriptions), userID))

		if len(subscriptions) > 0 {
			var creatorIDs []string
			for _, sub := range subscriptions {
				if sub.ContentCreatorID != "" {
					creatorIDs = append(creatorIDs, sub.ContentCreatorID)
					utils.LogSuccess(fmt.Sprintf("User %s is subscribed to creator %s", userID, sub.ContentCreatorID))
				}
			}

			if len(creatorIDs) > 0 {
				query = query.Where("user_id IN ? AND is_free = ?", creatorIDs, false)
				utils.LogSuccess(fmt.Sprintf("Filtering for %d creators with paid posts", len(creatorIDs)))
			} else {
				// Si aucun abonnement actif, ne retourner aucun post
				query = query.Where("1 = 0")
				utils.LogSuccess("No valid creator IDs found, returning no posts")
			}
		} else {
			// Si aucun abonnement actif, ne retourner aucun post
			query = query.Where("1 = 0")
			utils.LogSuccess("No active subscriptions found, returning no posts")
		}
	}

	// Afficher le user qui a créé le post
	query = query.Preload("User")

	// Filtre par catégories (multiple)
	print("Categories received in GetAllPosts:", c.QueryArray("categories"))
	if categories := c.QueryArray("categories"); len(categories) > 0 {
		query = query.Joins("JOIN post_categories ON posts.id = post_categories.post_id").
			Where("post_categories.category_id IN (?)", categories).
			Group("posts.id")
	}

	// Exclure les posts reportés par l'utilisateur connecté
	if exists && userID != nil {
		var reportedPostIds []string
		if err := db.DB.Model(&models.Report{}).
			Where("reported_by = ?", userID).
			Pluck("post_id", &reportedPostIds).Error; err == nil && len(reportedPostIds) > 0 {
			query = query.Where("posts.id NOT IN (?)", reportedPostIds)
			utils.LogSuccess("Filtered out posts reported by user " + userID.(string))
		}
	}

	// Compter le nombre total de posts pour la pagination
	var total int64
	if err := query.Model(&models.Post{}).Count(&total).Error; err != nil {
		utils.LogError(err, "Error counting posts in GetAllPosts")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting posts: " + err.Error()})
		return
	}

	// Pagination
	limit := 10
	if limitParam := c.Query("limit"); limitParam != "" {
		fmt.Sscanf(limitParam, "%d", &limit)
		if limit <= 0 {
			limit = 10
		}
	}

	page := 1
	if pageParam := c.Query("page"); pageParam != "" {
		fmt.Sscanf(pageParam, "%d", &page)
		if page <= 0 {
			page = 1
		}
	}

	offset := (page - 1) * limit
	query = query.Limit(limit).Offset(offset)
	if err := query.Find(&posts).Error; err != nil {
		utils.LogError(err, "Error retrieving posts in GetAllPosts")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving posts: " + err.Error()})
		return
	}

	var response []models.PostResponse = make([]models.PostResponse, 0, len(posts))
	for _, post := range posts {
		// Compter le nombre de likes
		var likesCount int64
		db.DB.Model(&models.Like{}).Where("post_id = ?", post.ID).Count(&likesCount)
		// Compter le nombre de commentaires
		var commentsCount int64
		db.DB.Model(&models.Comment{}).Where("post_id = ?", post.ID).Count(&commentsCount)

		// Compter le nombre de reports
		var reportsCount int64
		db.DB.Model(&models.Report{}).Where("post_id = ?", post.ID).Count(&reportsCount)

		// Vérifier si l'utilisateur actuel a liké ce post
		var isLikedByUser bool
		if exists && userID != nil {
			var likeCount int64
			db.DB.Model(&models.Like{}).Where("post_id = ? AND user_id = ?", post.ID, userID).Count(&likeCount)
			isLikedByUser = likeCount > 0
		}

		// Créer la réponse pour ce post
		postResponse := models.PostResponse{
			ID: post.ID, Name: post.Name, Description: post.Description, PictureURL: post.PictureURL,
			IsFree:     post.IsFree,
			Enable:     post.Enable,
			Categories: post.Categories,
			CreatedAt:  post.CreatedAt,
			UpdatedAt:  post.UpdatedAt,
			User: models.UserInfo{
				ID:             post.User.ID,
				UserName:       post.User.UserName,
				ProfilePicture: post.User.ProfilePicture,
			},
			LikesCount:     int(likesCount),
			CommentsCount:  int(commentsCount),
			ReportsCount:   int(reportsCount),
			CommentEnabled: post.User.CommentsEnable,
			MessageEnabled: post.User.MessageEnable,
			IsLikedByUser:  isLikedByUser,
		}

		response = append(response, postResponse)
	}

	if !exists {
		userID = "0"
	}
	utils.LogSuccessWithUser(userID, "Posts retrieved successfully in GetAllPosts")
	utils.LogSuccess("Posts retrieved successfully in GetAllPosts")

	// Renvoyer les posts avec les informations de pagination
	c.JSON(http.StatusOK, gin.H{
		"posts": response,
		"pagination": gin.H{
			"total":       total,
			"limit":       limit,
			"page":        page,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// @Summary Get a post by ID
// @Description Retrieve a post by its ID
// @Tags posts
// @Produce json
// @Param id path string true "Post ID"
// @Success 200 {object} models.PostResponse
// @Failure 404 {object} map[string]string "error: Post not found"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /posts/{id} [get]
func GetPostByID(c *gin.Context) {
	var post models.Post
	postID := c.Param("id")

	// Récupérer l'ID utilisateur s'il est connecté
	userID, exists := c.Get("user_id")

	// Vérifier si l'utilisateur a reporté ce post
	var userHasReported bool = false
	var reportCount int64
	if exists && userID != nil {
		if err := db.DB.Model(&models.Report{}).
			Where("post_id = ? AND reported_by = ?", postID, userID).
			Count(&reportCount).Error; err == nil && reportCount > 0 {
			userHasReported = true
			utils.LogSuccess("User " + userID.(string) + " has reported post " + postID)
		}
	}

	// Si l'utilisateur a reporté ce post et qu'il n'est pas l'auteur ou un admin, renvoyer une erreur 404
	if userHasReported && exists && userID != nil {
		var isAuthorOrAdmin bool = false
		var userRole string
		roleInterface, roleExists := c.Get("user_role")
		if roleExists {
			userRole = roleInterface.(string)
		}

		// Vérifier si l'utilisateur est l'auteur du post ou un admin
		if err := db.DB.Model(&models.Post{}).
			Where("id = ? AND user_id = ?", postID, userID).
			Count(&reportCount).Error; err == nil && reportCount > 0 {
			isAuthorOrAdmin = true
		} else if userRole == string(models.AdminRole) {
			isAuthorOrAdmin = true
		}

		// Si l'utilisateur n'est ni l'auteur, ni un admin, renvoyer 404
		if !isAuthorOrAdmin {
			utils.LogError(nil, "Post reported by user, access denied in GetPostByID")
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
	}

	if err := db.DB.Preload("Categories").Preload("User").First(&post, "id = ?", postID).Error; err != nil {
		utils.LogError(err, "Post not found in GetPostByID")
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Compter le nombre de likes
	var likesCount int64
	db.DB.Model(&models.Like{}).Where("post_id = ?", post.ID).Count(&likesCount)

	// Compter le nombre de commentaires
	var commentsCount int64
	db.DB.Model(&models.Comment{}).Where("post_id = ?", post.ID).Count(&commentsCount)
	// Compter le nombre de reports
	var reportsCount int64
	db.DB.Model(&models.Report{}).Where("post_id = ?", post.ID).Count(&reportsCount)

	// Vérifier si l'utilisateur actuel a liké ce post
	var isLikedByUser bool
	if exists && userID != nil {
		var likeCount int64
		db.DB.Model(&models.Like{}).Where("post_id = ? AND user_id = ?", post.ID, userID).Count(&likeCount)
		isLikedByUser = likeCount > 0
	}

	// Créer la réponse pour ce post
	postResponse := models.PostResponse{
		ID: post.ID, Name: post.Name, Description: post.Description, PictureURL: post.PictureURL,
		IsFree:     post.IsFree,
		Enable:     post.Enable,
		Categories: post.Categories,
		CreatedAt:  post.CreatedAt,
		UpdatedAt:  post.UpdatedAt,
		User: models.UserInfo{
			ID:             post.User.ID,
			UserName:       post.User.UserName,
			ProfilePicture: post.User.ProfilePicture,
		},
		LikesCount:     int(likesCount),
		CommentsCount:  int(commentsCount),
		ReportsCount:   int(reportsCount),
		CommentEnabled: post.User.CommentsEnable,
		MessageEnabled: post.User.MessageEnable,
		IsLikedByUser:  isLikedByUser,
	}

	utils.LogSuccess("Post retrieved successfully in GetPostByID")
	c.JSON(http.StatusOK, postResponse)
}

// @Summary Update a post
// @Description Update a post with the provided information
// @Tags posts
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Post ID"
// @Param name formData string false "Post name"
// @Param description formData string false "Post description"
// @Param isFree formData boolean false "Is the post free"
// @Param enable formData boolean false "Is the post enabled"
// @Param categories formData []string false "Category IDs"
// @Param file formData file false "Post picture"
// @Security BearerAuth
// @Success 200 {object} models.Post
// @Failure 400 {object} map[string]string "error: Invalid input"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 404 {object} map[string]string "error: Post not found"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /posts/{id} [put]
func UpdatePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.LogError(nil, "User not found in token in UpdatePost")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in token"})
		return
	}

	var post models.Post
	postID := c.Param("id")

	if err := db.DB.Preload("Categories").First(&post, "id = ?", postID).Error; err != nil {
		utils.LogError(err, "Post not found in UpdatePost")
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Vérifier que l'utilisateur est propriétaire du post ou admin
	userRole, _ := c.Get("user_role")
	if post.UserID != userID.(string) && userRole.(string) != string(models.AdminRole) {
		utils.LogError(nil, "Not authorized to update this post in UpdatePost")
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to update this post"})
		return
	}

	var input models.PostUpdate
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.LogError(err, "Invalid JSON in UpdatePost")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	post.Name = input.Name
	post.IsFree = input.IsFree
	post.Description = input.Description

	categoryIDs := input.Categories
	var categories []models.Category
	if err := db.DB.Where("id IN ?", categoryIDs).Find(&categories).Error; err != nil {
		utils.LogError(err, "Error finding categories in UpdatePost")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding categories: " + err.Error()})
		return
	}

	if len(categories) > 0 {
		if err := db.DB.Model(&post).Association("Categories").Replace(&categories); err != nil {
			utils.LogError(err, "Error updating categories in UpdatePost")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating categories: " + err.Error()})
			return
		}
	}

	if err := db.DB.Save(&post).Error; err != nil {
		utils.LogError(err, "Error updating post in UpdatePost")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating post: " + err.Error()})
		return
	}

	if err := db.DB.Preload("Categories").First(&post, "id = ?", post.ID).Error; err != nil {
		utils.LogError(err, "Error retrieving updated post in UpdatePost")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving updated post: " + err.Error()})
		return
	}

	utils.LogSuccess("Post updated successfully in UpdatePost")
	c.JSON(http.StatusOK, post)
}

// @Summary Delete a post
// @Description Delete a post by its ID
// @Tags posts
// @Produce json
// @Param id path string true "Post ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string "message: Post deleted successfully"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 404 {object} map[string]string "error: Post not found"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /posts/{id} [delete]
func DeletePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.LogError(nil, "User not found in token in DeletePost")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in token"})
		return
	}

	var post models.Post
	postID := c.Param("id")

	if err := db.DB.First(&post, "id = ?", postID).Error; err != nil {
		utils.LogError(err, "Post not found in DeletePost")
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Vérifier que l'utilisateur est propriétaire du post ou admin
	userRole, _ := c.Get("user_role")
	if post.UserID != userID.(string) && userRole.(string) != string(models.AdminRole) {
		utils.LogError(nil, "Not authorized to delete this post in DeletePost")
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete this post"})
		return
	}

	if post.PictureURL != "" {
		_ = utils.DeleteImage(post.PictureURL)
	}

	// Supprimer tous les rapports associés à ce post
	if err := db.DB.Where("post_id = ?", postID).Delete(&models.Report{}).Error; err != nil {
		utils.LogError(err, "Error deleting post reports in DeletePost")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting post reports: " + err.Error()})
		return
	}

	// Supprimer tous les commentaires associés à ce post
	if err := db.DB.Where("post_id = ?", postID).Delete(&models.Comment{}).Error; err != nil {
		utils.LogError(err, "Error deleting post comments in DeletePost")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting post comments: " + err.Error()})
		return
	}

	// Supprimer tous les likes associés à ce post
	if err := db.DB.Where("post_id = ?", postID).Delete(&models.Like{}).Error; err != nil {
		utils.LogError(err, "Error deleting post likes in DeletePost")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting post likes: " + err.Error()})
		return
	}

	// Supprimer les associations avec les catégories
	if err := db.DB.Model(&post).Association("Categories").Clear(); err != nil {
		utils.LogError(err, "Error removing post categories in DeletePost")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error removing post categories: " + err.Error()})
		return
	}

	// Supprimer le post
	if err := db.DB.Delete(&post).Error; err != nil {
		utils.LogError(err, "Error deleting post in DeletePost")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting post: " + err.Error()})
		return
	}

	utils.LogSuccess("Post deleted successfully in DeletePost")
	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

// @Summary Get post statistics (Admin)
// @Description Get statistics about posts by day and by category
// @Tags posts
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "total: total number of posts, daily_data: array of daily post creation data, category_data: array of posts by category"
// @Failure 400 {object} map[string]string "error: Invalid date parameters"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: Error retrieving statistics"
// @Router /posts/statistics [get]
func GetPostsStatistics(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		utils.LogError(nil, "start_date or end_date missing in GetPostsStatistics")
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required (format YYYY-MM-DD)"})
		return
	}

	var startDate, endDate time.Time
	var err error

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
		utils.LogError(err, "Invalid start_date format in GetPostsStatistics: "+startDateStr)
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
		utils.LogError(err, "Invalid end_date format in GetPostsStatistics: "+endDateStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format (YYYY-MM-DD)"})
		return
	}

	if endDate.Before(startDate) {
		utils.LogError(nil, "end_date before start_date in GetPostsStatistics")
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_date must be after start_date"})
		return
	}

	// Calcul du nombre total de posts créés dans la période
	var total int64
	err = db.DB.Model(&models.Post{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate.Add(24*time.Hour)).
		Count(&total).Error
	if err != nil {
		utils.LogError(err, "Error calculating total post count in GetPostsStatistics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error calculating total post count"})
		return
	}

	// Récupération des données quotidiennes pour le graphique
	type DailyData struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}

	var dailyData []DailyData
	err = db.DB.Model(&models.Post{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ? AND created_at <= ?", startDate, endDate.Add(24*time.Hour)).
		Group("DATE(created_at)").
		Order("date").
		Scan(&dailyData).Error
	if err != nil {
		utils.LogError(err, "Error fetching daily post data in GetPostsStatistics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching daily post data"})
		return
	}

	// Récupération des données par catégorie
	type CategoryData struct {
		CategoryID   string `json:"category_id"`
		CategoryName string `json:"category_name"`
		Count        int64  `json:"count"`
	}

	var categoryData []CategoryData
	err = db.DB.Table("categories").
		Select("categories.id as category_id, categories.name as category_name, COUNT(post_categories.post_id) as count").
		Joins("LEFT JOIN post_categories ON categories.id = post_categories.category_id").
		Joins("LEFT JOIN posts ON post_categories.post_id = posts.id").
		Where("posts.created_at >= ? AND posts.created_at <= ?", startDate, endDate.Add(24*time.Hour)).
		Group("categories.id, categories.name").
		Order("count DESC").
		Scan(&categoryData).Error
	if err != nil {
		utils.LogError(err, "Error fetching category post data in GetPostsStatistics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching category post data"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		userID = "0"
	}
	utils.LogSuccessWithUser(userID, "Post statistics retrieved successfully in GetPostsStatistics")
	c.JSON(http.StatusOK, gin.H{
		"total":         total,
		"daily_data":    dailyData,
		"category_data": categoryData,
	})
}
