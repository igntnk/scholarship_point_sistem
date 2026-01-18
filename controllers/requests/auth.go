package requests

type ChangePassword struct {
	Password string `json:"password"`
}

type SingIn struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
