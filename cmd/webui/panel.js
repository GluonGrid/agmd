// agmd board: the side panel that shows, edits and creates a task. Each field
// saves on its own (PATCH with only that field), so an agent's concurrent
// change to another field is never overwritten.
(function () {
  "use strict";

  var App = window.App;
  var esc = App.esc;
  var panel;
  var openId = null;
  var mode = null; // "view" | "create"

  var STATUS_OPTIONS = [["pending", "Pending"], ["in_progress", "In progress"], ["completed", "Done"]];
  var PRIORITY_OPTIONS = [[0, "P0 · critical"], [1, "P1 · high"], [2, "P2 · normal"], [3, "P3 · low"], [4, "P4 · backlog"]];
  var TYPE_OPTIONS = [["task", "task"], ["bug", "bug"], ["feature", "feature"], ["chore", "chore"]];

  function options(list, value) {
    return list.map(function (o) {
      return '<option value="' + esc(o[0]) + '"' + (String(o[0]) === String(value) ? " selected" : "") + ">" + esc(o[1]) + "</option>";
    }).join("");
  }

  function q(sel) { return panel.querySelector(sel); }

  function show() {
    panel.classList.add("open");
    panel.setAttribute("aria-hidden", "false");
  }

  function close() {
    openId = null;
    mode = null;
    panel.classList.remove("open");
    panel.setAttribute("aria-hidden", "true");
    App.render();
  }

  // ---------- View / edit ----------
  function open(id) {
    var t = App.state.byId[id];
    if (!t) return;
    openId = id;
    mode = "view";
    panel.innerHTML =
      '<div class="panel-head"><span class="panel-ref" id="p-ref"></span><span class="spacer"></span>' +
      '<button class="icon-btn" id="p-run" title="Run with Claude Code">▶</button>' +
      '<button class="icon-btn" id="p-delete" title="Delete task">🗑</button>' +
      '<button class="icon-btn" id="p-close" title="Close (Esc)">×</button></div>' +
      '<div class="panel-body">' +
      '<input class="panel-title" id="p-subject" type="text" placeholder="Title">' +
      '<div class="props">' +
      '<label for="p-status">Status</label><select id="p-status"></select>' +
      '<label>Agent</label><div class="claim-box" id="p-claim"></div>' +
      '<label for="p-priority">Priority</label><select id="p-priority"></select>' +
      '<label for="p-type">Type</label><select id="p-type"></select>' +
      '<label for="p-feature">Feature</label><input id="p-feature" type="text" placeholder="None">' +
      '<label>Depends on</label><div class="deps"><span id="p-deps"></span><input id="p-dep-add" type="text" list="p-dep-list" placeholder="+ add"><datalist id="p-dep-list"></datalist></div>' +
      "</div>" +
      '<div class="section-label">Description</div>' +
      '<textarea class="desc" id="p-content" placeholder="What needs doing? Markdown is fine."></textarea>' +
      '<div class="section-label">Comments <span class="count" id="p-comment-count"></span></div>' +
      '<div id="p-comments"></div>' +
      '<div class="compose"><textarea id="p-comment" placeholder="Add a comment"></textarea>' +
      '<div class="row"><span>⌘↵ to send</span><button class="btn" id="p-send">Comment</button></div></div>' +
      "</div>";
    bindView();
    fill(t, true);
    show();
    App.render();
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
    q("#p-ref").textContent = t.project + " / " + t.name;
    set(q("#p-subject"), t.subject || "");
    function setSelect(el, list, value) {
      if (!initial && el === active) return;
      el.innerHTML = options(list, value);
    }
    var status = q("#p-status");
    setSelect(status, STATUS_OPTIONS, t.status);
    status.disabled = t.status === "completed";
    status.title = t.status === "completed" ? "Completed tasks stay done" : "";
    setSelect(q("#p-priority"), PRIORITY_OPTIONS, t.priority);
    setSelect(q("#p-type"), TYPE_OPTIONS, t.type || "task");
    set(q("#p-feature"), t.feature || "");
    set(q("#p-content"), t.content || "");
    q("#p-run").classList.toggle("hidden", t.status === "completed");

    var pending = t.pending_dependencies || [];
    q("#p-deps").innerHTML = (t.depends_on || []).map(function (d) {
      var waiting = pending.indexOf(d) >= 0;
      return '<span class="dep' + (waiting ? " pending" : "") + '" title="' + (waiting ? "Not done yet" : "Done") + '">' +
        esc(d) + (waiting ? " ⏳" : " ✓") + '<button data-dep="' + esc(d) + '" title="Remove dependency">×</button></span>';
    }).join("");
    q("#p-dep-list").innerHTML = App.state.tasks.filter(function (x) {
      return x.project === t.project && x.id !== t.id && (t.depends_on || []).indexOf(x.name) < 0;
    }).map(function (x) { return '<option value="' + esc(x.name) + '">' + esc(x.subject) + "</option>"; }).join("");

    var c = App.claimInfo(t);
    var a = t.assignment;
    var claim = q("#p-claim");
    if (!c) {
      claim.innerHTML = '<span class="claim-detail">' + (t.computed_status === "blocked"
        ? "Unclaimed. Agents can claim it once its dependencies are done."
        : t.status === "completed" ? "Nobody." : "Unclaimed. An agent can pick it up with <code>agmd task next --claim</code>.") + "</span>";
    } else {
      claim.innerHTML = App.claimHTML(t) +
        (a ? '<div class="claim-detail">Last heartbeat ' + esc(App.relTime(a.updated_at)) +
          (a.worktree_path ? ' · <code>' + esc(a.worktree_path) + "</code>" : "") + "</div>" +
          '<div><button class="btn" id="p-release">Release claim</button></div>'
          : '<div class="claim-detail">' + esc(c.title) + "</div>");
      var rel = q("#p-release");
      if (rel) rel.addEventListener("click", function () { release(t); });
    }

    var comments = t.comments || [];
    q("#p-comment-count").textContent = comments.length || "";
    q("#p-comments").innerHTML = comments.map(function (cm) {
      return '<div class="comment"><div class="by"><b>' + esc(cm.author) + "</b> · " + esc(App.relTime(cm.created_at)) +
        '</div><div class="body">' + esc(cm.body) + "</div></div>";
    }).join("");
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
      if (!t || !confirm("Delete task '" + t.name + "' and its comments?")) return;
      App.mutate("DELETE", App.taskUrl(t)).then(function () {
        App.toast("Deleted " + t.name);
        close();
      }).catch(function (err) { App.toast(err.message, "error"); });
    });

    var subject = q("#p-subject");
    saveOnBlur(subject, "subject", function (t) { return t.subject; });
    subject.addEventListener("keydown", function (e) { if (e.key === "Enter") subject.blur(); });
    saveOnBlur(q("#p-feature"), "feature", function (t) { return t.feature; });
    var content = q("#p-content");
    saveOnBlur(content, "content", function (t) { return t.content; });
    content.addEventListener("keydown", function (e) { if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) content.blur(); });

    q("#p-status").addEventListener("change", function (e) { App.moveTask(current(), e.target.value); });
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
      var t = current();
      patch({ depends_on: (t.depends_on || []).concat([depAdd.value.trim()]) }).then(function () { depAdd.value = ""; });
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

  function release(t) {
    var c = App.claimInfo(t);
    if (c && (c.kind === "active" || c.kind === "run") &&
      !confirm(c.who + " checked in " + c.when + ". Release its claim anyway? The task goes back to the queue.")) return;
    App.mutate("DELETE", App.taskUrl(t, "/claim"))
      .then(function () { App.toast("Released " + c.who + "'s claim"); })
      .catch(function (err) { App.toast(err.message, "error"); });
  }

  // ---------- Create ----------
  function openCreate() {
    openId = null;
    mode = "create";
    var st = App.state;
    var project = st.project || st.currentProject || (st.projects[0] && st.projects[0].name) || "";
    var names = st.projects.map(function (p) { return p.name; });
    if (project && names.indexOf(project) < 0) names.unshift(project);
    panel.innerHTML =
      '<div class="panel-head"><span class="panel-ref">New task</span><span class="spacer"></span>' +
      '<button class="icon-btn" id="p-close" title="Close (Esc)">×</button></div>' +
      '<div class="panel-body">' +
      '<input class="panel-title" id="c-subject" type="text" placeholder="Task title">' +
      '<div class="props">' +
      '<label for="c-project">Project</label><select id="c-project">' +
      names.map(function (n) { return '<option' + (n === project ? " selected" : "") + ">" + esc(n) + "</option>"; }).join("") +
      '<option value="__new">New project…</option></select>' +
      '<label for="c-priority">Priority</label><select id="c-priority">' + options(PRIORITY_OPTIONS, 2) + "</select>" +
      '<label for="c-type">Type</label><select id="c-type">' + options(TYPE_OPTIONS, "task") + "</select>" +
      '<label for="c-feature">Feature</label><input id="c-feature" type="text" placeholder="None">' +
      '<label for="c-deps">Depends on</label><input id="c-deps" type="text" list="c-dep-list" placeholder="task names, comma separated"><datalist id="c-dep-list"></datalist>' +
      "</div>" +
      '<div class="section-label">Description</div>' +
      '<textarea class="desc" id="c-content" placeholder="What needs doing? Markdown is fine."></textarea>' +
      "</div>" +
      '<div class="panel-foot"><span class="err" id="c-err"></span><button class="btn" id="c-cancel">Cancel</button><button class="btn primary" id="c-create">Create task</button></div>';

    function fillDeps() {
      var p = q("#c-project").value;
      q("#c-dep-list").innerHTML = st.tasks.filter(function (x) { return x.project === p && x.status !== "completed"; })
        .map(function (x) { return '<option value="' + esc(x.name) + '">' + esc(x.subject) + "</option>"; }).join("");
    }
    q("#c-project").addEventListener("change", function (e) {
      if (e.target.value !== "__new") { fillDeps(); return; }
      var name = (prompt("Project name") || "").trim();
      if (!name) { e.target.value = project; return; }
      var opt = document.createElement("option");
      opt.textContent = name;
      e.target.insertBefore(opt, e.target.lastElementChild);
      e.target.value = name;
      fillDeps();
    });
    fillDeps();
    q("#p-close").addEventListener("click", close);
    q("#c-cancel").addEventListener("click", close);
    q("#c-create").addEventListener("click", create);
    panel.querySelectorAll("input, textarea").forEach(function (el) {
      el.addEventListener("keydown", function (e) {
        if (e.key === "Enter" && (e.metaKey || e.ctrlKey || el.id === "c-subject")) { e.preventDefault(); create(); }
      });
    });
    show();
    App.render();
    q("#c-subject").focus();
  }

  function create() {
    var subject = q("#c-subject").value.trim();
    if (!subject) { q("#c-err").textContent = "Give the task a title."; q("#c-subject").focus(); return; }
    var btn = q("#c-create");
    btn.disabled = true;
    var body = {
      project: q("#c-project").value,
      subject: subject,
      content: q("#c-content").value,
      priority: Number(q("#c-priority").value),
      type: q("#c-type").value,
      feature: q("#c-feature").value.trim(),
      depends_on: q("#c-deps").value.split(",").map(function (s) { return s.trim(); }).filter(Boolean)
    };
    App.mutate("POST", "/api/tasks", body).then(function (t) {
      App.toast("Created " + t.name);
      if (App.state.project && App.state.project !== t.project) App.state.project = t.project;
      open(t.id);
    }).catch(function (err) {
      q("#c-err").textContent = err.message;
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

  document.addEventListener("DOMContentLoaded", function () { panel = App.$("panel"); });
})();
