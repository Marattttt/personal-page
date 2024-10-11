package models

type AuthReq struct {
	Login string `json:"login"`
	Pass  string `json:"pass"`
}
