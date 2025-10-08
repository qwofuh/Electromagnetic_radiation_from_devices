package ds

import (
	"database/sql"
	"time"
)

type Emission struct {
	ID            int          `gorm:"primaryKey"`
	Status        string       `gorm:"type:varchar(20);not null;default:'черновик';check:status IN ('черновик', 'удалён', 'сформирован', 'завершён', 'отклонён')"`
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
