package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
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
	City      string
	Country   string
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

	// Create table if it does not exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS visitors (
		ip TEXT PRIMARY KEY,
		latitude REAL,
		longitude REAL,
		city TEXT,
		country TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// Add missing columns if necessary
	addColumnIfNotExists("city", "TEXT")
	addColumnIfNotExists("country", "TEXT")
	addColumnIfNotExists("timestamp", "DATETIME DEFAULT CURRENT_TIMESTAMP")

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Routes
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/stats", statsPageHandler)
	http.HandleFunc("/api/visitors", apiVisitorsHandler)
	http.HandleFunc("/api/stats", apiStatsHandler)
	http.HandleFunc("/api/statistics", apiStatisticsHandler)
	http.HandleFunc("/api/visitor_types", apiVisitorTypesHandler)
	http.HandleFunc("/api/trends", apiTrendsHandler)

	log.Println("Starting server on :8905...")
	log.Fatal(http.ListenAndServe(":8905", nil))
}

// Helper function to add columns if they don't exist
func addColumnIfNotExists(columnName string, columnType string) {
	query := fmt.Sprintf(`ALTER TABLE visitors ADD COLUMN %s %s`, columnName, columnType)
	_, err := db.Exec(query)
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		log.Printf("Failed to add column %s to visitors table: %v\n", columnName, err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)

	ip := getRealIP(r)
	if ip == "" {
		log.Println("Failed to extract IP")
		return
	}
	log.Printf("Visitor IP: %s\n", ip)

	latitude, longitude, city, country := fetchGeolocationFromIPInfo(ip)
	log.Printf("Inserting into DB: IP %s, Latitude %f, Longitude %f, City %s, Country %s\n", ip, latitude, longitude, city, country)

	_, err = db.Exec(`INSERT OR IGNORE INTO visitors (ip, latitude, longitude, city, country) VALUES (?, ?, ?, ?, ?)`, ip, latitude, longitude, city, country)
	if err != nil {
		log.Printf("Error inserting into DB: %v\n", err)
	}
}

func statsPageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/stats.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func apiVisitorsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT ip, latitude, longitude, city, country FROM visitors`)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var visitors []Visitor
	for rows.Next() {
		var visitor Visitor
		if err := rows.Scan(&visitor.IP, &visitor.Latitude, &visitor.Longitude, &visitor.City, &visitor.Country); err != nil {
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
	err := db.QueryRow(`SELECT COUNT(DISTINCT ip) FROM visitors`).Scan(&unique)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error querying unique visitor count: %v\n", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"unique_visitors": unique,
	})
}

func apiStatisticsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT country, COUNT(*) as count FROM visitors GROUP BY country`)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error querying statistics: %v\n", err)
		return
	}
	defer rows.Close()

	var labels []string
	var counts []int

	for rows.Next() {
		var country string
		var count int
		if err := rows.Scan(&country, &count); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		labels = append(labels, country)
		counts = append(counts, count)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"labels": labels,
		"counts": counts,
	})
}

func apiVisitorTypesHandler(w http.ResponseWriter, r *http.Request) {
	var unique, returning int
	err := db.QueryRow(`SELECT COUNT(DISTINCT ip) FROM visitors`).Scan(&unique)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = db.QueryRow(`SELECT COUNT(ip) - COUNT(DISTINCT ip) FROM visitors`).Scan(&returning)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"unique_visitors":    unique,
		"returning_visitors": returning,
	})
}

func apiTrendsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT DATE(timestamp) as date, COUNT(*) as count FROM visitors GROUP BY DATE(timestamp) ORDER BY date`)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var dates []string
	var visitorCounts []int
	for rows.Next() {
		var date string
		var count int
		if err := rows.Scan(&date, &count); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		dates = append(dates, date)
		visitorCounts = append(visitorCounts, count)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"dates":          dates,
		"visitor_counts": visitorCounts,
	})
}

func getRealIP(r *http.Request) string {
	ip := r.Header.Get("CF-Connecting-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}
	log.Printf("Extracted IP: %s\n", ip)
	return ip
}

func fetchGeolocationFromIPInfo(ip string) (float64, float64, string, string) {
	if ip == "" || ip == "127.0.0.1" {
		log.Println("Using default location for invalid IP")
		return 37.7749, -122.4194, "San Francisco", "United States"
	}

	token := os.Getenv("IPINFO_TOKEN")
	if token == "" {
		log.Println("IPINFO_TOKEN environment variable not set")
		return 37.7749, -122.4194, "San Francisco", "United States"
	}

	url := fmt.Sprintf("https://ipinfo.io/%s?token=%s", ip, token)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to fetch geolocation for IP %s: %v\n", ip, err)
		return 37.7749, -122.4194, "San Francisco", "United States"
	}
	defer resp.Body.Close()

	var result struct {
		Loc     string `json:"loc"`
		City    string `json:"city"`
		Country string `json:"country"`
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body for IP %s: %v\n", ip, err)
		return 37.7749, -122.4194, "San Francisco", "United States"
	}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Failed to parse geolocation response for IP %s: %v\n", ip, err)
		return 37.7749, -122.4194, "San Francisco", "United States"
	}

	locParts := strings.Split(result.Loc, ",")
	if len(locParts) != 2 {
		log.Printf("Invalid location format for IP %s: %s\n", ip, result.Loc)
		return 37.7749, -122.4194, "San Francisco", "United States"
	}
	var latitude, longitude float64
	fmt.Sscanf(locParts[0], "%f", &latitude)
	fmt.Sscanf(locParts[1], "%f", &longitude)

	log.Printf("Geolocation for IP %s: Latitude %f, Longitude %f, City %s, Country %s\n", ip, latitude, longitude, result.City, result.Country)
	return latitude, longitude, result.City, result.Country
}
