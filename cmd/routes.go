package main

import (
	"fmt"
	"log"
	"mime"
	"net/http"
	"strconv"

	"github.com/dimfu/castaway/assets"
	"github.com/dimfu/castaway/internal/store"
	"github.com/dimfu/castaway/internal/websocket"
	"github.com/dimfu/castaway/views"
	"github.com/gin-gonic/gin"
)

func (a *app) setupRoutes() {
	hub := websocket.NewHub(a.store)
	go hub.Run()

	// serve static files from the embeded assets
	assets := assets.Assets
	a.engine.StaticFS("/public", http.FS(assets))

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

	a.engine.POST("/detect-mime", func(ctx *gin.Context) {
		fileHeader, err := ctx.FormFile("chunk")
		if err != nil {
			ctx.JSON(400, gin.H{"error": "chunk not provided"})
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			ctx.JSON(500, gin.H{"error": "failed to open chunk"})
			return
		}
		defer file.Close()

		buf := make([]byte, 512)
		n, err := file.Read(buf)
		if err != nil && err.Error() != "EOF" {
			ctx.JSON(500, gin.H{"error": "failed to read chunk"})
			return
		}

		mimeType := http.DetectContentType(buf[:n])
		exts, err := mime.ExtensionsByType(mimeType)
		if err != nil {
			ctx.JSON(500, gin.H{"error": "failed to get mime type extension"})
			return
		}

		ctx.JSON(200, gin.H{
			"mime": mimeType,
			"ext":  exts[0],
		})
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
			ctx.Redirect(http.StatusMovedPermanently, "/")
			return
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
		ctx.Header("Content-Type", r.FileInfo.Type)
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
	})
}
