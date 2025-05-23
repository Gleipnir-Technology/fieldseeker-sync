var map;
onload = (event) => {
	const bounds = parseBoundsFromHash();
	console.log("Fitting bounds", bounds);
	map = L.map('map').fitBounds(bounds);

	const tiles = L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
		maxZoom: 19,
		attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
	}).addTo(map);
	map.on("moveend", onMoveEnd);
	getMarkersForBounds(map.getBounds());
}

function parseBoundsFromHash() {
	const hash = window.location.hash;
	const params = new URLSearchParams(
		hash.substring(1)
	);
	try {
		const bounds = L.latLngBounds(
			L.latLng(
				parseFloat(params.get("north")),
				parseFloat(params.get("east")),
			),
			L.latLng(
				parseFloat(params.get("south")),
				parseFloat(params.get("west")),
			)
		)
		console.log("From hash", bounds);
		return bounds;
	} catch(e) {
		return L.latLngBounds(
			L.latLng(
				36.129001,
				-118.391418,
			),
			L.latLng(
				36.789491,
				-120.16845,
			)
		);
	}
}

function onMoveEnd(e) {
	let bounds = map.getBounds()
	setHashToBounds(bounds)
	console.log(bounds.getEast(), bounds.getNorth(), bounds.getWest(), bounds.getSouth())
	//getMarkersForBounds(bounds)
}

function paramsFromBounds(bounds) {
	const params = new URLSearchParams({
		west: bounds.getWest(),
		north: bounds.getNorth(),
		east: bounds.getEast(),
		south: bounds.getSouth(),
	});
	return params;
}

function setHashToBounds(bounds) {
	const params = paramsFromBounds(bounds);
	window.location.hash = params.toString();
}

async function getMarkersForBounds(bounds) {
	getServiceRequestsForBounds(bounds);
	getTrapDataForBounds(bounds);
}

async function getServiceRequestsForBounds(bounds) {
	const params = paramsFromBounds(bounds);
	const url = "/api/service-request?" + params.toString();
	const response = await fetch(url);
	if (!response.ok) {
		throw new Error(`Response status: ${response.status}`);
	}
	const json = await response.json();
	for(let i = 0; i < json.length; i++) {
		const r = json[i];
		L.marker([r.lat, r.long]).addTo(map);
		//console.log(r.lat, r.long);
	}
}

async function getTrapDataForBounds(bounds) {
	const params = paramsFromBounds(bounds);
	const url = "/api/trap-data?" + params.toString();
	const response = await fetch(url);
	if (!response.ok) {
		throw new Error(`Response status: ${response.status}`);
	}
	const json = await response.json();
	for(let i = 0; i < json.length; i++) {
		const r = json[i];
		L.marker([r.lat, r.long]).addTo(map);
		//console.log(r.lat, r.long);
	}
}
