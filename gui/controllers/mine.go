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

// MineHandler ...
func MineHandler(blockchain *bc.Blockchain, me string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			MineGet(w, r, blockchain)
		case http.MethodPost:
			MinePost(w, r, blockchain, me)
		}
	}
}

// MineGet ...
func MineGet(w http.ResponseWriter, r *http.Request, blockchain *bc.Blockchain) {

	t, err := template.ParseFiles("gui/views/layout.gohtml", "gui/views/mine.gohtml")
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

// MinePost ...
func MinePost(w http.ResponseWriter, r *http.Request, blockchain *bc.Blockchain, me string) {

	previousBblock := blockchain.GetPreviousBlock()
	proof := blockchain.ProofOfWork(previousBblock.Proof)
	previousHash, err := previousBblock.Hash()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// transaction fee, sending money to myself
	t := bc.NewTransaction(blockchain.Address, me, 1)
	blockchain.AddTransaction(t)

	block := blockchain.CreateBlock(proof, previousHash)

	flashMsg := fmt.Sprintf("New block with index %d mined! We found the "+
		"nounce %d.", block.Index, block.Proof)
	formData := url.Values{
		"flash": {flashMsg},
	}

	req, err := http.NewRequest(http.MethodPost, "/mine", strings.NewReader(formData.Encode()))
	if err != nil {
		RenderHTTPError(w, "failed to POST status: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(formData.Encode())))

	MineGet(w, req, blockchain)
}
