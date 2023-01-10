package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

type DatabaseConfig struct {
	Username string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"dbname"`
}

var dbc string

func logError(err error) {
	if err != nil {
		log.Println(err)
	}
}

func main() {
	logFile, err := os.OpenFile("errors.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Println(err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	mux := http.NewServeMux()
	// route naar url
	mux.HandleFunc("/", RootHandler)

	mux.HandleFunc("/Locatie", LocatieHandler)

	mux.HandleFunc("/Login", LoginHandler)

	mux.HandleFunc("/Booking", BookingHandler)

	// read json file
	configBytes, err := ioutil.ReadFile("Db.json")
	if err != nil {
		logError(err)
	}

	// setting up new struct
	var config DatabaseConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		log.Println(err)
	}

	// make connection database
	dbc = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.Username, config.Password, config.Host, config.Port, config.Database)
	db, err := sql.Open("mysql", dbc)
	if err != nil {
		logError(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Println(err)
	}

	http.ListenAndServe(":8080", mux)
}
func checkLogin(r *http.Request) bool {
	// get usename and password
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Connect to the database
	db, err := sql.Open("mysql", dbc)
	if err != nil {
		log.Println(err)
		return false
	}
	defer db.Close()

	// Execute a SELECT query to check if the username and password are correct
	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE username=? AND password=?", username, password).Scan(&userID)
	if err != nil {
		log.Println(err)
		return false
	}

	// If the username and password are correct, insert a new session into the database
	_, err = db.Exec("INSERT INTO sessions (user_id, start_time, end_time) VALUES (?, NOW(), NOW())", userID)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}
func RootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `
		<h1>Fonteyn Vakantieparken</h1>
		<p>Vakantie parken voor een onvergeetelijke vakantie.</p>
		<ul>
			<li><a href="/Locatie">Locatie</a></li>
			<li><a href="/Login">Login</a></li>
		</ul>
	`)
}

func LocatieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `
		<h1>Locatie</h1>
		<p>Wij zijn beschikbaar in meerder landen.</p>
		<ul>
			<li><a href="/">Home</a></li>
			<li><a href="/Login">Login</a></li>
		</ul>
	`)
}

func BookingHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	if !checkLogin(r) {
		http.Redirect(w, r, "/Login", http.StatusSeeOther)
		return
	}
}
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// display login
	fmt.Fprintln(w, `
        <h1>Fonteyn Vakantieparken</h1>
        <h2>Inloggen</h2>
        <form action="/Login" method="POST">
            <label for="username">Gebruikersnaam:</label><br>
            <input type="text" id="username" name="username"><br>
            <label for="password">Wachtwoord:</label><br>
            <input type="password" id="password" name="password"><br><br>
            <input type="submit" value="Inloggen">
        </form>
    `)
	if r.Method == "POST" {
		if checkLogin(r) {
			http.Redirect(w, r, "/Booking", http.StatusSeeOther)
			return
		} else {
			fmt.Fprintln(w, "Invalid username or password.")
		}
	}
}
