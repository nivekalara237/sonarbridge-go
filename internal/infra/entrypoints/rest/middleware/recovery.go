package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

/**
 Sert à empêcher qu'un panic fasse planter tout ton serveur HTTP.

 En cas de bug, tu obtiens la trace complète de la pile d'appels,
	ce qui facilite énormément le diagnostic tout en évitant que le
	serveur ne s'arrête.
*/

func Recovery(enabled bool, next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if enabled {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("panic: %v\n%s", err, debug.Stack())

					http.Error(
						w,
						"Internal Server Error",
						http.StatusInternalServerError,
					)
				}
			}()
		}

		next.ServeHTTP(w, r)
	})
}
