package facades

import "framework-back/nucleo-de-diagnostico/router"

func Route() *router.Router {
	return &router.RouterManager
}
