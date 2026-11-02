(function () {
  "use strict";

  Klowna.register(function (scope) {
    var toggle = scope.querySelector(".app-menu-toggle");
    var sidebar = scope.querySelector(".app-sidebar");

    if (!toggle || !sidebar) return;

    toggle.addEventListener("click", function () {
      var open = sidebar.classList.toggle("is-open");
      toggle.setAttribute("aria-expanded", String(open));
    });
  });
})();
