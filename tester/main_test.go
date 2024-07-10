package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type CreateTokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func TestTokens(t *testing.T) {
	userID := 1
	jsonBody := []byte(`{"user_id":` + fmt.Sprint(userID) + `}`)
	req, err := http.NewRequest("POST", "http://localhost:5000/tokens/create", bytes.NewBuffer(jsonBody))
	assert.Nil(t, err)

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	var tokens CreateTokensResponse
	err = json.Unmarshal(body, &tokens)
	assert.Nil(t, err)

	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)

	ValidateTokens(t, tokens)
}

type ValidateResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func ValidateTokens(t *testing.T, tokens CreateTokensResponse) {
	fmt.Println(tokens)

	validReq, err := http.NewRequest("GET", "http://localhost:5000/tokens/verify", nil)
	assert.Nil(t, err)

	cookies := []http.Cookie{
		{
			Name:  "access_token",
			Value: tokens.AccessToken,
		},
		{
			Name:  "refresh_token",
			Value: tokens.RefreshToken,
		},
	}

	validReq.AddCookie(&cookies[0])
	validReq.AddCookie(&cookies[1])

	client := &http.Client{}

	//VALID REQUEST
	resp, err := client.Do(validReq)
	assert.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// body, err := io.ReadAll(resp.Body)
	// assert.Nil(t, err)

	// var result ValidateResponse

	// err = json.Unmarshal(body, &result)
	// assert.Nil(t, err)

	//ACCESS TOKEN IS INVALID
	fullyInvalidAccessTokenReq, err := http.NewRequest("GET", "http://localhost:5000/tokens/verify", nil)
	assert.Nil(t, err)

	cookies[0] = http.Cookie{
		Name:  "access_token",
		Value: "fully_invalid_access_token",
	}

	fullyInvalidAccessTokenReq.AddCookie(&cookies[0])
	fullyInvalidAccessTokenReq.AddCookie(&cookies[1])

	resp, err = client.Do(fullyInvalidAccessTokenReq)
	assert.Nil(t, err)
	assert.Equal(t, 401, resp.StatusCode)

}
