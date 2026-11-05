// agmd board: state, sidebar, header counts, columns, cards, keyboard
// navigation, drag and drop, filters and live updates. ui.js defines App and
// its primitives; panel.js (task panel) and runs.js (Claude runs) build on it.
(function () {
  "use strict";

  var App = window.App;
  var $ = App.$;
  var esc = App.esc;
  var icon = App.icon;

  var state = (App.state = {
    tasks: [],
    byId: {},
    projects: [],
    currentProject: "",
    ttl: 0, // seconds a claim survives without a heartbeat; 0 = never expires
    project: null, // selected project; "" = all projects
    filters: { q: "", show: "", feature: "" },
    selected: {},
    focusId: null, // the card that holds the roving tab stop
    version: null,
    failures: 0,
    offline: false,
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

  var COL_ICONS = {
    blocked: '<svg class="col-icon" viewBox="0 0 16 16" aria-hidden="true"><circle cx="8" cy="8" r="6.2" fill="none" stroke="var(--fg-dim)" stroke-width="1.6"/><path d="M4 12 12 4" stroke="var(--fg-dim)" stroke-width="1.6"/></svg>',
    ready: '<svg class="col-icon" viewBox="0 0 16 16" aria-hidden="true"><circle cx="8" cy="8" r="6.2" fill="none" stroke="var(--fg-dim)" stroke-width="1.6"/></svg>',
    in_progress: '<svg class="col-icon" viewBox="0 0 16 16" aria-hidden="true"><circle cx="8" cy="8" r="6.2" fill="none" stroke="var(--accent)" stroke-width="1.6"/><path d="M8 3.6a4.4 4.4 0 0 1 0 8.8z" fill="var(--accent)"/></svg>',
    completed: '<svg class="col-icon" viewBox="0 0 16 16" aria-hidden="true"><circle cx="8" cy="8" r="7" fill="var(--green)"/><path d="m5 8.2 2 2 4-4.2" fill="none" stroke="var(--surface)" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/></svg>'
  };

  // ---------- Claims and attention ----------
  // claimInfo describes who holds a task: a live agent, a stale claim, a
  // board-started Claude run, nobody despite in_progress, or the agent that
  // held it when it was completed.
  App.claimInfo = function (t) {
    var a = t.assignment;
    if (a && t.status === "completed") {
      return { kind: "done", who: "Done by " + a.workflow_id, when: "", title: a.workflow_id + " held this task when it was completed." };
    }
    if (!a) {
      if (t.status !== "in_progress") return null;
      return { kind: "orphan", who: "No agent", when: "", title: "In progress, but no agent holds it, so 'agmd task next --claim' will not hand it out. Re-queue it to make it claimable." };
    }
    var age = (Date.now() - Date.parse(a.updated_at)) / 1000;
    var stale = state.ttl > 0 && age >= state.ttl;
    var isRun = a.workflow_id.indexOf("web-") === 0;
    var who = isRun ? "Claude run " + a.workflow_id.slice(4, 12) : a.workflow_id;
    return {
      kind: stale ? "stale" : isRun ? "run" : "active",
      who: who,
      when: stale ? "stale · " + App.age(a.updated_at) : App.relTime(a.updated_at),
      title: stale
        ? who + " has not checked in for " + App.age(a.updated_at) + ". Another agent may take this task over."
        : who + " is working on this. Last heartbeat " + App.relTime(a.updated_at) + ".",
      runId: isRun ? a.workflow_id.slice(4) : ""
    };
  };

  App.claimHTML = function (t) {
    var c = App.claimInfo(t);
    if (!c) return "";
    var inner = '<span class="dot" aria-hidden="true"></span><span class="who">' + esc(c.who) + "</span>" +
      (c.when ? '<span class="when">' + esc(c.when) + "</span>" : "");
    if (c.runId && App.Runs && App.Runs.has(c.runId)) {
      return '<button type="button" class="claim run" data-run="' + esc(c.runId) + '" title="Open the run console" tabindex="-1">' + inner + "</button>";
    }
    return '<div class="claim ' + c.kind + '" title="' + esc(c.title) + '">' + inner + "</div>";
  };

  function taskByName(project, name) {
    for (var i = 0; i < state.tasks.length; i++) {
      var t = state.tasks[i];
      if (t.project === project && t.name === name) return t;
    }
    return null;
  }

  function isStuck(t) {
    var c = App.claimInfo(t);
    return !!c && (c.kind === "stale" || c.kind === "orphan");
  }

  // stuckBlockers lists the dependencies of a blocked task that are themselves
  // stuck (stale claim or in progress with no agent).
  App.stuckBlockers = function (t) {
    return (t.pending_dependencies || []).filter(function (name) {
      var dep = taskByName(t.project, name);
      return dep && isStuck(dep);
    });
  };

  // attention is why a task needs the human, or "" when it doesn't.
  App.attention = function (t) {
    if (isStuck(t)) return App.claimInfo(t).kind;
    if (t.computed_status === "blocked" && App.stuckBlockers(t).length) return "blocked-stuck";
    return "";
  };

  function isWorking(t) {
    var c = App.claimInfo(t);
    return !!c && (c.kind === "active" || c.kind === "run");
  }

  App.tagsHTML = function (t, withProject) {
    var tags = [];
    if (t.priority !== 2) tags.push('<span class="tag prio p' + t.priority + '" title="Priority ' + t.priority + '">P' + t.priority + "</span>");
    if (t.type === "bug") tags.push('<span class="tag">' + icon("bug") + "bug</span>");
    else if (t.type && t.type !== "task") tags.push('<span class="tag">' + esc(t.type) + "</span>");
    if (t.feature) tags.push('<span class="tag feature" title="Feature">' + icon("hash") + esc(t.feature) + "</span>");
    if (t.comments && t.comments.length) {
      var n = t.comments.length;
      tags.push('<span class="tag plain" title="' + n + " comment" + (n === 1 ? "" : "s") + '">' + icon("comment") + n + "</span>");
    }
    if (withProject) tags.push('<span class="tag plain">' + projectIcon(t.project) + esc(t.project) + "</span>");
    return tags.join("");
  };

  function projectIcon(name) {
    var p = state.projects.find(function (x) { return x.name === name; });
    return p && p.emoji ? '<span class="emoji" aria-hidden="true">' + esc(p.emoji) + "</span>" : icon("folder");
  }
  App.projectIcon = projectIcon;

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
    }).catch(function (err) {
      if (!state.offline) App.toast(err.message, "error");
    });
  };

  // mutate runs a request, then reloads the board so computed statuses
  // (blocked/ready) and counts reflect the change everywhere.
  App.mutate = function (method, url, body) {
    return App.api(method, url, body).then(function (res) {
      return App.loadState().then(function () { return res; });
    });
  };

  function label(t) { return "“" + (t.subject || t.name) + "”"; }

  // releaseClaim gives a task back to the queue. A live agent may still be
  // working on it, so that case asks first; a stale claim does not.
  App.releaseClaim = function (t) {
    var c = App.claimInfo(t);
    var live = c && (c.kind === "active" || c.kind === "run");
    var ask = live
      ? App.confirm({
          title: "Release " + c.who + "’s claim?",
          body: c.who + " checked in " + App.relTime(t.assignment.updated_at) + " and may still be working on " + label(t) + ". It goes back to Ready, where any agent can claim it.",
          confirmLabel: "Release claim",
          danger: true
        })
      : Promise.resolve(true);
    return ask.then(function (ok) {
      if (!ok) return false;
      return App.mutate("DELETE", App.taskUrl(t, "/claim")).then(function () {
        App.toast(label(t) + " is back in Ready.");
        return true;
      });
    }).catch(function (err) { App.toast(err.message, "error"); return false; });
  };

  // moveTask changes a task's status, guarding the moves that cannot be
  // undone (completion) or that take work from an agent (release).
  App.moveTask = function (t, status) {
    if (!t || t.status === status) return Promise.resolve(false);
    if (status === "pending" && t.assignment) return App.releaseClaim(t);
    var ask = status === "completed"
      ? App.confirm({
          title: "Mark " + label(t) + " as done?",
          body: "Done is final: the task can't be reopened." + (t.assignment ? " " + t.assignment.workflow_id + " still holds it." : "") + " Tasks waiting on it become ready.",
          confirmLabel: "Mark as done"
        })
      : Promise.resolve(true);
    return ask.then(function (ok) {
      if (!ok) return false;
      return App.mutate("PATCH", App.taskUrl(t), { status: status }).then(function () {
        if (status === "completed") App.toast(label(t) + " is done.");
        else if (status === "pending") App.toast(label(t) + " is back in the queue.");
        return true;
      });
    }).catch(function (err) { App.toast(err.message, "error"); return false; });
  };

  App.selectedTasks = function () {
    return Object.keys(state.selected).map(function (id) { return state.byId[id]; }).filter(Boolean);
  };
  App.clearSelection = function () { state.selected = {}; render(); };

  // ---------- Filtering ----------
  function filtersActive() {
    var f = state.filters;
    return !!(f.q.trim() || f.show || f.feature);
  }

  function visibleTasks() {
    var f = state.filters;
    var q = f.q.trim().toLowerCase();
    return state.tasks.filter(function (t) {
      if (state.project && t.project !== state.project) return false;
      if (f.feature && t.feature !== f.feature) return false;
      if (f.show === "working" && !isWorking(t)) return false;
      if (f.show === "attention" && !App.attention(t)) return false;
      if (f.show === "blocked" && t.computed_status !== "blocked") return false;
      if (f.show === "unclaimed" && (t.assignment || t.status === "completed")) return false;
      if (!q) return true;
      var hay = [t.subject, t.name, t.content, t.feature, t.project, t.assignment && t.assignment.workflow_id]
        .concat((t.comments || []).map(function (c) { return c.body + " " + c.author; }))
        .join("\n").toLowerCase();
      return hay.indexOf(q) >= 0;
    });
  }

  function inScope(t) { return !state.project || t.project === state.project; }

  // ---------- URL state ----------
  function writeHash() {
    var p = new URLSearchParams();
    if (state.project) p.set("project", state.project);
    if (state.filters.show) p.set("show", state.filters.show);
    if (state.filters.feature) p.set("feature", state.filters.feature);
    if (state.filters.q) p.set("q", state.filters.q);
    var hash = p.toString();
    history.replaceState(null, "", hash ? "#" + hash : location.pathname);
  }

  function readHash() {
    var p = new URLSearchParams(location.hash.slice(1));
    if (p.has("project")) state.project = p.get("project");
    state.filters.show = p.get("show") || "";
    state.filters.feature = p.get("feature") || "";
    state.filters.q = p.get("q") || "";
  }

  // ---------- Rendering ----------
  var render = (App.render = function () {
    // Rebuilding the board mid-drag would drop the dragged card.
    if (state.dragId) { state.renderAfterDrag = true; return; }
    renderSidebar();
    renderHeader();
    renderBoard();
    writeHash();
  });

  function projectStats(project) {
    var open = 0, working = 0, attention = 0;
    state.tasks.forEach(function (t) {
      if (project && t.project !== project) return;
      if (t.status !== "completed") open++;
      if (isWorking(t)) working++;
      if (App.attention(t)) attention++;
    });
    return { open: open, working: working, attention: attention };
  }

  function navBadges(s) {
    return (s.working ? '<span class="live" title="' + s.working + ' being worked on"><span class="sr-only">' + s.working + " being worked on</span></span>" : "") +
      (s.attention ? '<span class="attn-badge" title="' + s.attention + ' need attention">' + s.attention + '<span class="sr-only"> need attention</span></span>' : "") +
      '<span class="nav-count" title="Open tasks">' + s.open + "</span>";
  }

  function renderSidebar() {
    var all = projectStats("");
    $("nav-all").innerHTML = '<span class="nav-icon">' + icon("grid") + '</span><span class="nav-name">All projects</span>' + navBadges(all);
    $("nav-all").setAttribute("aria-current", state.project === "" ? "page" : "false");
    $("project-list").innerHTML = state.projects.map(function (p) {
      return '<button type="button" class="nav-item" data-project="' + esc(p.name) + '" aria-current="' + (state.project === p.name ? "page" : "false") + '" title="' + esc(p.name) + '">' +
        '<span class="nav-icon">' + projectIcon(p.name) + '</span><span class="nav-name">' + esc(p.name) + "</span>" + navBadges(projectStats(p.name)) + "</button>";
    }).join("");

    var m = $("m-project");
    m.innerHTML = '<option value="">All projects</option>' + state.projects.map(function (p) {
      return '<option value="' + esc(p.name) + '">' + esc((p.emoji ? p.emoji + " " : "") + p.name) + "</option>";
    }).join("");
    m.value = state.project || "";
  }

  function renderHeader() {
    var title = $("board-title");
    title.innerHTML = state.project ? projectIcon(state.project) + "<span>" + esc(state.project) + "</span>" : "<span>All projects</span>";

    var scoped = state.tasks.filter(inScope);
    var open = scoped.filter(function (t) { return t.status !== "completed"; }).length;
    var working = scoped.filter(isWorking).length;
    var stale = scoped.filter(function (t) { return App.attention(t) === "stale"; }).length;
    var orphan = scoped.filter(function (t) { return App.attention(t) === "orphan"; }).length;
    var waiting = scoped.filter(function (t) { return App.attention(t) === "blocked-stuck"; }).length;
    var attention = stale + orphan + waiting;
    var blocked = scoped.filter(function (t) { return t.computed_status === "blocked"; }).length;
    var why = [stale && stale + " stale claim" + (stale === 1 ? "" : "s"), orphan && orphan + " in progress with no agent", waiting && waiting + " waiting on stuck work"].filter(Boolean).join(", ");

    function stat(show, n, text, cls, tip) {
      var on = state.filters.show === show;
      return '<button type="button" class="stat' + (cls ? " " + cls : "") + (on ? " on" : "") + '" data-show="' + show + '" aria-pressed="' + on + '"' + (tip ? ' title="' + esc(tip) + '"' : "") + ">" + n + " " + text + "</button>";
    }
    $("status-line").innerHTML =
      '<span class="stat static">' + open + " open</span>" +
      stat("working", working, "working", working ? "working" : "") +
      stat("attention", attention, attention === 1 ? "needs attention" : "need attention", attention ? "attention" : "", why) +
      stat("blocked", blocked, "blocked", "");

    var features = {};
    scoped.forEach(function (t) { if (t.feature) features[t.feature] = true; });
    var names = Object.keys(features).sort();
    var sel = $("f-feature");
    if (names.indexOf(state.filters.feature) < 0) state.filters.feature = "";
    sel.innerHTML = '<option value="">All features</option>' + names.map(function (n) {
      return '<option value="' + esc(n) + '">' + esc(n) + "</option>";
    }).join("");
    sel.value = state.filters.feature;
    sel.classList.toggle("hidden", names.length === 0);
    $("f-show").value = state.filters.show;
    if ($("f-search").value !== state.filters.q) $("f-search").value = state.filters.q;
    $("btn-clear").classList.toggle("hidden", !filtersActive());

    var n = App.selectedTasks().length;
    $("btn-run").classList.toggle("hidden", n === 0);
    $("run-count").textContent = n ? n : "";
    $("conn").classList.toggle("hidden", !state.offline);
  }

  // In progress lists what needs the human first: stale claims, then work
  // nobody holds, then live agents. Other columns go by priority.
  var CLAIM_RANK = { stale: 0, orphan: 1, run: 2, active: 2 };
  function sortCards(colKey, cards) {
    return cards.sort(function (a, b) {
      if (colKey === "in_progress") {
        var ra = CLAIM_RANK[(App.claimInfo(a) || {}).kind], rb = CLAIM_RANK[(App.claimInfo(b) || {}).kind];
        if (ra !== rb) return ra - rb;
      }
      if (colKey === "blocked") {
        var sa = App.attention(a) ? 0 : 1, sb = App.attention(b) ? 0 : 1;
        if (sa !== sb) return sa - sb;
      }
      return a.priority - b.priority || a.name.localeCompare(b.name);
    });
  }

  function renderBoard() {
    var hadFocus = document.activeElement && document.activeElement.closest && document.activeElement.closest("#board .card");
    var focusedId = hadFocus ? hadFocus.dataset.id : null;
    var tasks = visibleTasks();
    var openId = App.Panel && App.Panel.openId();
    var showProject = !state.project;
    var filtered = filtersActive();
    $("board").innerHTML = COLUMNS.map(function (col) {
      var cards = sortCards(col.key, tasks.filter(function (t) { return t.computed_status === col.key; }));
      var collapsed = !!state.collapsed[col.key];
      var empty = filtered ? "No matching tasks" : col.key === "ready" ? "Nothing ready. Press <kbd>n</kbd> to add a task." : "No tasks";
      var body = cards.length
        ? cards.map(function (t) { return cardHTML(t, t.id === openId, showProject); }).join("")
        : '<p class="col-empty">' + empty + "</p>";
      var heading = COL_ICONS[col.key] + '<span class="col-label">' + col.label + '</span><span class="count">' + cards.length + "</span>";
      var head = col.key === "completed"
        ? '<button type="button" class="col-head col-toggle" aria-expanded="' + !collapsed + '" title="' + (collapsed ? "Show" : "Hide") + ' done tasks">' + heading + '<span class="toggle-icon">' + icon(collapsed ? "chevronLeft" : "chevronRight") + "</span></button>"
        : '<h2 class="col-head">' + heading + "</h2>";
      return '<section class="column' + (collapsed ? " collapsed" : "") + '" data-col="' + col.key + '" aria-label="' + col.label + ", " + cards.length + ' tasks">' +
        head + '<div class="col-body">' + body + "</div></section>";
    }).join("");
    updateRovingFocus(focusedId);
  }

  function cardHTML(t, open, showProject) {
    var done = t.status === "completed";
    var desc = (t.content || "").split("\n").find(function (l) { return l.trim(); }) || "";
    var attn = App.attention(t);
    var cls = "card" + (open ? " open" : "") + (state.selected[t.id] ? " selected" : "") + (done ? " done" : "") + (attn ? " attn" : "");
    var waits = "";
    if (t.computed_status === "blocked") {
      var deps = t.pending_dependencies || [];
      var stuck = App.stuckBlockers(t);
      var shown = deps.slice(0, 2).join(", ") + (deps.length > 2 ? " +" + (deps.length - 2) : "");
      waits = '<p class="waits' + (stuck.length ? " stuck" : "") + '">' + icon(stuck.length ? "alert" : "wait") +
        "<span>Waits on <b>" + esc(shown) + "</b>" + (stuck.length ? " · " + esc(stuck.length === deps.length && deps.length === 1 ? "stuck" : stuck.join(", ") + " stuck") : "") + "</span></p>";
    }
    return '<article class="' + cls + '" data-id="' + esc(t.id) + '" draggable="' + !done + '"' + (done ? ' title="Done is final"' : "") + ">" +
      '<input type="checkbox" class="select-box" tabindex="-1" aria-label="Select ' + esc(t.subject || t.name) + ' for a Claude run"' + (state.selected[t.id] ? " checked" : "") + (done ? " disabled" : "") + ">" +
      '<button type="button" class="card-title" tabindex="-1" aria-keyshortcuts="x">' + esc(t.subject || t.name) + "</button>" +
      (desc ? '<p class="card-desc">' + esc(desc.replace(/^#+\s*/, "")) + "</p>" : "") +
      waits +
      '<div class="card-meta">' + App.tagsHTML(t, showProject) + '<code class="card-ref" title="Name agents use">' + esc(t.name) + "</code></div>" +
      App.claimHTML(t) +
      "</article>";
  }

  // ---------- Keyboard navigation ----------
  // One card at a time holds the tab stop (roving tabindex); arrows or
  // h/j/k/l move it, Enter opens, x selects for a run.
  function cardButtons() { return Array.prototype.slice.call(document.querySelectorAll("#board .column:not(.collapsed) .card-title")); }

  function updateRovingFocus(refocusId) {
    var buttons = cardButtons();
    if (!buttons.length) return;
    var id = refocusId || state.focusId;
    var target = buttons.find(function (b) { return b.closest(".card").dataset.id === id; }) || buttons[0];
    buttons.forEach(function (b) { b.tabIndex = b === target ? 0 : -1; });
    state.focusId = target.closest(".card").dataset.id;
    if (refocusId) target.focus({ preventScroll: true });
  }

  function focusCard(btn) {
    if (!btn) return;
    cardButtons().forEach(function (b) { b.tabIndex = -1; });
    btn.tabIndex = 0;
    state.focusId = btn.closest(".card").dataset.id;
    btn.focus();
    btn.closest(".card").scrollIntoView({ block: "nearest", inline: "nearest" });
  }

  function moveFocus(dx, dy) {
    var current = document.activeElement && document.activeElement.classList.contains("card-title") ? document.activeElement : null;
    if (!current) {
      var stop = cardButtons().find(function (b) { return b.tabIndex === 0; });
      focusCard(stop || cardButtons()[0]);
      return;
    }
    var col = current.closest(".column");
    var inCol = Array.prototype.slice.call(col.querySelectorAll(".card-title"));
    var idx = inCol.indexOf(current);
    if (dy) { focusCard(inCol[Math.max(0, Math.min(inCol.length - 1, idx + dy))]); return; }
    var cols = Array.prototype.slice.call(document.querySelectorAll("#board .column:not(.collapsed)"));
    for (var i = cols.indexOf(col) + dx; i >= 0 && i < cols.length; i += dx) {
      var targets = cols[i].querySelectorAll(".card-title");
      if (targets.length) { focusCard(targets[Math.min(idx, targets.length - 1)]); return; }
    }
  }

  function toggleSelect(id) {
    var t = state.byId[id];
    if (!t || t.status === "completed") return;
    if (state.selected[id]) delete state.selected[id]; else state.selected[id] = true;
    render();
  }

  // ---------- Events ----------
  function onBoardClick(e) {
    var toggle = e.target.closest(".col-toggle");
    var col = e.target.closest(".column");
    if (toggle || (col && col.classList.contains("collapsed"))) {
      var key = col.dataset.col;
      state.collapsed[key] = !state.collapsed[key];
      App.storage.set("agmd-collapsed", JSON.stringify(state.collapsed));
      renderBoard();
      var again = document.querySelector('#board .column[data-col="' + key + '"] .col-toggle');
      if (again) again.focus();
      return;
    }
    var card = e.target.closest(".card");
    if (!card) return;
    var id = card.dataset.id;
    var run = e.target.closest(".claim.run");
    if (run) { App.Runs.openConsole(run.dataset.run); return; }
    if (e.target.closest(".select-box") || e.shiftKey || e.metaKey || e.ctrlKey) {
      if (e.target.closest(".select-box")) e.preventDefault();
      toggleSelect(id);
      return;
    }
    state.focusId = id;
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
    if (targetStatus(col.dataset.col) === t.status) return;
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

  function isTyping(el) {
    var tag = (el.tagName || "").toLowerCase();
    return tag === "input" || tag === "textarea" || tag === "select" || el.isContentEditable;
  }

  var KEY_MOVES = { ArrowDown: [0, 1], j: [0, 1], ArrowUp: [0, -1], k: [0, -1], ArrowLeft: [-1, 0], h: [-1, 0], ArrowRight: [1, 0], l: [1, 0] };

  function onKey(e) {
    if (document.querySelector("dialog[open]") || $("run-modal").classList.contains("open")) return;
    if (e.key === "Escape") {
      if (e.target === $("f-search")) {
        $("f-search").value = "";
        state.filters.q = "";
        e.target.blur();
        render();
        return;
      }
      if (App.Panel.isOpen()) { App.Panel.close(); return; }
      if (isTyping(e.target)) { e.target.blur(); return; }
      if (Object.keys(state.selected).length) App.clearSelection();
      return;
    }
    if (isTyping(e.target) || e.metaKey || e.ctrlKey || e.altKey) return;
    var onCard = e.target.classList && e.target.classList.contains("card-title");
    var move = KEY_MOVES[e.key];
    if (move && (onCard || e.key.length === 1)) {
      e.preventDefault();
      moveFocus(move[0], move[1]);
    } else if (e.key === "x" && onCard) {
      e.preventDefault();
      toggleSelect(e.target.closest(".card").dataset.id);
    } else if (e.key === "/" || e.key === "f") {
      e.preventDefault();
      $("f-search").focus();
    } else if (e.key === "n") {
      e.preventDefault();
      App.Panel.openCreate();
    }
  }

  // ---------- Theme ----------
  var THEMES = ["system", "light", "dark"];
  var THEME_META = { system: ["monitor", "System theme"], light: ["sun", "Light theme"], dark: ["moon", "Dark theme"] };
  function applyTheme(name) {
    if (name === "system") delete document.documentElement.dataset.theme;
    else document.documentElement.dataset.theme = name;
    $("theme-icon").innerHTML = icon(THEME_META[name][0]);
    $("theme-label").textContent = THEME_META[name][1];
  }

  // ---------- Live updates ----------
  // The board polls a cheap fingerprint of the task files. Two failures in a
  // row mean the agmd server is gone: say so instead of showing stale data as
  // if it were live.
  function poll() {
    if (document.hidden) return;
    App.api("GET", "/api/changes").then(function (data) {
      state.failures = 0;
      if (state.offline) {
        state.offline = false;
        App.toast("Reconnected to agmd.");
        App.loadState();
      } else if (state.version !== null && data.version !== state.version) {
        App.loadState();
      }
      state.version = data.version;
    }).catch(function () {
      state.failures++;
      if (state.failures >= 2 && !state.offline) {
        state.offline = true;
        renderHeader();
      }
    });
  }

  // ---------- Boot ----------
  document.addEventListener("DOMContentLoaded", function () {
    App.fillIcons();
    try { Object.assign(state.collapsed, JSON.parse(App.storage.get("agmd-collapsed") || "{}")); } catch (e) { /* ignore */ }
    var theme = App.storage.get("agmd-theme");
    applyTheme(THEMES.indexOf(theme) >= 0 ? theme : "system");
    readHash();
    $("keys").innerHTML = "<kbd>n</kbd> new · <kbd>/</kbd> filter · <kbd>j</kbd><kbd>k</kbd> move · <kbd>x</kbd> select · <kbd>esc</kbd> close";

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
    board.addEventListener("focusin", function (e) {
      if (e.target.classList.contains("card-title")) state.focusId = e.target.closest(".card").dataset.id;
    });

    document.querySelector(".nav").addEventListener("click", function (e) {
      var item = e.target.closest(".nav-item");
      if (!item) return;
      state.project = item.dataset.project;
      render();
    });
    $("m-project").addEventListener("change", function (e) { state.project = e.target.value; render(); });
    $("status-line").addEventListener("click", function (e) {
      var b = e.target.closest("button[data-show]");
      if (!b) return;
      state.filters.show = state.filters.show === b.dataset.show ? "" : b.dataset.show;
      render();
    });
    $("f-search").addEventListener("input", function (e) { state.filters.q = e.target.value; render(); });
    $("f-show").addEventListener("change", function (e) { state.filters.show = e.target.value; render(); });
    $("f-feature").addEventListener("change", function (e) { state.filters.feature = e.target.value; render(); });
    $("btn-clear").addEventListener("click", function () {
      state.filters = { q: "", show: "", feature: "" };
      render();
    });
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
