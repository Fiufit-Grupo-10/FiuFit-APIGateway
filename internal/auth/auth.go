package auth

type SignUpModel struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthorizedModel struct {
	UID          string `json:"uid"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type Service interface {
	CreateUser(data SignUpModel) (UserModel, error)
	// May change it's return type in the future depending in the
	// info needed when validatin users
	VerifyToken(token string) (string, error) 
}

type UserModel struct {
	UID      string `json:"uid"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

