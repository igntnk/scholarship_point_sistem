package requests

type Role struct {
	UUID    string       `json:"uuid"`
	Name    string       `json:"name"`
	Members []RoleMember `json:"members"`
}

type RoleMember struct {
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	SecondName string `json:"second_name"`
	Patronymic string
}
