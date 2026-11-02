(function () {
  "use strict";

  function modal() {
    return document.getElementById("invoice-modal");
  }

  function body() {
    return document.getElementById("invoice-modal-body");
  }

  function close() {
    var m = modal();
    if (m) m.hidden = true;
  }

  function open(id) {
    var m = modal();
    var target = body();
    if (!m || !target || !id) return;
    target.innerHTML = '<p class="muted">Cargando factura…</p>';
    m.hidden = false;
    fetch("/app/facturas/" + encodeURIComponent(id), {
      headers: { "HX-Request": "true" },
    })
      .then(function (response) {
        if (!response.ok) throw new Error("invoice");
        return response.text();
      })
      .then(function (html) {
        target.innerHTML = html;
      })
      .catch(function () {
        target.innerHTML = '<p class="muted">No se pudo cargar la factura.</p>';
      });
  }

  function edit(id) {
    var m = modal();
    var target = body();
    if (!m || !target || !id) return;
    target.innerHTML = '<p class="muted">Cargando…</p>';
    m.hidden = false;
    fetch("/app/facturas/" + encodeURIComponent(id) + "/editar", {
      headers: { "HX-Request": "true" },
    })
      .then(function (response) {
        if (!response.ok) throw new Error("invoice");
        return response.text();
      })
      .then(function (html) {
        target.innerHTML = html;
      })
      .catch(function () {
        target.innerHTML = '<p class="muted">No se pudo cargar la factura.</p>';
      });
  }

  function addLine() {
    var target = body();
    if (!target) return;
    var list = target.querySelector("[data-invoice-lines]");
    if (!list) return;
    var row = document.createElement("div");
    row.className = "invoice-edit-line";
    row.innerHTML =
      '<input name="line_description" maxlength="255" placeholder="Descripción">' +
      '<input name="line_qty" type="number" min="1" step="1" value="1">' +
      '<input name="line_price" type="number" min="0" step="0.01" value="0.00">' +
      '<button type="button" class="btn btn-icon btn-danger" data-line-remove title="Eliminar línea">' +
      '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M4 7h16"/><path d="M9 7V5a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/><path d="M6 7l1 13a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2l1-13"/><path d="M10 11v6M14 11v6"/></svg>' +
      "</button>";
    list.appendChild(row);
  }

  function submitEdit(form) {
    var id = form.getAttribute("action");
    if (!id) return;
    fetch(id, {
      method: "POST",
      body: new URLSearchParams(new FormData(form)).toString(),
      headers: { "Content-Type": "application/x-www-form-urlencoded", "HX-Request": "true" },
    })
      .then(function (response) {
        if (!response.ok) throw new Error("invoice");
        return response.text();
      })
      .then(function (html) {
        var target = body();
        if (target) target.innerHTML = html;
      })
      .catch(function () {
        if (window.alert) window.alert("No se pudo guardar la factura.");
      });
  }

  window.openInvoice = open;

  document.addEventListener("click", function (event) {
    if (event.target.closest("[data-invoice-close]")) {
      close();
      return;
    }
    var view = event.target.closest("[data-invoice-view]");
    if (view) {
      event.preventDefault();
      open(view.getAttribute("data-invoice-view"));
      return;
    }
    var editBtn = event.target.closest("[data-invoice-edit]");
    if (editBtn) {
      event.preventDefault();
      edit(editBtn.getAttribute("data-invoice-edit"));
      return;
    }
    if (event.target.closest("[data-line-add]")) {
      addLine();
      return;
    }
    var removeBtn = event.target.closest("[data-line-remove]");
    if (removeBtn) {
      var list = removeBtn.closest("[data-invoice-lines]");
      if (list && list.querySelectorAll(".invoice-edit-line:not(.invoice-edit-line-head)").length > 1) {
        removeBtn.closest(".invoice-edit-line").remove();
      }
      return;
    }
  });

  document.addEventListener("submit", function (event) {
    var form = event.target.closest("[data-invoice-edit-form]");
    if (!form) return;
    event.preventDefault();
    submitEdit(form);
  });

  document.addEventListener("keydown", function (event) {
    if (event.key === "Escape") close();
  });

  Klowna.register(function (scope) {
    if (!scope.querySelectorAll) return;
    scope.querySelectorAll("[data-invoice-id]").forEach(function (row) {
      if (row.dataset.invoiceBound === "1") return;
      row.dataset.invoiceBound = "1";
      row.addEventListener("dblclick", function (event) {
        if (event.target.closest("button, a")) return;
        open(row.getAttribute("data-invoice-id"));
      });
    });
  });
})();
