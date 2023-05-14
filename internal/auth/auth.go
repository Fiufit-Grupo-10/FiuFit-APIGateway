package auth

type SignUpModel struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Federated bool   `json:"is_federated"`
	UID       string `json:"uid"`
}

type Service interface {
	CreateUser(data SignUpModel) (UserModel, error)
	// May change it's return type in the future depending in the
	// info needed when validatin users
	VerifyToken(token string) (string, error)
	GetUser(uid string) (UserModel, error)
}

type UserModel struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	UID      string `json:"uid"`
}
