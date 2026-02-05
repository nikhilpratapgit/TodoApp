package handler

import (
	"database/sql"
	"errors"
	//"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nikhilpratapgit/TodoApp/database/dbHelper"
	"github.com/nikhilpratapgit/TodoApp/middleware"
	"github.com/nikhilpratapgit/TodoApp/models"
	"github.com/nikhilpratapgit/TodoApp/utils"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var registerUser models.RegisterUser

	if parseErr := utils.ParseBody(r.Body, &registerUser); parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "failed to parse request body")
		return
	}

	if err := utils.Validate.Struct(registerUser); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "validation failed")
		return
	}

	exists, existsErr := dbHelper.IsUserExists(registerUser.Email)
	if existsErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, existsErr, "failed to check user existence")
		return
	}
	if exists {
		utils.RespondError(w, http.StatusBadRequest, nil, "user already exists")
		return
	}

	hashPassword, err := utils.HashPassword(registerUser.Password)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "failed while hashing password")
		return
	}

	userID, saveErr := dbHelper.CreateUser(registerUser.Name, registerUser.Email, hashPassword)
	if saveErr != nil {
		utils.RespondError(w, http.StatusNotFound, saveErr, "failed to create user")
		return
	}

	sessionID, sessionErr := dbHelper.CreateUserSession(userID)
	if sessionErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, sessionErr, "failed to create user session")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, struct {
		Message string `json:"message"`
		Token   string `json:"token"`
	}{
		Message: "user created successfully",
		Token:   sessionID,
	})
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var req models.LoginUser

	if parseErr := utils.ParseBody(r.Body, &req); parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "invalid request body")
		return
	}
	if err := utils.Validate.Struct(req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "validation failed")
		return
	}

	userID, userErr := dbHelper.GetUserByEmail(req.Email, req.Password)
	if userErr != nil {
		utils.RespondError(w, http.StatusUnauthorized, userErr, "failed to find user")
		return
	}

	sessionID, sessionErr := dbHelper.CreateUserSession(userID)
	if sessionErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, sessionErr, "failed to create user session")
		return
	}
	token, err := utils.GenerateJWT(userID, sessionID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to generate token")
		return
	}
	utils.RespondJSON(w, http.StatusCreated, struct {
		Token string `json:"token"`
	}{
		Token: token,
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	userCtx := middleware.UserContext(r)
	sessionID := userCtx.SessionID

	if err := dbHelper.DeleteSessionByToken(sessionID); err != nil {
		utils.RespondError(w, http.StatusUnauthorized, err, "Invalid user")
		return
	}
	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "User Logout Successfully",
	})
}

func CreateTodo(w http.ResponseWriter, r *http.Request) {
	var todoRequest models.CreateTodo

	userCtx := middleware.UserContext(r)
	userID := userCtx.UserID

	if parseErr := utils.ParseBody(r.Body, &todoRequest); parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "failed to parse body")
		return
	}

	if err := utils.Validate.Struct(todoRequest); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "validation failed")
		return
	}

	if todoRequest.ExpiringAt.Before(time.Now()) {
		utils.RespondError(w, http.StatusBadRequest, nil, "provided time and date is wrong")
		return
	}

	todo, err := dbHelper.CreateTodo(userID, todoRequest.Name, todoRequest.Description, todoRequest.ExpiringAt)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to create todo")
		return
	}
	utils.RespondJSON(w, http.StatusCreated, todo)
}
func GetAllTodos(w http.ResponseWriter, r *http.Request) {
	completeStr := r.URL.Query().Get("status")
	expiringAtStr := r.URL.Query().Get("expiringAt")
	search := r.URL.Query().Get("search")

	userCtx := middleware.UserContext(r)
	userID := userCtx.UserID

	//model make complete to string
	var complete bool
	complete = utils.ParseBool(completeStr)

	expiringAt, err := utils.ParseExpiringAt(expiringAtStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "invalid time")
	}

	todos, err := dbHelper.GetTodos(userID, search, expiringAt, complete)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch todos")
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Todos []models.Todos `json:"todos"`
	}{
		Todos: todos,
	})
}
func GetTodoById(w http.ResponseWriter, r *http.Request) {
	todoID := chi.URLParam(r, "id")
	if todoID == "" {
		utils.RespondError(w, http.StatusBadRequest, nil, "todo id is required")
		return
	}

	userCtx := middleware.UserContext(r)
	userID := userCtx.UserID
	todo, err := dbHelper.GetTodoByID(todoID, userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to fetch todo")
		return
	}
	utils.RespondJSON(w, http.StatusOK, todo)
}

