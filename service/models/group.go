package models

type Group struct {
	UUID      string
	Name      string
	Roles     []SimpleRole
	Resources []Resource
}

type SimpleGroup struct {
	UUID string
	Name string
}

type UpdateGroup struct {
	UUID                string
	Name                string
	CreateRoleUUIDs     []string
	DeleteRoleUUIDs     []string
	CreateResourceUUIDs []string
	DeleteResourceUUIDs []string
}
