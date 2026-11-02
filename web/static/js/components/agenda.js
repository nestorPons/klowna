(function () {
  "use strict";

  function pad(n) {
    return n < 10 ? "0" + n : String(n);
  }

  function isoFromDayMinutes(dateStr, minutes) {
    var total = 8 * 60 + minutes;
    var h = Math.floor(total / 60);
    var m = total % 60;
    return new Date(dateStr + "T" + pad(h) + ":" + pad(m) + ":00Z").toISOString();
  }

  function normalizeTime(value) {
    var parts = String(value || "").trim().match(/^(\d{1,2}):(\d{2})$/);
    if (!parts) return value;
    return pad(Number(parts[1])) + ":" + parts[2];
  }

  function isoFromDateTime(dateStr, timeStr) {
    return new Date(dateStr + "T" + normalizeTime(timeStr) + ":00Z").toISOString();
  }

  var escapeBound = false;
  var createOutsideBound = false;
  var restored = false;
  var storageKey = "klowna.agenda.state";
  var iconsKey = "klowna.agenda.icons";

  function readState() {
    try {
      var raw = localStorage.getItem(storageKey);
      return raw ? JSON.parse(raw) : null;
    } catch (e) {
      return null;
    }
  }

  function writeState(view, date, staff) {
    try {
      localStorage.setItem(storageKey, JSON.stringify({ view: view, date: date, staff: staff || "" }));
    } catch (e) {}
  }

  function readIcons() {
    try {
      return localStorage.getItem(iconsKey) !== "0";
    } catch (e) {
      return true;
    }
  }

  function writeIcons(visible) {
    try {
      localStorage.setItem(iconsKey, visible ? "1" : "0");
    } catch (e) {}
  }

  function applyIcons(agenda, visible) {
    agenda.classList.toggle("is-icons-hidden", !visible);
    var toggle = agenda.querySelector("[data-icons-toggle]");
    if (toggle) toggle.checked = visible;
  }

  function applyLayout(root) {
    if (!root || !root.querySelectorAll) return;
    root.querySelectorAll(".agenda-grid").forEach(function (grid) {
      var cols = parseInt(grid.dataset.cols, 10);
      if (cols > 0) {
        grid.style.gridTemplateColumns = "58px repeat(" + cols + ", 1fr)";
      }
    });
    root.querySelectorAll(".agenda-event[data-top][data-height]").forEach(function (el) {
      el.style.top = el.dataset.top + "%";
      el.style.height = el.dataset.height + "%";
    });
  }

  function syncTotal(picker) {
    if (!picker) return;
    var total = 0;
    picker.querySelectorAll('input[name="service_id"]:checked').forEach(function (i) {
      total += Number(i.dataset.duration) || 0;
    });
    var out = picker.parentElement.querySelector("[data-svc-total]");
    if (out) out.textContent = total;
    syncPayment(picker.closest("form"));
  }

  function money(value) {
    try {
      return new Intl.NumberFormat("es-ES", { style: "currency", currency: "EUR" }).format(value || 0);
    } catch (e) {
      return (Number(value) || 0).toFixed(2) + " €";
    }
  }

  function syncPayment(form) {
    if (!form) return;
    var status = form.querySelector('[name="status"]');
    var box = form.querySelector("[data-appt-payment]");
    if (box) box.hidden = !status || status.value !== "abonada";

    var base = 0;
    form.querySelectorAll('[data-svc-picker] input[name="service_id"]:checked').forEach(function (i) {
      base += Number(i.dataset.price) || 0;
    });
    var vat = Number(form.dataset.apptVat) || 0;
    var total = base * (1 + vat / 100);
    form.querySelectorAll("[data-appt-total]").forEach(function (el) {
      el.textContent = money(total);
    });
    syncDeposit(form);
  }

  // --- Señal de reserva ---
  function selectedDeposit(form) {
    var total = 0;
    form.querySelectorAll('[data-svc-picker] input[name="service_id"]:checked').forEach(function (i) {
      total += Number(i.dataset.deposit) || 0;
    });
    return Math.round(total * 100) / 100;
  }

  function syncDeposit(form) {
    if (!form) return;
    var box = form.querySelector("[data-appt-deposit]");
    if (!box) return;
    var due = selectedDeposit(form);
    box.hidden = due <= 0;
    if (due <= 0) return;

    var paid = Number(form.dataset.depositPaid) || 0;
    var remaining = Math.max(0, Math.round((due - paid) * 100) / 100);
    var check = box.querySelector('[name="deposit_paid"]');
    var dueEl = box.querySelector("[data-appt-deposit-due]");
    var statusEl = box.querySelector("[data-appt-deposit-status]");

    if (dueEl) dueEl.textContent = money(remaining > 0 ? remaining : paid);
    if (statusEl) {
      statusEl.textContent = remaining > 0
        ? "Pendiente: " + money(remaining)
        : "Cobrada: " + money(paid);
    }
    if (check) {
      if (remaining <= 0) {
        check.checked = true;
        check.disabled = true;
      } else {
        check.disabled = false;
        if (form.dataset.depositTouched !== "1") check.checked = false;
      }
    }
  }

  function closedRanges(el) {
    if (!el || !el.dataset || !el.dataset.closed) return [];
    return el.dataset.closed.split(",").filter(Boolean).map(function (part) {
      var bounds = part.split("-");
      return [Number(bounds[0]), Number(bounds[1])];
    });
  }

  function overlapsClosed(el, start, end) {
    var ranges = closedRanges(el);
    for (var i = 0; i < ranges.length; i++) {
      if (start < ranges[i][1] && end > ranges[i][0]) return true;
    }
    return false;
  }

  Klowna.register(function () {
    var agenda = document.getElementById("agenda");
    if (!agenda) return;

    applyIcons(agenda, readIcons());

    if (!restored) {
      restored = true;
      var params = new URLSearchParams(window.location.search);
      var hasParams = params.has("view") || params.has("date") || params.has("staff");
      var saved = readState();
      if (saved && !hasParams) {
        var sView = saved.view || agenda.dataset.view;
        var sDate = saved.date || "";
        var sStaff = saved.staff || "";
        if (sView !== agenda.dataset.view || (sDate && sDate !== agenda.dataset.date) || String(sStaff) !== String(agenda.dataset.staff || "")) {
          htmx.ajax("GET", buildAgendaURL(sView, sDate, sStaff), {
            target: "#agenda",
            swap: "outerHTML",
          });
          return;
        }
      }
    }

    applyLayout(agenda);
    if (agenda.dataset.bound === "1") return;
    agenda.dataset.bound = "1";

    function buildAgendaURL(view, date, staff) {
      var url = "/app/agenda?view=" + encodeURIComponent(view) + "&date=" + encodeURIComponent(date);
      if (staff && staff !== "0") url += "&staff=" + encodeURIComponent(staff);
      return url;
    }

    function syncURL() {
      var p = new URLSearchParams(window.location.search);
      p.set("view", agenda.dataset.view || "day");
      if (agenda.dataset.date) p.set("date", agenda.dataset.date); else p.delete("date");
      var staff = agenda.dataset.staff;
      if (staff && staff !== "0") p.set("staff", staff); else p.delete("staff");
      var qs = p.toString();
      try {
        history.replaceState(null, "", window.location.pathname + (qs ? "?" + qs : ""));
      } catch (e) {}
      writeState(agenda.dataset.view, agenda.dataset.date, staff);
    }

    syncURL();

    var modal = agenda.querySelector("#appt-modal");
    var form = agenda.querySelector("[data-appt-form]");
    var dragData = null;
    var overEl = null;

    agenda.querySelectorAll("[data-svc-picker]").forEach(syncTotal);
    agenda.querySelectorAll("form").forEach(syncPayment);

    function refresh() {
      var view = agenda.dataset.view || "day";
      var date = agenda.dataset.date || "";
      htmx.ajax("GET", buildAgendaURL(view, date, agenda.dataset.staff), {
        target: "#agenda",
        swap: "outerHTML",
      });
    }

    function postJSON(url, body, done) {
      fetch(url, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      })
        .then(function (r) { return r.json(); })
        .then(function (res) {
          if (res && res.ok) {
            done && done(res);
          } else if (res && res.error && window.alert) {
            window.alert(res.error);
          }
        })
        .catch(function () {});
    }

    // --- Modal de edición ---
    function openModal(el) {
      if (!form || !modal) return;
      var data = el.dataset;
      var start = data.start;

      form.querySelector('[name="id"]').value = data.id;
      form.querySelector('[name="patient"]').value = data.patient;
      form.querySelector('[name="assignee_id"]').value = (data.assignee === "0" || !data.assignee) ? "" : data.assignee;
      form.querySelector('[name="date"]').value = start.slice(0, 10);
      form.querySelector('[name="time"]').value = start.slice(11, 16);
      form.querySelector('[name="status"]').value = data.status || "pendiente";
      form.querySelector('[name="notes"]').value = data.notes || "";

      form.dataset.depositPaid = data.depositPaid || "0";
      form.dataset.depositTouched = "0";
      var depositMethod = form.querySelector('[name="deposit_method"]');
      if (depositMethod && data.depositMethod) depositMethod.value = data.depositMethod;

      var services = [];
      try { services = JSON.parse(data.services || "[]"); } catch (e) { services = []; }
      var selected = {};
      services.forEach(function (s) {
        if (s && s.service_id) selected[String(s.service_id)] = true;
      });

      var picker = form.querySelector("[data-svc-picker]");
      if (picker) {
        picker.querySelectorAll('input[name="service_id"]').forEach(function (i) {
          i.checked = !!selected[i.value];
        });
        syncTotal(picker);
      }
      syncPayment(form);

      modal.hidden = false;
    }

    function closeModal() {
      var m = modal || document.getElementById("appt-modal");
      if (m) m.hidden = true;
    }

    // --- Abrir "Nueva cita" desde una celda horaria ---
    function openCreate(date, time, assignee) {
      var details = agenda.querySelector(".app-create");
      if (!details) return;
      var createForm = details.querySelector("form");
      if (!createForm) return;
      var dateInput = createForm.querySelector('input[name="date"]');
      var timeInput = createForm.querySelector('input[name="time"]');
      var assigneeSelect = createForm.querySelector('select[name="assignee_id"]');
      if (dateInput && date) dateInput.value = date;
      if (timeInput && time) timeInput.value = time;
      if (assigneeSelect) assigneeSelect.value = assignee || "";
      details.open = true;
      if (timeInput) {
        try { timeInput.focus({ preventScroll: true }); } catch (err) { timeInput.focus(); }
      }
    }

    if (!escapeBound) {
      escapeBound = true;
      document.addEventListener("keydown", function (e) {
        if (e.key === "Escape") {
          var m = document.getElementById("appt-modal");
          if (m) m.hidden = true;
          var create = document.querySelector("#agenda .app-create");
          if (create) create.open = false;
        }
      });
    }

    if (!createOutsideBound) {
      createOutsideBound = true;
      document.addEventListener("click", function (e) {
        var create = document.querySelector("#agenda .app-create");
        if (!create || !create.open) return;
        if (e.target.closest("#agenda .app-create")) return;
        if (e.target.closest("[data-add-slot]")) return;
        create.open = false;
      });
    }

    function deleteAppointment() {
      var id = form.querySelector('[name="id"]').value;
      if (!id) return;
      fetch("/app/appointments/" + id, { method: "DELETE" })
        .then(function (r) { return r.json(); })
        .then(function (res) {
          if (res && res.ok) {
            closeModal();
            refresh();
          } else if (window.alert) {
            window.alert((res && res.error) || "No se pudo eliminar la cita.");
          }
        })
        .catch(function () {
          if (window.alert) window.alert("No se pudo eliminar la cita.");
        });
    }

    function createInvoice() {
      var id = form.querySelector('[name="id"]').value;
      if (!id) return;
      var button = agenda.querySelector("[data-appt-invoice]");
      if (button) button.disabled = true;
      fetch("/app/appointments/" + id + "/factura", { method: "POST" })
        .then(function (r) { return r.json(); })
        .then(function (res) {
          if (button) button.disabled = false;
          if (res && res.ok) {
            closeModal();
            if (window.openInvoice) window.openInvoice(res.invoice_id);
          } else if (window.alert) {
            window.alert((res && res.error) || "No se pudo crear la factura.");
          }
        }, function () {
          if (button) button.disabled = false;
          if (window.alert) window.alert("No se pudo crear la factura.");
        });
    }

    function submitForm(e) {
      e.preventDefault();
      var id = form.querySelector('[name="id"]').value;
      var date = form.querySelector('[name="date"]').value;
      var time = form.querySelector('[name="time"]').value;
      var assignee = form.querySelector('[name="assignee_id"]').value;
      var status = form.querySelector('[name="status"]').value;
      var notes = form.querySelector('[name="notes"]').value;
      var payment = form.querySelector('[name="payment_method"]');
      var paymentMethod = payment ? payment.value : "";
      var depositCheck = form.querySelector('[name="deposit_paid"]');
      var depositMethod = form.querySelector('[name="deposit_method"]');
      var depositPaid = depositCheck ? depositCheck.checked : false;

      var services = [];
      form.querySelectorAll('[data-svc-picker] input[name="service_id"]:checked').forEach(function (i) {
        services.push(Number(i.value));
      });

      if (!services.length) {
        if (window.confirm("No hay servicios seleccionados. ¿Eliminar la cita?")) {
          deleteAppointment();
        }
        return;
      }

      postJSON("/app/appointments/" + id, {
        assignee_id: assignee === "" ? null : Number(assignee),
        starts_at: isoFromDateTime(date, time),
        status: status,
        notes: notes,
        services: services,
        payment_method: paymentMethod,
        deposit_paid: depositPaid,
        deposit_method: depositMethod ? depositMethod.value : "",
      }, function (res) {
        closeModal();
        refresh();
        if (res && res.invoice_id && window.openInvoice) {
          window.openInvoice(res.invoice_id);
        }
      });
    }

    if (form) form.addEventListener("submit", submitForm);

    // --- Delegación de eventos ---
    agenda.addEventListener("change", function (e) {
      var staffSelect = e.target.closest("[data-staff-filter]");
      if (staffSelect) {
        var view = agenda.dataset.view || "week";
        var date = agenda.dataset.date || "";
        var url = "/app/agenda?view=" + view + "&date=" + date;
        if (staffSelect.value) url += "&staff=" + encodeURIComponent(staffSelect.value);
        htmx.ajax("GET", url, { target: "#agenda", swap: "outerHTML" });
        return;
      }
      var iconsToggle = e.target.closest("[data-icons-toggle]");
      if (iconsToggle) {
        var visible = iconsToggle.checked;
        writeIcons(visible);
        applyIcons(agenda, visible);
        return;
      }
      var statusSelect = e.target.closest("[data-appt-status]");
      if (statusSelect) {
        syncPayment(statusSelect.closest("form"));
        return;
      }
      var depositToggle = e.target.closest('[name="deposit_paid"]');
      if (depositToggle) {
        var depositForm = depositToggle.closest("form");
        if (depositForm) depositForm.dataset.depositTouched = depositToggle.checked ? "1" : "0";
        return;
      }
      var input = e.target.closest('input[name="service_id"]');
      if (!input) return;
      syncTotal(input.closest("[data-svc-picker]"));
    });

    agenda.addEventListener("click", function (e) {
      var addBtn = e.target.closest("[data-add-slot]");
      if (addBtn) {
        var body = addBtn.closest(".agenda-week-lane-body, .agenda-col-body");
        var holder = addBtn.closest("[data-day]");
        var date = (holder && holder.dataset.day) || agenda.dataset.date || "";
        var assignee = body ? body.dataset.assignee : "";
        if (!assignee) {
          var col = addBtn.closest(".agenda-col");
          if (col) assignee = col.dataset.assignee;
        }
        openCreate(date, addBtn.dataset.time, assignee === "0" ? "" : assignee);
        return;
      }
      if (e.target.closest("[data-modal-close]")) {
        closeModal();
        return;
      }
      if (e.target.closest("[data-appt-invoice]")) {
        createInvoice();
        return;
      }
      if (e.target.closest("[data-appt-delete]")) {
        if (window.confirm("¿Eliminar esta cita?")) deleteAppointment();
      }
    });

    // Doble clic sobre una cita -> modo edición
    agenda.addEventListener("dblclick", function (e) {
      var laneHead = e.target.closest(".agenda-week-lane-head");
      if (laneHead) {
        var assignee = laneHead.dataset.assignee;
        var staffSelect = agenda.querySelector("[data-staff-filter]");
        if (staffSelect && assignee && assignee !== "0") {
          e.preventDefault();
          staffSelect.value = assignee;
          staffSelect.dispatchEvent(new Event("change", { bubbles: true }));
        }
        return;
      }
      var editEl = e.target.closest(".agenda-event.is-editable, .agenda-chip.is-editable");
      if (!editEl) return;
      e.preventDefault();
      openModal(editEl);
    });

    // --- Arrastrar y soltar ---
    agenda.addEventListener("dragstart", function (e) {
      var el = e.target.closest(".agenda-event, .agenda-chip");
      if (!el || !el.classList.contains("is-editable")) return;
      dragData = {
        id: el.dataset.id,
        start: el.dataset.start,
        end: el.dataset.end,
        assignee: el.dataset.assignee,
      };
      el.classList.add("dragging");
      e.dataTransfer.effectAllowed = "move";
      try { e.dataTransfer.setData("text/plain", el.dataset.id); } catch (err) {}
    });

    agenda.addEventListener("dragend", function (e) {
      var el = e.target.closest(".agenda-event, .agenda-chip");
      if (el) el.classList.remove("dragging");
      dragData = null;
      if (overEl) { overEl.classList.remove("is-over"); overEl = null; }
    });

    agenda.addEventListener("dragover", function (e) {
      if (!dragData) return;
      var target = e.target.closest ? e.target.closest("[data-drop]") : null;
      if (!target) return;
      e.preventDefault();
      e.dataTransfer.dropEffect = "move";
      if (target !== overEl) {
        if (overEl) overEl.classList.remove("is-over");
        overEl = target;
        target.classList.add("is-over");
      }
    });

    agenda.addEventListener("dragleave", function (e) {
      var target = e.target.closest ? e.target.closest("[data-drop]") : null;
      if (target === overEl) {
        target.classList.remove("is-over");
        overEl = null;
      }
    });

    agenda.addEventListener("drop", function (e) {
      var target = e.target.closest ? e.target.closest("[data-drop]") : null;
      if (!target || !dragData) return;
      e.preventDefault();

      var body;
      var weekLane = target.closest(".agenda-week-lane-body");
      var colBody = target.closest(".agenda-col-body");
      if (weekLane) {
        var rect = weekLane.getBoundingClientRect();
        var offsetY = e.clientY - rect.top;
        var minutes = (offsetY / rect.height) * 12 * 60;
        minutes = Math.max(0, Math.min(12 * 60 - 15, minutes));
        minutes = Math.round(minutes / 15) * 15;
        var start = new Date(dragData.start);
        var dur = new Date(dragData.end) - start;
        if (overlapsClosed(weekLane, minutes, minutes + Math.round(dur / 60000))) {
          if (overEl) { overEl.classList.remove("is-over"); overEl = null; }
          dragData = null;
          if (window.alert) window.alert("Ese horario está cerrado para este empleado.");
          return;
        }
        var newStart = isoFromDayMinutes(weekLane.dataset.day, minutes);
        var newEnd = new Date(new Date(newStart).getTime() + dur);
        var laneAssignee = weekLane.dataset.assignee;
        body = {
          assignee_id: (laneAssignee === "0" || !laneAssignee) ? null : Number(laneAssignee),
          starts_at: newStart,
          ends_at: newEnd.toISOString(),
        };
      } else if (colBody) {
        var col = target.closest(".agenda-col");
        var dayEl = target.closest(".agenda-day");
        var assignee = col.dataset.assignee;
        var rect2 = colBody.getBoundingClientRect();
        var offsetY2 = e.clientY - rect2.top;
        var minutes2 = (offsetY2 / rect2.height) * 12 * 60;
        minutes2 = Math.max(0, Math.min(12 * 60 - 15, minutes2));
        minutes2 = Math.round(minutes2 / 15) * 15;
        var start2 = new Date(dragData.start);
        var dur2 = new Date(dragData.end) - start2;
        if (overlapsClosed(colBody, minutes2, minutes2 + Math.round(dur2 / 60000))) {
          if (overEl) { overEl.classList.remove("is-over"); overEl = null; }
          dragData = null;
          if (window.alert) window.alert("Ese horario está cerrado para este empleado.");
          return;
        }
        var newStart2 = isoFromDayMinutes(dayEl.dataset.day, minutes2);
        var newEnd2 = new Date(new Date(newStart2).getTime() + dur2);
        body = {
          assignee_id: (assignee === "0" || !assignee) ? null : Number(assignee),
          starts_at: newStart2,
          ends_at: newEnd2.toISOString(),
        };
      } else {
        var day = target.dataset.day;
        var s = new Date(dragData.start);
        var d = new Date(dragData.end) - s;
        var ns = new Date(day + "T" + pad(s.getUTCHours()) + ":" + pad(s.getUTCMinutes()) + ":00Z");
        var ne = new Date(ns.getTime() + d);
        var asg = (target.dataset.assignee !== undefined) ? target.dataset.assignee : dragData.assignee;
        body = {
          assignee_id: (asg === "0" || !asg) ? null : Number(asg),
          starts_at: ns.toISOString(),
          ends_at: ne.toISOString(),
        };
      }

      var id = dragData.id;
      if (overEl) { overEl.classList.remove("is-over"); overEl = null; }
      dragData = null;
      postJSON("/app/appointments/" + id + "/move", body, refresh);
    });
  });
})();
