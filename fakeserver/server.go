package fakeserver

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// BindTo is the IP and port we're binding to
const BindTo = "127.0.0.1:8297"

type server struct {
	router *mux.Router
}

// Start the server
func Start() {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	r := mux.NewRouter()

	s := &server{router: r}
	s.routes()

	// Run our server in a goroutine so that it doesn't block.
	fmt.Printf("testtrack server binding to %s\n", BindTo)
	log.Fatal(http.ListenAndServe(BindTo, cors.New(cors.Options{
		AllowCredentials: true,
		AllowedHeaders:   []string{"authorization"},
		AllowOriginFunc: func(origin string) bool {
			dotTest := strings.HasSuffix(origin, ".test")
			localhost := origin == "localhost"
			ip := net.ParseIP(origin)
			return dotTest || localhost || (ip != nil && ip.IsLoopback())
		},
	}).Handler(r)))
}

func (s *server) handleGet(pattern string, responseFunc func() (interface{}, error)) {
	s.router.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
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

func (s *server) handlePost(pattern string, actionFunc func(*http.Request) error) {
	s.router.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		err := actionFunc(r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}).Methods("POST")
}
