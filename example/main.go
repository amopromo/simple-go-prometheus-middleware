package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	sgpm "github.com/amopromo/simple-go-prometheus-middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	addr := os.Getenv("SGPM_LISTEN_ON")

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello")
	})

	middleware := sgpm.Middleware(sgpm.Config{
		Prefix:          "sgpm_example",
		HandlerLabel:    "handler",
		MethodLabel:     "methond",
		StatusCodeLabel: "code",
		DurationBuckets: prometheus.DefBuckets,
	})

	h := middleware(mux)

	go func() {
		log.Printf("Serving metrics at: %s", "0.0.0.0:9090")
		log.Println(http.ListenAndServe("0.0.0.0:9090", promhttp.Handler()))
	}()

	log.Println("Listening on: ", addr)
	log.Fatal(http.ListenAndServe(addr, h))
}
