package server

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"

	"golang.org/x/oauth2"

	"github.com/faqiharifian/moslem-prayer-gcal-sync/config"
	"github.com/faqiharifian/moslem-prayer-gcal-sync/constant"
	"github.com/faqiharifian/moslem-prayer-gcal-sync/prayertime"
	"github.com/faqiharifian/moslem-prayer-gcal-sync/server/handler"
)

type Server struct {
	srv *http.Server
}

func New(cfg config.Config, prayerTime *prayertime.PrayerTime, tokCh chan *oauth2.Token) *http.Server {
	server := &http.Server{Addr: cfg.Host}
	callbackHandler := handler.NewCallbackHandler(cfg, prayerTime, tokCh)
	http.Handle(constant.CallbackPath, callbackHandler)
	return server
}

func Start(ctx context.Context, srv *http.Server) {
	ctx, cancelFn := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	go func() {
		defer cancelFn()
		_ = srv.ListenAndServe()
	}()

	<-ctx.Done()

	srv.Close()
}
