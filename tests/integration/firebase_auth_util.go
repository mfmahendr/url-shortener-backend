package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"

	"firebase.google.com/go/v4/auth"
)

func createTestUserAndToken(ctx context.Context, authClient *auth.Client, email string, claims *map[string]any) (uid string, customToken string, err error) {
	userRecord, err := authClient.CreateUser(ctx, (&auth.UserToCreate{}).
		Email(email).
		EmailVerified(false).
		Password("password123").
		Disabled(false))
	if err != nil {
		return
	}
	uid = userRecord.UID

	if claims == nil {
		customToken, err = authClient.CustomToken(ctx, uid)
	} else {
		customToken, err = authClient.CustomTokenWithClaims(ctx, uid, *claims)
	}
	if err != nil {
		return
	}

		reqBody := map[string]interface{}{
		"token":             customToken,
		"returnSecureToken": true,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	fbAuthHost := os.Getenv("FIREBASE_AUTH_EMULATOR_HOST")
	resp, err := http.Post(
		"http://"+fbAuthHost+"/identitytoolkit.googleapis.com/v1/accounts:signInWithCustomToken?key=fake-api-key",
		"application/json",
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var res struct {
		IDToken string `json:"idToken"`
	}
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return "", "", err
	}

	return uid, res.IDToken, nil
}
