(function () {
  "use strict";

  var delay = 180;

  function submit(form) {
    if (form.dataset.autosaveBusy === "1" || form.dataset.autosaveDirty !== "1") return;
    form.dataset.autosaveDirty = "";
    form.dataset.autosaveBusy = "1";
    form.classList.add("is-saving");
    if (typeof form.requestSubmit === "function") {
      form.requestSubmit();
    } else {
      form.submit();
    }
  }

  function schedule(form) {
    if (!form || form.dataset.autosaveBusy === "1") return;
    window.clearTimeout(form._autosaveTimer);
    form._autosaveTimer = window.setTimeout(function () {
      if (!form.isConnected) return;
      if (form.contains(document.activeElement)) return;
      submit(form);
    }, delay);
  }

  Klowna.register(function (scope) {
    if (!scope.querySelectorAll) return;
    scope.querySelectorAll("form[data-autosave]").forEach(function (form) {
      if (form.dataset.autosaveBound === "1") return;
      form.dataset.autosaveBound = "1";

      form.addEventListener("input", function () {
        form.dataset.autosaveDirty = "1";
      });

      form.addEventListener("change", function () {
        form.dataset.autosaveDirty = "1";
        schedule(form);
      });

      form.addEventListener("focusout", function () {
        schedule(form);
      });

      form.addEventListener("submit", function () {
        form.dataset.autosaveDirty = "";
        form.classList.add("is-saving");
      });

      form.addEventListener("htmx:afterRequest", function () {
        form.dataset.autosaveBusy = "";
        form.classList.remove("is-saving");
      });
    });
  });
})();
