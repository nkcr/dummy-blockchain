package controllers

import (
	bc "dummy-blockchain/blockchain"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
)

// NodeHandler ...
func NodeHandler(blockchain *bc.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			NodeGet(w, r, blockchain)
		case http.MethodPost:
			NodePost(w, r, blockchain)
		}
	}
}

// NodeGet ...
func NodeGet(w http.ResponseWriter, r *http.Request, blockchain *bc.Blockchain) {

	t, err := template.ParseFiles("gui/views/layout.gohtml", "gui/views/node.gohtml")
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
		Title: "Node",
		Flash: flashStr,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// NodePost ...
func NodePost(w http.ResponseWriter, r *http.Request, blockchain *bc.Blockchain) {

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	host := r.PostForm.Get("host")
	if host == "" {
		RenderHTTPError(w, "'Host' field not found", http.StatusBadRequest)
		return
	}

	portStr := r.PostForm.Get("port")
	if portStr == "" {
		RenderHTTPError(w, "'Port' field not found", http.StatusBadRequest)
		return
	}

	port, err := strconv.ParseInt(portStr, 10, 64)
	if err != nil {
		RenderHTTPError(w, "Failed to convert port: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	node := bc.NewNode(host, int(port))
	blockchain.AddNode(node)

	flashMsg := fmt.Sprintf("New node added!")
	formData := url.Values{
		"flash": {flashMsg},
	}

	req, err := http.NewRequest(http.MethodPost, "/node", strings.NewReader(formData.Encode()))
	if err != nil {
		RenderHTTPError(w, "failed to POST status: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(formData.Encode())))

	NodeGet(w, req, blockchain)
}
