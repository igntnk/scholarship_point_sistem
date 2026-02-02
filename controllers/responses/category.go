package responses

type CreateCategory struct {
	Uuid string `json:"uuid"`
}

type Category struct {
	UUID   string           `json:"uuid,omitempty"`
	Name   string           `json:"name"`
	Points float32          `json:"points"`
	Values []CategoryValues `json:"values,omitempty"`
}

type CategoryValues struct {
	Name   string  `json:"name"`
	Points float32 `json:"points"`
}

type Subcategory struct {
	UUID            string   `json:"uuid"`
	Name            string   `json:"name"`
	SelectedValue   string   `json:"selected_value"`
	Points          float32  `json:"points"`
	AvailableValues []string `json:"available_values"`
}
