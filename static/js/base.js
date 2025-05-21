var map;
onload = (event) => {
	console.log("hey")
	map = L.map('map').setView([36.111, -118.0], 13);

	const tiles = L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
		maxZoom: 19,
		attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
	}).addTo(map);
	map.on("moveend", onMoveEnd);
	updateMarkers(map.getBounds())
}

function onMoveEnd(e) {
	let bounds = map.getBounds()
	console.log(bounds.getSouthEast(), bounds.getNorthWest())
	//updateMarkers(bounds)
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
	for(let i = 0; i < json.length; i++) {
		const r = json[i];
		L.marker([r.lat, r.long]).addTo(map).bindPopup(r.target).openPopup();
		console.log(r.lat, r.long);
	}
}
