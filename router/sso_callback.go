package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type OAuthData struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
	HD            string `json:"hd"`
}

func OnAuthCallback(w http.ResponseWriter, req *http.Request, RandomString, allowedDomains string, oAuth2Config *oauth2.Config) {
	state := req.FormValue("state")
	code := req.FormValue("code")
	data, err := getUserData(state, code, RandomString, oAuth2Config)
	if err != nil {
		log.Errorf("Error when getting user data: %q", err.Error())
	}
	ctx := context.WithValue(req.Context(), 0, data)
	r2 := req.WithContext(ctx)
	*req = *r2

	fmt.Printf("Context callback: %+v", req.Context())
}

func getUserData(state, code, RandomString string, oAuth2Config *oauth2.Config) (*OAuthData, error) {
	if state != RandomString {
		return nil, errors.New("invalid user state")
	}
	token, err := oAuth2Config.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var oAuthData OAuthData
	parse_err := json.Unmarshal(data, &oAuthData)
	if parse_err != nil {
		return nil, parse_err
	}
	return &oAuthData, nil
}
