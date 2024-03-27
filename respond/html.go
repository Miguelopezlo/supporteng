package respond

import "net/http"

func HTML(w http.ResponseWriter, status int, body string) {
	Respond(w, status, "text/html; charset=utf-8", []byte(body))
}
