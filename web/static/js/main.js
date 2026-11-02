(function () {
  "use strict";

  var header = document.querySelector(".site-header");

  function onScroll() {
    if (!header) return;
    header.classList.toggle("is-scrolled", window.scrollY > 8);
  }

  onScroll();
  window.addEventListener("scroll", onScroll, { passive: true });

  var toggle = document.querySelector(".nav-toggle");
  var nav = document.getElementById("primary-nav");

  if (toggle && nav) {
    toggle.addEventListener("click", function () {
      var open = toggle.getAttribute("aria-expanded") === "true";
      toggle.setAttribute("aria-expanded", String(!open));
      nav.classList.toggle("is-open", !open);
      document.body.classList.toggle("nav-open", !open);
    });

    nav.querySelectorAll("a").forEach(function (link) {
      link.addEventListener("click", function () {
        toggle.setAttribute("aria-expanded", "false");
        nav.classList.remove("is-open");
        document.body.classList.remove("nav-open");
      });
    });
  }

  document.querySelectorAll(".nav-item.has-dropdown > .nav-link").forEach(function (button) {
    button.addEventListener("click", function (event) {
      if (!window.matchMedia("(max-width: 880px)").matches) return;
      event.preventDefault();
      var item = button.parentElement;
      var open = item.classList.toggle("is-open");
      button.setAttribute("aria-expanded", String(open));
    });
  });

  var tabs = Array.prototype.slice.call(document.querySelectorAll("[data-tab]"));

  tabs.forEach(function (tab) {
    tab.addEventListener("click", function () {
      tabs.forEach(function (other) {
        other.classList.remove("is-active");
        other.setAttribute("aria-selected", "false");
      });
      tab.classList.add("is-active");
      tab.setAttribute("aria-selected", "true");

      document.querySelectorAll(".tab-panel").forEach(function (panel) {
        panel.classList.remove("is-active");
      });

      var panel = document.getElementById(tab.dataset.tab);
      if (panel) panel.classList.add("is-active");
    });
  });

  function resetCaptcha(form) {
    var widget = form.querySelector(".cf-turnstile");
    if (widget && window.turnstile) {
      try {
        window.turnstile.reset(widget);
      } catch (err) {
        /* el widget aún no está listo */
      }
    }
  }

  document.querySelectorAll("[data-ajax-form]").forEach(function (form) {
    var status = form.querySelector("[data-form-status]");

    form.addEventListener("submit", function (event) {
      event.preventDefault();

      if (!form.reportValidity()) return;

      var button = form.querySelector('button[type="submit"]');
      var originalLabel = button ? button.textContent : "";

      if (button) {
        button.disabled = true;
        button.textContent = "Enviando…";
      }

      if (status) {
        status.textContent = "";
        status.className = "form-status";
      }

      var payload = Object.fromEntries(new FormData(form).entries());
      var consent = form.querySelector('[name="consent"]');
      payload.consent = Boolean(consent && consent.checked);

      fetch("/api/leads", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Accept": "application/json"
        },
        body: JSON.stringify(payload)
      })
        .then(function (response) {
          return response.json().catch(function () {
            return {};
          }).then(function (data) {
            if (!response.ok || !data.ok) {
              throw new Error(data.error || "No hemos podido enviar el formulario.");
            }
            return data;
          });
        })
        .then(function () {
          form.reset();
          if (status) {
            status.textContent = "¡Gracias! Te contactaremos en menos de 24 h laborables.";
            status.classList.add("is-success");
          }
        })
        .catch(function (error) {
          if (status) {
            status.textContent = error.message || "No hemos podido enviar el formulario.";
            status.classList.add("is-error");
          }
        })
        .finally(function () {
          resetCaptcha(form);
          if (button) {
            button.disabled = false;
            button.textContent = originalLabel;
          }
        });
    });
  });
})();
