package models

type User struct {
	Id       int    `json:"id" db:"id"`
	Role     Role   `json:"role" db:"role"`
	Login    string `json:"login" db:"login"`
	PassHash string `json:"-" db:"pass_hash"`
	Salt     string `json:"-" db:"pass_salt"`
}
