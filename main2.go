package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
)

const cookieName = "user"

type User struct {
	Id       int
	Username string
	Password string
}

type Server struct {
	Users    []*User
	Sessions sync.Map
}

func NewServer() *Server {
	var server Server
	server.Users = []*User{
		&User{1, "john", "secret"},
	}
	return &server
}

func (server *Server) Run() error {
	mux := http.NewServeMux()
	// route naar url
	mux.HandleFunc("/", server.RootHandler)

	mux.HandleFunc("/Locatie", server.LocatieHandler)

	mux.HandleFunc("/Login", server.Login)

	mux.HandleFunc("/Booking", server.Booking)

	return http.ListenAndServe(":8080", mux)
}

func (server *Server) getUser(username, password string) *User {
	for _, u := range server.Users {
		if u.Username == username && u.Password == password {
			return u
		}
	}
	return nil
}

func (server *Server) checkLogin(r *http.Request) *User {
	// get usename and password
	username := r.FormValue("username")
	password := r.FormValue("password")

	user := server.getUser(username, password)
	if user == nil {
		return nil
	}
	fmt.Printf("user %d %s\n", user.Id, user.Username)
	return user
}

func (server *Server) checkSession(r *http.Request) *User {
	for _, cookie := range r.Cookies() {
		if cookie.Name == cookieName {
			v, ok := server.Sessions.Load(cookie.Value)
			if !ok {
				return nil
			}
			user, ok := v.(*User)
			if ok {
				return user
			}
		}
	}
	return nil
}

func (server *Server) RootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `
		<h1>Fonteyn Vakantieparken</h1>
		<p>Vakantie parken voor een onvergeetelijke vakantie.</p>
		<ul>
			<li><a href="/Locatie">Locatie</a></li>
			<li><a href="/Login">Login</a></li>
		</ul>
	`)
}

func (server *Server) LocatieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `
		<h1>Locatie</h1>
		<p>Wij zijn beschikbaar in meerder landen.</p>
		<ul>
			<li><a href="/">Home</a></li>
			<li><a href="/Login">Login</a></li>
		</ul>
	`)
}

func (server *Server) Booking(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	user := server.checkSession(r)
	if user == nil {
		http.Redirect(w, r, "/Login", http.StatusSeeOther)
		return
	}
	fmt.Fprintf(w, "Hello %s\n", user.Username)
}

func (server *Server) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
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
	} else if r.Method == "POST" {
		user := server.checkLogin(r)
		if user != nil {
			bytes := make([]byte, 24)
			rand.Read(bytes)
			sessionId := base64.RawURLEncoding.EncodeToString(bytes)
			server.Sessions.Store(sessionId, user)
			http.SetCookie(w, &http.Cookie{Name: cookieName, Value: sessionId})
			http.Redirect(w, r, "/Booking", http.StatusSeeOther)
			return
		} else {
			fmt.Fprintln(w, "Invalid username or password.")
		}
	}
}

func main() {
	server := NewServer()
	err := server.Run()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}
