package dbHelper

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/nikhilpratapgit/TodoApp/database"
	"github.com/nikhilpratapgit/TodoApp/models"
	"github.com/nikhilpratapgit/TodoApp/utils"
)

func IsUserExists(email string) (bool, error) {
	SQL := `SELECT count(*) > 0
			FROM users
			WHERE email = TRIM(LOWER($1))
			  AND archived_at IS NULL;`

	var exists bool
	err := database.Todo.Get(&exists, SQL, email)
	return exists, err
}
func CreateUser(name, email, password string) (string, error) {
	SQL := `INSERT INTO users(name, email, password)
			VALUES ($1, TRIM(LOWER($2)), $3) RETURNING id;`
	var userID string
	err := database.Todo.Get(&userID, SQL, name, email, password)
	return userID, err
}
func CreateUserSession(userID string) (string, error) {
	SQL := `INSERT INTO user_session(user_id)
			VALUES ($1) RETURNING id;`
	var sessionID string
	err := database.Todo.Get(&sessionID, SQL, userID)
	if err != nil {
		return "", err
	}
	return sessionID, nil
}
func GetUserByEmail(email, password string) (string, error) {
	SQL := `
		SELECT id, password
		FROM users
		WHERE email = $1 AND archived_at IS NULL
		RETURNING id,password;
	`

	var user models.UserAuth
	err := database.Todo.Get(&user, SQL, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("no user exist")
		}
		return "", err
	}
	if err := utils.CheckPassword(user.Password, password); err != nil {
		return "", errors.New("invalid credentials")
	}
	return user.ID, nil
}
func DeleteSessionByToken(Token string) error {
	SQL := `UPDATE user_session
			SET archived_at = NOW()
			WHERE id = $1
			AND archived_at IS NULL
			`

	result, err := database.Todo.Exec(SQL, Token)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("Invalid Session")
	}
	return nil
}

//	func GetUserBySession(Token string) (string, error) {
//		SQL := `SELECT user_id FROM user_session
//	           where id =$1 AND archived_at IS NULL `
//		var userID string
//		err := database.Todo.Get(&userID, SQL, Token)
//		if err != nil {
//			if errors.Is(err, sql.ErrNoRows) {
//				return "", errors.New("invalid or expired session")
//			}
//			return "", err
//		}
//		return userID, nil
//	}
func CreateTodo(userID, name, description string, expiringAt time.Time) (*models.Todos, error) {
	SQL := `INSERT INTO todos (user_id,name,description,expiring_at) 
			VALUES ($1,$2,$3,$4) RETURNING id,created_at;`
	todo := &models.Todos{
		UserId:      userID,
		Name:        name,
		Description: description,
		ExpiringAt:  expiringAt,
	}
	err := database.Todo.QueryRow(SQL, userID, name, description, expiringAt).Scan(&todo.Id, &todo.CreatedAt)
	if err != nil {
		return nil, err
	}
	return todo, nil
}
func GetTodos(userID, name string, date time.Time, complete bool) ([]models.Todos, error) {
	SQL := `
			SELECT id,
			       user_id,
			       name,
			       description,
			       complete,
			       expiring_at,
				   created_at
			FROM todos
			WHERE user_id =$1
			AND (
			    $2::boolean IS NULL or complete=$2
			)
			AND (
			    $3::TIMESTAMPTZ IS NULL or expiring_at<=$3
			)
			AND (
			    $4::TEXT IS NULL OR name LIKE'%'||$4||'%'
			)
			order by expiring_at
			`
	var todos []models.Todos

	err := database.Todo.Select(&todos, SQL, userID, complete, date, name)
	if err != nil {
		return nil, err
	}
	return todos, nil
}
func GetTodoByID(todoID, userID string) (*models.Todos, error) {
	SQL := `SELECT id,user_id,name,description,complete,expiring_at,created_at
			FROM todos where id = $1 
			AND user_id=$2 `
	var todo models.Todos

	err := database.Todo.Get(&todo, SQL, todoID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("todo not found")
		}
		return nil, err
	}
	return &todo, nil
}
func DeleteTodoById(userID, todoID string) error {
	SQL := `DELETE FROM todos WHERE id=$1 AND user_id =$2;`

	_, err := database.Todo.Exec(SQL, todoID, userID)
	if err != nil {
		return err
	}
	return nil
}
func UpdateTodoById(name, description, complete string, expiringAt string, todoID, userID string) error {
	SQL := `UPDATE todos 
			SET name=$1,description=$2,complete=$3,expiring_at=$4
			WHERE id=$5 
			and user_id=$6;`

	_, err := database.Todo.Exec(
		SQL, name, description, complete, expiringAt, todoID, userID)

	if err != nil {
		return err
	}
	return nil
}

//	func CompleteTodos(userID string) ([]models.Todos, error) {
//		SQL := `SELECT id, user_id,name,description,complete,expiring_at FROM todos
//	           WHERE user_id=$1 AND complete = TRUE;`
//
//		var todos []models.Todos
//
//		err := database.Todo.Select(&todos, SQL, userID)
//		if err != nil {
//			return nil, err
//		}
//		return todos, nil
//	}
//
//	func IncompleteTodos(userID string) ([]models.Todos, error) {
//		SQL := `
//			SELECT id,user_id,name,description,complete,expiring_at FROM todos
//	       WHERE user_id=$1 AND complete=FALSE;
//
// `
//
//		var todos []models.Todos
//
//		err := database.Todo.Select(&todos, SQL, userID)
//		if err != nil {
//			return nil, err
//		}
//		return todos, nil
//	}
//
//	func UpcomingTodos(userID string, days int) ([]models.Todos, error) {
//		SQL := `SELECT id, user_id, name, description, complete, expiring_at FROM todos
//	           WHERE user_id = $1
//		  		AND expiring_at between  NOW()
//		    	AND  NOW() + ($2 || 'days')::interval `
//
//		var todos []models.Todos
//
//		err := database.Todo.Select(&todos, SQL, userID, days)
//		fmt.Println(err)
//		if err != nil {
//			return nil, err
//		}
//
//		return todos, nil
//	}
func ValidateSession(sessionID string) (uuid.UUID, error) {
	SQL := `SELECT user_id from user_session where id=$1 AND archived_at IS NULL;`

	var userID uuid.UUID

	err := database.Todo.Get(&userID, SQL, sessionID)

	if err != nil {
		return uuid.Nil, errors.New("invalid session")
	}

	return userID, nil
}
