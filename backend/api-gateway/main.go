package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool
var verifier *oidc.IDTokenVerifier

func main() {
	ctx := context.Background()

	// --- Keycloak issuer and clientID ---
	issuer := os.Getenv("KEYCLOAK_ISSUER")
	if issuer == "" {
		issuer = "http://localhost:8081/realms/open-mission-control"
	}
	clientID := os.Getenv("KEYCLOAK_CLIENT_ID")
	if clientID == "" {
		clientID = "omc-api" // backend client
	}

	// Setup OIDC provider
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		log.Fatalf("Failed to get OIDC provider: %v", err)
	}
	verifier = provider.Verifier(&oidc.Config{ClientID: clientID})

	// --- Database connection ---
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Auto-switch depending on environment
		host := "localhost"
		if os.Getenv("DOCKER_ENV") == "true" {
			host = "postgres"
		}
		dsn = fmt.Sprintf("postgres://omc:omc@%s:5432/omc?sslmode=disable", host)
	}

	db, err = pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()

	// Run DB migration (create missions table if not exists)
	if err := migrateDB(ctx, db); err != nil {
		log.Fatalf("DB migration failed: %v\n", err)
	}

	// --- Router ---
	r := mux.NewRouter()

	// Healthchecks
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	}).Methods("GET")

	r.HandleFunc("/healthz/db", func(w http.ResponseWriter, r *http.Request) {
		c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := db.Ping(c); err != nil {
			http.Error(w, "db not ok", http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "db ok")
	}).Methods("GET")

	// Protected routes
	r.HandleFunc("/me", handleMe).Methods("GET")
	r.HandleFunc("/missions", listMissions).Methods("GET")
	r.HandleFunc("/missions", createMission).Methods("POST")
	r.HandleFunc("/missions/{id}", updateMission).Methods("PUT")
	r.HandleFunc("/missions/{id}", deleteMission).Methods("DELETE")

	// CORS so frontend can call backend
	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:3000"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Authorization", "Content-Type"}),
	)

	log.Println("API Gateway running on :8080")
	log.Fatal(http.ListenAndServe(":8080", cors(r)))
}

// --- Migration step ---
func migrateDB(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS missions (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)
	`)
	return err
}

// --- Helpers ---
func verifyToken(r *http.Request) (map[string]interface{}, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing Authorization header")
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		return nil, fmt.Errorf("invalid Authorization header")
	}
	idToken, err := verifier.Verify(r.Context(), token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}
	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims")
	}
	return claims, nil
}

func hasRole(claims map[string]interface{}, role string) bool {
	realmAccess := claims["realm_access"].(map[string]interface{})
	roles := realmAccess["roles"].([]interface{})
	for _, r := range roles {
		if r.(string) == role {
			return true
		}
	}
	return false
}

// --- Handlers ---
func handleMe(w http.ResponseWriter, r *http.Request) {
	claims, err := verifyToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	json.NewEncoder(w).Encode(claims)
}

func listMissions(w http.ResponseWriter, r *http.Request) {
	claims, err := verifyToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if !hasRole(claims, "default-roles-open-mission-control") && !hasRole(claims, "user") {
		http.Error(w, "forbidden: insufficient role", http.StatusForbidden)
		return
	}

	rows, err := db.Query(r.Context(), "SELECT id, name, status FROM missions ORDER BY id")
	if err != nil {
		http.Error(w, "failed to query missions: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var id int
		var name, status string
		if err := rows.Scan(&id, &name, &status); err != nil {
			http.Error(w, "failed to scan mission: "+err.Error(), http.StatusInternalServerError)
			return
		}
		result = append(result, map[string]interface{}{
			"id":     id,
			"name":   name,
			"status": status,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func createMission(w http.ResponseWriter, r *http.Request) {
	claims, err := verifyToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if !hasRole(claims, "admin") {
		http.Error(w, "forbidden: admin role required", http.StatusForbidden)
		return
	}

	var mission struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&mission); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	var id int
	err = db.QueryRow(r.Context(),
		"INSERT INTO missions (name, status) VALUES ($1, $2) RETURNING id",
		mission.Name, mission.Status).Scan(&id)
	if err != nil {
		http.Error(w, "failed to insert mission: "+err.Error(), http.StatusInternalServerError)
		return
	}

	missionJSON := map[string]interface{}{"id": id, "name": mission.Name, "status": mission.Status}
	json.NewEncoder(w).Encode(missionJSON)
}

func updateMission(w http.ResponseWriter, r *http.Request) {
	claims, err := verifyToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if !hasRole(claims, "admin") {
		http.Error(w, "forbidden: admin role required", http.StatusForbidden)
		return
	}

	id := mux.Vars(r)["id"]
	var update struct {
		Name   *string `json:"name"`
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(r.Context(),
		"UPDATE missions SET name = COALESCE($1, name), status = COALESCE($2, status), updated_at = NOW() WHERE id = $3",
		update.Name, update.Status, id)
	if err != nil {
		http.Error(w, "failed to update mission: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func deleteMission(w http.ResponseWriter, r *http.Request) {
	claims, err := verifyToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if !hasRole(claims, "admin") {
		http.Error(w, "forbidden: admin role required", http.StatusForbidden)
		return
	}

	id := mux.Vars(r)["id"]
	_, err = db.Exec(r.Context(), "DELETE FROM missions WHERE id=$1", id)
	if err != nil {
		http.Error(w, "failed to delete mission: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
