package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"bitbucket.org/truora-tests/supporteng/app"
	"bitbucket.org/truora-tests/supporteng/model"
	"bitbucket.org/truora-tests/supporteng/respond"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
)

type Users struct {
}

func validateAmount(amount float64) bool {
	return amount >= 0
}
func validateUsername(fromUsername, toUsername string) bool {
	return fromUsername == toUsername
}
func isUsernameInToken(claims jwt.MapClaims, username string) bool {
	value, ok := claims[username].(string)
	if !ok || value == "" {

		return false

	}
	return true
}
func (c Users) Setup(a *app.App) error {
	type NewUserRequest struct {
		Username string `json:"Username"`
		Password string `json:"Password"`
	}

	a.Router.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		result, err := a.Users.List(r.Context(), model.Pager{})
		if err != nil {
			respond.Err(w, err)
			return
		}

		respond.JSONObjects(w, 200, result)
	})

	a.Router.Post("/users", func(w http.ResponseWriter, r *http.Request) {
		nureq := NewUserRequest{}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			respond.JSON(w, 400, "Failed to read body")
			return
		}

		err = json.Unmarshal(body, &nureq)
		if err != nil {
			log.Printf("error parsing body: %s", err)

			respond.JSON(w, 400, "Invalid JSON body")
			return
		}

		if len(nureq.Password) == 0 || len(nureq.Username) == 0 {
			respond.JSON(w, 400, "Username/Password cannot be empty")
			return
		}

		if len(nureq.Password) > 64 || len(nureq.Username) > 64 {
			respond.JSON(w, 400, "Username or Password is too long")

			return
		}

		u := model.User{
			ID:       ulid.Make().String(),
			Username: nureq.Username,
			Password: nureq.Password,
			Money:    5.0,
		}

		err = a.Users.Insert(r.Context(), u)
		if err != nil {
			respond.Err(w, err)
			return
		}

		respond.JSON(w, 201, map[string]any{"Message": "ok", "User": u.Public()})
	})

	a.Router.Get("/users/{username}", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		username := chi.URLParam(r, "username")
		user, err := a.Users.FindBy(r.Context(), map[string]any{"username": username})
		if err != nil {
			respond.Err(w, err)
			return
		}
		claims := jwt.MapClaims{}
		p := jwt.NewParser()
		// A check for the token structure was added.
		_, _, err = p.ParseUnverified(token, claims)
		if err != nil {
			log.Printf("error parsing jwt %#v: %s", token, err)

			respond.JSON(w, 403, map[string]any{"error": fmt.Sprintf("invalid token: %#v", token)})
			return
		}
		// A check was added to verify that the username is in the token.
		if !isUsernameInToken(claims, username) {
			respond.JSON(w, 403, map[string]any{"error": fmt.Sprintf("token doesn't match with username: %#v", token)})
			return
		}

		respond.JSON(w, 200, user)
	})

	a.Router.Get("/users/{id}.{format:(html|tsv)}", func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "id")
		//The username was added to the path to perform the respective token verification.
		username := chi.URLParam(r, "username")
		token := r.URL.Query().Get("token")
		user, err := a.Users.FindBy(r.Context(), map[string]any{"id": userID})
		if err != nil {
			respond.Err(w, err)
			return
		}
		claims := jwt.MapClaims{}
		p := jwt.NewParser()

		// A check for the token structure was added.
		_, _, err = p.ParseUnverified(token, claims)
		if err != nil {
			log.Printf("error parsing jwt %#v: %s", token, err)

			respond.JSON(w, 400, map[string]any{"error": fmt.Sprintf("invalid token: %#v", token)})
			return
		}
		// A check was added to verify that the username is in the token.
		if !isUsernameInToken(claims, username) {
			respond.JSON(w, 403, map[string]any{"error": fmt.Sprintf("token doesn't match with username: %#v", token)})
			return
		}

		switch chi.URLParam(r, "format") {
		case "html":
			body := fmt.Sprintf("<!doctype html>\n<html><body>ID: %s Username: %s Amount: %f</body></html>", user.ID, user.Username, user.Money)

			respond.HTML(w, 200, body)
		case "tsv":
			tsv := fmt.Sprintf("ID\tUsername\tAmount\n%s\t%s\t%f\n", user.ID, user.Username, user.Money)

			respond.Respond(w, 200, "text/tsv", []byte(tsv))
		}
	})
	//login endpoint modified to transfer the sensible info with the body
	//the Http method changed from get to post
	a.Router.Post("/users/login", func(w http.ResponseWriter, r *http.Request) {
		//i recicled the logic from "Create user" to this method
		nureq := NewUserRequest{}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			respond.JSON(w, 400, "Failed to read body")
			return
		}

		err = json.Unmarshal(body, &nureq)
		if err != nil {
			log.Printf("error parsing body: %s", err)

			respond.JSON(w, 400, "Invalid JSON body")
			return
		}

		username := nureq.Username
		password := nureq.Password

		user, err := a.Users.FindBy(r.Context(), map[string]any{"username": username})
		if err != nil {
			respond.Err(w, err)
			return
		}

		if user.Password != password {
			respond.Err(w, respond.ForbiddenErr)
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": user.Username,
		})

		ss, err := token.SignedString([]byte(user.Password))
		if err != nil {
			respond.Err(w, err)
			return
		}

		respond.JSON(w, 200, map[string]any{"token": ss})
	})

	a.Router.Get("/users/{username}/transfer", func(w http.ResponseWriter, r *http.Request) {
		fromUsername := chi.URLParam(r, "username")
		toUsername := r.URL.Query().Get("to")
		token := r.URL.Query().Get("token")
		amount, _ := strconv.ParseFloat(r.URL.Query().Get("amount"), 64)
		//It validates that the transfer amount is greater than 0.
		if !validateAmount(amount) {
			respond.JSON(w, 400, map[string]any{"error": fmt.Sprintf("invalid amount: %#v", amount)})
			return
		}
		// It validates that the source user is different from the destination user.
		if validateUsername(fromUsername, toUsername) {
			respond.JSON(w, 400, map[string]any{"error": fmt.Sprintf("The target username is the same as the source username: %#v", fromUsername)})
			return
		}

		claims := jwt.MapClaims{}
		p := jwt.NewParser()
		// A check for the token structure was added.
		_, _, err := p.ParseUnverified(token, claims)
		if err != nil {
			log.Printf("error parsing jwt %#v: %s", token, err)

			respond.JSON(w, 400, map[string]any{"error": fmt.Sprintf("invalid token: %#v", token)})
			return
		}

		// A check was added to verify that the username is in the token.
		if !isUsernameInToken(claims, fromUsername) {
			respond.JSON(w, 403, map[string]any{"error": fmt.Sprintf("token doesn't match with username: %#v", token)})
			return
		}

		from, err := a.Users.FindBy(r.Context(), map[string]any{"username": fromUsername})
		if err != nil {
			respond.Err(w, err)
			return
		}

		to, err := a.Users.FindBy(r.Context(), map[string]any{"username": toUsername})
		if err != nil {
			respond.Err(w, err)
			return
		}

		funds := from.Money
		from.Money -= amount
		if from.Money < 0 {
			respond.Err(w, respond.NewRequestError(400, fmt.Sprintf("Sorry, not enough funds to transfer (%v)", funds)))
			return
		}

		to.Money += amount

		err = a.Users.Update(r.Context(), *to, "money")
		if err != nil {
			respond.Err(w, err)
			return
		}

		err = a.Users.Update(r.Context(), *from, "money")
		if err != nil {
			respond.Err(w, err)
			return
		}

		respond.JSON(w, 200, map[string]any{"user": from})
	})

	return nil
}
