package responses

type Group struct {
	UUID      string       `json:"uuid"`
	Name      string       `json:"name"`
	Roles     []SimpleRole `json:"roles"`
	Resources []Resource   `json:"resources"`
}

type SimpleGroup struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}
