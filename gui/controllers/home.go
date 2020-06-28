package controllers

import (
	bc "dummy-blockchain/blockchain"
	"net/http"
	"text/template"
)

// HomeHandler ...
func HomeHandler(blockchain *bc.Blockchain, me string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			HomeGet(w, r, blockchain, me)
		case http.MethodPost:
			// to post flash
			HomeGet(w, r, blockchain, me)
		}
	}
}

// HomeGet ...
func HomeGet(w http.ResponseWriter, r *http.Request, blockchain *bc.Blockchain, me string) {

	if r.URL.Path != "/" {
		RenderHTTPError(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	t, err := template.ParseFiles("gui/views/layout.gohtml", "gui/views/home.gohtml")
	if err != nil {
		RenderHTTPError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type viewData struct {
		Title string
		BC    *bc.Blockchain
	}

	p := &viewData{
		Title: "Home",
		BC:    blockchain,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// RenderHTTPError ...
func RenderHTTPError(w http.ResponseWriter, message string, code int) {

	var viewData = struct {
		Title   string
		Message string
		Code    int
	}{
		"Something bad happened",
		message,
		code,
	}

	t, err := template.ParseFiles("gui/views/error.gohtml")
	if err != nil {
		http.Error(w, message, code)
		return
	}

	err = t.ExecuteTemplate(w, "error", viewData)
	if err != nil {
		http.Error(w, message, code)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
}
