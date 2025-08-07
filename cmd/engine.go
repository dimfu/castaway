package main

import (
	"net/http"

	"github.com/dimfu/castaway/internal/renderer"
	"github.com/dimfu/castaway/internal/websocket"
	"github.com/dimfu/castaway/views"
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
	hub := websocket.NewHub()
	go hub.Run()

	a.engine.GET("/ws", func(ctx *gin.Context) {
		if err := websocket.Serve(ctx, hub); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	})

	a.engine.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "", views.Index())
	})

	a.engine.POST("init-upload", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
		})
	})
}

func (a *app) run() error {
	// TODO: add flag to configure the port or just use system environment
	return a.engine.Run(":8080")
}
