package template

import "net/http"

type Executable interface {
	ExecuteTemplate(rw http.ResponseWriter)
}
