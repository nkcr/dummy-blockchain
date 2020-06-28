package controllers

import (
	bc "dummy-blockchain/blockchain"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
)

// TransactionHandler ...
func TransactionHandler(blockchain *bc.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			TransactionNew(w, r, blockchain)
		case http.MethodPost:
			TransactionPost(w, r, blockchain)
		}
	}
}

// AddTransactionHandler ...
func AddTransactionHandler(blockchain *bc.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			AddTransactionPost(w, r, blockchain)
		}
	}
}

// TransactionNew ...
func TransactionNew(w http.ResponseWriter, r *http.Request, blockchain *bc.Blockchain) {

	t, err := template.ParseFiles("gui/views/layout.gohtml", "gui/views/transaction.gohtml")
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
		BC    *bc.Blockchain
		Flash string
	}

	p := &viewData{
		Title: "Home",
		BC:    blockchain,
		Flash: flashStr,
	}

	err = t.ExecuteTemplate(w, "layout", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// TransactionPost is called by the HTML form. It transforms the HTML arguments
// into JSON and call the REST endpoint
func TransactionPost(w http.ResponseWriter, r *http.Request, blockchain *bc.Blockchain) {

	err := r.ParseForm()
	if err != nil {
		RenderHTTPError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sender := r.PostForm.Get("sender")
	if sender == "" {
		RenderHTTPError(w, "'Sender' field not found", http.StatusBadRequest)
		return
	}

	receiver := r.PostForm.Get("receiver")
	if receiver == "" {
		RenderHTTPError(w, "'Receiver' field not found", http.StatusBadRequest)
		return
	}

	amountStr := r.PostForm.Get("amount")
	if amountStr == "" {
		RenderHTTPError(w, "'Amount' field not found", http.StatusBadRequest)
		return
	}

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		RenderHTTPError(w, "Failed to convert amount: "+err.Error(),
			http.StatusBadRequest)
		return
	}

	transaction := bc.NewTransaction(sender, receiver, int(amount))

	index := blockchain.AddTransaction(transaction)

	flashMsg := fmt.Sprintf("New transaction added to the pool. "+
		"The transaction should be added in block #%d", index)
	formData := url.Values{
		"flash": {flashMsg},
	}

	req, err := http.NewRequest(http.MethodPost, "/transaction", strings.NewReader(formData.Encode()))
	if err != nil {
		RenderHTTPError(w, "failed to POST status: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(formData.Encode())))

	TransactionNew(w, req, blockchain)
}

// AddTransactionPost is called by REST request
func AddTransactionPost(w http.ResponseWriter, r *http.Request, blockchain *bc.Blockchain) {

	var transaction bc.Transaction
	err := json.NewDecoder(r.Body).Decode(&transaction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		fmt.Println(err.Error())
		return
	}

	index := blockchain.AddTransaction(&transaction)

	var resp = struct {
		Message    string
		BlockIndex int
	}{
		"Transaction added",
		index,
	}

	respJSON, err := json.MarshalIndent(resp, "", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(respJSON)
}
