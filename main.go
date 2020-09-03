package main

import (
	"bytes"
	"context"
	"github.com/Riskified/worker_runner/internal/healthcheck"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sync"
	"syscall"
)



func startWorker(stopChan <-chan struct{}, wg *sync.WaitGroup, command string, args ...string) {
	cmd := exec.Command(command, args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	go func() {
		err := cmd.Run()
		if err != nil {
			log.WithField("error", err).Error("cmd.Run() failed")
		}
		cmd.Wait()
		wg.Done()
		log.Info("worker ended")
	}()

	<-stopChan
	log.Info("send worker SIGTERM")
	cmd.Process.Signal(syscall.SIGTERM)
}

func main() {
	if len(os.Args) == 1 {
		log.Fatal("command not found. usages: worker_runner COMMANDS ARGS")
	}

	stopChan := signals.SetupSignalHandler()
	
	log.Info("start worker")
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go startWorker(stopChan, wg, os.Args[1], os.Args[2:]...)

	log.Info("start server")
	srv := healthcheck.StartServerAsync(8000)

	wg.Wait()
	log.Info("shutdown server")
	srv.Shutdown(context.TODO())
	log.Info("server ended")

}