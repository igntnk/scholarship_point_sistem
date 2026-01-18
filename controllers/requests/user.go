package requests

type CreateUser struct {
	Name            string `json:"name"`
	SecondName      string `json:"second_name"`
	Patronymic      string `json:"patronymic"`
	GradebookNumber string `json:"gradebook_number"`
	BirthDate       string `json:"birth_date"`
	Email           string `json:"email"`
	PhoneNumber     string `json:"phone_number"`
	Password        string `json:"password"`
}

type UpdateUser struct {
	UUID            string `json:"uuid"`
	Name            string `json:"name"`
	SecondName      string `json:"second_name"`
	Patronymic      string `json:"patronymic"`
	GradebookNumber string `json:"gradebook_number"`
	BirthDate       string `json:"birth_date"`
	Email           string `json:"email"`
	PhoneNumber     string `json:"phone_number"`
}
