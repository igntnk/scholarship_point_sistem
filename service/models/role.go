package models

type Role struct {
	UUID    string
	Name    string
	Members []RoleMember
}

type RoleMember struct {
	UUID       string
	Name       string
	SecondName string
	Patronymic string
}

type SimpleRole struct {
	UUID string
	Name string
}

type UpdateRole struct {
	UUID              string
	Name              string
	CreateMemberUUIDs []string
	DeleteMemberUUIDs []string
}
