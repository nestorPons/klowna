(function () {
  "use strict";

  window.Klowna = window.Klowna || { components: [] };

  window.Klowna.register = function (init) {
    window.Klowna.components.push(init);
  };
})();
