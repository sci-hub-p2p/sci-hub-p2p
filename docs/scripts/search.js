$(document).ready(function () {
  $("a[href^='https://']").attr("target", "_blank");
  $("a[href^='http://']").attr("target", "_blank");
});
