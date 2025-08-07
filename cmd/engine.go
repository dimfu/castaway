package main

import (
	"fmt"
	"net/http"

	"github.com/dimfu/castaway/internal/renderer"
	"github.com/dimfu/castaway/internal/store"
	"github.com/dimfu/castaway/internal/websocket"
	"github.com/dimfu/castaway/views"
	"github.com/gin-gonic/gin"
)

type app struct {
	engine *gin.Engine
	store  *store.Store
}

func newApp() *app {
	app := &app{
		store: store.New(),
	}
	app.engine = gin.Default()

	ginHtmlRenderer := app.engine.HTMLRender
	app.engine.HTMLRender = &renderer.HTMLTemplRenderer{FallbackHtmlRenderer: ginHtmlRenderer}

	app.setupRoutes()

	return app
}

func (a *app) setupRoutes() {
	hub := websocket.NewHub(a.store)
	go hub.Run()

	a.engine.GET("/ws/:key", func(ctx *gin.Context) {
		key := ctx.Param("key")
		if err := websocket.Serve(ctx, hub, key); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	})

	a.engine.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "", views.Index())
	})

	a.engine.POST("/init-upload", func(ctx *gin.Context) {
		secret := ctx.PostForm("secret")
		r, err := a.store.AddToRegistry(secret, "test_file.txt")
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		scheme := "http"
		if ctx.Request.TLS != nil {
			scheme = "https"
		}

		// only scheme + host (e.g., http://localhost:8080)
		baseURL := fmt.Sprintf("%s://%s", scheme, ctx.Request.Host)

		ctx.JSON(http.StatusOK, gin.H{
			"status": "Waiting for reciever",
			"url":    fmt.Sprintf("%s/dl/%s", baseURL, r.Key),
		})
	})

	a.engine.GET("/dl/:key", func(ctx *gin.Context) {
		key := ctx.Param("key")
		r, err := a.store.FindRegistry(key)
		if err != nil {
			ctx.AbortWithError(http.StatusNotFound, err)
		}
		ctx.HTML(http.StatusOK, "", views.Download(r.Filename, key))
	})
}

func (a *app) run() error {
	// TODO: add flag to configure the port or just use system environment
	return a.engine.Run(":8080")
}
