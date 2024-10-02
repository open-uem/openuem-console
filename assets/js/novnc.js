"use strict";

// RFB holds the API to connect and communicate with a VNC server
import RFB from "./core/rfb.js";

let rfb;
let desktopName;

// When this function is called we have
// successfully connected to a server
function connectedToServer(e) {
  status("Connected to " + desktopName);
}

// This function is called when we are disconnected
function disconnectedFromServer(e) {
  console.log(e);
  if (e.detail.clean) {
    status("Disconnected");
  } else {
    status("Wrong authentication");

    let inputs = document
      .getElementById("vncPIN")
      .getElementsByTagName("input");
    for (let i = 0; i < inputs.length; i++) {
      inputs[i].value = "";
    }
    document.getElementById("vncConnectPanel").classList.remove("uk-hidden");
  }
}

// When this function is called we have received
// a desktop name from the server
function updateDesktopName(e) {
  desktopName = e.detail.name;
}

// Show a status text in the top bar
function status(text) {
  document.getElementById("status").textContent = text;
}

// This function extracts the value of one variable from the
// query string. If the variable isn't defined in the URL
// it returns the default value instead.
function readQueryVariable(name, defaultValue) {
  // A URL with a query parameter can look like this:
  // https://www.example.com?myqueryparam=myvalue
  //
  // Note that we use location.href instead of location.search
  // because Firefox < 53 has a bug w.r.t location.search
  const re = new RegExp(".*[?&]" + name + "=([^&#]*)"),
    match = document.location.href.match(re);

  if (match) {
    // We have to decode the URL since want the cleartext value
    return decodeURIComponent(match[1]);
  }

  return defaultValue;
}

htmx.onLoad(function () {
  document
    .getElementById("connectVNC")
    ?.addEventListener("click", connectToServer);
});

function connectToServer() {
  // Read parameters specified in the URL query string
  // By default, use the host and port of server that served this file
  /* const host = readQueryVariable('host', window.location.hostname); */

  const host =
    document.getElementById("vncHostname").value || "lothlorien.openuem.eu";
  const port = document.getElementById("vncPort").value || "1443";
  const password =
    document.getElementById("vncPIN").getElementsByTagName("input")[6].value ||
    "";
  /* const path = readQueryVariable('path', 'websockify'); */
  const path = readQueryVariable("path", "ws");

  // | | |         | | |
  // | | | Connect | | |
  // v v v         v v v

  status("Connecting...");

  // Build the websocket URL used to connect
  let url;
  if (window.location.protocol === "https:") {
    url = "wss";
  } else {
    url = "ws";
  }
  url += "://" + host;
  if (port) {
    url += ":" + port;
  }
  url += "/" + path;

  // Creating a new RFB object will start a new connection
  rfb = new RFB(document.getElementById("screen"), url, {
    credentials: { password: password },
  });

  // Add listeners to important events from the RFB module
  rfb.addEventListener("connect", connectedToServer);
  rfb.addEventListener("disconnect", disconnectedFromServer);
  rfb.addEventListener("desktopname", updateDesktopName);

  // Set parameters that can be changed on an active connection
  rfb.viewOnly = readQueryVariable("view_only", false);
  rfb.scaleViewport = readQueryVariable("scale", false);

  document.getElementById("vncConnectPanel").classList.add("uk-hidden");
}
