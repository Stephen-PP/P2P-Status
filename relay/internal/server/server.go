package server

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	migrations "github.com/stephen-pp/p2p-status/relay/migrations"

	_ "github.com/mattn/go-sqlite3"
)

type RelayServer struct {
	Db *sql.DB
	// Server field
	Server *http.Server
}

func StartRelayServer() {
	// Start SQLite connection
	conn, err := sql.Open("sqlite3", os.Getenv("DB_FILE"))
	if err != nil {
		panic(err)
	}

	// Apply SQLite migrations
	err = migrations.ApplyMigrations(conn)
	if err != nil {
		panic(err)
	}

	// Create server
	router := http.NewServeMux()
	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  6 * time.Second,
		WriteTimeout: 6 * time.Second,
	}

	// Create relay struct
	relayServer := RelayServer{
		Db:     conn,
		Server: server,
	}

	// Register all routes now
	router.HandleFunc("GET /health", relayServer.HealthCheck)
}
