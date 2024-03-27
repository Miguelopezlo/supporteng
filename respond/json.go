package respond

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, v any) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)

	if err := enc.Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Respond(w, status, "application/json", buf.Bytes())
}

func JSONObject[T Object](w http.ResponseWriter, status int, val T) {
	JSON(w, status, val.Public())
}

func JSONObjects[T Object](w http.ResponseWriter, status int, vals []T) {
	list := make([]map[string]any, 0, len(vals))

	for _, v := range vals {
		list = append(list, v.Public())
	}

	JSON(w, status, list)
}

func Err(w http.ResponseWriter, err error) {
	if rerr, ok := err.(Error); ok {
		JSON(w, rerr.StatusCode(), map[string]any{"error": rerr.Error()})
		return
	}

	log.Printf("Unhandled error: %s", err)

	JSON(w, 500, map[string]any{"error": "Something went wrong. Please retry later"})
}
