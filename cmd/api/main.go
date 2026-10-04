// Role: Application entry point, composition root, and lifecycle coordinator.
// Connects with: internal/database, internal/repository, and internal/handlers.
// Responsibilities:
// - Parses runtime configurations and establishes the PostgreSQL connection pool.
// - Assembles dependencies (DB -> Repository -> Handlers -> Router).
// - Listens for OS termination signals (SIGINT, SIGTERM) to execute clean graceful shutdown.

package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"task-service-pg/internal/database"
	"task-service-pg/internal/handlers"
	"task-service-pg/internal/repository"
)

func main() {
	// 1. Intercept OS termination signals; therefore, the runtime can start an orderly shutdown instead of abruptly killing active workers.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// 2. Load the local .env file.
	// We ignore loading errors here because containerized production environments inject environment variables directly into the system shell.
	_ = godotenv.Load()

	// 3. Retrieve database credentials securely
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}

	// 4. Connect to PostgreSQL with a strict startup context timeout
	initCtx, cancelInit := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelInit()

	db, err := database.NewPostgresDB(initCtx, dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Ensure the database schema is migrated before accepting incoming HTTP traffic
	if err := database.Migrate(initCtx, db); err != nil {
		log.Fatalf("failed to migrate database schema: %v", err)
	}

	// 5. Dependency Injection Assembly
	// The repository requires *sql.DB; the HTTP handler requires the repository interface.
	repo := repository.NewPostgresTaskRepository(db)
	handler := handlers.NewTaskHandler(repo)

	// 6. Router initialization using Go 1.22+ method-matching ServeMux
	mux := http.NewServeMux()
	mux.HandleFunc("POST /tasks", handler.CreateTask)
	mux.HandleFunc("GET /tasks/{id}", handler.GetTask)
	mux.HandleFunc("GET /tasks", handler.ListTasks)
	mux.HandleFunc("PUT /tasks/{id}", handler.UpdateTask)
	mux.HandleFunc("DELETE /tasks/{id}", handler.DeleteTask)

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	server := &http.Server{
		Addr:         port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 7. Launch HTTP listener in a separate goroutine so it does not block signal listening
	go func() {
		log.Printf("starting server on port %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed to listen to the server: %v", err)
		}
	}()

	// 8. Block execution until SIGINT or SIGTERM is intercepted
	sig := <-stop
	log.Printf("signal received %v: initiating graceful shutdown", sig)

	// 9. Grant active connections up to 10 seconds to finish in-flight processing
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
	}

	log.Println("server exited cleanly")
}