func DeleteTodoById(w http.ResponseWriter, r *http.Request) {
	todoID := chi.URLParam(r, "id")
	if todoID == "" {
		utils.RespondError(w, http.StatusBadRequest, nil, "required todo ID")
		return
	}

	userCtx := middleware.UserContext(r)
	userID := userCtx.UserID

	err := dbHelper.DeleteTodoById(userID, todoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.RespondError(w, http.StatusNotFound, err, "todo not found")
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to delete todo")
		return
	}
	utils.RespondJSON(w, http.StatusOK, "todo deleted successfully")
}
func UpdateTodoById(w http.ResponseWriter, r *http.Request) {
	todoID := chi.URLParam(r, "id")
	if todoID == "" {
		utils.RespondError(w, http.StatusBadRequest, nil, "required Todo id")
		return
	}

	var todo models.UpdateTodoRequest
	if err := utils.ParseBody(r.Body, &todo); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "invalid request body")
		return
	}
	if err := utils.Validate.Struct(todo); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "validation failed")
		return
	}

	userCtx := middleware.UserContext(r)
	userID := userCtx.UserID

	err := dbHelper.UpdateTodoById(todo.Name, todo.Description, todo.Complete, todo.ExpiringAt, todoID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.RespondError(w, http.StatusNotFound, err, "todo not found")
			return
		}
		utils.RespondError(w, http.StatusInternalServerError, err, "failed to update todo")
		return
	}
	utils.RespondJSON(w, http.StatusOK, "updated successfully")
}

//func CompleteTodo(w http.ResponseWriter, r *http.Request) {
//	userCtx := middleware.UserContext(r)
//	userTokenID := userCtx.SessionID
//
//	userID, err := dbHelper.GetUserBySession(userTokenID)
//	if err != nil {
//		utils.RespondError(w, http.StatusUnauthorized, err, "invalid session")
//		return
//	}
//
//	Todos, err := dbHelper.CompleteTodos(userID)
//	if err != nil {
//		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to fetch todos")
//		return
//	}
//
//	utils.RespondJSON(w, http.StatusOK, struct {
//		Todos []models.Todos `json:"todos"`
//	}{
//		Todos: Todos,
//	})
//
//}
//func IncompleteTodo(w http.ResponseWriter, r *http.Request) {
//	userCtx := middleware.UserContext(r)
//	userTokenID := userCtx.SessionID
//	userID, err := dbHelper.GetUserBySession(userTokenID)
//	if err != nil {
//		utils.RespondError(w, http.StatusNotFound, err, "invalid session")
//		return
//	}
//	Todos, err := dbHelper.IncompleteTodos(userID)
//	if err != nil {
//		utils.RespondError(w, http.StatusInternalServerError, err, "failed to fetch todos")
//	}
//	utils.RespondJSON(w, http.StatusOK, struct {
//		Todos []models.Todos `json:"todos"`
//	}{Todos: Todos})
//}
//func UpcomingTodos(w http.ResponseWriter, r *http.Request) {
//	daysParam := r.URL.Query().Get("days")
//	var days int
//	if daysParam == "" {
//		days = 0
//	} else {
//		days, err := strconv.Atoi(daysParam)
//		//fmt.Println(daysParam)
//		if err != nil || days < 0 {
//			utils.RespondError(w, http.StatusBadRequest, nil, "days must be a positive number")
//			return
//		}
//	}
//
//	userCtx := middleware.UserContext(r)
//	userTokenID := userCtx.SessionID
//
//	userID, err := dbHelper.GetUserBySession(userTokenID)
//	if err != nil {
//		utils.RespondError(w, http.StatusNotFound, err, "invalid session")
//		return
//	}
//	Todos, err := dbHelper.UpcomingTodos(userID, days)
//	if err != nil {
//		utils.RespondError(w, http.StatusInternalServerError, err, "failed to fetch upcoming Todos")
//		return
//	}
//	utils.RespondJSON(w, http.StatusOK, struct {
//		Todos []models.Todos `json:"todos"`
//	}{Todos: Todos})
//
//}
