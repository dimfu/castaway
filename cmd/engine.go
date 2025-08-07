package main

import (
	"net/http"

	"github.com/dimfu/castaway/internal/renderer"
	"github.com/gin-gonic/gin"
)

type app struct {
	engine *gin.Engine
}

func newApp() *app {
	app := &app{}
	app.engine = gin.Default()

	ginHtmlRenderer := app.engine.HTMLRender
	app.engine.HTMLRender = &renderer.HTMLTemplRenderer{FallbackHtmlRenderer: ginHtmlRenderer}

	app.setupRoutes()

	return app
}

func (a *app) setupRoutes() {
	a.engine.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello World")
	})
}

func (a *app) run() error {
	// TODO: add flag to configure the port or just use system environment
	return a.engine.Run(":8080")
}
