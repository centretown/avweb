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

function toggleClass(id, toggle, klass) {
  if (toggle) {
    toggle = false;
    htmx.removeClass("#" + id, klass);
  } else {
    toggle = true;
    htmx.addClass("#" + id, klass);
  }
  return toggle;
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
  let clockFmt = new Intl.DateTimeFormat("en-US", {
    weekday: "short",
    month: "short",
    day: "numeric", // dateStyle: "full",
    hour: "numeric",
    minute: "numeric",
    timeZone: "America/New_York",
  });

  document.getElementById("clock").innerHTML = clockFmt.format(today);
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

function minMax(...lists) {
  var val = { min: 64000, max: -64000 };
  for (let list of lists) {
    for (let x of list) {
      if (val.max < x) val.max = x;
      if (val.min > x) val.min = x;
    }
  }
  return val;
}

const rain = "rgba(31, 144, 255, 255)";
const snow = "rgba(255, 255, 255, 255)";

function codeToColor(code) {
  if ((code >= 71 && code <= 77) || (code >= 85 && code <= 86)) {
    return snow;
  }
  return rain;
}

function clearCanvas(canvasId) {
  let canvas = document.getElementById(canvasId);
  let ctx = canvas.getContext("2d");
  ctx.clearRect(0, 0, canvas.width, canvas.height);
}

function showBars(canvasId, values, codes) {
  let canvas = document.getElementById(canvasId);
  let ctx = canvas.getContext("2d");
  let xStep = canvas.width / values.length;
  let yStep = 5;
  let precision = (7 * 0.1) / values.length;
  for (let i = 0; i < values.length; i++) {
    let x = i * xStep;
    let yStart = 1;
    let y = yStart;
    ctx.beginPath();
    ctx.strokeStyle = codeToColor(codes[i]);

    for (let value = values[i]; value > 0.0; value = value - precision) {
      ctx.moveTo(x, y);
      ctx.setLineDash([1, 1]);
      ctx.lineWidth = 1;
      ctx.lineTo(x + xStep - 2, y);
      x = i * xStep;
      y += yStep;
      if (y >= canvas.height) {
        yStart++;
        y = yStart;
      }
    }

    ctx.stroke();
  }
}

function showGraph(canvasId, color, min, max, values, lineWidth = 2) {
  let canvas = document.getElementById(canvasId);
  let ctx = canvas.getContext("2d");
  let xStep = canvas.width / (values.length - 1);
  let height = canvas.height;
  let yStep = height / values.length;
  if (max > min) yStep = height / (max - min);

  ctx.beginPath();
  ctx.strokeStyle = color;
  ctx.lineWidth = lineWidth;
  let x = 0;
  let y = height;
  ctx.font = "10px sans-serif";
  for (let val of values) {
    y = (max - val) * yStep;
    if (x == 0) ctx.moveTo(x, y);
    else ctx.lineTo(x, y);
    x += xStep;
  }
  ctx.setLineDash([]);
  ctx.stroke();
}

function toggleItem(sel, prefix, index, canvasId, list) {
  let listItems = list[index];
  let item = listItems.get(sel);
  item.selected = toggleClass(
    prefix + "-" + sel,
    item.selected,
    "title-selected",
  );
  if (item.selected) {
    htmx.removeClass("#" + prefix + "-max-" + sel, "hide");
    htmx.removeClass("#" + prefix + "-min-" + sel, "hide");
  } else {
    htmx.addClass("#" + prefix + "-max-" + sel, "hide");
    htmx.addClass("#" + prefix + "-min-" + sel, "hide");
  }
  drawItems(index, canvasId, list);
}

function drawItems(index, canvasId, list) {
  clearCanvas(canvasId);
  let listItems = list[index];
  let items = listItems.entries();
  for (let [key, item] of items) {
    if (item.selected && item.draw !== undefined) {
      item.draw(canvasId);
    }
  }
  showMinMax(canvasId);
}

function showMinMax(canvasId) {
  let canvas = document.getElementById(canvasId);
  let ctx = canvas.getContext("2d");
  let height = canvas.height;

  ctx.beginPath();
  ctx.lineWidth = 2;
  ctx.strokeStyle = "yellow";
  ctx.setLineDash([2, 2]);
  ctx.moveTo(0, 1);
  ctx.lineTo(canvas.width, 1);
  ctx.moveTo(0, canvas.height - 1);
  ctx.lineTo(canvas.width, canvas.height - 1);
  ctx.stroke();
}

function showFullDate(dt) {
  if (dt == undefined) {
    let dt = new Date();
  }
  let options = {
    timeStyle: "short",
    dateStyle: "full",
    timeZone: "America/New_York",
  };
  let fmt = new Intl.DateTimeFormat("en-US", options);
  return fmt.format(dt);
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
