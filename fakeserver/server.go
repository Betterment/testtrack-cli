package fakeserver

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Force every biz logic operation acquire the same lock so nobody's reading or
// writing inconsistent state from/to the filesystem
var mutex sync.Mutex

type server struct {
	router *mux.Router
}

// Start the server
func Start(listenOn string) {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	r := mux.NewRouter()

	s := &server{router: r}
	s.routes()

	// Run our server in a goroutine so that it doesn't block.
	fmt.Printf("testtrack server listening on %s\n", listenOn)
	log.Fatal(http.ListenAndServe(listenOn, cors.New(cors.Options{
		AllowCredentials: true,
		AllowedHeaders:   []string{"authorization"},
		AllowOriginFunc: func(origin string) bool {
			allowedOrigins, ok := os.LookupEnv("TESTTRACK_ALLOWED_ORIGINS")
			if ok {
				for _, allowedOrigin := range strings.Split(allowedOrigins, ",") {
					allowedOrigin = strings.Trim(allowedOrigin, " ")
					if strings.HasSuffix(origin, allowedOrigin) {
						return true
					}
				}
			} else {
				// .test cannot be registered so we allow it by default
				if strings.HasSuffix(origin, ".test") {
					return true
				}
			}
			if origin == "localhost" {
				return true
			}
			ip := net.ParseIP(origin)
			if ip != nil && ip.IsLoopback() {
				return true
			}
			return false
		},
	}).Handler(r)))
}

func (s *server) handleGet(pattern string, responseFunc func() (interface{}, error)) {
	s.router.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		result, err := responseFunc()
		mutex.Unlock()
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
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}).Methods("GET")
}

func (s *server) handlePost(pattern string, actionFunc func(*http.Request) (interface{}, error)) {
	s.router.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		result, err := actionFunc(r)
		mutex.Unlock()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if result == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		bytes, err := json.Marshal(result)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}).Methods("POST")
}

func (s *server) handlePostReturnNoContent(pattern string, actionFunc func(*http.Request) error) {
	s.handlePost(pattern, func(r *http.Request) (interface{}, error) {
		err := actionFunc(r)
		return nil, err
	})
}
