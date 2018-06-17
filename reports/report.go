package reports

import (
	"github.com/gorilla/mux"
)

var RegisteredReports map[string]*Report

type Report interface {
	Name() string
	IsAutorized(user string) bool
	RegisterHandlers(*mux.Router) error
}
