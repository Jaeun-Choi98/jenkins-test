package main

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	server := &http.Server{
		Addr:    ":8088",
		Handler: NewController(),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	server.Shutdown(ctx)
	wg.Wait()
	log.Println("server shut down")
	os.Exit(0)
}

func NewController() http.Handler {
	r := gin.Default()
	r.Use(SpaHandlerRoot("build", "index.html"))
	return r
}

func SpaHandlerRoot(staticPath, indexPath string) gin.HandlerFunc {

	return func(c *gin.Context) {
		url := c.Request.URL.Path
		ext := filepath.Ext(url)

		blockedExt := map[string]bool{
			".git": true,
			".ini": true,
			".txt": true,
		}
		if blockedExt[ext] {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		cleanPath := filepath.Clean(url)
		absPath := filepath.Join(staticPath, cleanPath)

		_, err := os.Stat(absPath)

		if err != nil {

			if errors.Is(err, fs.ErrNotExist) {
				http.ServeFile(c.Writer, c.Request, filepath.Join(staticPath, indexPath))
				return
			} else {
				http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		http.FileServer(http.Dir(staticPath)).ServeHTTP(c.Writer, c.Request)
	}
}
