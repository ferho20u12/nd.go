package routerContract

import "github.com/gin-gonic/gin"

type Resource struct {
	Index   bool
	Store   bool
	Show    bool
	Update  bool
	Destroy bool
}

type ResourceController interface {
	// Resources() Resource
	Index(ctx *gin.Context)
	Store(ctx *gin.Context)
	Show(ctx *gin.Context)
	Update(ctx *gin.Context)
	Destroy(ctx *gin.Context)
}
