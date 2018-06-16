package app

import (
	"encoding/gob"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type LoginTemplate struct {
	Msg string
}

var templates *template.Template

func init() {
	var err error
	templates, err = template.ParseGlob("page/**/*.html")
	if err != nil {
		log.Panic(err)
	}

	templates.ParseGlob("page/*.html")

	log.Print("Defined Templates")
	log.Print(templates.DefinedTemplates())

	gob.Register(LoginTemplate{})

}

func TemplateServer(dir http.Dir) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print(r.URL.Path)
		if len(r.URL.Path) == 1 {
			err := templates.ExecuteTemplate(w, "index.html", nil)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		} else {
			split := strings.Split(r.URL.Path, "/")
			file := split[len(split)-1]
			log.Print(file)
			var data interface{}

			switch file {
			case "login.html":
				loginTemplate, err := store.Get(r, "LoginTemplate")
				if err == nil {
					flashes := loginTemplate.Flashes()
					if len(flashes) > 0 {
						data = flashes[0].(LoginTemplate)
					}
					log.Printf("Login template data loaded %+v", data)
					loginTemplate.Save(r, w)
				}
			}

			err := templates.ExecuteTemplate(w, file, data)
			if err != nil {
				http.NotFound(w, r)
			}
		}
	})

}
