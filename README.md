# go-map

[DEMO](https://devmaks.biz)
[TEST](https://beta.devmaks.biz)

This project visualizes visitor data on an interactive world map using [Leaflet.js](https://leafletjs.com/). Visitor IPs are captured, geolocated using the [ipinfo.io](https://ipinfo.io) API, and displayed as markers on the map along with their city, country, and coordinates.

### @TODO
- lockdown api endpoints

## Features
- Captures visitor IP addresses.
- Geolocates visitors to determine their city, country, and coordinates.
- Displays visitors as pins on an interactive world map.

## Technologies Used
- Backend: [Go (Golang)](https://go.dev/)
- Database: SQLite
- Frontend: HTML, CSS, [Leaflet.js](https://leafletjs.com/)
- IP Geolocation: [ipinfo.io](https://ipinfo.io)
- Containerization: Docker, Docker Compose
- Web Server: Nginx (optional for production)

## Requirements
- Go 1.23 or higher
- Docker and Docker Compose (optional)
- SQLite
- An [ipinfo.io](https://ipinfo.io) API token for IP geolocation.

## Setup Instructions

1. clone
2. `touch .env`
3. `IPINFO_TOKEN=<your-ipinfo-io-token>`
4. `docker-compose up --build`


## API Endpoints

### `/api/visitors`
**Method:** GET  
**Description:** Returns a list of all visitors, including their IP address, city, country, latitude, and longitude.  
**Response Example:**
```json
[
    {
        "IP": "2602:79:xxxx:xxxx:xxxx:xxxx:xxxx:xxxx",
        "Latitude": 39.9524,
        "Longitude": -75.1636,
        "City": "Philadelphia",
        "Country": "US"
    },
    {
        "IP": "185.195.xxx.xxx",
        "Latitude": 51.5085,
        "Longitude": -0.1257,
        "City": "London",
        "Country": "GB"
    }
]
```

### `/api/stats`
**Method:** GET  
**Description:** Returns the number of unique visitors.  

**Response Example:**
```json
{
    "unique_visitors": 5
}
