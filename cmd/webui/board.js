// agmd board: state, API, sidebar, columns, cards, drag and drop, filters and
// live updates. panel.js (task detail) and runs.js (Claude runs) build on the
// App object defined here.
(function () {
  "use strict";

  var App = (window.App = {});
  var state = (App.state = {
    tasks: [],
    byId: {},
    projects: [],
    currentProject: "",
    ttl: 0, // seconds a claim survives without a heartbeat; 0 = never expires
    project: null, // selected project; "" = all projects
    filters: { q: "", claim: "", feature: "" },
    selected: {},
    version: null,
    dragId: null,
    collapsed: { completed: true },
    runs: {}
  });

  var COLUMNS = [
    { key: "blocked", label: "Blocked", status: "pending" },
    { key: "ready", label: "Ready", status: "pending" },
    { key: "in_progress", label: "In progress", status: "in_progress" },
    { key: "completed", label: "Done", status: "completed" }
  ];

  var ICONS = {
    blocked: '<svg class="col-icon" viewBox="0 0 16 16"><circle cx="8" cy="8" r="6.2" fill="none" stroke="var(--amber)" stroke-width="1.6"/><path d="M4 12 12 4" stroke="var(--amber)" stroke-width="1.6"/></svg>',
    ready: '<svg class="col-icon" viewBox="0 0 16 16"><circle cx="8" cy="8" r="6.2" fill="none" stroke="var(--fg-dim)" stroke-width="1.6"/></svg>',
    in_progress: '<svg class="col-icon" viewBox="0 0 16 16"><circle cx="8" cy="8" r="6.2" fill="none" stroke="var(--accent)" stroke-width="1.6"/><path d="M8 3.6a4.4 4.4 0 0 1 0 8.8z" fill="var(--accent)"/></svg>',
    completed: '<svg class="col-icon" viewBox="0 0 16 16"><circle cx="8" cy="8" r="7" fill="var(--green)"/><path d="m5 8.2 2 2 4-4.2" fill="none" stroke="var(--surface)" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/></svg>'
  };

  var $ = (App.$ = function (id) { return document.getElementById(id); });

  // ---------- Utilities ----------
  var esc = (App.esc = function (s) {
    return String(s == null ? "" : s).replace(/[&<>"']/g, function (c) {
      return { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c];
    });
  });

  App.toast = function (msg, kind) {
    var el = document.createElement("div");
    el.className = "toast" + (kind === "error" ? " error" : "");
    el.textContent = msg;
    $("toasts").appendChild(el);
    setTimeout(function () { el.remove(); }, kind === "error" ? 5000 : 2600);
  };

  App.relTime = function (iso) {
    var t = Date.parse(iso);
    if (isNaN(t)) return "";
    var s = Math.max(0, (Date.now() - t) / 1000);
    if (s < 45) return "just now";
    if (s < 3600) return Math.round(s / 60) + "m ago";
    if (s < 86400) return Math.round(s / 3600) + "h ago";
    return Math.round(s / 86400) + "d ago";
  };

  App.storage = {
    get: function (k) { try { return localStorage.getItem(k); } catch (e) { return null; } },
    set: function (k, v) { try { localStorage.setItem(k, v); } catch (e) { /* private mode */ } }
  };

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
          throw new Error((data && data.error && data.error.message) || ("Request failed (HTTP " + r.status + ")"));
        }
        return data;
      });
    });
  };

  App.taskUrl = function (t, suffix) {
    return "/api/tasks/" + encodeURIComponent(t.project) + "/" + encodeURIComponent(t.id) + (suffix || "");
  };

  // ---------- Claims ----------
  // claimInfo describes who is working on a task: an agent with a live claim,
  // a stale claim, a board-started Claude run, or nobody despite in_progress.
  App.claimInfo = function (t) {
    var a = t.assignment;
    if (a && t.status === "completed") {
      return { kind: "done", who: "Done by " + a.workflow_id, when: "", title: a.workflow_id + " held this task when it was completed" };
    }
    if (!a) {
      if (t.status === "in_progress") {
        return { kind: "orphan", who: "No agent", when: "", title: "In progress but not claimed: 'agmd task next --claim' won't hand it out. Move it back to Ready to queue it again." };
      }
      return null;
    }
    var age = (Date.now() - Date.parse(a.updated_at)) / 1000;
    var stale = state.ttl > 0 && age >= state.ttl;
    var isRun = a.workflow_id.indexOf("web-") === 0;
    var who = isRun ? "Claude run " + a.workflow_id.slice(4, 12) : a.workflow_id;
    var kind = stale ? "stale" : (isRun ? "run" : "active");
    var title = stale
      ? who + " has not checked in for " + App.relTime(a.updated_at).replace(" ago", "") + ". Another agent may take this task over."
      : who + " is working on this (last heartbeat " + App.relTime(a.updated_at) + ")";
    return { kind: kind, who: who, when: stale ? "stale" : App.relTime(a.updated_at), title: title, runId: isRun ? a.workflow_id.slice(4) : "" };
  };

  App.claimHTML = function (t) {
    var c = App.claimInfo(t);
    if (!c) return "";
    return '<div class="claim ' + c.kind + '" title="' + esc(c.title) + '"' + (c.runId ? ' data-run="' + esc(c.runId) + '"' : "") + ">" +
      '<span class="dot"></span><span class="who">' + esc(c.who) + "</span>" +
      (c.when ? '<span class="when">' + esc(c.when) + "</span>" : "") + "</div>";
  };

  App.tagsHTML = function (t, withProject) {
    var tags = [];
    if (t.priority !== 2) tags.push('<span class="tag p' + t.priority + '">P' + t.priority + "</span>");
    if (t.type && t.type !== "task") tags.push('<span class="tag ' + esc(t.type) + '">' + esc(t.type) + "</span>");
    if (t.feature) tags.push('<span class="tag"># ' + esc(t.feature) + "</span>");
    if (t.computed_status === "blocked") {
      var n = (t.pending_dependencies || []).length;
      tags.push('<span class="tag blocked" title="Waiting on ' + esc((t.pending_dependencies || []).join(", ")) + '">⏳ ' + n + " dep" + (n === 1 ? "" : "s") + "</span>");
    }
    if (t.comments && t.comments.length) {
      tags.push('<span class="tag" title="' + t.comments.length + ' comment(s)"><svg viewBox="0 0 16 16"><path d="M3 3h10a1 1 0 0 1 1 1v6a1 1 0 0 1-1 1H7l-3 2.5V11H3a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1z" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linejoin="round"/></svg>' + t.comments.length + "</span>");
    }
    if (withProject) tags.push('<span class="tag">' + esc(projectLabel(t.project)) + "</span>");
    return tags.join("");
  };

  function projectLabel(name) {
    var p = state.projects.find(function (x) { return x.name === name; });
    return (p && p.emoji ? p.emoji + " " : "") + name;
  }

  // ---------- Data ----------
  App.loadState = function () {
    return App.api("GET", "/api/state").then(function (data) {
      state.tasks = data.tasks || [];
      state.projects = data.projects || [];
      state.currentProject = data.currentProject || "";
      state.ttl = data.claimTtlSeconds || 0;
      state.byId = {};
      state.tasks.forEach(function (t) { state.byId[t.id] = t; });
      Object.keys(state.selected).forEach(function (id) { if (!state.byId[id]) delete state.selected[id]; });
      if (state.project === null) {
        var known = state.projects.some(function (p) { return p.name === state.currentProject; });
        state.project = known ? state.currentProject : "";
      }
      render();
      if (App.Panel) App.Panel.refresh();
    }).catch(function (err) { App.toast(err.message, "error"); });
  };

  // mutate runs a request, then reloads the board so computed statuses
  // (blocked/ready) and counts reflect the change everywhere.
  App.mutate = function (method, url, body) {
    return App.api(method, url, body).then(function (res) {
      return App.loadState().then(function () { return res; });
    });
  };

  // moveTask changes a task's status. Moving a claimed task back to the queue
  // also releases its claim, so it can be handed out again.
  App.moveTask = function (t, status) {
    if (t.status === status) return Promise.resolve();
    var req = status === "pending" && t.assignment
      ? App.mutate("DELETE", App.taskUrl(t, "/claim"))
      : App.mutate("PATCH", App.taskUrl(t), { status: status });
    return req.then(function () {
      if (status === "pending" && t.assignment) App.toast("Released " + App.claimInfo(t).who + "'s claim");
    }).catch(function (err) { App.toast(err.message, "error"); });
  };

  App.selectedTasks = function () {
    return Object.keys(state.selected).map(function (id) { return state.byId[id]; }).filter(Boolean);
  };
  App.clearSelection = function () { state.selected = {}; render(); };

  // ---------- Filtering ----------
  function visibleTasks() {
    var f = state.filters;
    var q = f.q.trim().toLowerCase();
    return state.tasks.filter(function (t) {
      if (state.project && t.project !== state.project) return false;
      if (f.feature && t.feature !== f.feature) return false;
      if (f.claim) {
        var c = App.claimInfo(t);
        var kind = c ? c.kind : "";
        if (f.claim === "claimed" && !(kind === "active" || kind === "run" || kind === "stale")) return false;
        if (f.claim === "stale" && kind !== "stale") return false;
        if (f.claim === "orphan" && kind !== "orphan") return false;
        if (f.claim === "unclaimed" && t.assignment) return false;
      }
      if (!q) return true;
      var hay = [t.subject, t.name, t.content, t.feature, t.project, t.assignment && t.assignment.workflow_id]
        .concat((t.comments || []).map(function (c) { return c.body + " " + c.author; }))
        .join("\n").toLowerCase();
      return hay.indexOf(q) >= 0;
    });
  }

  // ---------- Rendering ----------
  var render = (App.render = function () {
    // Rebuilding the board mid-drag would drop the dragged card.
    if (state.dragId) { state.renderAfterDrag = true; return; }
    renderSidebar();
    renderHeader();
    renderBoard();
  });

  function openCount(project) {
    return state.tasks.filter(function (t) {
      return t.status !== "completed" && (!project || t.project === project);
    }).length;
  }

  function hasLiveClaim(project) {
    return state.tasks.some(function (t) {
      if (t.project !== project) return false;
      var c = App.claimInfo(t);
      return c && (c.kind === "active" || c.kind === "run");
    });
  }

  function renderSidebar() {
    $("count-all").textContent = openCount("");
    $("nav-all").classList.toggle("active", state.project === "");
    $("project-list").innerHTML = state.projects.map(function (p) {
      var icon = p.emoji ? esc(p.emoji) : "▢";
      return '<button class="nav-item' + (state.project === p.name ? " active" : "") + '" data-project="' + esc(p.name) + '" title="' + esc(p.name) + '">' +
        '<span class="nav-icon">' + icon + '</span><span class="nav-name">' + esc(p.name) + "</span>" +
        (hasLiveClaim(p.name) ? '<span class="live" title="An agent is working here"></span>' : "") +
        '<span class="nav-count">' + openCount(p.name) + "</span></button>";
    }).join("");
  }

  function renderHeader() {
    $("board-title").textContent = state.project ? projectLabel(state.project) : "All projects";
    var tasks = visibleTasks();
    var open = tasks.filter(function (t) { return t.status !== "completed"; }).length;
    var working = tasks.filter(function (t) {
      var c = App.claimInfo(t);
      return c && (c.kind === "active" || c.kind === "run");
    }).length;
    $("board-sub").textContent = open + " open" + (working ? " · " + working + " being worked on" : "");

    var features = {};
    state.tasks.forEach(function (t) {
      if (t.feature && (!state.project || t.project === state.project)) features[t.feature] = true;
    });
    var names = Object.keys(features).sort();
    var sel = $("f-feature");
    if (names.indexOf(state.filters.feature) < 0) state.filters.feature = "";
    sel.innerHTML = '<option value="">All features</option>' + names.map(function (n) {
      return '<option value="' + esc(n) + '">' + esc(n) + "</option>";
    }).join("");
    sel.value = state.filters.feature;
    sel.classList.toggle("hidden", names.length === 0);

    var n = App.selectedTasks().length;
    $("btn-run").classList.toggle("hidden", n === 0);
    $("run-count").textContent = n ? "(" + n + ")" : "";
  }

  function renderBoard() {
    var tasks = visibleTasks();
    var openId = App.Panel && App.Panel.openId();
    var showProject = !state.project;
    var html = COLUMNS.map(function (col) {
      var cards = tasks.filter(function (t) { return t.computed_status === col.key; })
        .sort(function (a, b) { return a.priority - b.priority || a.name.localeCompare(b.name); });
      var collapsed = !!state.collapsed[col.key];
      var body = cards.length
        ? cards.map(function (t) { return cardHTML(t, t.id === openId, showProject); }).join("")
        : '<div class="col-empty">' + (col.key === "ready" ? "Nothing ready. Press n to add a task." : "No tasks") + "</div>";
      return '<section class="column' + (collapsed ? " collapsed" : "") + '" data-col="' + col.key + '">' +
        '<div class="col-head">' + ICONS[col.key] + "<span>" + col.label + '</span><span class="count">' + cards.length + "</span>" +
        (col.key === "completed" ? '<button class="icon-btn collapse-btn" title="Collapse">»</button>' : "") +
        "</div>" +
        '<div class="col-body">' + body + "</div></section>";
    }).join("");
    $("board").innerHTML = html;
  }

  function cardHTML(t, open, showProject) {
    var done = t.status === "completed";
    var desc = (t.content || "").split("\n").find(function (l) { return l.trim(); }) || "";
    var cls = "card" + (open ? " open" : "") + (state.selected[t.id] ? " selected" : "") + (done ? " done" : "");
    return '<article class="' + cls + '" data-id="' + esc(t.id) + '" draggable="' + (done ? "false" : "true") + '"' +
      (done ? ' title="Completed tasks stay done"' : "") + ">" +
      '<span class="select-box" title="Select for a Claude run">' + (state.selected[t.id] ? "✓" : "") + "</span>" +
      '<div class="card-top"><span class="card-title">' + esc(t.subject || t.name) + '</span><span class="card-ref">' + esc(t.name) + "</span></div>" +
      (desc ? '<div class="card-desc">' + esc(desc.replace(/^#+\s*/, "")) + "</div>" : "") +
      '<div class="card-meta">' + App.tagsHTML(t, showProject) + "</div>" +
      App.claimHTML(t) +
      "</article>";
  }

  // ---------- Events ----------
  function onBoardClick(e) {
    var col = e.target.closest(".column");
    if (col && (e.target.closest(".collapse-btn") || col.classList.contains("collapsed"))) {
      var key = col.dataset.col;
      state.collapsed[key] = !state.collapsed[key];
      App.storage.set("agmd-collapsed", JSON.stringify(state.collapsed));
      renderBoard();
      return;
    }
    var card = e.target.closest(".card");
    if (!card) return;
    var id = card.dataset.id;
    var run = e.target.closest(".claim.run");
    if (run && App.Runs && App.Runs.has(run.dataset.run)) {
      App.Runs.openConsole(run.dataset.run);
      return;
    }
    if (e.target.closest(".select-box") || e.shiftKey || e.metaKey || e.ctrlKey) {
      if (state.selected[id]) delete state.selected[id]; else state.selected[id] = true;
      render();
      return;
    }
    App.Panel.open(id);
  }

  function targetStatus(colKey) {
    var col = COLUMNS.find(function (c) { return c.key === colKey; });
    return col ? col.status : null;
  }

  function onDragStart(e) {
    var card = e.target.closest(".card");
    if (!card) return;
    state.dragId = card.dataset.id;
    card.classList.add("dragging");
    e.dataTransfer.effectAllowed = "move";
    e.dataTransfer.setData("text/plain", state.dragId);
  }

  function onDragOver(e) {
    var col = e.target.closest(".column");
    var t = state.byId[state.dragId];
    if (!col || !t) return;
    var status = targetStatus(col.dataset.col);
    if (status === t.status) return;
    e.preventDefault();
    document.querySelectorAll(".column.drop-ok").forEach(function (c) { if (c !== col) c.classList.remove("drop-ok"); });
    col.classList.add("drop-ok");
  }

  function onDrop(e) {
    var col = e.target.closest(".column");
    var t = state.byId[state.dragId];
    endDrag();
    if (!col || !t) return;
    e.preventDefault();
    if (col.dataset.col === "completed" && state.collapsed.completed) {
      state.collapsed.completed = false;
    }
    App.moveTask(t, targetStatus(col.dataset.col));
  }

  function endDrag() {
    state.dragId = null;
    document.querySelectorAll(".drop-ok, .dragging").forEach(function (el) { el.classList.remove("drop-ok", "dragging"); });
    if (state.renderAfterDrag) {
      state.renderAfterDrag = false;
      render();
    }
  }

  function isTyping(e) {
    var tag = (e.target.tagName || "").toLowerCase();
    return tag === "input" || tag === "textarea" || tag === "select" || e.target.isContentEditable;
  }

  function onKey(e) {
    if (e.key === "Escape") {
      if (document.querySelector(".modal-scrim.open")) return;
      if (isTyping(e)) { e.target.blur(); return; }
      if (App.Panel.isOpen()) { App.Panel.close(); return; }
      if (state.filters.q) { $("f-search").value = ""; state.filters.q = ""; render(); return; }
      if (Object.keys(state.selected).length) App.clearSelection();
      return;
    }
    if (isTyping(e) || e.metaKey || e.ctrlKey || e.altKey) return;
    if (e.key === "/" || e.key === "f") { e.preventDefault(); $("f-search").focus(); }
    else if (e.key === "n") { e.preventDefault(); App.Panel.openCreate(); }
  }

  // ---------- Theme ----------
  var THEMES = ["system", "light", "dark"];
  function applyTheme(name) {
    if (name === "system") delete document.documentElement.dataset.theme;
    else document.documentElement.dataset.theme = name;
    $("theme-label").textContent = name === "system" ? "System theme" : name === "light" ? "Light theme" : "Dark theme";
    $("theme-icon").textContent = name === "system" ? "◐" : name === "light" ? "☀" : "☾";
  }

  // ---------- Live updates ----------
  function poll() {
    if (document.hidden) return;
    App.api("GET", "/api/changes").then(function (data) {
      if (state.version !== null && data.version !== state.version) App.loadState();
      state.version = data.version;
    }).catch(function () { /* server restarting */ });
  }

  // ---------- Boot ----------
  document.addEventListener("DOMContentLoaded", function () {
    try { Object.assign(state.collapsed, JSON.parse(App.storage.get("agmd-collapsed") || "{}")); } catch (e) { /* ignore */ }
    var theme = App.storage.get("agmd-theme");
    applyTheme(THEMES.indexOf(theme) >= 0 ? theme : "system");

    var board = $("board");
    board.addEventListener("click", onBoardClick);
    board.addEventListener("dragstart", onDragStart);
    board.addEventListener("dragover", onDragOver);
    board.addEventListener("drop", onDrop);
    board.addEventListener("dragend", endDrag);
    board.addEventListener("dragleave", function (e) {
      var col = e.target.closest(".column");
      if (col && !col.contains(e.relatedTarget)) col.classList.remove("drop-ok");
    });

    document.querySelector(".nav").addEventListener("click", function (e) {
      var item = e.target.closest(".nav-item");
      if (!item) return;
      state.project = item.dataset.project;
      render();
    });
    $("f-search").addEventListener("input", function (e) { state.filters.q = e.target.value; render(); });
    $("f-claim").addEventListener("change", function (e) { state.filters.claim = e.target.value; render(); });
    $("f-feature").addEventListener("change", function (e) { state.filters.feature = e.target.value; render(); });
    $("btn-new").addEventListener("click", function () { App.Panel.openCreate(); });
    $("btn-run").addEventListener("click", function () { App.Runs.openModal(App.selectedTasks()); });
    $("btn-theme").addEventListener("click", function () {
      var cur = document.documentElement.dataset.theme || "system";
      var next = THEMES[(THEMES.indexOf(cur) + 1) % THEMES.length];
      applyTheme(next);
      App.storage.set("agmd-theme", next);
    });
    document.addEventListener("keydown", onKey);

    App.loadState().then(poll);
    setInterval(poll, 1500);
    // Claim ages ("3m ago", stale) move with time even when nothing changes.
    setInterval(function () { render(); if (App.Panel) App.Panel.refresh(); }, 30000);
  });
})();
