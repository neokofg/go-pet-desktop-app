package model

type Credential struct {
	ID       int    `json:"id"`
	Service  string `json:"service"`
	Username string `json:"username"`
	Password string `json:"password"`
	Notes    string `json:"notes"`
}
