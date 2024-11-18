document.addEventListener("DOMContentLoaded", function () {
    const map = L.map("map").setView([20, 0], 2); // Initialize map with world view

    // Add OpenStreetMap tiles
    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
        maxZoom: 18,
    }).addTo(map);

    // Fetch visitors data from the API
    fetch("/api/visitors")
        .then((res) => res.json())
        .then((data) => {
            // Iterate through each visitor
            data.forEach((visitor) => {
                const { latitude, longitude, IP } = visitor; // Use correct property names

                // Validate latitude and longitude
                if (
                    latitude >= -90 &&
                    latitude <= 90 &&
                    longitude >= -180 &&
                    longitude <= 180
                ) {
                    // Add a marker to the map
                    L.marker([latitude, longitude]).addTo(map)
                        .bindPopup(`Visitor at [${latitude}, ${longitude}]`);
                } else {
                    console.warn("Invalid visitor data (out of range):", visitor);
                }
            });
        })
        .catch((error) => {
            console.error("Error fetching visitor data:", error);
        });

    // Fetch and display stats
    fetch("/api/stats")
        .then((res) => res.json())
        .then((data) => {
            document.getElementById("unique-visitors").textContent = data.unique_visitors;
        })
        .catch((error) => {
            console.error("Error fetching stats data:", error);
        });
});
