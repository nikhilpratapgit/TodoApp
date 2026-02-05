package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nikhilpratapgit/TodoApp/handler"
	"github.com/nikhilpratapgit/TodoApp/middleware"

	//"github.com/nikhilpratapgit/TodoApp/utils"
	"github.com/nikhilpratapgit/TodoApp/utils"
)

type Server struct {
	chi.Router
	server *http.Server
}

func SetupRoutes() *Server {
	router := chi.NewRouter()
	router.Route("/v1", func(v1 chi.Router) {
		v1.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			utils.RespondJSON(w, http.StatusOK, map[string]string{
				"status": "Server is Running",
			})
		})

		//public
		v1.Post("/register", handler.RegisterUser)
		v1.Post("/login", handler.LoginUser)

		v1.Group(func(v1 chi.Router) {
			v1.Use(middleware.Auth)
			v1.Route("/user", func(user chi.Router) {
				user.Group(userRoutes)
			})
			//private
			v1.Get("/todos", handler.GetAllTodos)
			v1.Get("/todo/{id}", handler.GetTodoById)
			v1.Post("/todo", handler.CreateTodo)
			v1.Put("/todo/{id}", handler.UpdateTodoById)
			v1.Delete("/todo/{id}", handler.DeleteTodoById)
			//v1.Get("/todos-complete", handler.CompleteTodo)
			//v1.Get("/todos-incomplete", handler.IncompleteTodo)
			//v1.Get("/upcoming-todos", handler.UpcomingTodos)
		})
	})
	return &Server{
		Router: router,
	}
}
