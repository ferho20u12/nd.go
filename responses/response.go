package responses

var (
	Handler *Response
)

func Boot() {
	Handler = &Response{
		Api:  Api{},
		File: File{},
		Auth: Auth{},
	}
}

type Response struct {
	Api  Api
	File File
	Auth Auth
}
