package responses

import "github.com/gin-gonic/gin"

type File struct {
	FileName string `json:"fileName,omitempty"`
	Status   int    `json:"status"`
}

func (f *File) OK(ctx *gin.Context, statusCode int, file File) {
	file.Status = 200
	ctx.JSON(statusCode, file)
	ctx.Next()
}

func (f *File) Error(ctx *gin.Context, statusCode int, e Error) {
	ctx.JSON(statusCode, e)
	ctx.Abort()
}
