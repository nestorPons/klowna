(function () {
  "use strict";

  function boot(scope) {
    if (!window.Klowna) return;
    window.Klowna.components.forEach(function (init) {
      try {
        init(scope);
      } catch (err) {
        if (window.console) console.error(err);
      }
    });
  }

  function onReady() {
    boot(document);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", onReady);
  } else {
    onReady();
  }

  document.addEventListener("htmx:afterSettle", function (event) {
    boot(event.target);
  });
})();
