package foundation

import (
	"github.com/ferho20u12/nd.go/configuration"
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
