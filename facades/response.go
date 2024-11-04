package facades

import "framework-back/nucleo-de-diagnostico/responses"

func Response() *responses.Response {
	return responses.Handler
}
