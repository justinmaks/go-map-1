package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestAddColumnIfNotExists(t *testing.T) {
	// Mock the global `db` variable
	var err error
	db, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create the initial table
	_, err = db.Exec(`CREATE TABLE visitors (ip TEXT PRIMARY KEY)`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Call the function
	addColumnIfNotExists("latitude", "REAL")
	addColumnIfNotExists("longitude", "REAL")

	// Verify the columns
	columns, err := db.Query(`PRAGMA table_info(visitors)`)
	if err != nil {
		t.Fatalf("Failed to query table info: %v", err)
	}
	defer columns.Close()

	var foundLat, foundLong bool
	for columns.Next() {
		var colID int
		var name, dataType string
		var notNull, pk int
		var dfltValue sql.NullString // Use sql.NullString for nullable column

		err := columns.Scan(&colID, &name, &dataType, &notNull, &dfltValue, &pk)
		if err != nil {
			t.Fatalf("Failed to scan table info: %v", err)
		}

		if name == "latitude" {
			foundLat = true
		}
		if name == "longitude" {
			foundLong = true
		}
	}

	if !foundLat || !foundLong {
		t.Errorf("Expected columns latitude and longitude to be added")
	}
}

func TestGetRealIP(t *testing.T) {
	tests := []struct {
		name           string
		headers        map[string]string
		remoteAddr     string
		expectedResult string
	}{
		{"Cloudflare Header", map[string]string{"CF-Connecting-IP": "1.1.1.1"}, "", "1.1.1.1"},
		{"X-Forwarded-For Header", map[string]string{"X-Forwarded-For": "2.2.2.2"}, "", "2.2.2.2"},
		{"RemoteAddr", nil, "3.3.3.3:8080", "3.3.3.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			req.RemoteAddr = tt.remoteAddr

			result := getRealIP(req)
			if result != tt.expectedResult {
				t.Errorf("Expected %s, got %s", tt.expectedResult, result)
			}
		})
	}
}

func TestAPIVisitorsHandler(t *testing.T) {
	// Mock the global `db` variable
	var err error
	db, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create schema and insert test data
	_, err = db.Exec(`CREATE TABLE visitors (
		ip TEXT PRIMARY KEY,
		latitude REAL,
		longitude REAL,
		city TEXT,
		country TEXT
	)`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}
	_, err = db.Exec(`INSERT INTO visitors (ip, latitude, longitude, city, country) VALUES ('1.1.1.1', 37.7749, -122.4194, 'San Francisco', 'United States')`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Perform the request
	req := httptest.NewRequest("GET", "/api/visitors", nil)
	rr := httptest.NewRecorder()
	apiVisitorsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}

	var visitors []Visitor
	if err := json.Unmarshal(rr.Body.Bytes(), &visitors); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if len(visitors) != 1 {
		t.Errorf("Expected 1 visitor, got %d", len(visitors))
	}
	if visitors[0].City != "San Francisco" {
		t.Errorf("Expected city San Francisco, got %s", visitors[0].City)
	}
}

func TestAPIStatsHandler(t *testing.T) {
	// Mock the global `db` variable
	var err error
	db, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create schema and insert test data
	_, err = db.Exec(`CREATE TABLE visitors (
		ip TEXT PRIMARY KEY,
		latitude REAL,
		longitude REAL,
		city TEXT,
		country TEXT
	)`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}
	_, err = db.Exec(`INSERT INTO visitors (ip, latitude, longitude, city, country) VALUES ('1.1.1.1', 37.7749, -122.4194, 'San Francisco', 'United States')`)
	_, err = db.Exec(`INSERT INTO visitors (ip, latitude, longitude, city, country) VALUES ('2.2.2.2', 40.7128, -74.0060, 'New York', 'United States')`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Perform the request
	req := httptest.NewRequest("GET", "/api/stats", nil)
	rr := httptest.NewRecorder()
	apiStatsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}

	var stats map[string]int
	if err := json.Unmarshal(rr.Body.Bytes(), &stats); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if stats["unique_visitors"] != 2 {
		t.Errorf("Expected 2 unique visitors, got %d", stats["unique_visitors"])
	}
}
