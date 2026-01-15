package models

type SimpleUser struct {
	UUID            string `json:"uuid"`
	Name            string `json:"name"`
	SecondName      string `json:"second_name"`
	Patronymic      string `json:"patronymic"`
	GradeBookNumber string `json:"gradebook_number"`
	BirthDate       string `json:"birth_date"`
	Email           string `json:"email"`
	PhoneNumber     string `json:"phone_number"`
}
