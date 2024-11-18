package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Visitor struct {
	IP        string
	Latitude  float64
	Longitude float64
}

var db *sql.DB

func main() {
	var err error
	// Initialize the database
	db, err = sql.Open("sqlite3", "./db/database.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create table if not exists
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS visitors (
		ip TEXT PRIMARY KEY,
		latitude REAL,
		longitude REAL
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Routes
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/visitors", apiVisitorsHandler)
	http.HandleFunc("/api/stats", apiStatsHandler)

	log.Println("Starting server on :8905...")
	log.Fatal(http.ListenAndServe(":8905", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)

	// Get the IP from the headers
	ip := getRealIP(r)
	if ip == "" {
		log.Println("Failed to extract IP")
		return
	}
	log.Printf("Visitor IP: %s\n", ip)

	// Get geolocation from ipinfo.io
	latitude, longitude := fetchGeolocationFromIPInfo(ip)

	// Insert into DB
	_, err = db.Exec(`INSERT OR IGNORE INTO visitors (ip, latitude, longitude) VALUES (?, ?, ?)`, ip, latitude, longitude)
	if err != nil {
		log.Printf("Error inserting into DB: %v\n", err)
	}
}

func apiVisitorsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT ip, latitude, longitude FROM visitors`)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var visitors []Visitor
	for rows.Next() {
		var visitor Visitor
		if err := rows.Scan(&visitor.IP, &visitor.Latitude, &visitor.Longitude); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		visitors = append(visitors, visitor)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(visitors)
}

func apiStatsHandler(w http.ResponseWriter, r *http.Request) {
	var unique int
	err := db.QueryRow(`SELECT COUNT(*) FROM visitors`).Scan(&unique)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"unique_visitors": unique,
	})
}

func getRealIP(r *http.Request) string {
	ip := r.Header.Get("CF-Connecting-IP") // Cloudflare header
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For") // Nginx proxy
	}
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0] // Fallback to direct connection IP
	}
	log.Printf("Extracted IP: %s\n", ip)
	return ip
}

func fetchGeolocationFromIPInfo(ip string) (float64, float64) {
	if ip == "" || ip == "127.0.0.1" {
		log.Println("Using default location for invalid IP")
		return 37.7749, -122.4194 // Default to San Francisco
	}

	// Replace with your ipinfo.io token
	token := os.Getenv("IPINFO_TOKEN")
	if token == "" {
		log.Println("IPINFO_TOKEN environment variable not set")
		return 37.7749, -122.4194 // Default to San Francisco
	}

	// Make the request to ipinfo.io
	url := fmt.Sprintf("https://ipinfo.io/%s?token=%s", ip, token)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to fetch geolocation for IP %s: %v\n", ip, err)
		return 37.7749, -122.4194 // Default location on failure
	}
	defer resp.Body.Close()

	// Parse the response
	var result struct {
		Loc string `json:"loc"` // Example: "39.9524,-75.1636"
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Failed to parse geolocation response for IP %s: %v\n", ip, err)
		return 37.7749, -122.4194
	}

	// Split the "loc" field into latitude and longitude
	locParts := strings.Split(result.Loc, ",")
	if len(locParts) != 2 {
		log.Printf("Invalid location format for IP %s: %s\n", ip, result.Loc)
		return 37.7749, -122.4194
	}

	// Parse latitude and longitude
	var latitude, longitude float64
	fmt.Sscanf(locParts[0], "%f", &latitude)
	fmt.Sscanf(locParts[1], "%f", &longitude)

	log.Printf("Geolocation for IP %s: Latitude %f, Longitude %f\n", ip, latitude, longitude)
	return latitude, longitude
}
