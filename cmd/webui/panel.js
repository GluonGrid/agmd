// agmd board: the side panel that shows, edits and creates a task. Each field
// saves on its own (PATCH with only that field), so an agent's concurrent
// change to another field is never overwritten.
(function () {
  "use strict";

  var App = window.App;
  var esc = App.esc;
  var icon = App.icon;
  var panel;
  var openId = null;
  var mode = null; // "view" | "create"
  var returnFocus = null;

  var PRIORITY_OPTIONS = [[0, "P0 · critical"], [1, "P1 · high"], [2, "P2 · normal"], [3, "P3 · low"], [4, "P4 · backlog"]];
  var TYPE_OPTIONS = [["task", "task"], ["bug", "bug"], ["feature", "feature"], ["chore", "chore"]];

  // The status menu speaks the board's language: a pending task is shown as
  // Ready or Blocked, matching its column.
  function statusOptions(t) {
    var pending = t.computed_status === "blocked" ? "Blocked (waiting on " + (t.pending_dependencies || []).join(", ") + ")" : "Ready";
    return [["pending", pending], ["in_progress", "In progress"], ["completed", "Done"]];
  }

  function options(list, value) {
    return list.map(function (o) {
      return '<option value="' + esc(o[0]) + '"' + (String(o[0]) === String(value) ? " selected" : "") + ">" + esc(o[1]) + "</option>";
    }).join("");
  }

  function q(sel) { return panel.querySelector(sel); }

  function show(focusEl) {
    returnFocus = returnFocus || document.activeElement;
    panel.inert = false;
    panel.classList.add("open");
    document.body.classList.add("panel-open");
    (focusEl || panel).focus({ preventScroll: true });
  }

  function close() {
    // Blurring first lets a field with unsaved text save itself.
    if (panel.contains(document.activeElement)) document.activeElement.blur();
    var id = openId;
    openId = null;
    mode = null;
    panel.classList.remove("open");
    panel.inert = true;
    document.body.classList.remove("panel-open");
    App.render();
    var card = id && document.querySelector('#board .card[data-id="' + CSS.escape(id) + '"] .card-title');
    var target = card || (returnFocus && document.contains(returnFocus) ? returnFocus : null);
    returnFocus = null;
    if (target) target.focus();
  }

  // ---------- View / edit ----------
  function open(id) {
    var t = App.state.byId[id];
    if (!t) return;
    if (mode === null) returnFocus = document.activeElement;
    openId = id;
    mode = "view";
    panel.setAttribute("aria-label", "Task " + (t.subject || t.name));
    panel.innerHTML =
      '<div class="panel-head"><span class="panel-ref" id="p-ref"></span><span class="spacer"></span>' +
      '<button type="button" class="icon-btn" id="p-run" title="Run with Claude Code" aria-label="Run with Claude Code">' + icon("play") + "</button>" +
      '<button type="button" class="icon-btn" id="p-close" title="Close (Esc)" aria-label="Close panel">' + icon("x") + "</button></div>" +
      '<div class="panel-body">' +
      '<label class="sr-only" for="p-subject">Title</label>' +
      '<input class="panel-title" id="p-subject" type="text" placeholder="Title">' +
      '<div class="props">' +
      '<label for="p-status">Status</label><select id="p-status"></select>' +
      '<span class="prop-label" id="p-claim-label">Agent</span><div class="claim-box" id="p-claim" aria-labelledby="p-claim-label"></div>' +
      '<label for="p-priority">Priority</label><select id="p-priority"></select>' +
      '<label for="p-type">Type</label><select id="p-type"></select>' +
      '<label for="p-feature">Feature</label><input id="p-feature" type="text" placeholder="None">' +
      '<label for="p-dep-add">Depends on</label><div class="deps"><span id="p-deps"></span><input id="p-dep-add" type="text" list="p-dep-list" placeholder="Add a task name" aria-describedby="p-deps-hint"><datalist id="p-dep-list"></datalist></div>' +
      "</div>" +
      '<p class="sr-only" id="p-deps-hint">Press Enter to add the dependency.</p>' +
      '<label class="section-label" for="p-content">Description</label>' +
      '<textarea class="desc" id="p-content" placeholder="What needs doing? Markdown is fine."></textarea>' +
      '<h3 class="section-label">Comments <span class="count" id="p-comment-count"></span></h3>' +
      '<div id="p-comments"></div>' +
      '<div class="compose"><label class="sr-only" for="p-comment">Add a comment</label><textarea id="p-comment" placeholder="Add a comment"></textarea>' +
      '<div class="compose-row"><span class="hint">' + App.modKey + "+Enter to send</span><button type=\"button\" class=\"btn\" id=\"p-send\">Comment</button></div></div>" +
      "</div>" +
      '<div class="panel-foot"><button type="button" class="btn danger" id="p-delete">' + icon("trash") + "Delete task</button></div>";
    bindView();
    fill(t, true);
    show();
    App.render();
    // Wait for the board to finish making room for the panel.
    setTimeout(function () {
      var card = document.querySelector('#board .card[data-id="' + CSS.escape(id) + '"]');
      if (card) card.scrollIntoView({ block: "nearest", inline: "nearest" });
    }, 220);
  }

  // fill copies the task into the panel. On refresh it leaves alone any field
  // that has focus or unsaved edits.
  function fill(t, initial) {
    var active = document.activeElement;
    function set(el, value) {
      if (!initial && (el === active || el.dataset.dirty)) return;
      el.value = value;
      delete el.dataset.dirty;
    }
    function setSelect(el, list, value) {
      if (!initial && el === active) return;
      el.innerHTML = options(list, value);
    }
    var done = t.status === "completed";
    q("#p-ref").innerHTML = App.projectIcon(t.project) + "<span>" + esc(t.project) + '</span><span aria-hidden="true">/</span><code>' + esc(t.name) + "</code>";
    set(q("#p-subject"), t.subject || "");
    var status = q("#p-status");
    setSelect(status, statusOptions(t), t.status);
    status.disabled = done;
    status.title = done ? "Done is final" : "";
    setSelect(q("#p-priority"), PRIORITY_OPTIONS, t.priority);
    setSelect(q("#p-type"), TYPE_OPTIONS, t.type || "task");
    set(q("#p-feature"), t.feature || "");
    set(q("#p-content"), t.content || "");
    q("#p-run").classList.toggle("hidden", done || t.computed_status === "blocked");

    var pending = t.pending_dependencies || [];
    var stuck = App.stuckBlockers(t);
    q("#p-deps").innerHTML = (t.depends_on || []).map(function (d) {
      var waiting = pending.indexOf(d) >= 0;
      var state = stuck.indexOf(d) >= 0 ? "stuck" : waiting ? "pending" : "done";
      var text = { stuck: "not done, stuck", pending: "not done yet", done: "done" }[state];
      return '<span class="dep ' + state + '">' + icon(state === "done" ? "check" : state === "stuck" ? "alert" : "wait") + esc(d) +
        '<span class="sr-only">, ' + text + '</span><button type="button" data-dep="' + esc(d) + '" aria-label="Remove dependency on ' + esc(d) + '">' + icon("x") + "</button></span>";
    }).join("");
    q("#p-dep-list").innerHTML = App.state.tasks.filter(function (x) {
      return x.project === t.project && x.id !== t.id && (t.depends_on || []).indexOf(x.name) < 0;
    }).map(function (x) { return '<option value="' + esc(x.name) + '">' + esc(x.subject) + "</option>"; }).join("");

    fillClaim(t);

    var comments = t.comments || [];
    q("#p-comment-count").textContent = comments.length || "";
    q("#p-comments").innerHTML = comments.length ? comments.map(function (cm) {
      var human = cm.author === "user";
      return '<article class="comment' + (human ? " human" : "") + '"><p class="by"><b' + (human ? "" : ' class="agent"') + ">" + esc(human ? "You" : cm.author) + "</b> · " +
        '<time datetime="' + esc(cm.created_at) + '">' + esc(App.relTime(cm.created_at)) + '</time></p><div class="body">' + esc(cm.body) + "</div></article>";
    }).join("") : '<p class="muted">No comments yet. Agents post progress here with <code>agmd task comment</code>.</p>';
  }

  function fillClaim(t) {
    var c = App.claimInfo(t);
    var a = t.assignment;
    var box = q("#p-claim");
    var detail = "", action = "";
    if (!c) {
      detail = t.computed_status === "blocked" ? "Unclaimed. Agents can claim it once its dependencies are done."
        : t.status === "completed" ? "Nobody." : "Unclaimed. An agent can pick it up with <code>agmd task next --claim</code>.";
      box.innerHTML = '<p class="claim-detail">' + detail + "</p>";
      return;
    }
    if (c.kind === "orphan") {
      detail = "In progress, but no agent holds it, so it won't be handed out.";
      action = '<button type="button" class="btn" id="p-requeue">' + icon("requeue") + "Re-queue</button>";
    } else if (c.kind === "stale") {
      detail = c.who + " stopped checking in " + App.relTime(a.updated_at) + ". Another agent may take it over automatically.";
      action = '<button type="button" class="btn" id="p-release">' + icon("requeue") + "Release and re-queue</button>";
    } else if (c.kind === "done") {
      detail = "";
    } else {
      action = '<button type="button" class="btn subtle" id="p-release">Release claim</button>';
    }
    if (a && a.worktree_path) detail += (detail ? " " : "") + "Worktree <code>" + esc(a.worktree_path) + "</code>";
    box.innerHTML = App.claimHTML(t) + (detail ? '<p class="claim-detail">' + detail + "</p>" : "") + (action ? '<div class="claim-actions">' + action + "</div>" : "");
    var rel = q("#p-release");
    if (rel) rel.addEventListener("click", function () { App.releaseClaim(current()); });
    var rq = q("#p-requeue");
    if (rq) rq.addEventListener("click", function () { App.moveTask(current(), "pending"); });
  }

  function current() { return App.state.byId[openId]; }

  function patch(fields) {
    var t = current();
    if (!t) return Promise.resolve();
    return App.mutate("PATCH", App.taskUrl(t), fields).catch(function (err) {
      App.toast(err.message, "error");
      App.loadState();
    });
  }

  // saveOnBlur PATCHes a text field when it loses focus with a changed value.
  function saveOnBlur(el, field, read) {
    el.addEventListener("input", function () { el.dataset.dirty = "1"; });
    el.addEventListener("blur", function () {
      var t = current();
      if (!t || !el.dataset.dirty) return;
      delete el.dataset.dirty;
      var value = el.value.trim();
      if (value === (read(t) || "").trim()) return;
      var body = {};
      body[field] = value;
      patch(body);
    });
  }

  function bindView() {
    q("#p-close").addEventListener("click", close);
    q("#p-run").addEventListener("click", function () { App.Runs.openModal([current()]); });
    q("#p-delete").addEventListener("click", function () {
      var t = current();
      if (!t) return;
      App.confirm({
        title: "Delete “" + (t.subject || t.name) + "”?",
        body: "This removes the task file, its claim and its comments. Agents will no longer see " + t.name + ". This can't be undone.",
        confirmLabel: "Delete task",
        danger: true
      }).then(function (ok) {
        if (!ok) return;
        return App.mutate("DELETE", App.taskUrl(t)).then(function () {
          App.toast("Deleted “" + (t.subject || t.name) + "”.");
          close();
        });
      }).catch(function (err) { App.toast(err.message, "error"); });
    });

    var subject = q("#p-subject");
    saveOnBlur(subject, "subject", function (t) { return t.subject; });
    subject.addEventListener("keydown", function (e) { if (e.key === "Enter") subject.blur(); });
    saveOnBlur(q("#p-feature"), "feature", function (t) { return t.feature; });
    var content = q("#p-content");
    saveOnBlur(content, "content", function (t) { return t.content; });
    content.addEventListener("keydown", function (e) { if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) content.blur(); });

    q("#p-status").addEventListener("change", function (e) {
      var t = current();
      App.moveTask(t, e.target.value).then(function (moved) {
        if (!moved && current()) fill(current(), true);
      });
    });
    q("#p-priority").addEventListener("change", function (e) { patch({ priority: Number(e.target.value) }); });
    q("#p-type").addEventListener("change", function (e) { patch({ type: e.target.value }); });

    q("#p-deps").addEventListener("click", function (e) {
      var btn = e.target.closest("button[data-dep]");
      if (!btn) return;
      var t = current();
      patch({ depends_on: (t.depends_on || []).filter(function (d) { return d !== btn.dataset.dep; }) });
    });
    var depAdd = q("#p-dep-add");
    depAdd.addEventListener("keydown", function (e) {
      if (e.key !== "Enter" || !depAdd.value.trim()) return;
      e.preventDefault();
      var t = current();
      App.mutate("PATCH", App.taskUrl(t), { depends_on: (t.depends_on || []).concat([depAdd.value.trim()]) })
        .then(function () { depAdd.value = ""; depAdd.removeAttribute("aria-invalid"); })
        .catch(function (err) {
          depAdd.setAttribute("aria-invalid", "true");
          App.toast(err.message, "error");
        });
    });

    var box = q("#p-comment");
    function send() {
      var t = current();
      var body = box.value.trim();
      if (!t || !body) return;
      q("#p-send").disabled = true;
      App.mutate("POST", App.taskUrl(t, "/comments"), { body: body })
        .then(function () { box.value = ""; })
        .catch(function (err) { App.toast(err.message, "error"); })
        .finally(function () { q("#p-send").disabled = false; });
    }
    q("#p-send").addEventListener("click", send);
    box.addEventListener("keydown", function (e) { if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) send(); });
  }

  // ---------- Create ----------
  // Mirrors the server's slug rule so the user sees the name agents will use.
  function slugFor(subject) {
    var s = subject.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, "");
    if (s.length > 48) s = s.slice(0, 48).replace(/-+$/, "");
    return s || "task";
  }

  function validateProjectName(name) {
    if (!name) return "Give the project a name.";
    if (name.charAt(0) === ".") return "A project name can't start with a dot.";
    if (/[\/\\]/.test(name)) return "A project name can't contain / or \\.";
    return "";
  }

  function openCreate() {
    if (mode === null) returnFocus = document.activeElement;
    openId = null;
    mode = "create";
    var st = App.state;
    var project = st.project || st.currentProject || (st.projects[0] && st.projects[0].name) || "";
    var names = st.projects.map(function (p) { return p.name; });
    if (project && names.indexOf(project) < 0) names.unshift(project);
    panel.setAttribute("aria-label", "New task");
    panel.innerHTML =
      '<div class="panel-head"><span class="panel-ref">New task</span><span class="spacer"></span>' +
      '<button type="button" class="icon-btn" id="p-close" title="Close (Esc)" aria-label="Close panel">' + icon("x") + "</button></div>" +
      '<form class="panel-body" id="c-form" novalidate>' +
      '<label class="sr-only" for="c-subject">Title</label>' +
      '<input class="panel-title" id="c-subject" type="text" placeholder="Task title" aria-describedby="c-slug" required>' +
      '<p class="slug-preview" id="c-slug"></p>' +
      '<div class="props">' +
      '<label for="c-project">Project</label><select id="c-project">' +
      names.map(function (n) { return "<option" + (n === project ? " selected" : "") + ">" + esc(n) + "</option>"; }).join("") +
      '<option value="__new">New project…</option></select>' +
      '<label for="c-priority">Priority</label><select id="c-priority">' + options(PRIORITY_OPTIONS, 2) + "</select>" +
      '<label for="c-type">Type</label><select id="c-type">' + options(TYPE_OPTIONS, "task") + "</select>" +
      '<label for="c-feature">Feature</label><input id="c-feature" type="text" placeholder="None">' +
      '<label for="c-deps">Depends on</label><div><input id="c-deps" type="text" list="c-dep-list" placeholder="Task names, comma separated" aria-describedby="c-deps-error"><datalist id="c-dep-list"></datalist><p class="field-error" id="c-deps-error"></p></div>' +
      "</div>" +
      '<label class="section-label" for="c-content">Description</label>' +
      '<textarea class="desc" id="c-content" placeholder="What needs doing? Markdown is fine."></textarea>' +
      "</form>" +
      '<div class="panel-foot"><p class="err" id="c-err" role="alert"></p><button type="button" class="btn" id="c-cancel">Cancel</button><button type="submit" form="c-form" class="btn primary" id="c-create">Create task</button></div>';

    var subject = q("#c-subject");
    function updateSlug() {
      var v = subject.value.trim();
      q("#c-slug").innerHTML = v ? "Agents will call it <code>" + esc(slugFor(v)) + "</code>" : "";
    }
    function fillDeps() {
      var p = q("#c-project").value;
      q("#c-dep-list").innerHTML = st.tasks.filter(function (x) { return x.project === p && x.status !== "completed"; })
        .map(function (x) { return '<option value="' + esc(x.name) + '">' + esc(x.subject) + "</option>"; }).join("");
    }
    subject.addEventListener("input", updateSlug);
    q("#c-project").addEventListener("change", function (e) {
      if (e.target.value !== "__new") { fillDeps(); return; }
      App.prompt({ title: "New project", label: "Project name", placeholder: "my-app", confirmLabel: "Use project", validate: validateProjectName }).then(function (name) {
        if (!name) { e.target.value = project; return; }
        var opt = document.createElement("option");
        opt.textContent = name;
        e.target.insertBefore(opt, e.target.lastElementChild);
        e.target.value = name;
        fillDeps();
      });
    });
    fillDeps();
    q("#p-close").addEventListener("click", close);
    q("#c-cancel").addEventListener("click", close);
    q("#c-form").addEventListener("submit", function (e) { e.preventDefault(); create(); });
    panel.querySelectorAll("#c-form input, #c-form textarea").forEach(function (el) {
      el.addEventListener("keydown", function (e) {
        if (e.key === "Enter" && (e.metaKey || e.ctrlKey || el.id === "c-subject")) { e.preventDefault(); create(); }
      });
    });
    show(subject);
    App.render();
  }

  function create() {
    var subject = q("#c-subject");
    var err = q("#c-err");
    var depsField = q("#c-deps");
    err.textContent = "";
    q("#c-deps-error").textContent = "";
    subject.removeAttribute("aria-invalid");
    depsField.removeAttribute("aria-invalid");
    if (!subject.value.trim()) {
      err.textContent = "Give the task a title.";
      subject.setAttribute("aria-invalid", "true");
      subject.focus();
      return;
    }
    var btn = q("#c-create");
    btn.disabled = true;
    var body = {
      project: q("#c-project").value,
      subject: subject.value.trim(),
      content: q("#c-content").value,
      priority: Number(q("#c-priority").value),
      type: q("#c-type").value,
      feature: q("#c-feature").value.trim(),
      depends_on: depsField.value.split(",").map(function (s) { return s.trim(); }).filter(Boolean)
    };
    App.mutate("POST", "/api/tasks", body).then(function (t) {
      App.toast("Created " + t.name + ".");
      if (App.state.project && App.state.project !== t.project) App.state.project = t.project;
      open(t.id);
    }).catch(function (e) {
      if (/dependency/.test(e.message)) {
        q("#c-deps-error").textContent = e.message.charAt(0).toUpperCase() + e.message.slice(1) + ".";
        depsField.setAttribute("aria-invalid", "true");
        depsField.focus();
      } else {
        err.textContent = e.message;
      }
      btn.disabled = false;
    });
  }

  App.Panel = {
    open: open,
    openCreate: openCreate,
    close: close,
    isOpen: function () { return mode !== null; },
    openId: function () { return openId; },
    refresh: function () {
      if (mode !== "view") return;
      var t = current();
      if (!t) { close(); return; }
      fill(t, false);
    }
  };

  document.addEventListener("DOMContentLoaded", function () {
    panel = App.$("panel");
    panel.inert = true;
  });
})();
