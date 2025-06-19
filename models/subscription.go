package models

import (
	"time"
)

type SubscriptionStatus string

const (
	SubscriptionActive   SubscriptionStatus = "ACTIVE"
	SubscriptionCanceled SubscriptionStatus = "CANCELED"
	SubscriptionPending  SubscriptionStatus = "PENDING"
)

type Subscription struct {
	ID                   string             `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID               string             `json:"userId" gorm:"column:user_id;type:uuid;references:ID;foreignKey:fk_subscription"`
	User                 User               `json:"user,omitempty" gorm:"foreignKey:UserID"`
	ContentCreatorID     string             `json:"contentCreatorId" gorm:"type:uuid;not null"`
	Status               SubscriptionStatus `json:"status" gorm:"type:varchar(20);default:'PENDING'"`
	StripeSubscriptionId string             `json:"stripeSubscriptionId"`
	StartDate            time.Time          `json:"startDate"`
	EndDate              *time.Time         `json:"endDate"`
	CreatedAt            time.Time          `json:"createdAt"`
	UpdatedAt            time.Time          `json:"updatedAt"`
}
