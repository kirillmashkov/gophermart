package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/kirillmashkov/gophermart/internal/app"
	"github.com/kirillmashkov/gophermart/internal/httpserver/router"
	"github.com/kirillmashkov/gophermart/internal/logger"
	"github.com/kirillmashkov/gophermart/internal/model"
)

func main() {
	err := logger.Initialize()
	if err != nil {
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
		log.SetPrefix("ERROR: ")
		log.Println("Can't init logger")
		panic(err)
	}

	flag.Parse()
	err = app.Initialize()
	if err != nil {
		panic(err)
	}

	defer app.Close()

	err = http.ListenAndServe(app.ServerConf.Host, router.Serv())
	if err != nil {
		panic(err)
	}

	if model.OrderNumChanToCreate != nil {
		close(model.OrderNumChanToCreate)
	}

	if model.RequestToAccrualChan != nil {
		close(model.RequestToAccrualChan)
	}

}
