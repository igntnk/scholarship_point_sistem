package models

type Category struct {
	UUID                string           `json:"uuid,omitempty"`
	Name                string           `json:"name"`
	Points              float32          `json:"points"`
	SubcategoriesAmount int              `json:"subcategories_amount"`
	Values              []CategoryValues `json:"values,omitempty"`
}

type CategoryValues struct {
	Name   string  `json:"name"`
	Points float32 `json:"points"`
}
