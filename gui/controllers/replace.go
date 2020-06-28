package controllers

import (
	bc "dummy-blockchain/blockchain"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
)

// ReplaceHandler ...
func ReplaceHandler(blockchain *bc.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ReplaceGet(w, r, blockchain)
		case http.MethodPost:
			ReplacePost(w, r, blockchain)
		}
	}
}

// ReplaceGet ...
func ReplaceGet(w http.ResponseWriter, r *http.Request, blockchain *bc.Blockchain) {

	t, err := template.ParseFiles("gui/views/layout.gohtml", "gui/views/replace.gohtml")
	if err != nil {
		RenderHTTPError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flashStr := ""
	err = r.ParseForm()
	if err == nil {
		flashStr = r.PostForm.Get("flash")
	}

	type viewData struct {
		Title string
		Flash string
	}

	p := &viewData{
		Title: "Home",
		Flash: flashStr,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ReplacePost ...
func ReplacePost(w http.ResponseWriter, r *http.Request, blockchain *bc.Blockchain) {

	replaced, err := blockchain.ReplaceChain()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var flashMsg string
	if replaced {
		flashMsg = "the chain has been replaced by a longer one"
	} else {
		flashMsg = "we already have the longest chain possible, nothing changed"
	}

	formData := url.Values{
		"flash": {flashMsg},
	}

	req, err := http.NewRequest(http.MethodPost, "/replace", strings.NewReader(formData.Encode()))
	if err != nil {
		RenderHTTPError(w, "failed to POST status: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(formData.Encode())))

	ReplaceGet(w, req, blockchain)
}
