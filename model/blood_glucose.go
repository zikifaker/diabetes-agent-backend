package model

import "time"

const (
	DiningStatusFasting         = "fasting"
	DiningStatusBeforeBreakfast = "before_breakfast"
	DiningStatusAfterBreakfast  = "after_breakfast"
	DiningStatusBeforeLunch     = "before_lunch"
	DiningStatusAfterLunch      = "after_lunch"
	DiningStatusBeforeDinner    = "before_dinner"
	DiningStatusAfterDinner     = "after_dinner"
	DiningStatusBedtime         = "bedtime"
	DiningStatusRandom          = "random"

	DiabetesTypeType1       = "type1"
	DiabetesTypeType2       = "type2"
	DiabetesTypeGestational = "gestational"
	DiabetesTypeOther       = "other"
	DiabetesTypeNone        = "none"

	ReportTypeWeekly  = "weekly"
	ReportTypeMonthly = "monthly"
	ReportTypeYearly  = "yearly"
)

type BloodGlucoseRecord struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null" json:"updated_at"`
	UserEmail    string    `gorm:"not null;index:idx_email_measured_at" json:"user_email"`
	Value        float32   `gorm:"not null" json:"value"`
	MeasuredAt   time.Time `gorm:"not null;index:idx_email_measured_at" json:"measured_at"`
	DiningStatus string    `gorm:"not null;type:enum('fasting','before_breakfast','after_breakfast','before_lunch','after_lunch','before_dinner','after_dinner','bedtime','random')" json:"dining_status"`
	Notes        string    `gorm:"type:text" json:"notes"`
}

func (BloodGlucoseRecord) TableName() string {
	return "blood_glucose_record"
}

type HealthProfile struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt     time.Time `gorm:"not null" json:"updated_at"`
	UserEmail     string    `gorm:"not null;index:idx_email" json:"user_email"`
	DiabetesType  string    `gorm:"not null type:enum('type1','type2', 'gestational', 'other', 'none')" json:"diabetes_type"`
	TargetLow     float32   `gorm:"not null" json:"target_low"`
	TargetHigh    float32   `gorm:"not null" json:"target_high"`
	Medication    string    `gorm:"type:text" json:"medication"`
	Complications string    `gorm:"type:text" json:"complications"`
}

func (HealthProfile) TableName() string {
	return "health_profile"
}

type HealthReport struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt  time.Time `gorm:"not null" json:"updated_at"`
	UserEmail  string    `gorm:"not null;index:idx_email_start_at" json:"user_email"`
	StartAt    time.Time `gorm:"not null;index:idx_email_start_at" json:"start_at"`
	EndAt      time.Time `gorm:"not null" json:"end_at"`
	ReportType string    `gorm:"not null type:enum('weekly','monthly','yearly')" json:"report_type"`
	Content    string    `gorm:"type:text" json:"content"`
}

func (HealthReport) TableName() string {
	return "health_report"
}
