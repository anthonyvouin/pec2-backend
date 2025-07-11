package models

import (
	"time"
)

type Role string
type Sexe string

const (
	AdminRole      Role = "ADMIN"
	UserRole       Role = "USER"
	ContentCreator Role = "CONTENT_CREATOR"
)

const (
	Male   Sexe = "MAN"
	Female Sexe = "WOMAN"
	Other  Sexe = "OTHER"
)

type User struct {
	ID                   string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email                string     `json:"email" binding:"required,email" gorm:"uniqueIndex"`
	Password             string     `json:"password" binding:"required,min=6"`
	UserName             string     `json:"userName" binding:"required" gorm:"uniqueIndex"`
	FirstName            string     `json:"firstName" binding:"required"`
	LastName             string     `json:"lastName" binding:"required"`
	BirthDayDate         time.Time  `json:"birthDayDate" binding:"required"`
	Sexe                 Sexe       `json:"sexe" binding:"required"`
	Role                 Role       `json:"role"`
	Bio                  string     `json:"bio"`
	ProfilePicture       string     `json:"profilePicture"`
	StripeCustomerId     string     `json:"stripeCustomerId"`
	Enable               bool       `json:"enable"`
	SubscriptionEnable   bool       `json:"subscriptionEnable"`
	CommentsEnable       bool       `json:"commentsEnable"`
	MessageEnable        bool       `json:"messageEnable"`
	EmailVerifiedAt      *time.Time `json:"emailVerifiedAt"`
	Siret                string     `json:"siret"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
	DeletedAt            *time.Time `json:"deletedAt,omitempty" gorm:"index"`
	ConfirmationCode     string     `json:"confirmationCode"`
	ConfirmationCodeEnd  time.Time  `json:"ConfirmationCodeEnd"`
	ResetPasswordCode    string     `json:"resetPasswordCode"`
	ResetPasswordCodeEnd time.Time  `json:"resetPasswordCodeEnd"`
}

type UserLogin struct {
	ID                 string    `json:"id"`
	Email              string    `json:"email"`
	UserName           string    `json:"userName"`
	Role               Role      `json:"role"`
	Bio                string    `json:"bio"`
	ProfilePicture     string    `json:"profilePicture"`
	FirstName          string    `json:"firstName"`
	LastName           string    `json:"lastName"`
	BirthDayDate       time.Time `json:"birthDayDate"`
	Sexe               Sexe      `json:"sexe"`
	CommentsEnable     bool      `json:"commentsEnable"`
	MessageEnable      bool      `json:"messageEnable"`
	SubscriptionEnable bool      `json:"subscriptionEnable"`
}

func (User) TableName() string {
	return "users"
}

type LiteUser struct {
	ID             string    `json:"id"`
	UserName       string    `json:"username"`
	ProfilePicture string    `json:"profile_picture"`
	BirthDayDate   time.Time `json:"birthDayDate"`
	Sexe           Sexe      `json:"sexe"`
}

// UserCreate model for create a user
// @Description model for create a user
type UserCreate struct {
	Email        string    `json:"email" binding:"required,email" example:"utilisateur@exemple.com"`
	Password     string    `json:"password" binding:"required,min=6" example:"Motdepasse123"`
	UserName     string    `json:"userName" binding:"required" example:"utilisateur123"`
	FirstName    string    `json:"firstName" binding:"required" example:"Jean"`
	LastName     string    `json:"lastName" binding:"required" example:"Dupont"`
	BirthDayDate time.Time `json:"birthDayDate" binding:"required" example:"1990-01-01T00:00:00Z"`
	Sexe         Sexe      `json:"sexe" binding:"required" example:"MAN"`
}

// PasswordUpdate modèle pour la mise à jour du mot de passe
// @Description modèle pour mettre à jour le mot de passe d'un utilisateur
type PasswordUpdate struct {
	OldPassword string `json:"oldPassword" binding:"required" example:"AncienMotdepasse123"`
	NewPassword string `json:"newPassword" binding:"required,min=6" example:"NouveauMotdepasse123"`
}

type UserUpdateFormData struct {
	UserName     string    `form:"userName"`
	Bio          string    `form:"bio"`
	FirstName    string    `form:"firstName"`
	Email        string    `form:"email"`
	LastName     string    `form:"lastName"`
	BirthDayDate time.Time `form:"birthDayDate"`
	Sexe         Sexe      `form:"sexe"`
}

type UserFollow struct {
	ID         string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	FollowerID string    `gorm:"type:uuid;not null" json:"followerId"`
	Follower   User      `gorm:"foreignKey:FollowerID;references:ID" json:"follower"`
	FollowedID string    `gorm:"type:uuid;not null" json:"followedId"`
	Followed   User      `gorm:"foreignKey:FollowedID;references:ID" json:"followed"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

type StatisticsUsers struct {
	Subscribers []User
}
