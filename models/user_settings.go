package models

import (
	"time"
)

type UserSettings struct {
	ID             string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID         string    `json:"userId" gorm:"uniqueIndex" binding:"required"`
	User           User      `json:"-" gorm:"foreignKey:UserID"`
	CommentEnabled bool      `json:"commentEnabled" gorm:"default:true"`
	MessageEnabled bool      `json:"messageEnabled" gorm:"default:true"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (UserSettings) TableName() string {
	return "user_settings"
}
