package models

type UserData struct {
	Foods []Food `json:"foods"`
	Poops []Poop `json:"poops"`
}

func NewUserData() *UserData {
	return &UserData{
		Foods: []Food{},
		Poops: []Poop{},
	}
}
