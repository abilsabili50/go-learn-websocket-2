var app = {};
app.ws = undefined;
app.container = undefined;

app.init = function () {
	if (!window.WebSocket) {
		alert("Your browser doesn't support WebSocket");
		return;
	}

	var username = prompt("Enter your username please:") || "No name";
	document.querySelector(".username").innerHTML = username;
	var room = prompt("Please write the room you wanna go to:") || "";

	app.container = document.querySelector(".container");

	app.ws = new WebSocket("ws://localhost:8080/ws?username=" + username);

	app.ws.onopen = function () {
		app.ws.send(
			JSON.stringify({
				type: "login",
				room: room,
			})
		);
		var message = "<b>me</b>: connected";
		app.print(message);
	};

	app.ws.onmessage = function (event) {
		var res = JSON.parse(event.data);

		var message = "";
		switch (res.Type) {
			case "New User":
				message = "User <b>" + res.From + "</b>: connected";
				break;
			case "Leave":
				message = "User <b>" + res.From + "</b>: disconnected";
				break;
			default:
				message = "<b>" + res.From + "</b>: " + res.Message;
		}

		app.print(message);
	};

	app.ws.onclose = function () {
		var message = "<b>me</b>: disconnected";
		app.print(message);
	};

	window.onunload = function () {
		app.ws.send(
			JSON.stringify({
				type: "disconnect",
				room,
			})
		);
	};

	app.print = function (message) {
		var element = document.createElement("p");
		element.innerHTML = message;
		app.container.append(element);
	};

	app.doSendMessage = function () {
		var messageRaw = document.querySelector(".input-message").value;
		app.ws.send(
			JSON.stringify({
				type: "chat",
				message: messageRaw,
				room,
			})
		);

		var message = "<b>me</b>: " + messageRaw;
		app.print(message);

		document.querySelector(".input-message").value = "";
	};
};

window.onload = app.init;
