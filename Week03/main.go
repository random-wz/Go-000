package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var Stop bool

func main() {
	var stop = make(chan bool)
	var done = make(chan error, 2)
	go HttpServer1(stop, done)
	go HttpServer2(stop, done)
	for i := 0; i < cap(done); i++ {
		if err := <-done; err != nil {
			close(stop)
			log.Println(err)
		}
	}
	if Stop {
		close(stop)
		log.Println("Good Bye!!!")
	}
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-sig:
		close(stop)
		log.Println(err)
	case <-stop:
		log.Println("Good Bye!!!")
	}
}

type test struct{}

func (test) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("Hello World1"))
}

func HttpServer1(stop chan bool, done chan error) {
	var handler http.Handler = test{}
	done <- server(":80", handler, stop)
}

func HttpServer2(stop chan bool, done chan error) {
	var handler http.Handler = test{}
	done <- server(":8080", handler, stop)
}

func server(addr string, handle http.Handler, stop chan bool) error {
	s := http.Server{
		Handler: handle,
		Addr:    addr,
	}

	go func() {
		<-stop
		_ = s.Shutdown(context.Background())
	}()
	return s.ListenAndServe()
}
