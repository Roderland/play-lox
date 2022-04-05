package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"plugin"
	"sync"

	"github.com/gin-gonic/gin"
)

var lock sync.Mutex

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: play-lox [glox.so]")
		os.Exit(64)
	}

	play, buffer := loadPlugin(os.Args[1])

	r := playground(play, buffer)

	// Listen and Server in 0.0.0.0:8080
	_ = r.Run(":8080")
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

func playground(play func(string), buffer *bytes.Buffer) *gin.Engine {
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

		lock.Lock()
		play(req.Code)
		result := buffer.String()
		lock.Unlock()

		c.JSON(http.StatusOK, gin.H{"result": result})
	})

	return r
}
