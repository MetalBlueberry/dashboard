package app

import (
	"encoding/gob"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type LoginTemplate struct {
	Msg  string
	Type string
}

type IndexTemplate struct {
	Reports map[string]string
}

var Templates *template.Template

func init() {
	var err error
	Templates, err = template.ParseGlob("page/**/*.html")
	if err != nil {
		log.Panic(err)
	}

	Templates.ParseGlob("page/*.html")

	log.Print("Defined Templates")
	log.Print(Templates.DefinedTemplates())

	gob.Register(LoginTemplate{})

}

func TemplateServer() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print(r.URL.Path)
		if len(r.URL.Path) == 1 {
			err := Templates.ExecuteTemplate(w, "index.html", nil)
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

			err := Templates.ExecuteTemplate(w, file, data)
			if err != nil {
				http.NotFound(w, r)
			}
		}
	})
}
