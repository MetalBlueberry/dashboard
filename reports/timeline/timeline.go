package timeline

import (
	"github.com/gorilla/mux"
	"github.com/metalblueberry/dashboard/app"
	"net/http"
)

type Timeline struct {
	InternalName   string
	AutorizedUsers map[string]bool
	Query          string
	updating       bool
}

func LoadFromData(rawdata interface{}) Timeline {
	data := rawdata.(map[string]interface{})
	t := Timeline{}
	t.InternalName = data["InternalName"].(string)
	t.Query = data["Query"].(string)
	t.AutorizedUsers = make(map[string]bool)
	for key, value := range data["AutorizedUsers"].(map[string]interface{}) {
		t.AutorizedUsers[key] = value.(bool)
	}

	return t
}

func (t Timeline) Name() string {
	return t.InternalName
}

func (t Timeline) IsAutorized(user string) bool {
	value, ok := t.AutorizedUsers[user]
	return ok && value
}

func (t Timeline) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if t.updating {
		app.Templates.ExecuteTemplate(w, "timeline-updating.html", nil)

		return
	}
	app.Templates.ExecuteTemplate(w, "timeline.html", t)
	app.TemplateServer().ServeHTTP(w, r)
}

func (t Timeline) RegisterHandlers(router *mux.Router) error {
	router.Path("/reports/" + t.InternalName).Handler(t)
	return nil
}
