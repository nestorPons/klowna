(function () {
  "use strict";

  function normalize(value) {
    var digits = String(value || "").replace(/\D/g, "");
    if (!digits) return "";
    if (digits.length <= 2) {
      digits = digits.padStart(2, "0") + "00";
    } else {
      digits = digits.slice(0, 4).padStart(4, "0");
    }
    var hour = Math.min(23, Number(digits.slice(0, 2)));
    var minute = Math.min(59, Number(digits.slice(2, 4)));
    return (hour < 10 ? "0" + hour : String(hour)) + ":" + (minute < 10 ? "0" + minute : String(minute));
  }

  function bind(input) {
    if (!input || input.dataset.time24Bound === "1") return;
    input.dataset.time24Bound = "1";
    input.addEventListener("blur", function () {
      if (input.value.trim() === "") return;
      input.value = normalize(input.value);
    });
  }

  Klowna.register(function (scope) {
    if (!scope || !scope.querySelectorAll) return;
    scope.querySelectorAll("[data-time24]").forEach(bind);
  });
})();
