package contacts

import (
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"
	"pec2-backend/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// @Summary Create a new contact request
// @Description Submit a new contact request with the provided information
// @Tags contacts
// @Accept json
// @Produce json
// @Param contact body models.ContactCreate true "Contact information"
// @Success 201 {object} map[string]interface{} "message: Contact request submitted successfully, id: contact ID"
// @Failure 400 {object} map[string]interface{} "error: Invalid input"
// @Failure 500 {object} map[string]interface{} "error: Error message"
// @Router /contact [post]
func CreateContact(c *gin.Context) {
	var contactInput models.ContactCreate

	if err := c.ShouldBindJSON(&contactInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input: " + err.Error(),
		})
		return
	}

	if !utils.ValidateEmail(contactInput.Email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email format",
		})
		return
	}

	contact := models.Contact{
		FirstName:   contactInput.FirstName,
		LastName:    contactInput.LastName,
		Email:       contactInput.Email,
		Subject:     contactInput.Subject,
		Message:     contactInput.Message,
		SubmittedAt: time.Now(),
	}

	result := db.DB.Create(&contact)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Contact request submitted successfully",
		"id":      contact.ID,
	})
}

// @Summary Get all contact requests (Admin)
// @Description Retrieves a list of all contact requests (Admin access only)
// @Tags contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string][]models.Contact
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 403 {object} map[string]string "error: Forbidden - Admin access required"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /contacts [get]
func GetAllContacts(c *gin.Context) {
	var contacts []models.Contact
	result := db.DB.Order("submitted_at DESC").Find(&contacts)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, contacts)
}
