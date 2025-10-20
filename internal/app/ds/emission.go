package ds

import (
	"database/sql"
	"time"
)

type Emission struct {
	ID            int          `gorm:"primaryKey"`
	Status        string       `gorm:"type:varchar(20);not null;default:'черновик';check:status IN ('черновик', 'удален', 'сформирован', 'завершен', 'отклонен')"`
	CreateAt      time.Time    `gorm:"not null;default:now()"`
	UpdateAt      time.Time    `gorm:"default:now()"`
	FinishAt      sql.NullTime `gorm:"default:null"`
	Distance      float64      `gorm:"type:numeric(10,2)"`
	TotalEmission float64      `gorm:"type:numeric(10,2)"`
	CreatorID     int          `gorm:"not null"`
	ModeratorID   int

	Creator   User `gorm:"foreignKey:CreatorID"`
	Moderator User `gorm:"foreignKey:ModeratorID"`
}

type EmissionResponse struct {
	ID            int        `json:"id"`
	Status        string     `json:"status"`
	CreateAt      time.Time  `json:"create_at"`
	UpdateAt      time.Time  `json:"update_at"`
	FinishAt      *time.Time `json:"finish_at,omitempty"`
	Distance      float64    `json:"distance"`
	TotalEmission float64    `json:"total_emission"`
}

type EmissionUpdateRequest struct {
	Distance *float64 `json:"distance,omitempty"`
}

type EmissionWithDevices struct {
	ID            int                   `json:"id"`
	Status        string                `json:"status"`
	CreateAt      time.Time             `json:"create_at"`
	UpdateAt      time.Time             `json:"update_at"`
	FinishAt      *time.Time            `json:"finish_at,omitempty"`
	Distance      float64               `json:"distance"`
	TotalEmission float64               `json:"total_emission"`
	CreatorID     int                   `json:"creator_id"`
	ModeratorID   *int                  `json:"moderator_id,omitempty"`
	Devices       []DeviceInCalculation `json:"devices"`
}

type EmissionCompleteRequest struct {
	Status      string `json:"status" binding:"required,oneof=завершен отклонен"`
	ModeratorID int    `json:"moderator_id" binding:"required"`
}
