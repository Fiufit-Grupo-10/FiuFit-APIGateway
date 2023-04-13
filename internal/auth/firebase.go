package auth

import (
	"context"

	"firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

type Firebase struct {
	app        *firebase.App
	authClient *auth.Client
}

//type FirebaseAuth

func GetFirebase(ctx context.Context) (*Firebase, error) {
	// TODO change to env
	opt := option.WithCredentialsFile("firebase.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}
	return &Firebase{app: app, authClient: client}, nil
}

func (f *Firebase) CreateUser(userData SignUpModel) (UserModel, error) {
	params := (&auth.UserToCreate{}).
		DisplayName(userData.Username).
		Email(userData.Email).
		EmailVerified(false).
		Password(userData.Password).
		Disabled(false)
	ctx := context.Background()
	u, err := f.authClient.CreateUser(ctx, params)
	if err != nil {
		return UserModel{}, err
	}
	user := UserModel{
		UID:      u.UID,
		Username: u.DisplayName,
		Email:    u.Email,
	}
	return user, err
}

func (f *Firebase) VerifyToken(token string) (string, error) {
	ctx := context.Background()
	// TODO: Decide whether or not to use CheckRevoked
	tokenData, err := f.authClient.VerifyIDToken(ctx, token)
	if err != nil {
		return "", err
	}
	return tokenData.UID, nil
}
