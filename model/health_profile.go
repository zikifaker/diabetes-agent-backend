package model

import "time"

const (
	DiabetesTypeType1       = "type1"
	DiabetesTypeType2       = "type2"
	DiabetesTypeGestational = "gestational"
	DiabetesTypeOther       = "other"
	DiabetesTypeNone        = "none"
)

type HealthProfile struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt     time.Time `gorm:"not null" json:"updated_at"`
	UserEmail     string    `gorm:"not null;index:idx_email" json:"user_email"`
	DiabetesType  string    `gorm:"not null type:enum('type1','type2', 'gestational', 'other', 'none')" json:"diabetes_type"`
	Medication    string    `gorm:"type:text" json:"medication"`
	Complications string    `gorm:"type:text" json:"complications"`
}

func (HealthProfile) TableName() string {
	return "health_profile"
}
