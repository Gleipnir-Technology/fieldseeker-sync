function onAudioMetadata(event) {
	console.log("Got audio metadata");
	const audioPlayer = document.getElementById('audio');
	const duration = audioPlayer.duration;
	const intervals = 10; // Number of marks
	for (let i = 0; i <= intervals; i++) {
		const mark = document.createElement('div');
		mark.classList.add('ruler-mark');
		const label = document.createElement('div');
		label.classList.add('ruler-label');
		// Format time in minutes:seconds
		const time = (i * duration / intervals).toFixed(0);
		const minutes = Math.floor(time / 60);
		const seconds = time % 60;
		label.textContent = `${ minutes }:${ seconds.toString().padStart(2, '0') }`;
		mark.appendChild(label);
		rulerMarks.appendChild(mark);
	}
}

function initAudio() {
	console.log("Setting up audio ruler events");
	const audioPlayer = document.getElementById('audio');
	const rulerMarks = document.getElementById('rulerMarks');
	audioPlayer.addEventListener('loadedmetadata', onAudioMetadata);
	// If we were so fast that its already loaded
	if (audioPlayer.readyState >= 2) {
		onAudioMetadata(null);
	}
	// Optional: Seek to clicked position on ruler
	rulerMarks.addEventListener(
		'click',
		function (e) {
			const ruler = e.currentTarget;
			const clickPosition = e.clientX - ruler.getBoundingClientRect().left;
			const rulerWidth = ruler.offsetWidth;
			const seekTime = (clickPosition / rulerWidth) * audioPlayer.duration;
			audioPlayer.currentTime = seekTime;
		}
	);
}

window.addEventListener("load", initAudio);
