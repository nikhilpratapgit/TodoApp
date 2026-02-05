package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/nikhilpratapgit/TodoApp/handler"
)

func userRoutes(r chi.Router) {
	r.Group(func(user chi.Router) {
		user.Delete("/logout", handler.Logout)
	})
}
