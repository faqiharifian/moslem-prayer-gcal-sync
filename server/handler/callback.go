package handler

import (
	"net/http"

	"golang.org/x/oauth2"

	"github.com/faqiharifian/moslem-prayer-gcal-sync/config"
	"github.com/faqiharifian/moslem-prayer-gcal-sync/prayertime"
)

type CallbackHandler struct {
	oauthCfg   *oauth2.Config
	prayerTime *prayertime.PrayerTime
	tokCh      chan *oauth2.Token
}

func NewCallbackHandler(cfg config.Config, prayerTime *prayertime.PrayerTime, tokCh chan *oauth2.Token) *CallbackHandler {
	return &CallbackHandler{oauthCfg: cfg.Oauth2, prayerTime: prayerTime, tokCh: tokCh}
}

func (h *CallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	errMsg := r.URL.Query().Get("error")
	if errMsg != "" {
		http.Error(w, errMsg+": "+r.FormValue("error_description"), http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")

	tok, err := h.oauthCfg.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "failed to retreive token, please try again after sometime", http.StatusBadRequest)
		return
	}

	h.tokCh <- tok

	http.Redirect(w, r, "https://calendar.google.com", http.StatusTemporaryRedirect)
}
