package main

import (
	"flag"
	"fmt"

	"github.com/dimfu/castaway/internal/renderer"
	"github.com/dimfu/castaway/internal/store"
	"github.com/gin-gonic/gin"
)

const DEFAULT_CHUNK_SIZE = 2 // mb
const DEFAULT_PORT = 8080

type app struct {
	engine *gin.Engine
	store  *store.Store
}

var (
	chunksize = flag.Int64("chunksize", DEFAULT_CHUNK_SIZE, "chunk limit size for the registry, default to 2 MiB")
	port      = flag.Int64("port", DEFAULT_PORT, fmt.Sprintf("using port %d by default\n", DEFAULT_PORT))
)

//go:generate npm --prefix .. run build
func main() {
	flag.Parse()

	app := &app{
		store: store.New(*chunksize * 1024 * 1024), // convert to MiB
	}

	// disable gin's routing logs
	gin.SetMode("release")
	app.engine = gin.New()
	app.engine.Use(gin.Recovery())

	ginHtmlRenderer := app.engine.HTMLRender
	app.engine.HTMLRender = &renderer.HTMLTemplRenderer{FallbackHtmlRenderer: ginHtmlRenderer}

	app.setupRoutes()

	if err := app.engine.Run(fmt.Sprintf(":%d", *port)); err != nil {
		panic(err)
	}
}
