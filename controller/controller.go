package controller

import "bitbucket.org/truora-tests/supporteng/app"

type Controller interface {
	Setup(app *app.App) error
}
