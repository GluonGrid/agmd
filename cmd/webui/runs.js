// agmd board: running tasks with Claude Code. A run is one `claude` session
// started by the server; its tasks are claimed as workflow web-<session-id>.
(function () {
  "use strict";

  var App = window.App;
  var esc = App.esc;
  var $ = App.$;
  var runs = App.state.runs;
  var pending = []; // tasks shown in the run modal
  var activeRun = null;
  var timer = null;

  var PERM_WARN = {
    bypassPermissions: { cls: "danger", text: "Unattended: Claude edits files and runs commands with no confirmation prompts. Only start tasks you trust on this machine." },
    acceptEdits: { cls: "danger", text: "Claude auto-applies file edits but still prompts for other sensitive actions." },
    plan: { cls: "", text: "Read-only: Claude plans the work without editing. Run again in another mode to apply it." },
    "default": { cls: "", text: "Claude prompts as usual, so the run may stop on a question it can't show here." }
  };

  function updateWarn() {
    var w = PERM_WARN[$("run-perm").value] || PERM_WARN["default"];
    $("run-warn").className = "warn " + w.cls;
    $("run-warn").textContent = w.text;
  }

  // Blocked and finished tasks can't be worked on yet, so a run skips them.
  function openModal(tasks) {
    tasks = tasks.filter(Boolean);
    pending = tasks.filter(function (t) { return t.status !== "completed" && t.computed_status !== "blocked"; });
    var skipped = tasks.length - pending.length;
    if (!pending.length) {
      App.toast(skipped ? "Nothing to run: blocked and done tasks can't be started." : "Select tasks to run first.", "error");
      return;
    }
    $("run-tasklist").innerHTML = pending.map(function (t) {
      return "<li><code>" + esc(t.project + "/" + t.name) + "</code> " + esc(t.subject) + "</li>";
    }).join("") + (skipped ? '<li class="muted">' + skipped + " blocked or done task" + (skipped === 1 ? "" : "s") + " left out</li>" : "");
    updateWarn();
    returnFocus = document.activeElement;
    $("run-modal").classList.add("open");
    $("run-perm").focus();
  }

  var returnFocus = null;
  function closeModal() {
    $("run-modal").classList.remove("open");
    if (returnFocus && document.contains(returnFocus)) returnFocus.focus();
    returnFocus = null;
  }

  function launch() {
    var tasks = pending.map(function (t) { return { project: t.project, id: t.id }; });
    $("run-go").disabled = true;
    App.api("POST", "/api/run", { tasks: tasks, permissionMode: $("run-perm").value }).then(function (info) {
      closeModal();
      App.clearSelection();
      runs[info.sessionId] = info;
      App.toast("Started run " + info.sessionId.slice(0, 8) + " for " + tasks.length + " task" + (tasks.length === 1 ? "" : "s"));
      openConsole(info.sessionId);
      ensurePolling();
      App.loadState();
    }).catch(function (err) {
      App.toast(err.message, "error");
    }).finally(function () {
      $("run-go").disabled = false;
    });
  }

  // ---------- Console ----------
  function statusHTML(r) {
    if (!r.done) return '<span class="st-dot running"></span> running';
    return r.exitError ? '<span class="st-dot fail"></span> exited' : '<span class="st-dot ok"></span> finished';
  }

  function openConsole(id) {
    var info = runs[id];
    if (!info) return;
    activeRun = id;
    var con = $("run-console");
    con.classList.remove("min");
    $("run-title").textContent = id.slice(0, 8) + " · " + (info.tasks || []).map(function (t) { return t.name; }).join(", ");
    $("run-meta").innerHTML = "<div>mode " + esc(info.permissionMode) + " · cwd " + esc(info.cwd) + "</div><div>transcript " + esc(info.transcriptPath) + "</div>";
    $("run-stat").innerHTML = statusHTML(info);
    $("run-events").innerHTML = '<div class="run-empty">Loading transcript…</div>';
    con.classList.add("open");
    refreshEvents();
    renderTray();
  }

  function closeConsole() {
    activeRun = null;
    $("run-console").classList.remove("open", "min");
    renderTray();
  }

  function renderEvents(events) {
    var box = $("run-events");
    var atBottom = box.scrollHeight - box.scrollTop - box.clientHeight < 40;
    if (!events || !events.length) {
      box.innerHTML = '<div class="run-empty">No output yet…</div>';
      return;
    }
    box.innerHTML = events.map(function (ev) {
      var who = ev.kind === "summary" ? "summary" : (ev.role || ev.kind || "");
      return '<div class="run-ev"><span class="who">' + esc(who) + "</span>" + esc(ev.text || "") + "</div>";
    }).join("");
    if (atBottom) box.scrollTop = box.scrollHeight;
  }

  function refreshEvents() {
    var id = activeRun;
    if (!id) return;
    App.api("GET", "/api/run/" + encodeURIComponent(id)).then(function (data) {
      if (activeRun !== id) return;
      merge(data);
      renderEvents(data.events || []);
      $("run-stat").innerHTML = statusHTML(runs[id]);
    }).catch(function () { /* transient */ });
  }

  function merge(data) {
    var prev = runs[data.sessionId] || {};
    Object.keys(data).forEach(function (k) { if (k !== "events") prev[k] = data[k]; });
    runs[data.sessionId] = prev;
  }

  // ---------- Runs list ----------
  function list() {
    return Object.keys(runs).map(function (id) { return runs[id]; })
      .sort(function (a, b) { return (b.startedAt || "").localeCompare(a.startedAt || ""); });
  }

  function renderTray() {
    var all = list();
    var running = all.filter(function (r) { return !r.done; }).length;
    $("btn-runs").classList.toggle("hidden", all.length === 0);
    $("runs-badge").textContent = running ? running + " running" : String(all.length);
    $("runs-list").innerHTML = all.length ? all.map(function (r) {
      var names = (r.tasks || []).map(function (t) { return t.name; }).join(", ");
      return '<div class="run-item' + (r.sessionId === activeRun ? " active" : "") + '" data-id="' + esc(r.sessionId) + '">' +
        '<span class="st-dot ' + (!r.done ? "running" : r.exitError ? "fail" : "ok") + '"></span>' +
        '<div class="ri-main"><div class="ri-title">' + esc(r.sessionId.slice(0, 8)) + '</div><div class="ri-sub">' + esc(names) + "</div></div>" +
        '<span class="ri-stat">' + esc(App.relTime(r.startedAt)) + "</span></div>";
    }).join("") : '<div class="runs-empty">No runs yet.</div>';
  }

  function refreshList() {
    return App.api("GET", "/api/runs").then(function (data) {
      var finished = false;
      (data.runs || []).forEach(function (r) {
        var prev = runs[r.sessionId];
        if (prev && !prev.done && r.done) {
          finished = true;
          App.toast("Run " + r.sessionId.slice(0, 8) + (r.exitError ? " exited with an error." : " finished."), r.exitError ? "error" : undefined);
        }
        merge(r);
      });
      renderTray();
      if (finished && activeRun) refreshEvents();
      if (!list().some(function (r) { return !r.done; })) stopPolling();
    }).catch(function () { /* transient */ });
  }

  function ensurePolling() {
    if (!timer) timer = setInterval(function () { refreshList(); refreshEvents(); }, 1500);
  }
  function stopPolling() {
    if (timer) { clearInterval(timer); timer = null; }
  }

  function copyResume() {
    var info = runs[activeRun];
    if (!info) return;
    var cmd = info.resumeHint || "claude --resume " + activeRun;
    var done = function () { App.toast("Copied: " + cmd); };
    if (navigator.clipboard) navigator.clipboard.writeText(cmd).then(done, function () { App.toast(cmd); });
    else App.toast(cmd);
  }

  App.Runs = {
    openModal: openModal,
    openConsole: openConsole,
    has: function (id) { return !!runs[id]; }
  };

  document.addEventListener("DOMContentLoaded", function () {
    var perm = App.storage.get("agmd-run-mode");
    $("run-perm").value = PERM_WARN[perm] ? perm : "acceptEdits";
    $("run-perm").addEventListener("change", function () {
      App.storage.set("agmd-run-mode", $("run-perm").value);
      updateWarn();
    });
    $("run-cancel").addEventListener("click", closeModal);
    $("run-go").addEventListener("click", launch);
    $("run-modal").addEventListener("click", function (e) { if (e.target === e.currentTarget) closeModal(); });
    document.addEventListener("keydown", function (e) {
      if (e.key === "Escape" && $("run-modal").classList.contains("open")) closeModal();
    });
    $("run-close").addEventListener("click", closeConsole);
    $("run-min").addEventListener("click", function () { $("run-console").classList.toggle("min"); });
    $("run-resume").addEventListener("click", copyResume);
    $("btn-runs").addEventListener("click", function () { $("runs-tray").classList.toggle("hidden"); });
    $("runs-list").addEventListener("click", function (e) {
      var item = e.target.closest(".run-item");
      if (!item) return;
      openConsole(item.dataset.id);
      $("runs-tray").classList.add("hidden");
    });
    document.addEventListener("click", function (e) {
      if (!e.target.closest("#runs-tray") && !e.target.closest("#btn-runs")) $("runs-tray").classList.add("hidden");
    });
    refreshList().then(function () {
      if (list().some(function (r) { return !r.done; })) ensurePolling();
    });
  });
})();
