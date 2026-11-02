(function () {
  "use strict";

  function money(value) {
    try {
      return new Intl.NumberFormat("es-ES", { style: "currency", currency: "EUR" }).format(value || 0);
    } catch (e) {
      return (Number(value) || 0).toFixed(2) + " €";
    }
  }

  function sync(form) {
    if (!form) return;
    var input = form.querySelector("[data-cash-counted-input]");
    var counted = input ? parseFloat(input.value) : 0;
    if (isNaN(counted)) counted = 0;
    counted = Math.round(counted * 100) / 100;

    var expected = parseFloat(form.dataset.expected) || 0;
    var diff = Math.round((counted - expected) * 100) / 100;
    var diffEl = form.querySelector("[data-cash-diff]");
    if (diffEl) {
      diffEl.textContent = money(diff);
      diffEl.classList.toggle("caja-negative", diff < 0);
      diffEl.classList.toggle("caja-positive", diff > 0);
    }
    form.dataset.counted = String(counted);
    form.dataset.diff = String(diff);
  }

  Klowna.register(function (scope) {
    if (!scope.querySelectorAll) return;
    scope.querySelectorAll("[data-cash-form]").forEach(function (form) {
      if (form.dataset.cashBound === "1") return;
      form.dataset.cashBound = "1";
      form.addEventListener("input", function (event) {
        if (event.target.closest("[data-cash-counted-input]")) sync(form);
      });
      form.addEventListener("submit", function (event) {
        sync(form);
        var diff = parseFloat(form.dataset.diff) || 0;
        if (diff !== 0 && window.confirm) {
          var ok = window.confirm("El descuadre es de " + money(diff) + ". ¿Cerrar la caja de todos modos?");
          if (!ok) event.preventDefault();
        }
      });
      sync(form);
    });
  });
})();
