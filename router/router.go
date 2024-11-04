package router

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	middlewareContract "framework-back/nucleo-de-diagnostico/contracts/middlewares"
	routerContract "framework-back/nucleo-de-diagnostico/contracts/router"
	"framework-back/nucleo-de-diagnostico/foundation"

	"github.com/gookit/color"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	RouterManager Router
)

type Router struct {
	Engine      *gin.Engine
	Default     *gin.RouterGroup
	Middlewares []middlewareContract.Middleware
}

type RouteOptions struct {
	BasePath                  string
	GroupName                 string
	DontUseDefaultMiddlewares bool
	DisableLog                bool
	Middlewares               []middlewareContract.Middleware
	Resource                  *routerContract.Resource
}

type RouteOptionFunction struct {
	GroupName                 string
	PrefixName                string
	DontUseDefaultMiddlewares bool
	HttpMethod                string
	Function                  func(*gin.Context)
}

func Boot() {
	if name, ok := foundation.App.Config.Get("app.timezone", "America/Mexico_City").(string); ok {
		location, err := time.LoadLocation(name)
		if err != nil {
			color.Redln("error en la configuracion app.timezone : ", err)
			os.Exit(1)
		}
		time.Local = location
	}
	if mode, ok := foundation.App.Config.Get("app.env", "dev").(string); ok && strings.EqualFold(mode, "release") {
		gin.SetMode(gin.ReleaseMode)
	}
	var router = gin.New()
	if debug, ok := foundation.App.Config.Get("app.debug", true).(bool); ok && debug {
		router.Use(CustomLogger(colorable.NewColorableStdout()), gin.Recovery())
	}

	config := cors.DefaultConfig()
	config.AllowMethods = foundation.App.Config.Get("cors.allowed_methods", []string{"*"}).([]string)
	config.AllowOrigins = foundation.App.Config.Get("cors.allowed_origins", []string{"*"}).([]string)
	config.AllowHeaders = foundation.App.Config.Get("cors.allowed_headers", []string{"*"}).([]string)
	config.AllowCredentials = foundation.App.Config.Get("cors.supports_credentials", false).(bool)
	router.Use(cors.New(config))

	RouterManager.Engine = router
	RouterManager.Default = RouterManager.Engine.Group("/api")
}

func (router *Router) RegisterDefaultsMiddlewares(middlewares []middlewareContract.Middleware) {
	router.Middlewares = append(router.Middlewares, middlewares...)
}

func (router *Router) RegisterResource(option RouteOptions, controller routerContract.ResourceController) {
	r := RouterManager.Default
	if option.BasePath != "" {
		r = router.Engine.Group(option.BasePath)
	}
	r = r.Group(option.GroupName)
	if !option.DontUseDefaultMiddlewares {
		for _, middleware := range router.Middlewares {
			r.Use(middleware.Middleware)
		}
	}
	for _, middleware := range option.Middlewares {
		r.Use(middleware.Middleware)
	}
	// if option.DisableLog {
	// 	// r.Use(NoLoggingMiddleware())
	// }

	if option.Resource == nil {
		r.GET("/", controller.Index)
		r.POST("/", controller.Store)
		r.GET("/:id", controller.Show)
		r.PUT("/:id", controller.Update)
		r.DELETE("/:id", controller.Destroy)
	} else {
		if option.Resource.Index {
			r.GET("/", controller.Index)
		}
		if option.Resource.Store {
			r.POST("/", controller.Store)
		}
		if option.Resource.Show {
			r.GET("/:id", controller.Show)
		}
		if option.Resource.Update {
			r.PUT("/:id", controller.Update)
		}
		if option.Resource.Destroy {
			r.DELETE("/:id", controller.Destroy)
		}
	}
}

func (router *Router) RegisterFunctions(option RouteOptions, optionsFunctions []RouteOptionFunction) {
	r := RouterManager.Default
	if option.BasePath != "" {
		r = router.Engine.Group(option.BasePath)
	}
	r = r.Group(option.GroupName)

	if option.GroupName != "" {
		r = RouterManager.Default.Group(option.GroupName)
	}
	// if option.DisableLog {
	// 	// r.Use(NoLoggingMiddleware())
	// }
	for _, optionFunction := range optionsFunctions {
		httpMethod := optionFunction.HttpMethod
		function := optionFunction.Function
		prefixName := optionFunction.PrefixName
		if !option.DontUseDefaultMiddlewares && !optionFunction.DontUseDefaultMiddlewares {
			for _, middleware := range router.Middlewares {
				r.Use(middleware.Middleware)
			}
			for _, middleware := range option.Middlewares {
				r.Use(middleware.Middleware)
			}
		}

		if prefixName != "" {
			prefixName = "/" + prefixName
		}
		switch httpMethod {

		case http.MethodGet:
			r.GET(prefixName, function)
		case http.MethodPost:
			r.POST(prefixName, function)
		case http.MethodPut:
			r.PUT(prefixName, function)
		case http.MethodDelete:
			r.DELETE(prefixName, function)
		case http.MethodOptions:
			r.OPTIONS(prefixName, function)
		default:
			color.Redf("handler metodo incorrecto %s", httpMethod)
			os.Exit(1)
		}
	}
}

func (router *Router) Run() *gin.Engine {
	config := foundation.App.Config
	port, ok := config.Get("http.port", 8080).(int)
	if !ok {
		color.Redf("puerto mal declarado")
		os.Exit(1)
	}
	addr := fmt.Sprintf(":%d", port)
	if config.Get("http.tls.enable", false).(bool) {
		certFile := config.Get("http.tls.ssl.cert", "").(string)
		keyFile := config.Get("http.tls.ssl.key", "").(string)
		RouterManager.Engine.RunTLS(addr, certFile, keyFile)
	} else {
		RouterManager.Engine.Run(addr)
	}
	return RouterManager.Engine
}

const (
	green   = "\033[32m"
	yellow  = "\033[33m"
	red     = "\033[31m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	reset   = "\033[0m"
)

// CustomLogger retorna un middleware que escribe logs coloreados
func CustomLogger(out io.Writer) gin.HandlerFunc {
	isTerm := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	if !isTerm {
		return gin.Logger()
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		statusCode := c.Writer.Status()

		var statusColor, methodColor string
		switch {
		case statusCode >= 200 && statusCode < 300:
			statusColor = green
		case statusCode >= 300 && statusCode < 400:
			statusColor = cyan
		case statusCode >= 400 && statusCode < 500:
			statusColor = yellow
		default:
			statusColor = red
		}

		switch c.Request.Method {
		case "GET":
			methodColor = blue
		case "POST":
			methodColor = cyan
		case "PUT":
			methodColor = yellow
		case "DELETE":
			methodColor = red
		default:
			methodColor = reset
		}

		if raw != "" {
			path = path + "?" + raw
		}

		// Aquí pintamos el prefijo y la fecha en azul
		fmt.Fprintf(out, "%s[GIN] %v%s |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n",
			blue, end.Format("2006/01/02 - 15:04:05"), reset,
			statusColor, statusCode, reset,
			latency,
			c.ClientIP(),
			methodColor, c.Request.Method, reset,
			path,
		)
	}
}
