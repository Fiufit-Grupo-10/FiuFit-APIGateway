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
	VerifyToken(token string) (string, error)
}

type UserModel struct {
	UID      string `json:"uid"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

