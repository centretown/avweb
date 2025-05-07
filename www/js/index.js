const blank = "blank";
let swapID;

function doAction(action) {
  const slot = "slot-" + action;
  let target = document.getElementById(slot);
  if (target) {
    htmx.swap("#" + slot, "", { swapStyle: "delete" });
    return;
  }

  htmx.trigger("#" + action, "click");
}

var hideLeft = true;
function toggleMenu(id) {
  if (hideLeft) {
    hideLeft = false;
    htmx.removeClass("#" + id, "hide");
  } else {
    hideLeft = true;
    htmx.addClass("#" + id, "hide");
  }
}

var hideChat = true;
function toggleChat(id) {
  if (hideChat) {
    hideChat = false;
    htmx.removeClass("#" + id, "hide");
  } else {
    hideChat = true;
    htmx.addClass("#" + id, "hide");
  }
}

function startTime() {
  const today = new Date();
  const h = today.getHours();
  let m = today.getMinutes();
  m = m < 10 ? "0" + m : m;
  document.getElementById("clock").innerHTML = h + ":" + m;
  setTimeout(startTime, 1000 * (60 - today.getSeconds()));
}

// const chatId = "chat";
let drag_data = {};
let chat_data = {};
let slots = new Map();

function dragstartHandler(ev) {
  ev.dataTransfer.effectAllowed = "move";
  drag_data.offsetX = ev.offsetX;
  drag_data.offsetY = ev.offsetY;
}

function dragendHandler(ev) {
  const target = ev.target;
  const id = target.id;
  let data = {};
  data.X = ev.clientX - drag_data.offsetX;
  data.Y = ev.clientY - drag_data.offsetY;
  slots.set(id, data);
  target.style.left = data.X + "px";
  target.style.top = data.Y + "px";
  setdraggable(ev.target.id, false);
}

function addDragHandlers(id) {
  const target = document.getElementById(id);
  if (target !== undefined) {
    target.addEventListener("dragstart", dragstartHandler);
    target.addEventListener("dragend", dragendHandler);
  }
}

function removeDragHandlers(id) {
  const target = document.getElementById(id);
  if (target !== undefined) {
    target.removeEventListener("dragstart", dragstartHandler);
    target.removeEventListener("dragend", dragendHandler);
  }
}

function setdraggable(id, draggable) {
  const target = document.getElementById(id);
  if (target !== undefined) {
    document.getElementById(id).setAttribute("draggable", draggable);
    target.style.cursor = draggable ? "move" : "auto";
  }
}

const Anonymous = "Anonymous";
function postName() {
  const target = document.getElementById("postname");
  if (target === undefined) return Anonymous;
  if (target === "") return Anonymous;
  return target.value;
}

function clearMessage(id) {
  const target = document.getElementById(id);
  if (target !== undefined) {
    target.value = "";
  }
}

function currentSource() {
  const target = document.getElementById("source");
  if (target !== undefined) {
    console.log(target.src);
    return target.src;
  }
  return "target not found";
}

window.addEventListener("DOMContentLoaded", () => {
  addDragHandlers("chat");
});

window.addEventListener("htmx:load", function (evt) {
  const target = evt.detail.elt;
  let id = target.id;
  if (!id.startsWith("slot-")) {
    return;
  }
  if (slots.has(id)) {
    data = slots.get(id);
    target.style.left = data.X + "px";
    target.style.top = data.Y + "px";
  }
  addDragHandlers(id);
});

function minMax(lows, highs) {
  var val = { min: lows[0], max: highs[0] };
  for (let x of highs) if (val.max < x) val.max = x;
  for (let x of lows) if (val.min > x) val.min = x;
  return val;
}

let chartPadding = 20;
let chartMargin = 26;

function showGraph(canvasId, color, values, min, max) {
  let canvas = document.getElementById(canvasId);
  let ctx = canvas.getContext("2d");
  let xStep = canvas.width / (values.length - 1);
  let height = canvas.height - 2 * chartPadding;
  let yStep = height / (max - min);

  ctx.beginPath();
  ctx.strokeStyle = color;
  let x = chartMargin;
  let y = height;
  ctx.font = "10px sans-serif";
  for (let val of values) {
    y = (max - val) * yStep;
    if (x == 0) ctx.moveTo(x, y + chartPadding);
    else ctx.lineTo(x, y + chartPadding);
    x += xStep;
  }
  ctx.setLineDash([]);
  ctx.stroke();
}

function showTimes(canvasId, times, options) {
  let canvas = document.getElementById(canvasId);
  let ctx = canvas.getContext("2d");
  let xStep = (canvas.width - chartMargin) / (times.length - 1);
  let x = 0;
  let fmt = new Intl.DateTimeFormat("en-US", options);
  for (let t of times) {
    let day = new Date(t);
    console.log(fmt.format(day));
    ctx.fillText(fmt.format(day), x, 12);
    x += xStep;
  }
}

function showHours(canvasId, times) {
  let options = {
    hour: "numeric",
  };
  let intervals = [];
  let interval = times.length / 6;
  if (interval <= 1) intervals = times;
  else {
    for (let i = 0; i < times.length; i++) {
      if (i % interval == 0) {
        intervals.push(times[i]);
      }
    }
    intervals.push(times[times.length - 1]);
  }
  showTimes(canvasId, intervals, options);
}

function showDays(canvasId, times) {
  let options = {
    weekday: "short",
  };
  showTimes(canvasId, times, options);
}

function showMinMax(canvasId, min, max, times) {
  let canvas = document.getElementById(canvasId);
  let ctx = canvas.getContext("2d");
  let height = canvas.height - 2 * chartPadding;
  ctx.beginPath();
  ctx.setLineDash([5, 10]);
  ctx.moveTo(0, chartPadding);
  ctx.lineTo(canvas.width, chartPadding);
  ctx.fillStyle = "white";
  ctx.fillText(max, canvas.width / 2 - 10, chartPadding + 10 - 1);
  ctx.moveTo(0, height + chartPadding);
  ctx.lineTo(canvas.width, height + chartPadding);
  ctx.fillText(min, canvas.width / 2 - 10, height + chartPadding + 10 - 1);
  ctx.strokeStyle = "grey";
  ctx.stroke();
}

const gamepads = {};

function gamepadHandler(event, connected) {
  const gamepad = event.gamepad;
  // Note:
  // gamepad === navigator.getGamepads()[gamepad.index]

  if (connected) {
    gamepads[gamepad.index] = gamepad;
  } else {
    delete gamepads[gamepad.index];
  }
}

window.addEventListener("gamepaddisconnected", (e) => {
  console.log(
    "Gamepad disconnected from index %d: %s",
    e.gamepad.index,
    e.gamepad.id,
  );
});

window.addEventListener(
  "gamepaddisconnected",
  (e) => {
    gamepadHandler(e, false);
  },
  false,
);
