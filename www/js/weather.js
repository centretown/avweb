function showGraph(canvas, color, readings) {
  var c = document.getElementById(canvas);
  var ctx = c.getContext("2d");
  var x = 0;
  ctx.beginPath();
  ctx.strokeStyle = color;
  var i = 0;
  // var readings = [4.5, 3.3, 6.6, 2.1, 5.5];
  ctx.moveTo(0, 0);
  for (let y of readings) {
    ctx.lineTo(x, 100 - y * 10);
    x += 50;
  }
  ctx.stroke();
}
