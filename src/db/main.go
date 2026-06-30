package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/lib/pq" // Ensure you have the postgres driver
)

var db *sql.DB

func main() {
	// 1. Connect to your SQL Database
	var err error
	connStr := "postgres://username:password@localhost/dbname?sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// 2. Serve your HTML file at the root URL
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// 3. The API route your HTML calls: loadGamesHome() -> /api/games
	http.HandleFunc("/api/games", getGamesHandler)

	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getGamesHandler(w http.ResponseWriter, r *http.Request) {
	// Allow the HTML to talk to the Go server
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// Query your SQL table "Places" (from your schema)
	rows, err := db.Query("SELECT Name FROM Places LIMIT 10")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var games []map[string]string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		games = append(games, map[string]string{"name": name, "description": "A cool game!"})
	}

	// Send the data back to the HTML
	json.NewEncoder(w).Encode(games)
}