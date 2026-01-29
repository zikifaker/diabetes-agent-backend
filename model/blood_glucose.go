package model

import "time"

type BloodGlucoseRecord struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null" json:"updated_at"`
	UserEmail    string    `gorm:"not null;index:idx_email_measured_at" json:"user_email"`
	Value        float32   `gorm:"not null" json:"value"`
	MeasuredAt   time.Time `gorm:"not null;index:idx_email_measured_at" json:"measured_at"`
	DiningStatus string    `gorm:"not null;type:enum('fasting','before_breakfast','after_breakfast','before_lunch','after_lunch','before_dinner','after_dinner','bedtime','random')" json:"dining_status"`
}

func (BloodGlucoseRecord) TableName() string {
	return "blood_glucose_record"
}
