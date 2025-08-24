package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type db_interface struct {
	db *sql.DB
}

//

func InitNewDB() db_interface {
	var err error
	var db *sql.DB
	connStr := "host=localhost port=5432 user=postgres password=admin dbname=AuthDB sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to the database!")

	return db_interface{db}
}

// func (s *Supervisor) Login(username, password string) ret {

// 	// Retrieve user from the database
// 	var storedHash string
// 	query := `SELECT password_hash FROM users WHERE username = $1`
// 	err := s.db.QueryRow(query, username).Scan(&storedHash)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			log.Println(err)
// 			return ret{false, "User not found"}
// 		}
// 		log.Println(err)
// 		return ret{false, "Login query failed"}
// 	}

// 	// Compare the provided password with the stored hash
// 	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
// 	if err != nil {
// 		log.Println(err)
// 		return ret{false, err.Error()}
// 	}
// 	return ret{true, "Login success"}
// }

// func (s *Supervisor) SignUp(username, password, email string) ret {

// 	// Hash the password
// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// 	if err != nil {
// 		log.Println(err)
// 		return ret{false, "Problem with password hashing"}
// 	}

// 	// Insert user into the database
// 	query := `INSERT INTO users (username, password_hash, email) VALUES ($1, $2, $3)`
// 	_, err = s.db.Exec(query, username, string(hashedPassword), email)
// 	if err != nil {
// 		log.Println(err)
// 		return ret{false, "User already exist, try to login"}
// 	}
// 	return ret{true, "Sign-up success"}
// }

func main() {
	db_interface := InitNewDB()
	defer db_interface.db.Close()

	r := mux.NewRouter()
	// r.HandleFunc("/signup", db_interface.SingUp).Methods("POST")
	// r.HandleFunc("/login", db_interface.Login).Methods("POST")

	fmt.Println("PostgresSQL server is on localhost:8080")
	log.Fatal(http.ListenAndServe("localhost:8080", r))
}
