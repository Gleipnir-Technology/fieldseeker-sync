var map;
onload = (event) => {
	console.log("hey")
	map = L.map('map').setView([36.75, -119.77], 13);

	const tiles = L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
		maxZoom: 19,
		attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
	}).addTo(map);
	var marker = L.marker([51.5, -0.09]).addTo(map);
	map.on("moveend", onMoveEnd);
}

function onMoveEnd(e) {
	let bounds = map.getBounds()
	console.log(bounds.getSouthEast(), bounds.getNorthWest())
	updateMarkers(bounds)
}

async function updateMarkers(bounds) {
	const params = new URLSearchParams({
		maxX: bounds.getWest(),
		maxY: bounds.getNorth(),
		minX: bounds.getEast(),
		minY: bounds.getSouth(),
	});
	const url = "/api/service-request?" + params.toString();
	const response = await fetch(url);
	if (!response.ok) {
		throw new Error(`Response status: ${response.status}`);
	}
	const json = await response.json();
	console.log(json);
}
