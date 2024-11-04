package foundation

import (
	"framework-back/nucleo-de-diagnostico/configuration"
)

var (
	App Application
)

func init() {
	app := &Application{}
	app.Config = configuration.NewConfiguration(".env")
	App = *app
}

type Application struct {
	Config *configuration.Configuration
}

func NewApplication() Application {
	return App
}

func (app *Application) Boot() {

}
