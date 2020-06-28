package main

import (
	"context"
	"dummy-blockchain/blockchain"
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
	flag.StringVar(&listenAddr, "listen-addr", ":8080", "server listen address")
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

	// HTML endpoint
	mux.HandleFunc("/", controllers.HomeHandler(blockchain, ownerAddr))
	// REST endpoint
	mux.HandleFunc("/get_chain", controllers.GetChainHandler(blockchain))

	// HTML endpoint
	mux.HandleFunc("/transaction", controllers.TransactionHandler(blockchain))
	// REST endpoint
	mux.HandleFunc("/add_transaction", controllers.AddTransactionHandler(blockchain))

	// HTML endpoint
	mux.HandleFunc("/mine", controllers.MineHandler(blockchain, ownerAddr))
	// REST endpoint
	mux.HandleFunc("/mine_block", controllers.MineRESTHandler(blockchain, ownerAddr))

	// HTML endpoint
	mux.HandleFunc("/replace", controllers.ReplaceHandler(blockchain))
	// REST endpoint
	mux.HandleFunc("/replace_chain", controllers.ReplaceChainHandler(blockchain))

	// HTML endpoint
	mux.HandleFunc("/node", controllers.NodeHandler(blockchain))
	// REST endpoint
	mux.HandleFunc("/connect_node", controllers.ConnectNodesHandler(blockchain))

	mux.HandleFunc("/is_valid", isValidHandler(blockchain))

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

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "assets/images/favicon.ico")
}
