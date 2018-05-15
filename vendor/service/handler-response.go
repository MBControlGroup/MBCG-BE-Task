package service

import (
	"net/http"

	"github.com/unrolled/render"
)

func response(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func response_mem(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
