package facades

import "github.com/ferho20u12/nd.go/responses"

func Response() *responses.Response {
	return responses.Handler
}
