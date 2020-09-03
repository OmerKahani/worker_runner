package main

import (
	"bytes"
	"context"
	"flag"
	"github.com/Riskified/worker_runner/internal/healthcheck"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"syscall"
)



func startWorker(stopChan <-chan struct{}, command string, args ...string) (doneChan <-chan struct{}) {
	cmd := exec.Command(command, args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	done := make(chan struct{})

	go func() {
		err := cmd.Start()
		if err != nil {
			log.WithField("error", err).Error("Failed starting command")
			close(done)
		}
		log.Info("worker started")

		go func() {
			<-stopChan
			log.Info("send worker SIGTERM")
			cmd.Process.Signal(syscall.SIGTERM)
		}()

		cmd.Wait()
		close(done)
	}()



	return done
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "port for healthcheck")

	flag.Parse()
	if len(flag.Args()) == 0 {
		log.Fatal("command not found. usages: worker_runner COMMANDS ARGS")
	}

	log.Info(flag.Args())

	log.Info("start worker")
	done := startWorker(signals.SetupSignalHandler(), flag.Args()[0], flag.Args()[1:]...)

	log.Info("start server")
	srv := healthcheck.StartServerAsync(port)

	<- done
	log.Info("worker ended")
	log.Info("shutdown server")
	srv.Shutdown(context.TODO())
	log.Info("server ended")

}