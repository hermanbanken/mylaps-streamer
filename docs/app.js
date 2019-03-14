const result = document.getElementById("result");
const combinedresult = document.getElementById("combinedresult");
let runningSource = undefined;
let secondSource = undefined;

const isLoggedIn = document.cookie.split(";").find((s) => s.match(/accesstoken=.*/));
if (isLoggedIn) {
	runningSource = startStreaming("/api/stream/me", result);
	const lgn = document.querySelector(".login.btn");
	const btn = document.createElement("a");
	btn.innerHTML = '<a href="/api/public" class="btn">Maak mijn data publiek</a>'
	lgn.parentElement.insertBefore(btn, lgn.nextSibling);
} else {
	showSkaters();
}

function showSkaters() {
	if (!secondSource) {
		fetch("/api/streams").then((r) => r.json()).then((streams) => {
			if (streams.length) {
				secondSource = startStreaming("/api/stream/other?userid=" + streams[0].userId, combinedresult);
			}
		});
	}
}

/**
 * @param {String} url 
 * @param {HTMLElement} element 
 */
function startStreaming(url, element) {
	let source = undefined;
	function setup() {
		source = new EventSource(url);
		source.addEventListener("update", update);
		source.addEventListener("close", setup);
	}
	function update(e) {
		const data = JSON.parse(e.data);
		const id = e.lastEventId;
		if (data.type == "activity") {
			const div = document.createElement("div");
			div.innerHTML = '<div id="' + id + '" class="activity"><div class="info"><div class="track">' + data.location.name + ' <div class="length">' + data.location.trackLength + 'm</div></div><div class="date">' + data.start + '</div></div><div class="laps"><span /></div></div>';
			element.insertBefore(div, element.querySelector("h1").nextSibling);
		} else if (data.type == "lap") {
			const laps = element.querySelector('[id="' + id.replace(/:\d+$/, "")+'"]').querySelector(".laps");
			const height = Math.min(100, lapSeconds(data.duration)) * 1.5 + "px";
			const colorClass = speedClass(lapSeconds(data.duration)).join(" ");
			const div = document.createElement("div");
			div.innerHTML = '<div id="' + id + '" class="lap '+ colorClass +'" style="height: ' + height +'"><div class="box"><div class="duration">' + data.duration + '</div><div class="date">om ' + timeOnly(data.start) + '</div></div></div>';
			laps.insertBefore(div, laps.lastChild);
		} else if (data.type == "profile") {
			element.querySelector('h1').innerText = data.name;
		}
	}
	setup();
	return { stop: () => source.close() };
}

function timeOnly(date) {
	return new Date(date).toLocaleTimeString();
}

function speedClass(time) {
	const tags = [];
	if (time > 50) {
		tags.push("slow")
	}
	if (time < 30) {
		tags.push("ultra")
	}
	if (time < 33) {
		tags.push("fast")
	}
	if (time <= 50) {
		tags.push("ok")
	}
	return tags
}

function lapSeconds(duration) {
	// 1:56.537
	// 1m56.537s
	const [ms, s, m] = duration.replace(/s$/, "").split(/[^\d]/).reverse();
	let value = 0;
	value += parseInt(ms) / Math.pow(10, ms.length);
	value += parseInt(s);
	value += typeof m === "string" ? parseInt(m) * 60 : 0;
	return value;
}

// @source: http://www.3quarks.com/en/SegmentDisplay/index.html#sourceCode
function segments() {
	var display = new SegmentDisplay("display");
	display.pattern         = "##:##:##";
	display.displayAngle    = 3;
	display.digitHeight     = 23.5;
	display.digitWidth      = 14.5;
	display.digitDistance   = 2.5;
	display.segmentWidth    = 2.9;
	display.segmentDistance = 0.3;
	display.segmentCount    = 7;
	display.cornerType      = 3;
  display.colorOn         = "#ff291e";
  display.colorOff        = "#424542";
	display.draw();

  function animate() {
    var time    = new Date();
    var hours   = time.getHours();
    var minutes = time.getMinutes();
    var seconds = time.getSeconds();
    var value   = ((hours < 10) ? ' ' : '') + hours
                + ':' + ((minutes < 10) ? '0' : '') + minutes
                + ':' + ((seconds < 10) ? '0' : '') + seconds;
    display.setValue(value);
	}
	animate();
	window.setInterval(animate, 100);
}

segments();