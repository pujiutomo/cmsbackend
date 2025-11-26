package models

type Destinasi struct {
	Id     uint   `json:"id"`
	Name   string `json:"name" gorm:"unique; not null"`
	UserID string `json:"user_id"`
	User   User   `json:"user" gorm:"foreignKey:UserID"`
}
