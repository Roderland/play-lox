package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"plugin"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: play-lox [glox.so]")
		os.Exit(64)
	}

	r := playground()

	// Listen and Server in 0.0.0.0:8080
	_ = r.Run(":8080")
}

// playground
func playground() *gin.Engine {
	r := gin.Default()

	r.StaticFile("/", "./static/index.html")
	r.StaticFile("/index", "./static/index.html")
	r.StaticFile("/index.html", "./static/index.html")
	r.StaticFile("/style.css", "./static/style.css")
	r.StaticFile("/jquery.js", "./static/jquery.js")

	r.POST("/run", func(c *gin.Context) {
		req := &struct {
			Code string `json:"code"`
		}{}

		if err := c.BindJSON(req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{})
		}

		pipe := make(chan string, 1)
		go do(req.Code, pipe)

		select {
		case <-time.After(3 * time.Second):
			c.JSON(http.StatusOK, gin.H{"result": "拒绝执行非法代码"})
		case result := <-pipe:
			c.JSON(http.StatusOK, gin.H{"result": result})
		}

	})

	return r
}

func do(code string, pipe chan string) {
	play, buffer := loadPlugin(os.Args[1])
	go func(code string, pipe chan string) {
		play(code)
		pipe <- buffer.String()
		close(pipe)
	}(code, pipe)
}

//
// load the application "glox"  from a plugin file "glox.so"
//
func loadPlugin(filename string) (func(string), *bytes.Buffer) {
	p, err := plugin.Open(filename)
	if err != nil {
		log.Fatalf("cannot load plugin %v", filename)
	}
	xplay, err := p.Lookup("Play")
	if err != nil {
		log.Fatalf("cannot find Play in %v", filename)
	}
	play := xplay.(func(string))
	xbuf, err := p.Lookup("Buf")
	if err != nil {
		log.Fatalf("cannot find Buf in %v", filename)
	}
	buf := xbuf.(*bytes.Buffer)

	return play, buf
}
