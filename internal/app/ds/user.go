package ds

import "lab1/internal/app/role"

// User - пользователь
type User struct {
	ID       int       `gorm:"primaryKey" json:"id"`                     // уникальный идентификатор пользователя
	Login    string    `gorm:"varchar(25);unique;not null" json:"login"` // логин пользователя
	Password string    `gorm:"varchar(100);not null" json:"-"`           // пароль
	Role     role.Role `gorm:"type:int"`                                 // хранить enum как int
}

type RegisterReq struct {
	Login    string    `json:"login"`
	Password string    `json:"password"`
	Role     role.Role `json:"role,omitempty"` // опционально
}

// Структура ответа
type RegisterResp struct {
	Ok   bool      `json:"ok"`
	Role role.Role `json:"role,omitempty"` // добавляем поле роли
}
