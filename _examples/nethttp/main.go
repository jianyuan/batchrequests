package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jianyuan/batchrequests"
)

func main() {
	http.HandleFunc("/simpleget", func(w http.ResponseWriter, r *http.Request) {
		log.Println("hit /simpleget")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"hello":"world"}`)
	})

	log.Fatal(http.ListenAndServe(":8080", batchrequests.New("/batch", nil)))
}
