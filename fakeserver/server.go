package fakeserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Force every biz logic operation acquire the same lock so nobody's reading or
// writing inconsistent state from/to the filesystem
var mutex sync.Mutex

var logger *log.Logger

type server struct {
	router *mux.Router
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("%s - %s %s", r.RemoteAddr, r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func createCors() *cors.Cors {
	return cors.New(cors.Options{
		AllowCredentials: true,
		AllowedHeaders:   []string{"authorization"},
		AllowOriginFunc: func(origin string) bool {
			allowedOrigins, ok := os.LookupEnv("TESTTRACK_ALLOWED_ORIGINS")
			if ok {
				fmt.Println(origin)
				fmt.Println(allowedOrigins)
				for _, allowedOrigin := range strings.Split(allowedOrigins, ",") {
					allowedOrigin = strings.Trim(allowedOrigin, " ")
					if strings.HasSuffix(origin, allowedOrigin) {
						return true
					}
				}
			} else {
				fmt.Println(allowedOrigins)
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
	})
}

// Start the server
func Start(port int) {
	handler := CreateHandler()

	listenOn := fmt.Sprintf("127.0.0.1:%d", port)
	logger.Printf("testtrack server listening on %s", listenOn)
	logger.Fatalf("fatal - %s", http.ListenAndServe(listenOn, handler))
}

// CreateHandler (exposed for testing)
func CreateHandler() http.Handler {
	logger = log.New(os.Stdout, "", log.LstdFlags)

	r := mux.NewRouter()

	s := &server{router: r}
	s.routes()

	r.Use(loggingMiddleware)

	return createCors().Handler(r)
}

func (s *server) handleGet(pattern string, responseFunc func() (interface{}, error)) {
	s.router.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		result, err := responseFunc()
		mutex.Unlock()
		if err != nil {
			logger.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		bytes, err := json.Marshal(result)
		if err != nil {
			logger.Println(err)
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
			logger.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if result == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		bytes, err := json.Marshal(result)
		if err != nil {
			logger.Println(err)
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
