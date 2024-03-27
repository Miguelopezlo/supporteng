package app

import (
	"context"
	"fmt"
	"net/http"

	"bitbucket.org/truora-tests/supporteng/model"
	"github.com/chaisql/chai"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type App struct {
	DB     *chai.DB
	Router *chi.Mux

	Users *model.Storage[model.User]
}

func NewApp() (*App, error) {
	db, err := chai.Open("supporteng.db")
	if err != nil {
		return nil, err
	}

	app := &App{
		DB: db,
	}
	//error  configurating router handled
	err = app.setupRouter()
	if err != nil {
		return nil, err
	}

	app.Users = model.NewStorage[model.User](db)
	app.Users.CreateTable(context.Background())

	return app, nil
}

func (a *App) setupRouter() error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		//error writing "welcome" handled
		_, err := w.Write([]byte("welcome"))
		if err != nil {
			fmt.Println(err)
			return
		}
	})

	a.Router = r
	return nil
}

func (a App) Start() error {
	return http.ListenAndServe("127.0.0.1:4000", a.Router)
}
