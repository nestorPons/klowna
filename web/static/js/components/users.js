(function () {
  "use strict";

  function initials(name) {
    return (name || "").trim().split(/\s+/).slice(0, 2).map(function (word) {
      return word.charAt(0).toUpperCase();
    }).join("");
  }

  Klowna.register(function (scope) {
    var modal = scope.querySelector("#user-modal");
    var form = scope.querySelector("[data-user-form]");
    if (!modal || !form || form.dataset.bound === "1") return;
    form.dataset.bound = "1";

    var avatarInput = form.querySelector('[name="avatar_data"]');
    var fileInput = form.querySelector("[data-user-avatar-file]");
    var preview = form.querySelector("[data-user-avatar-preview]");
    var currentPass = form.querySelector("[data-user-current-pass]");
    var meta = form.querySelector("[data-user-meta]");

    function setPreview(avatar, name, color) {
      if (!preview) return;
      preview.style.setProperty("--staff", color || "#75162d");
      if (avatar) {
        preview.innerHTML = '<img src="' + avatar.replace(/"/g, "&quot;") + '" alt="">';
      } else {
        preview.textContent = initials(name) || "?";
      }
    }

    function closeModal() {
      modal.hidden = true;
    }

    function openModal(id) {
      fetch("/app/usuarios/" + encodeURIComponent(id), { headers: { "Accept": "application/json" } })
        .then(function (r) { return r.json(); })
        .then(function (res) {
          if (!res || !res.ok) throw new Error((res && res.error) || "No se pudo cargar el usuario.");
          form.querySelector('[name="id"]').value = res.id;
          form.querySelector('[name="name"]').value = res.name || "";
          form.querySelector('[name="email"]').value = res.email || "";
          form.querySelector('[name="color"]').value = res.color || "#75162d";
          form.querySelector('[name="icon"]').value = res.icon || "";
          form.querySelector('[name="password"]').value = "";
          form.querySelector('[name="current_password"]').value = "";
          avatarInput.value = res.avatar || "";
          if (currentPass) currentPass.hidden = !res.is_self;
          if (meta) meta.textContent = (res.is_manager ? "Gestor" : "Empleado") + (res.is_self ? " · tu perfil" : "");
          setPreview(res.avatar, res.name, res.color);
          modal.hidden = false;
          var nameInput = form.querySelector('[name="name"]');
          if (nameInput) nameInput.focus();
        })
        .catch(function (err) {
          if (window.alert) window.alert(err.message || "No se pudo cargar el usuario.");
        });
    }

    function resizeAvatar(file, done) {
      var reader = new FileReader();
      reader.onload = function () {
        var img = new Image();
        img.onload = function () {
          var size = 160;
          var canvas = document.createElement("canvas");
          canvas.width = size;
          canvas.height = size;
          var ctx = canvas.getContext("2d");
          var scale = Math.max(size / img.width, size / img.height);
          var w = img.width * scale;
          var h = img.height * scale;
          ctx.drawImage(img, (size - w) / 2, (size - h) / 2, w, h);
          done(canvas.toDataURL("image/jpeg", 0.85));
        };
        img.onerror = function () {
          if (window.alert) window.alert("No se pudo leer la imagen.");
        };
        img.src = reader.result;
      };
      reader.onerror = function () {
        if (window.alert) window.alert("No se pudo leer el archivo.");
      };
      reader.readAsDataURL(file);
    }

    form.addEventListener("submit", function (event) {
      event.preventDefault();
      var id = form.querySelector('[name="id"]').value;
      var payload = {
        name: form.querySelector('[name="name"]').value,
        email: form.querySelector('[name="email"]').value,
        color: form.querySelector('[name="color"]').value,
        icon: form.querySelector('[name="icon"]').value,
        avatar: avatarInput.value,
        password: form.querySelector('[name="password"]').value,
        current_password: form.querySelector('[name="current_password"]').value,
      };
      fetch("/app/usuarios/" + encodeURIComponent(id), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      })
        .then(function (r) { return r.json(); })
        .then(function (res) {
          if (res && res.ok) {
            window.location.reload();
            return;
          }
          if (window.alert) window.alert((res && res.error) || "No se pudo guardar el perfil.");
        })
        .catch(function () {
          if (window.alert) window.alert("No se pudo guardar el perfil.");
        });
    });

    if (fileInput) {
      fileInput.addEventListener("change", function () {
        var file = fileInput.files && fileInput.files[0];
        if (!file) return;
        if (file.size > 5 * 1024 * 1024) {
          if (window.alert) window.alert("La imagen no puede superar 5 MB.");
          fileInput.value = "";
          return;
        }
        resizeAvatar(file, function (data) {
          avatarInput.value = data;
          setPreview(data, form.querySelector('[name="name"]').value, form.querySelector('[name="color"]').value);
        });
        fileInput.value = "";
      });
    }

    var removeAvatar = form.querySelector("[data-user-avatar-remove]");
    if (removeAvatar) {
      removeAvatar.addEventListener("click", function () {
        avatarInput.value = "";
        setPreview("", form.querySelector('[name="name"]').value, form.querySelector('[name="color"]').value);
      });
    }

    document.addEventListener("click", function (event) {
      var trigger = event.target.closest("[data-user-edit]");
      if (trigger) {
        event.preventDefault();
        openModal(trigger.dataset.userEdit);
        return;
      }
      if (event.target.closest("[data-user-close]")) {
        closeModal();
      }
    });

    document.addEventListener("keydown", function (event) {
      if (event.key === "Escape" && !modal.hidden) closeModal();
    });
  });
})();
