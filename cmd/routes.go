package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/dimfu/castaway/internal/store"
	"github.com/dimfu/castaway/internal/websocket"
	"github.com/dimfu/castaway/views"
	"github.com/gin-gonic/gin"
)

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
		fileSize, _ := strconv.Atoi(ctx.PostForm("file_size"))
		r, err := a.store.AddToRegistry(secret, &store.FileInfo{
			Name: ctx.PostForm("file_name"),
			Size: fileSize,
			Type: ctx.PostForm("file_type"),
		})
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
			"chunks": r.BuildChunks(),
			"url":    fmt.Sprintf("%s/store/%s", baseURL, r.Key),
		})
	})

	// download page
	a.engine.GET("/store/:key", func(ctx *gin.Context) {
		key := ctx.Param("key")
		r, err := a.store.FindRegistry(key)
		if err != nil {
			ctx.AbortWithError(http.StatusNotFound, err)
		}
		ctx.HTML(http.StatusOK, "", views.Download(r.FileInfo.Name, key, r.Ready))
	})

	// direct download endpoint
	a.engine.GET("/dl/:key", func(ctx *gin.Context) {
		key := ctx.Param("key")
		r, err := a.store.FindRegistry(key)
		if err != nil {
			ctx.AbortWithError(http.StatusNotFound, err)
		}
		ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, r.FileInfo.Name))
		ctx.Header("Content-Type", "application/octet-stream")
		ctx.Header("Content-Length", fmt.Sprintf("%d", r.FileInfo.Size))

		transferred := 0
		for {
			chunk := r.DequeueChunk()
			_, err := ctx.Writer.Write(chunk)

			if err != nil {
				ctx.AbortWithError(http.StatusInternalServerError, err)
				log.Printf("Error writing chunk: %v", err)
				break
			}

			if len(chunk) == 0 {
				log.Println("No chunk to send")
				break
			}

			if f, ok := ctx.Writer.(http.Flusher); ok {
				f.Flush()
			}

			transferred += len(chunk)
			if transferred >= r.FileInfo.Size {
				log.Println("Finished transferring all data to client")
				break
			}
		}

		a.store.ClearRegistry(r.Key)
	})
}
