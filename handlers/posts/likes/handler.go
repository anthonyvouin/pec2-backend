package likes

import (
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"
	"pec2-backend/utils"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @Summary Toggle like on a post
// @Description Add or remove a like on a post
// @Tags posts
// @Produce json
// @Param id path string true "Post ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string "message: Like added/removed successfully"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 404 {object} map[string]string "error: Post not found"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /posts/{id}/like [post]
func ToggleLike(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.LogError(nil, "User not found in token in ToggleLike")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in token"})
		return
	}

	postID := c.Param("id")

	// Vérifier si le post existe
	var post models.Post
	if err := db.DB.First(&post, "id = ?", postID).Error; err != nil {
		utils.LogError(err, "Post not found in ToggleLike")
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	var like models.Like
	result := db.DB.Where("post_id = ? AND user_id = ?", postID, userID).First(&like)

	if result.Error == nil {
		// Le like existe déjà, on le supprime
		if err := db.DB.Delete(&like).Error; err != nil {
			utils.LogError(err, "Error removing like in ToggleLike")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error removing like: " + err.Error()})
			return
		}

		// Récupérer le nombre actuel de likes pour ce post
		var likesCount int64
		db.DB.Model(&models.Like{}).Where("post_id = ?", postID).Count(&likesCount)

		utils.LogSuccessWithUser(userID, "Like removed successfully in ToggleLike")
		c.JSON(http.StatusOK, gin.H{
			"message":    "Like removed successfully",
			"action":     "removed",
			"likesCount": likesCount,
		})
		return
	}

	// Le like n'existe pas, on le crée
	like = models.Like{
		PostID: postID,
		UserID: userID.(string),
	}

	if err := db.DB.Create(&like).Error; err != nil {
		utils.LogError(err, "Error adding like in ToggleLike")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error adding like: " + err.Error()})
		return
	}

	// Récupérer le nombre actuel de likes pour ce post
	var likesCount int64
	db.DB.Model(&models.Like{}).Where("post_id = ?", postID).Count(&likesCount)

	utils.LogSuccessWithUser(userID, "Like added successfully in ToggleLike")
	c.JSON(http.StatusOK, gin.H{
		"message":    "Like added successfully",
		"action":     "added",
		"likesCount": likesCount,
	})
}

// @Summary Get like statistics (Admin)
// @Description Get statistics about likes by day
// @Tags likes
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "total: total number of likes, daily_data: array of daily like creation data"
// @Failure 400 {object} map[string]string "error: Invalid date parameters"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: Error retrieving statistics"
// @Router /likes/statistics [get]
func GetLikesStatistics(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		utils.LogError(nil, "start_date or end_date missing in GetLikesStatistics")
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
		utils.LogError(err, "Invalid start_date format in GetLikesStatistics: "+startDateStr)
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
		utils.LogError(err, "Invalid end_date format in GetLikesStatistics: "+endDateStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format (YYYY-MM-DD)"})
		return
	}

	if endDate.Before(startDate) {
		utils.LogError(nil, "end_date before start_date in GetLikesStatistics")
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_date must be after start_date"})
		return
	}

	var total int64
	totalSession := db.DB.Session(&gorm.Session{PrepareStmt: false})
	err = totalSession.Model(&models.Like{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate.Add(24*time.Hour)).
		Count(&total).Error
	if err != nil {
		utils.LogError(err, "Error calculating total like count in GetLikesStatistics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error calculating total like count"})
		return
	}

	type DailyData struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}

	var dailyData []DailyData
	dailySession := db.DB.Session(&gorm.Session{PrepareStmt: false})
	err = dailySession.Model(&models.Like{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ? AND created_at <= ?", startDate, endDate.Add(24*time.Hour)).
		Group("DATE(created_at)").
		Order("date").
		Scan(&dailyData).Error
	if err != nil {
		utils.LogError(err, "Error fetching daily like data in GetLikesStatistics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching daily like data"})
		return
	}

	type MostLikedPost struct {
		PostID    string `json:"post_id"`
		PostName  string `json:"post_name"`
		LikeCount int64  `json:"like_count"`
	}

	var mostLikedPosts []MostLikedPost
	postsSession := db.DB.Session(&gorm.Session{PrepareStmt: false})

	type PostLikeCount struct {
		PostID    string `json:"post_id"`
		LikeCount int64  `json:"like_count"`
	}
	var postCounts []PostLikeCount

	err = postsSession.Table("likes").
		Select("post_id, COUNT(*) as like_count").
		Where("created_at >= ? AND created_at <= ?", startDate, endDate.Add(24*time.Hour)).
		Group("post_id").
		Order("like_count DESC").
		Limit(10).
		Scan(&postCounts).Error

	if err != nil {
		utils.LogError(err, "Error fetching post like counts in GetLikesStatistics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching post like counts"})
		return
	}

	for _, pc := range postCounts {
		var post models.Post
		if err := db.DB.Select("name").Where("id = ?", pc.PostID).First(&post).Error; err == nil {
			mostLikedPosts = append(mostLikedPosts, MostLikedPost{
				PostID:    pc.PostID,
				PostName:  post.Name,
				LikeCount: pc.LikeCount,
			})
		}
	}

	userID, exists := c.Get("user_id")
	if !exists {
		userID = "0"
	}
	utils.LogSuccessWithUser(userID, "Like statistics retrieved successfully in GetLikesStatistics")
	c.JSON(http.StatusOK, gin.H{
		"total":            total,
		"daily_data":       dailyData,
		"most_liked_posts": mostLikedPosts,
	})
}
