package ds

// User - пользователь
type User struct {
	ID          int    `gorm:"primaryKey" json:"id"`                               // уникальный идентификатор пользователя
	Login       string `gorm:"varchar(25);unique;not null" json:"login"`           // логин пользователя
	Password    string `gorm:"varchar(100);not null" json:"-"`                     // пароль
	IsModerator bool   `gorm:"boolean;not null;default:false" json:"is_moderator"` // признак модератора
}
