package fakeserver

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

type server http.Server

// Start the server
func Start() {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	r := mux.NewRouter()

	s := &server{
		Addr: "127.0.0.1:8297", // Integer("testtrack", 36).to_s[0..3] == 8297
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second,
		ReadTimeout:  time.Second,
		IdleTimeout:  time.Second,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}
	s.routes()

	srv := (*http.Server)(s)

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
	log.Printf("testtrack server running at %s", srv.Addr)

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}

func (s *server) handleGet(pattern string, responseFunc func() (interface{}, error)) {
	r := s.Handler.(*mux.Router)
	r.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		result, err := responseFunc()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		bytes, err := json.Marshal(result)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(bytes)
	}).Methods("GET")
}

func (s *server) handlePost(pattern string, actionFunc func([]byte) error) {
	r := s.Handler.(*mux.Router)
	r.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		requestBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = actionFunc(requestBytes)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}).Methods("POST")
}
