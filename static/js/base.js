var map;
var markers = {
	mosquitoSource: null,
	serviceRequest: null,
	trapData: null,
	types: {}
}


onload = (event) => {
	const bounds = parseBoundsFromHash();
	console.log("Fitting bounds", bounds);
	map = L.map('map').fitBounds(bounds);

	const tiles = L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
		maxZoom: 19,
		attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
	}).addTo(map);
	markers.mosquitoSource = L.layerGroup([]).addTo(map);
	markers.serviceRequest = L.layerGroup([]).addTo(map);
	markers.trapData = L.layerGroup([]).addTo(map);

	// Set up custom markers
	var MarkerIcon = L.Icon.extend({
		options: {
			// Numbers are taken from L.Marker.prototype.options.icon
			shadowUrl: '/static/img/marker-shadow.png',
			iconAnchor:   [12, 41],
			iconSize:     [25, 41],
			popupAnchor:  [1, -34],
			shadowSize:   [41, 41],
			shadowAnchor: [12, 42],
			tooltipAnchor: [16, -28],
		}
});
	markers.types.blue = new MarkerIcon({iconUrl: "/static/img/marker-blue.png"})
	markers.types.green = new MarkerIcon({iconUrl: "/static/img/marker-green.png"})
	markers.types.red = new MarkerIcon({iconUrl: "/static/img/marker-red.png"})
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
	getMarkersForBounds(bounds)
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
	getMosquitoSourcesForBounds(bounds);
	getServiceRequestsForBounds(bounds);
	getTrapDataForBounds(bounds);
}

async function getMosquitoSourcesForBounds(bounds) {
	const params = paramsFromBounds(bounds);
	const url = "/api/mosquito-source?" + params.toString();
	const response = await fetch(url);
	if (!response.ok) {
		throw new Error(`Response status: ${response.status}`);
	}
	const json = await response.json();
	markers.mosquitoSource.clearLayers();
	for(let i = 0; i < json.length; i++) {
		const r = json[i];
		var m = L.marker([r.location.longitude, r.location.latitude], {icon: markers.types.red})
		m.on("click", function(e) {
			showMosquitoSource(r);
		});
		markers.mosquitoSource.addLayer(m);
	}
	var count = document.getElementById("count-mosquito-source");
	count.innerHTML = json.length;
}

async function getServiceRequestsForBounds(bounds) {
	const params = paramsFromBounds(bounds);
	const url = "/api/service-request?" + params.toString();
	const response = await fetch(url);
	if (!response.ok) {
		throw new Error(`Response status: ${response.status}`);
	}
	const json = await response.json();
	markers.serviceRequest.clearLayers();
	for(let i = 0; i < json.length; i++) {
		const r = json[i];
		var m = L.marker([r.lat, r.long], {icon: markers.types.blue});
		m.on("click", function(e) {
			showServiceRequest(r);
		});
		markers.serviceRequest.addLayer(m);
	}
	var count = document.getElementById("count-service-request");
	count.innerHTML = json.length;
}

async function getTrapDataForBounds(bounds) {
	const params = paramsFromBounds(bounds);
	const url = "/api/trap-data?" + params.toString();
	const response = await fetch(url);
	if (!response.ok) {
		throw new Error(`Response status: ${response.status}`);
	}
	const json = await response.json();
	markers.trapData.clearLayers();
	for(let i = 0; i < json.length; i++) {
		const r = json[i];
		markers.trapData.addLayer(L.marker([r.lat, r.long], {icon: markers.types.green}).on("click", function(e) {
			showTrapData(r);
		}));
	}
	var count = document.getElementById("count-trap-data");
	count.innerHTML = json.length;
}

function showMosquitoSource(ms) {
	console.log("Mosquito Source", ms);
	var inspections = ("<table>" +
		"<tr><th>Created</th><th>Condition</th><th>Comments</th></tr>");
	for(let i = 0; i < ms.inspections.length; i++) {
		let insp = ms.inspections[i];
		inspections += (
			"<tr>" +
				"<td>" + insp.created + "</td>" +
				"<td>" + insp.condition + "</td>" +
				"<td>" + insp.comments + "</td>" +
			"</tr>");
	}
	var detail = document.getElementById("detail");
	detail.innerHTML = ("<h1>Mosquito Source</h1>" +
		"<table>" +
		"<tr><td>Access</td><td>" + ms.access + "</td>" +
		"<tr><td>Comments</td><td>" + ms.comments + "</td>" +
		"<tr><td>Description</td><td>" + ms.description + "</td>" +
		"<tr><td>Habitat</td><td>" + ms.habitat + "</td>" +
		"<tr><td>Latitude</td><td>" + ms.location.latitude + "</td>" +
		"<tr><td>Longitude</td><td>" + ms.location.longitude + "</td>" +
		"<tr><td>Status</td><td>" + ms.status + "</td>" +
		"<tr><td>Target</td><td>" + ms.target + "</td>" +
		"</table><h2>Inspections</h2>" + inspections
	);
}

function showServiceRequest(sr) {
	console.log("Service Request", sr);
	var detail = document.getElementById("detail");
	detail.innerHTML = ("<h1>Service Request</h1>" +
		"<table>" +
		"<tr><td>Address</td><td>" + sr.address + "</td>" +
		"<tr><td>City</td><td>" + sr.city + "</td>" +
		"<tr><td>Zip</td><td>" + sr.zip + "</td>" +
		"<tr><td>Priority</td><td>" + sr.priority + "</td>" +
		"<tr><td>Source</td><td>" + sr.source + "</td>" +
		"<tr><td>Status</td><td>" + sr.status + "</td>" +
		"<tr><td>Target</td><td>" + sr.target + "</td>" +
		"</table>"
	);
}

function showTrapData(sr) {
	console.log("Trap Data", sr);
	var detail = document.getElementById("detail");
	detail.innerHTML = ("<h1>Trap Data</h1>" +
		"<table>" +
		"<tr><td>Name</td><td>" + sr.name + "</td>" +
		"<tr><td>Latitude</td><td>" + sr.lat + "</td>" +
		"<tr><td>Longitude</td><td>" + sr.long + "</td>" +
		"</table>"
	);
}
