package requests

type Group struct {
	UUID      string     `json:"uuid"`
	Name      string     `json:"name"`
	Roles     []Role     `json:"roles"`
	Resources []Resource `json:"resources"`
}
