package app

import (
	"encoding/gob"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
)

type LoginTemplate struct {
	Msg  string
	Type string
}

type IndexTemplate struct {
	CsvFiles []string
}

var Templates *template.Template

func init() {

	ReloadTemplates()
	gob.Register(LoginTemplate{})
	go MonitorTemplates()

}
func ReloadTemplates() {
	var err error
	Templates, err = template.ParseGlob("page/**/*.html")
	if err != nil {
		log.Panic(err)
	}

	_, err = Templates.ParseGlob("page/*.html")
	if err != nil {
		log.Panic(err)
	}

	log.Print("Defined Templates")
	log.Print(Templates.DefinedTemplates())
}

func MonitorTemplates() {

	//DIRTY IMPLEMENTATION, only 2 level of recursion, I need a better idea :)
	files2, err := filepath.Glob("./*/*/*.html")
	if err != nil {
		log.Print(err)
		log.Print("Cant find html files, STOP monitor Templates for changes")
		return
	}

	files1, err := filepath.Glob("./*/*.html")
	if err != nil {
		log.Print(err)
		log.Print("Cant find html files, STOP monitor Templates for changes")
		return
	}

	// inotifywait -mq -e modify users.json
	cmd := exec.Command("inotifywait", "-mq", "-e", "modify")

	cmd.Args = append(cmd.Args, files1...)
	cmd.Args = append(cmd.Args, files2...)

	log.Print(cmd.Args)

	reader, err := cmd.StdoutPipe()
	if err != nil {
		log.Print(err)
		log.Print("Cant Read stdout, STOP monitor Templates for changes")
		return
	}

	buff := make([]byte, 50)

	err = cmd.Start()
	if err != nil {
		log.Print(err)
		log.Print("cant start, STOP monitor Templates for changes")
		return
	}

	for {
		_, err := reader.Read(buff)
		if err != nil {
			log.Print(err)
			log.Print("Error reading, STOP monitor Templates for changes")
			return
		}
		ReloadTemplates()
	}
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
			case "index.html":
				csvs, err := filepath.Glob("page/reports/table/*.csv")
				if err != nil {
					http.NotFound(w, r)
					return
				}
				data = IndexTemplate{
					CsvFiles: csvs,
				}
			}

			err := Templates.ExecuteTemplate(w, file, data)
			if err != nil {
				http.NotFound(w, r)
			}
		}
	})
}
