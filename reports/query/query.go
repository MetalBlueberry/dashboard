package query

import (
	"net/http"
)

type User interface {
}

type Query struct {
	internalName   string
	AutorizedUsers map[string]bool
}

func (q Query) Name() string {
	return q.internalName
}

func (q Query) IsAutorized(user string) bool {
	value, ok := q.AutorizedUsers[user]
	return ok && value
}

func (q Query) GetHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}
