package app

import (
	"encoding/json"
	"github.com/gorilla/sessions"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

var store = sessions.NewCookieStore([]byte("fkdiruasprjrkduufjasdlrntnasf"))

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
				if userData, exist := users[userName]; !exist || userData.Blocked {
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
				loginTemplate.AddFlash(LoginTemplate{Msg: "Unauthorized"})
				loginTemplate.Values["data"] = LoginTemplate{Msg: "Unauthorized"}
				err = loginTemplate.Save(r, w)
				if err != nil {
					log.Println(err)
				}
				log.Print(w.Header())
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

type User struct {
	Password string
	Blocked  bool
}

func (u User) checkPassword(password string) bool {
	return u.Password == password
}

var users map[string]User

func init() {
	ReloadUsers()

	go func() {
		// inotifywait -mq -e modify users.json
		cmd := exec.Command("inotifywait", "-mq", "-e", "modify", "users.json")

		reader, err := cmd.StdoutPipe()
		if err != nil {
			log.Print(err)
			log.Print("STOP user.json monitor for changes")
		}

		buff := make([]byte, 20)

		err = cmd.Start()
		if err != nil {
			log.Print(err)
			log.Print("STOP user.json monitor for changes")
		}

		for {
			_, err := reader.Read(buff)
			if err != nil {
				log.Print(err)
				log.Print("STOP user.json monitor for changes")
			}
			ReloadUsers()
		}
	}()

}

func ReloadUsers() {
	file, err := ioutil.ReadFile("users.json")
	if err != nil {
		log.Print(err)
	}
	if len(file) == 0 {
		return
	}
	err = json.Unmarshal(file, &users)
	if err != nil {
		log.Print(err)
	}
	log.Print("Reloaded users.json")
}

func doLogin(user, password string) bool {
	ReloadUsers()
	if data, exist := users[strings.ToLower(user)]; exist {
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
		// fmt.Fprint(w, "OK")
		session.Values["auth"] = true
		session.Values["user"] = r.FormValue("username")
		session.Save(r, w)

		log.Print(session.Values)
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		// fmt.Fprint(w, "ERROR")
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
