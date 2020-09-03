package healthcheck

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func healthcheck(w http.ResponseWriter, _ *http.Request) {
	log.Debug("got healthcheck")
	fmt.Fprintf(w, "ok\n")
}

const Work = "work"

type ServerStatus struct {
	Status string
	Error  error
}

func StartServerAsync(port int) *http.Server{
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port)}
	http.HandleFunc("/healthcheck", healthcheck)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.WithField("error", err).Error("error in ListenAndServe()")
		}
	}()

	return srv
}
