package main

import (
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	group, errCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		var handler http.Handler = HttpServer{}
		return Server(errCtx, ":80", handler)
	})
	group.Go(func() error {
		var handler http.Handler = HttpServer{}
		return Server(errCtx, ":8080", handler)
	})
	// 捕获信号
	var sig = make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	group.Go(func() error {
		select {
		case err := <-sig:
			cancel()
			return errors.New("Listen Signal exit: " + err.String())
		case <-errCtx.Done():
			return errors.New("context.Done exit")
		}
	})
	// 捕获err, group.Wait 会调用cancel,因此这里打印报错日志即可
	if err := group.Wait(); err != nil {
		log.Fatal(err.Error())
	}
}

type HttpServer struct{}

func (HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("Hello World1"))
}

func Server(ctx context.Context, addr string, handle http.Handler) error {
	s := http.Server{
		Handler: handle,
		Addr:    addr,
	}
	go func() {
		select {
		case <-ctx.Done():
			fmt.Printf("http server %s done\n", addr)
			_ = s.Shutdown(ctx)
		}
	}()
	return s.ListenAndServe()
}
