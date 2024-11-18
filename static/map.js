document.addEventListener("DOMContentLoaded", function () {
    // Initialize the map, centered on a default location (latitude: 20, longitude: 0)
    const map = L.map("map").setView([20, 0], 2); // Zoom level 2 for a world view

    // Add a tile layer (OpenStreetMap in this case)
    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
        maxZoom: 18,
    }).addTo(map);

    // Fetch visitor locations from the API
    fetch("/api/visitors")
    .then((res) => res.json())
    .then((data) => {
        data.forEach((visitor) => {
            if (visitor.Latitude && visitor.Longitude) {
                L.marker([visitor.Latitude, visitor.Longitude]).addTo(map)
                    .bindPopup(`Visitor at [${visitor.Latitude}, ${visitor.Longitude}]`);
            } else {
                console.warn("Invalid visitor data:", visitor);
            }
        });
    })
    .catch((error) => {
        console.error("Error fetching visitor data:", error);
    });


    // Fetch and display the statistics (e.g., unique visitors)
    fetch("/api/stats")
        .then((res) => res.json())
        .then((data) => {
            document.getElementById("unique-visitors").textContent = data.unique_visitors;
        })
        .catch((error) => {
            console.error("Error fetching stats data:", error);
        });
});
