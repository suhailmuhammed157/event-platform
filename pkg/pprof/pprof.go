package pprof

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
)

func Start(port string) {
	go func() {
		addr := fmt.Sprintf(":%s", port)

		log.Println("pprof listening on", addr)

		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Println("pprof server error:", err)
		}
	}()
}
