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
                const { Latitude, Longitude, IP } = visitor;

                // Validate latitude and longitude
                if (
                    typeof Latitude === "number" &&
                    typeof Longitude === "number" &&
                    !isNaN(Latitude) &&
                    !isNaN(Longitude)
                ) {
                    // Add a marker to the map
                    L.marker([Latitude, Longitude]).addTo(map)
                        .bindPopup(`Visitor at [${Latitude}, ${Longitude}]`);
                } else {
                    console.warn("Invalid visitor data:", visitor);
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
