// Please do not use this code in production 😋️ It's only purpose is to
// demonstrate OpenShift (Kubernetes) liveness and readiness probes.
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var (
	ready = true
	live  = true
)

func toggleLiveness(l *log.Logger, hostname string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		live = !live
		l.Printf("Toggling pod liveness to %t.\n", live)

		w.WriteHeader(http.StatusOK)
		io.WriteString(w, fmt.Sprintf("Pod %s liveness is now %t", hostname, live))
	}
}

func livenessProbe(log *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !live {
			log.Println("Liveness probe invoked. Pod is NOT live.")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		log.Println("Liveness probe invoked. Pod is live.")
		w.WriteHeader(http.StatusOK)
	}
}

func toggleReadyness(l *log.Logger, hostname string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		ready = !ready
		l.Printf("Toggling pod readyness to %t.\n", ready)

		w.WriteHeader(http.StatusOK)
		io.WriteString(w, fmt.Sprintf("Pod %s readyness is now %t", hostname, ready))
	}
}

func readynessProbe(l *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !ready {
			l.Println("Readyness probe invoked. Pod is NOT ready.")
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		l.Println("Readyness probe invoked. Pod is ready.")
		w.WriteHeader(http.StatusOK)
	}
}

func pod(hostname string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, fmt.Sprintf("request was handled by pod %s", hostname))
	}
}

func main() {
	hostname := os.Getenv("HOSTNAME")
	l := log.New(os.Stdout, hostname+" ", log.Ldate|log.Ltime)

	mux := http.NewServeMux()
	mux.HandleFunc("/live", livenessProbe(l))
	mux.HandleFunc("/live/toggle", toggleLiveness(l, hostname))
	mux.HandleFunc("/ready", readynessProbe(l))
	mux.HandleFunc("/ready/toggle", toggleReadyness(l, hostname))
	mux.HandleFunc("/pod", pod(hostname))

	l.Printf("Listening on port 8080...")
	l.Fatal(http.ListenAndServe(":8080", mux))
}
