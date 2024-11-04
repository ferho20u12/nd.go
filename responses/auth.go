package responses

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Auth struct {
	Data      any        `json:"data"`
	UserData  any        `json:"userData"`
	Message   string     `json:"message"`
	Token     string     `json:"token"`
	ExpiresAt *time.Time `json:"expiresAt"`
	Status    int        `json:"status"`
}

func (a *Auth) OK(ctx *gin.Context, auth Auth) {
	ctx.JSON(http.StatusOK, Auth{
		UserData:  auth.UserData,
		Data:      auth.Data,
		Message:   auth.Message,
		ExpiresAt: auth.ExpiresAt,
		Token:     auth.Token,
		Status:    200,
	})
}

func (a *Auth) Error(ctx *gin.Context, e Error) {
	e.ErrorDetail.LoadDetail()
	if e.Code != 0 {
		e.Status = e.Code
	}
	ctx.JSON(e.Code, e)
	ctx.Abort()
}
