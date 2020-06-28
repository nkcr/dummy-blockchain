package main

import (
	"context"
	"dummy-blockchain/blockchain"
	bc "dummy-blockchain/blockchain"
	"dummy-blockchain/gui/controllers"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/google/uuid"
)

type key int

const (
	requestIDKey key = 0
)

func main() {
	var listenAddr string
	flag.StringVar(&listenAddr, "listen-addr", ":5002", "server listen address")
	var ownerAddr string
	flag.StringVar(&ownerAddr, "owner", "alice", "owner address to which the "+
		"transaction fees are given")

	flag.Parse()

	address := uuid.New().String()
	address = strings.ReplaceAll(address, "-", "")
	blockchain := blockchain.NewBlockchain(address)

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Println("Server is starting...")

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./gui/assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	mux.HandleFunc("/", controllers.HomeHandler(blockchain, ownerAddr))
	// HTML form endpoint
	mux.HandleFunc("/transaction", controllers.TransactionHandler(blockchain))
	// REST endpoint
	mux.HandleFunc("/add_transaction", controllers.AddTransactionHandler(blockchain))

	mux.HandleFunc("/mine", controllers.MineHandler(blockchain, ownerAddr))

	mux.HandleFunc("/node", controllers.NodeHandler(blockchain))

	mux.HandleFunc("/replace", controllers.ReplaceHandler(blockchain))

	mux.HandleFunc("/mine_block", mineHandler(blockchain, ownerAddr))
	mux.HandleFunc("/get_chain", getChainHandler(blockchain))
	mux.HandleFunc("/is_valid", isValidHandler(blockchain))
	mux.HandleFunc("/connect_node", connectNodeHandler(blockchain))
	mux.HandleFunc("/replace_chain", replaceChainHandler(blockchain))

	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	server := &http.Server{
		Addr:         listenAddr,
		Handler:      tracing(nextRequestID)(logging(logger)(mux)),
		ErrorLog:     logger,
		ReadTimeout:  50 * time.Second,
		WriteTimeout: 600 * time.Second,
		// IdleTimeout:  150 * time.Second,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		logger.Println("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	lu := &url.URL{Scheme: "http"}
	if strings.HasPrefix(listenAddr, ":") {
		lu.Host = "localhost" + listenAddr
	} else {
		lu.Host = listenAddr
	}

	logger.Println("Server is ready to handle requests at", lu)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
	}

	<-done
	logger.Println("Server stopped")
}

func logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestID, ok := r.Context().Value(requestIDKey).(string)
				if !ok {
					requestID = "unknown"
				}
				logger.Println(requestID, r.Method, r.URL.Path, r.RemoteAddr,
					r.UserAgent())
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func tracing(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func isValidHandler(blockchain *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "only GET is allowed", http.StatusBadRequest)
			return
		}

		isValid, err := blockchain.IsCHainValid(blockchain.Chain)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var resp = struct {
			IsValid bool
		}{
			isValid,
		}

		respJSON, err := json.MarshalIndent(resp, "", "")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(respJSON)
	}
}

func getChainHandler(blockchain *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "only GET is allowed", http.StatusBadRequest)
			return
		}

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
}

func mineHandler(blockchain *blockchain.Blockchain, me string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "only Get is allowed", http.StatusBadRequest)
			return
		}

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
		var resp = struct {
			Message string
			Block   *bc.Block
		}{
			"You mined a new block!",
			block,
		}

		respJSON, err := json.MarshalIndent(resp, "", "")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(respJSON)
	}
}

func connectNodeHandler(blockchain *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "only POST is allowed", http.StatusBadRequest)
			return
		}

		type addNodeRequest struct {
			Nodes []*bc.Node
		}

		var addRequest addNodeRequest
		err := json.NewDecoder(req.Body).Decode(&addRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for _, node := range addRequest.Nodes {
			blockchain.AddNode(node)
		}

		var resp = struct {
			Message    string
			TotalNodes int
			Nodes      []*bc.Node
		}{
			"Nodes added",
			len(blockchain.Nodes),
			blockchain.Nodes,
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
}

func replaceChainHandler(blockchain *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "only GET is allowed", http.StatusBadRequest)
			return
		}

		replaced, err := blockchain.ReplaceChain()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var resp = struct {
			Message    string
			IsReplaced bool
			Blockchain []*bc.Block
		}{
			"Blockchain checked and replaced if shortest",
			replaced,
			blockchain.Chain,
		}

		respJSON, err := json.MarshalIndent(resp, "", "")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(respJSON)
	}
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "assets/images/favicon.ico")
}
