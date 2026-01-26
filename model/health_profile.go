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
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
	UserEmail string    `gorm:"not null;index:idx_email" json:"user_email"`

	Gender string  `gorm:"not null;type:enum('male','female','other')" json:"gender"`
	Age    int     `gorm:"not null" json:"age"`
	Height float32 `gorm:"not null" json:"height"`
	Weight float32 `gorm:"not null" json:"weight"`

	// 饮食偏好
	DietaryPreference string `gorm:"type:text" json:"dietary_preference"`
	// 是否吸烟
	SmokingStatus bool `gorm:"not null" json:"smoking_status"`
	// 运动等级
	ActivityLevel string `gorm:"not null;type:enum('sedentary','light','moderate','heavy')" json:"activity_level"`

	DiabetesType  string `gorm:"not null;type:enum('type1','type2', 'gestational', 'other', 'none')" json:"diabetes_type"`
	DiagnosisYear int    `json:"diagnosis_year"`

	// 治疗模式
	TherapyMode string `gorm:"not null;type:enum('lifestyle','oral_meds','insulin','combined')" json:"therapy_mode"`
	// 用药情况
	Medication string `gorm:"type:text" json:"medication"`

	// 药物/食物过敏史
	Allergies string `json:"allergies"`
	// 并发症情况
	Complications string `gorm:"type:text" json:"complications"`
}

func (HealthProfile) TableName() string {
	return "health_profile"
}
