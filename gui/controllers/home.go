package controllers

import (
	bc "dummy-blockchain/blockchain"
	"encoding/json"
	"net/http"
	"text/template"
)

// HomeHandler is the HTTP handler to view the chain
func HomeHandler(blockchain *bc.Blockchain, me string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			homeGet(w, r, blockchain, me)
		case http.MethodPost:
			// to post flash
			homeGet(w, r, blockchain, me)
		}
	}
}

// GetChainHandler is the REST handler to get the chain
func GetChainHandler(blockchain *bc.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getChainREST(w, r, blockchain)
		}
	}
}

func homeGet(w http.ResponseWriter, r *http.Request, blockchain *bc.Blockchain, me string) {

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

// RenderHTTPError is a utility function to render a user-friendly error
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

func getChainREST(w http.ResponseWriter, r *http.Request, blockchain *bc.Blockchain) {

	resp := bc.GetCHainResponse{
		Numblocks:  len(blockchain.Chain),
		Blockchain: blockchain,
	}

	respJSON, err := json.MarshalIndent(resp, "", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respJSON)
}
