package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var client = &http.Client{}

func doPost(t *testing.T, url string, jsonBody []byte, cookies ...map[string]string) *http.Response {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	assert.Nil(t, err)

	req.Header.Set("Content-Type", "application/json")

	for _, cookie := range cookies {
		for k, v := range cookie {
			req.AddCookie(&http.Cookie{
				Name:  k,
				Value: v,
			})
		}
	}

	resp, err := client.Do(req)
	assert.Nil(t, err)

	return resp
}

func doDelete(t *testing.T, url string, jsonBody []byte, cookies ...map[string]string) *http.Response {
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonBody))
	assert.Nil(t, err)

	req.Header.Set("Content-Type", "application/json")

	for _, cookie := range cookies {
		for k, v := range cookie {
			req.AddCookie(&http.Cookie{
				Name:  k,
				Value: v,
			})
		}
	}

	resp, err := client.Do(req)
	assert.Nil(t, err)

	return resp
}

func doGet(t *testing.T, url string, cookies ...map[string]string) *http.Response {
	req, err := http.NewRequest("GET", url, nil)
	assert.Nil(t, err)

	for _, cookie := range cookies {
		for k, v := range cookie {
			req.AddCookie(&http.Cookie{
				Name:  k,
				Value: v,
			})
		}
	}

	resp, err := client.Do(req)
	assert.Nil(t, err)

	return resp
}

func TestUnauthorized(t *testing.T) {
	t.Parallel()

	// create post
	jsonBody := []byte(`{"title":"post title", "content": "post content"}`)

	resp := doPost(t, "http://localhost:5000/posts/create", jsonBody)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// delete post
	jsonBody = []byte(`{"post_id": 1}`)
	resp = doDelete(t, "http://localhost:5000/posts/delete", jsonBody)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// get posts
	resp = doGet(t, "http://localhost:5000/posts/get")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	//signout
	resp = doPost(t, "http://localhost:5000/users/signout", []byte{})
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	//userinfo
	resp = doGet(t, "http://localhost:5000/users/info")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// like post
	jsonBody = []byte(`{"post_id": 1}`)
	resp = doPost(t, "http://localhost:5000/posts/like", jsonBody)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// unlike post
	jsonBody = []byte(`{"post_id": 1}`)
	resp = doPost(t, "http://localhost:5000/posts/unlike", jsonBody)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Post struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Likes     int       `json:"likes"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

func SignUpUser(t *testing.T, username string, password string, email string) Tokens {
	jsonBody := []byte(fmt.Sprintf(`{"username": "%s", "password": "%s", "email": "%s"}`, username, password, email))

	resp := doPost(t, "http://localhost:5000/users/signup", jsonBody)

	data := Tokens{}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "access_token" {
			data.AccessToken = cookie.Value
		}
		if cookie.Name == "refresh_token" {
			data.RefreshToken = cookie.Value
		}
	}

	assert.NotEmpty(t, data.AccessToken)
	assert.NotEmpty(t, data.RefreshToken)

	return data
}

func SignInUser(t *testing.T, username string, password string) Tokens {
	jsonBody := []byte(fmt.Sprintf(`{"username": "%s", "password": "%s"}`, username, password))

	resp := doPost(t, "http://localhost:5000/users/signin", jsonBody)

	data := Tokens{}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "access_token" {
			data.AccessToken = cookie.Value
		}
		if cookie.Name == "refresh_token" {
			data.RefreshToken = cookie.Value
		}
	}

	assert.NotEmpty(t, data.AccessToken)
	assert.NotEmpty(t, data.RefreshToken)

	return data
}

func TestAuthorized(t *testing.T) {
	t.Parallel()

	t.Run("Test sign up", testSignUp)

	t.Run("Test create post", testCreatePost)

	t.Run("Test delete post", testDeletePost)

	t.Run("Test sign in", testSignIn)

	t.Run("Test like post", testLike)

	t.Run("Test unlike post", testUnlike)

	t.Run("Test get info", testGetInfo)

	t.Run("Test get posts", testGetPosts)
}

func testSignUp(t *testing.T) {
	SignUpUser(t, "testuser", "testpassword", "testemail@gmail.com")
}

func testCreatePost(t *testing.T) {
	tokens := SignUpUser(t, "testCreatePost", "password", "email@email.com")

	jsonBody := []byte(`{"title":"post title", "content": "post content"}`)

	resp := doPost(t, "http://localhost:5000/posts/create", jsonBody, map[string]string{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})

	data := Post{}

	err := json.NewDecoder(resp.Body).Decode(&data)
	assert.Nil(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Equal(t, "post title", data.Title)
	assert.Equal(t, "post content", data.Content)
	assert.Equal(t, 0, data.Likes)
	assert.Equal(t, "testCreatePost", data.Author)

	assert.NotEmpty(t, data.ID)
	assert.NotEmpty(t, data.CreatedAt)
}

func testDeletePost(t *testing.T) {
	tokens := SignUpUser(t, "testDeletePost", "password", "email@email.com")

	jsonBody := []byte(`{"post_id": 1}`)

	//delete wrong post
	resp := doDelete(t, "http://localhost:5000/posts/delete", jsonBody, map[string]string{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	//delete unexistent post
	jsonBody = []byte(`{"post_id": 2}`)

	resp = doDelete(t, "http://localhost:5000/posts/delete", jsonBody, map[string]string{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	//create new post
	jsonBody = []byte(`{"title": "post title", "content": "post content"}`)
	resp = doPost(t, "http://localhost:5000/posts/create", jsonBody, map[string]string{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	//delete post
	jsonBody = []byte(`{"post_id": 2}`)
	resp = doDelete(t, "http://localhost:5000/posts/delete", jsonBody, map[string]string{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	//TODO: check if post was deleted
}

func testSignIn(t *testing.T) {
	tokens := SignInUser(t, "testuser", "testpassword")

	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
}

func testLike(t *testing.T) {
	tokens := SignInUser(t, "testuser", "testpassword")

	jsonBody := []byte(`{"post_id": 1}`)

	resp := doPost(t, "http://localhost:5000/posts/like", jsonBody, map[string]string{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	//TODO: check if post was liked
}

func testUnlike(t *testing.T) {
	tokens := SignInUser(t, "testuser", "testpassword")

	jsonBody := []byte(`{"post_id": 1}`)

	resp := doPost(t, "http://localhost:5000/posts/unlike", jsonBody, map[string]string{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	//TODO: check if post was unliked
}

func testGetInfo(t *testing.T) {
	tokens := SignInUser(t, "testuser", "testpassword")

	resp := doGet(t, "http://localhost:5000/users/info", map[string]string{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	//TODO: check if user info is correct
}

func testGetPosts(t *testing.T) {
	resp := doGet(t, "http://localhost:5000/posts/get")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	//TODO: check if posts are correct
}
