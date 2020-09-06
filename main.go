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



func startWorker(stopChan <-chan struct{}, command string, args ...string) (doneChan <-chan error) {
	cmd := exec.Command(command, args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	done := make(chan error)

	go func() {
		defer close(done)
		err := cmd.Start()
		if err != nil {
			log.WithField("error", err).Error("Failed starting command")
			done <- err
			return
		}
		log.Info("worker started")

		go func() {
			<-stopChan
			log.Info("send worker SIGTERM")
			cmd.Process.Signal(syscall.SIGTERM)
		}()

		err = cmd.Wait()
		done <- err
	}()

	return done
}

func main() {
	var port	int
	var debug	bool
	flag.IntVar(&port, "port", 8080, "port for healthcheck")
	flag.BoolVar(&debug, "debug", false, "print debug log")

	flag.Parse()
	if len(flag.Args()) == 0 {
		log.Fatal("command not found. usages: worker_runner COMMANDS ARGS")
	}

	if debug {
		log.SetLevel(log.DebugLevel)
	}

	log.Info("start worker")
	done := startWorker(signals.SetupSignalHandler(), flag.Args()[0], flag.Args()[1:]...)

	log.Info("start server")
	srv := healthcheck.StartServerAsync(port)

	err := <- done
	if err != nil {
		log.WithField("error", err).Error("Worker ended with error")
	}
	srv.Shutdown(context.TODO())
	log.Info("server ended")

	if err != nil {
		log.Exit(1)
	}


}