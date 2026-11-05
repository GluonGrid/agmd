// agmd board: shared UI primitives used by board.js, panel.js and runs.js:
// the App namespace, escaping, toasts, time formatting, the API client, the
// icon set, and accessible confirm/prompt dialogs.
(function () {
  "use strict";

  var App = (window.App = {});

  var $ = (App.$ = function (id) { return document.getElementById(id); });

  var esc = (App.esc = function (s) {
    return String(s == null ? "" : s).replace(/[&<>"']/g, function (c) {
      return { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c];
    });
  });

  App.isMac = /Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent);
  App.modKey = App.isMac ? "⌘" : "Ctrl";

  // ---------- Icons ----------
  // One 16px stroke family, drawn here so the board never relies on emoji or
  // font glyphs for controls.
  var PATHS = {
    play: '<path d="M5 3.5v9l7.5-4.5z" fill="currentColor" stroke="none"/>',
    trash: '<path d="M3 4.5h10M6.5 4.5V3h3v1.5M4.5 4.5l.6 8.2a1 1 0 0 0 1 .8h3.8a1 1 0 0 0 1-.8l.6-8.2"/>',
    x: '<path d="m4 4 8 8M12 4l-8 8"/>',
    plus: '<path d="M8 3v10M3 8h10"/>',
    sun: '<circle cx="8" cy="8" r="2.8"/><path d="M8 1.5v1.6M8 12.9v1.6M1.5 8h1.6M12.9 8h1.6M3.4 3.4l1.1 1.1M11.5 11.5l1.1 1.1M3.4 12.6l1.1-1.1M11.5 4.5l1.1-1.1"/>',
    moon: '<path d="M13 9.6A5.3 5.3 0 0 1 6.4 3a5.3 5.3 0 1 0 6.6 6.6z"/>',
    monitor: '<rect x="2" y="3" width="12" height="8" rx="1.2"/><path d="M6 13.5h4M8 11v2.5"/>',
    grid: '<rect x="2.5" y="2.5" width="4.5" height="4.5" rx="1"/><rect x="9" y="2.5" width="4.5" height="4.5" rx="1"/><rect x="2.5" y="9" width="4.5" height="4.5" rx="1"/><rect x="9" y="9" width="4.5" height="4.5" rx="1"/>',
    folder: '<path d="M2 4.5a1 1 0 0 1 1-1h3l1.5 1.5H13a1 1 0 0 1 1 1v6a1 1 0 0 1-1 1H3a1 1 0 0 1-1-1z"/>',
    chevronLeft: '<path d="m10 3.5-4.5 4.5 4.5 4.5"/>',
    chevronRight: '<path d="m6 3.5 4.5 4.5L6 12.5"/>',
    comment: '<path d="M3 3h10a1 1 0 0 1 1 1v6a1 1 0 0 1-1 1H7l-3 2.5V11H3a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1z"/>',
    wait: '<circle cx="8" cy="8" r="5.8"/><path d="M8 5v3.2l2 1.3"/>',
    alert: '<path d="M8 2.5 14 13H2z"/><path d="M8 6.5v3M8 11.2v.1"/>',
    check: '<path d="m3.5 8.3 3 3 6-6.3"/>',
    requeue: '<path d="M3 8a5 5 0 0 1 8.6-3.5L13 6M13 2.5V6H9.5M13 8a5 5 0 0 1-8.6 3.5L3 10M3 13.5V10h3.5"/>',
    copy: '<rect x="5" y="5" width="8.5" height="8.5" rx="1.2"/><path d="M3 10.5V3.5a1 1 0 0 1 1-1h6.5"/>',
    minimize: '<path d="M3.5 12h9"/>',
    bug: '<rect x="5" y="4.5" width="6" height="8.5" rx="3"/><path d="M8 7v6M2.5 8.5H5M11 8.5h2.5M3 4.5l2 1.5M13 4.5l-2 1.5M3.5 13l1.6-1.2M12.5 13l-1.6-1.2"/>',
    sparkle: '<path d="M8 2.5 9.3 6.7 13.5 8l-4.2 1.3L8 13.5 6.7 9.3 2.5 8l4.2-1.3z"/>',
    broom: '<path d="M10.5 2.5 7.5 8M5 8.5h5l1 5H4z"/>',
    hash: '<path d="M6 2.5 5 13.5M11 2.5l-1 11M2.5 6h11M2 10h11"/>',
    logo: '<rect x="2" y="2.5" width="12" height="11" rx="2"/><path d="M6.5 2.5v11"/>'
  };

  App.icon = function (name, cls) {
    return '<svg class="icon' + (cls ? " " + cls : "") + '" viewBox="0 0 16 16" aria-hidden="true" focusable="false">' + (PATHS[name] || "") + "</svg>";
  };

  // Fill [data-icon] placeholders in static markup.
  App.fillIcons = function (root) {
    (root || document).querySelectorAll("[data-icon]").forEach(function (el) {
      el.innerHTML = App.icon(el.dataset.icon);
    });
  };

  // ---------- Toasts ----------
  App.toast = function (msg, kind) {
    var el = document.createElement("div");
    el.className = "toast" + (kind === "error" ? " error" : "");
    el.setAttribute("role", kind === "error" ? "alert" : "status");
    el.textContent = msg;
    $("toasts").appendChild(el);
    setTimeout(function () { el.remove(); }, kind === "error" ? 6000 : 2800);
  };

  // ---------- Time ----------
  App.age = function (iso) {
    var t = Date.parse(iso);
    if (isNaN(t)) return "";
    var s = Math.max(0, (Date.now() - t) / 1000);
    if (s < 45) return "now";
    if (s < 3600) return Math.round(s / 60) + "m";
    if (s < 86400) return Math.round(s / 3600) + "h";
    return Math.round(s / 86400) + "d";
  };
  App.relTime = function (iso) {
    var a = App.age(iso);
    return a === "now" ? "just now" : a ? a + " ago" : "";
  };

  App.storage = {
    get: function (k) { try { return localStorage.getItem(k); } catch (e) { return null; } },
    set: function (k, v) { try { localStorage.setItem(k, v); } catch (e) { /* private mode */ } }
  };

  // ---------- API ----------
  // api resolves to the parsed JSON body, or rejects with the server's message.
  App.api = function (method, url, body) {
    var opts = { method: method, headers: { Accept: "application/json" } };
    if (body !== undefined) {
      opts.headers["Content-Type"] = "application/json";
      opts.body = JSON.stringify(body);
    }
    return fetch(url, opts).then(function (r) {
      return r.text().then(function (txt) {
        var data = null;
        try { data = txt ? JSON.parse(txt) : null; } catch (e) { data = null; }
        if (!r.ok) {
          var err = new Error((data && data.error && data.error.message) || ("The server answered " + r.status + "."));
          err.code = data && data.error && data.error.code;
          throw err;
        }
        return data;
      });
    });
  };

  App.taskUrl = function (t, suffix) {
    return "/api/tasks/" + encodeURIComponent(t.project) + "/" + encodeURIComponent(t.id) + (suffix || "");
  };

  // ---------- Dialogs ----------
  // confirm and prompt use one <dialog>, so focus is trapped, Esc cancels and
  // focus returns to where it was.
  function openDialog(opts, withInput) {
    return new Promise(function (resolve) {
      var dlg = $("dialog");
      var returnTo = document.activeElement;
      dlg.innerHTML =
        '<form method="dialog" class="dialog-form">' +
        '<h2 id="dialog-title">' + esc(opts.title) + "</h2>" +
        (opts.body ? '<p id="dialog-body">' + esc(opts.body) + "</p>" : "") +
        (withInput ? '<label class="field-label" for="dialog-input">' + esc(opts.label || "") + '</label><input id="dialog-input" type="text" autocomplete="off" placeholder="' + esc(opts.placeholder || "") + '"><p class="field-error" id="dialog-error" role="alert"></p>' : "") +
        '<div class="modal-actions"><button class="btn" value="cancel" type="submit">' + esc(opts.cancelLabel || "Cancel") + "</button>" +
        '<button class="btn ' + (opts.danger ? "danger-solid" : "primary") + '" value="ok" type="submit" id="dialog-ok">' + esc(opts.confirmLabel || "OK") + "</button></div></form>";
      dlg.setAttribute("aria-labelledby", "dialog-title");
      if (opts.body) dlg.setAttribute("aria-describedby", "dialog-body"); else dlg.removeAttribute("aria-describedby");
      var input = withInput ? $("dialog-input") : null;
      dlg.querySelector("form").addEventListener("submit", function (e) {
        var ok = e.submitter && e.submitter.value === "ok";
        if (ok && input && opts.validate) {
          var problem = opts.validate(input.value.trim());
          if (problem) {
            e.preventDefault();
            $("dialog-error").textContent = problem;
            input.setAttribute("aria-invalid", "true");
            input.focus();
            return;
          }
        }
        dlg.dataset.result = ok ? "ok" : "cancel";
      });
      dlg.addEventListener("close", function onClose() {
        dlg.removeEventListener("close", onClose);
        var ok = dlg.dataset.result === "ok";
        delete dlg.dataset.result;
        if (returnTo && document.contains(returnTo)) returnTo.focus();
        resolve(withInput ? (ok ? input.value.trim() : null) : ok);
      });
      dlg.showModal();
      (input || (opts.danger ? dlg.querySelector('[value="cancel"]') : $("dialog-ok"))).focus();
    });
  };

  App.confirm = function (opts) { return openDialog(opts, false); };
  App.prompt = function (opts) { return openDialog(opts, true); };
})();
