package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sync"
	"syscall"
)

func healthcheck(w http.ResponseWriter, req *http.Request) {
	log.Print("got healthcheck")
	fmt.Fprintf(w, "ok\n")
}

func startWorker(stopChan <-chan struct{}, wg *sync.WaitGroup, command string, args ...string) {
	cmd := exec.Command(command, args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	go func() {
		err := cmd.Run()
		if err != nil {
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}
		cmd.Wait()
		wg.Done()
		log.Print("worker ended")
	}()

	<-stopChan
	log.Print("send worker SIGTERM")
	cmd.Process.Signal(syscall.SIGTERM)
}

func startServerAsync() *http.Server{
	srv := &http.Server{Addr: ":8000"}
	http.HandleFunc("/healthcheck", healthcheck)
	go func() {
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("ListenAndServe(): %v", err)
	}
	}()

	return srv
}

func main() {
	if len(os.Args) == 1 {
		log.Fatal("command not found. usages: worker_runner COMMANDS ARGS")
	}
	stopChan := signals.SetupSignalHandler()

	wg := &sync.WaitGroup{}

	log.Print("start worker")
	wg.Add(1)
	go startWorker(stopChan, wg, os.Args[1], os.Args[2:]...)

	log.Print("start server")
	srv := startServerAsync()

	wg.Wait()
	log.Print("shutdown server")
	srv.Shutdown(context.TODO())
	log.Print("server ended")

}