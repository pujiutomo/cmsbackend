package models

import "golang.org/x/crypto/bcrypt"

type User struct {
	Id          uint   `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	Password    []byte `json:"-"`
	Phone       string `json:"phone"`
	DomainsId   string `json:"domains_id"`
	AccessRight string `json:"access_right"`
	Su          string `json:"su"`
}

func (user *User) SetPassword(password string) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	user.Password = hashedPassword
}

func (user *User) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword(user.Password, []byte(password))
}
