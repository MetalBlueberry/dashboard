package app

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/gorilla/sessions"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var Users map[string]User

type User struct {
	Password string
	Blocked  bool
}

func (u User) checkPassword(password string) bool {
	SERVER_SECRET, err := os.LookupEnv("SERVER_SECRET")
	if !err {
		panic("SERVER-SECRET secret is not defined")
	}

	hash := sha256.Sum256([]byte(password + SERVER_SECRET))
	log.Printf("hash %x", hash)
	return fmt.Sprintf("%x", hash) == strings.ToLower(u.Password)
}

var store sessions.Store
var SERVER_SECRET string

func init() {
	SERVER_SECRET, err := os.LookupEnv("SERVER_SECRET")
	if !err {
		panic("SERVER-SECRET secret is not defined")
	}

	store = sessions.NewCookieStore([]byte(SERVER_SECRET))
	ReloadUsers()
	go MonitorUsersFile()
}

func MonitorUsersFile() {
	// inotifywait -mq -e modify users.json
	cmd := exec.Command("inotifywait", "-mq", "-e", "modify", "users.json")

	reader, err := cmd.StdoutPipe()
	if err != nil {
		log.Print(err)
		log.Print("STOP user.json monitor for changes")
		return
	}

	buff := make([]byte, 20)

	err = cmd.Start()
	if err != nil {
		log.Print(err)
		log.Print("STOP user.json monitor for changes")
		return
	}

	for {
		_, err := reader.Read(buff)
		if err != nil {
			log.Print(err)
			log.Print("STOP user.json monitor for changes")
			return
		}
		ReloadUsers()
	}
}

func WithAuth(fn http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "login")
		if err != nil {
			log.Println(err)
		}

		if _, ok := session.Values["auth"]; ok == false {
			session.Values["auth"] = false
		}

		if userNameValue, ok := session.Values["user"]; ok && session.Values["auth"].(bool) {
			if userName, ok := userNameValue.(string); !ok {
				session.Values["auth"] = false
			} else {
				if userData, exist := Users[userName]; !exist || userData.Blocked {
					session.Values["auth"] = false
				}
			}
		}

		err = session.Save(r, w)
		if err != nil {
			log.Println(err)
		}

		if value, ok := session.Values["auth"].(bool); !(value && ok) {
			loginTemplate, err := store.New(r, "LoginTemplate")
			if err != nil {
				log.Println(err)
			} else {
				log.Print("Setting login template data")
				loginTemplate.AddFlash(LoginTemplate{Msg: "You must login first"})
				err = loginTemplate.Save(r, w)
				if err != nil {
					log.Println(err)
				}
			}

			http.Redirect(w, r, "/login/login.html", http.StatusFound)
			/*w.Write([]byte(`
			<script>
			alert('Please login')
			window.location='/login.html'
			</script>`))
			*/

			return
		}
		fn.ServeHTTP(w, r)
		return
	})
}

func ReloadUsers() {
	file, err := ioutil.ReadFile("users.json")
	if err != nil {
		log.Print(err)
	}
	if len(file) == 0 {
		return
	}
	err = json.Unmarshal(file, &Users)
	if err != nil {
		log.Print(err)
	}
	log.Print("Reloaded users.json")
}

func doLogin(user, password string) bool {
	ReloadUsers()
	if data, exist := Users[strings.ToLower(user)]; exist {
		if data.checkPassword(password) {
			return true
		}
	}
	return false
}

func Login(w http.ResponseWriter, r *http.Request) {
	log.Println("Atemp to login")

	session, err := store.Get(r, "login")
	if err != nil {
		log.Println(err)
	}

	if doLogin(r.FormValue("username"), r.FormValue("password")) {
		session.Values["auth"] = true
		session.Values["user"] = r.FormValue("username")
		session.Save(r, w)

		log.Print(session.Values)
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		loginTemplate, err := store.New(r, "LoginTemplate")
		if err != nil {
			log.Println(err)
		} else {
			log.Print("Setting login template data")
			loginTemplate.AddFlash(LoginTemplate{Msg: "User or password incorrect", Type: "alert-warning"})
			err = loginTemplate.Save(r, w)
			if err != nil {
				log.Println(err)
			}
			log.Print(w.Header())
		}

		http.Redirect(w, r, "/login/login.html", http.StatusFound)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "login")
	if err != nil {
		log.Println(err)
	}
	session.Options.MaxAge = -1
	session.Save(r, w)

	loginTemplate, err := store.New(r, "LoginTemplate")
	if err != nil {
		log.Println(err)
	} else {
		loginTemplate.AddFlash(LoginTemplate{Msg: "Logout successfully"})
		err = loginTemplate.Save(r, w)
		if err != nil {
			log.Println(err)
		}
		log.Print(w.Header())
	}

	http.Redirect(w, r, "/login/login.html", http.StatusFound)
}
