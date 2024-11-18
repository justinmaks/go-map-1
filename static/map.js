document.addEventListener("DOMContentLoaded", function () {
    const map = L.map("map").setView([20, 0], 2);

    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
        maxZoom: 18,
    }).addTo(map);

    fetch("/api/visitors")
        .then((res) => res.json())
        .then((data) => {
            data.forEach((visitor) => {
                if (visitor.latitude && visitor.longitude) {
                    L.marker([visitor.latitude, visitor.longitude]).addTo(map)
                        .bindPopup(`Visitor at [${visitor.latitude}, ${visitor.longitude}]`);
                } else {
                    console.warn("Invalid visitor data:", visitor);
                }
            });
        })
        .catch((error) => {
            console.error("Error fetching visitor data:", error);
        });

    fetch("/api/stats")
        .then((res) => res.json())
        .then((data) => {
            document.getElementById("unique-visitors").textContent = data.unique_visitors;
        })
        .catch((error) => {
            console.error("Error fetching stats data:", error);
        });
});
