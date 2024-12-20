package models

type User struct {
	Id       int    `json:"id" db:"user_id"`
	Role     Role   `json:"role" db:"role"`
	Login    string `json:"login" db:"login"`
	PassHash string `json:"-" db:"pass_hash"`
}
