package main

import (
	"fmt"
	"log"

	"bitbucket.org/truora-tests/supporteng/app"
	"bitbucket.org/truora-tests/supporteng/controller"
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	a, err := app.NewApp()
	if err != nil {
		log.Printf("failed to initialize the app: %s", err)
		return
	}

	// Setup Controllers
	panicIfErr(SetupController[controller.Users](a))
	//error  starting handled
	err = a.Start()
	if err != nil {
		fmt.Println("something was worng, app didnt start")
	}
}

func SetupController[T controller.Controller](a *app.App) error {
	var m T

	return m.Setup(a)
}
