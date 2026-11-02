(function () {
  "use strict";

  function luminance(hex) {
    var value = parseInt(hex.slice(1), 16);
    var channels = [(value >> 16) & 255, (value >> 8) & 255, value & 255].map(function (c) {
      var v = c / 255;
      return v <= 0.03928 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4);
    });
    return 0.2126 * channels[0] + 0.7152 * channels[1] + 0.0722 * channels[2];
  }

  function apply(input) {
    if (!/^#[0-9a-f]{6}$/i.test(input.value)) return;
    var root = document.documentElement;
    root.style.setProperty(input.dataset.var, input.value);
    if (input.name === "accent") {
      root.style.setProperty("--on-accent", luminance(input.value) > 0.24 ? "#2e1b1f" : "#fffdf8");
    }
    if (input.name === "bg") {
      root.style.setProperty("--scheme", luminance(input.value) < 0.24 ? "dark" : "light");
    }
  }

  function updateCode(input) {
    var code = input.parentNode.querySelector("code");
    if (code) code.textContent = input.value;
  }

  Klowna.register(function (scope) {
    var form = scope.querySelector("[data-theme-form]");
    if (!form) return;

    var inputs = Array.prototype.slice.call(form.querySelectorAll('input[type="color"][data-var]'));
    var byName = {};
    inputs.forEach(function (input) {
      byName[input.name] = input;
    });

    var custom = form.querySelector('input[name="preset"][value="custom"]');

    inputs.forEach(function (input) {
      input.addEventListener("input", function () {
        if (custom) custom.checked = true;
        apply(input);
        updateCode(input);
      });
    });

    form.querySelectorAll('input[name="preset"]').forEach(function (radio) {
      radio.addEventListener("change", function () {
        if (radio.value === "custom") return;
        inputs.forEach(function (input) {
          var value = radio.dataset[input.name];
          if (!value) return;
          input.value = value;
          apply(input);
          updateCode(input);
        });
      });
    });

    var reset = form.querySelector("[data-theme-reset]");
    if (reset) {
      reset.addEventListener("click", function () {
        var def = form.querySelector('input[name="preset"][value="bordeaux"]');
        if (!def) return;
        def.checked = true;
        def.dispatchEvent(new Event("change"));
      });
    }
  });
})();
