package main

// indexHTML is the single-page, mobile-friendly UI served at "/".
// It talks to the /api/run endpoint and renders a form per tool plus a
// built-in Help tab with how-to-run instructions.
const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover">
<meta name="theme-color" content="#0b1020">
<title>pentoolkit</title>
<style>
  :root {
    --bg:#0b1020; --panel:#151b30; --panel2:#1d2540; --text:#e7ecff;
    --muted:#9aa6c8; --accent:#4f8cff; --accent2:#2b6bff; --ok:#39d98a;
    --warn:#ffb020; --err:#ff5c72; --border:#2a3358;
  }
  * { box-sizing:border-box; -webkit-tap-highlight-color:transparent; }
  body { margin:0; font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;
    background:var(--bg); color:var(--text); line-height:1.5; }
  header { padding:16px 16px calc(16px + env(safe-area-inset-top));
    background:linear-gradient(160deg,#1a2350 0%,#0b1020 100%); border-bottom:1px solid var(--border); }
  h1 { margin:0; font-size:22px; letter-spacing:.5px; }
  .sub { color:var(--muted); font-size:13px; margin-top:2px; }
  .warn { margin:12px 16px; padding:10px 12px; background:rgba(255,176,32,.12);
    border:1px solid var(--warn); border-radius:10px; color:#ffd98a; font-size:13px; }
  .tabs { display:flex; gap:8px; padding:0 16px; margin-top:12px; }
  .tab { flex:1; padding:12px; text-align:center; background:var(--panel); border:1px solid var(--border);
    border-radius:12px 12px 0 0; color:var(--muted); font-weight:600; cursor:pointer; }
  .tab.active { background:var(--panel2); color:var(--text); border-bottom-color:var(--panel2); }
  main { padding:16px; padding-bottom:calc(24px + env(safe-area-inset-bottom)); }
  .card { background:var(--panel); border:1px solid var(--border); border-radius:14px; padding:16px; margin-bottom:16px; }
  label { display:block; font-size:13px; color:var(--muted); margin:12px 0 6px; }
  select, input[type=text], input[type=number] {
    width:100%; padding:14px; font-size:16px; color:var(--text); background:var(--panel2);
    border:1px solid var(--border); border-radius:12px; outline:none; }
  select:focus, input:focus { border-color:var(--accent); }
  .row { display:flex; align-items:center; gap:10px; margin-top:14px; }
  .row input[type=checkbox] { width:22px; height:22px; }
  button { width:100%; margin-top:18px; padding:16px; font-size:17px; font-weight:700; color:#fff;
    background:linear-gradient(180deg,var(--accent),var(--accent2)); border:none; border-radius:14px; cursor:pointer; }
  button:active { transform:translateY(1px); }
  button:disabled { opacity:.6; }
  .desc { font-size:13px; color:var(--muted); margin-top:6px; }
  pre { white-space:pre-wrap; word-break:break-word; background:#080c18; border:1px solid var(--border);
    border-radius:12px; padding:14px; font-size:13px; overflow:auto; max-height:60vh; margin:0; }
  .status { font-size:13px; margin:10px 0; }
  .status.ok { color:var(--ok); } .status.err { color:var(--err); }
  h2 { font-size:16px; margin:20px 0 8px; }
  h3 { font-size:14px; margin:16px 0 4px; color:var(--accent); }
  code { background:var(--panel2); padding:2px 6px; border-radius:6px; font-size:13px; }
  .cmd { display:block; background:#080c18; border:1px solid var(--border); border-radius:10px;
    padding:10px 12px; margin:6px 0; font-family:ui-monospace,Menlo,Consolas,monospace; font-size:13px; overflow:auto; }
  .hidden { display:none; }
  a { color:var(--accent); }
</style>
</head>
<body>
<header>
  <h1>pentoolkit</h1>
  <div class="sub">recon toolkit for authorized testing</div>
</header>
<div class="warn">⚠️ Only run these against systems you own or have written permission to test.</div>
<div class="tabs">
  <div class="tab active" id="tab-tools" onclick="showTab('tools')">Tools</div>
  <div class="tab" id="tab-threats" onclick="showTab('threats')">Threats</div>
  <div class="tab" id="tab-help" onclick="showTab('help')">Help</div>
</div>

<main>
  <section id="view-tools">
    <div class="card">
      <label for="tool">Choose a tool</label>
      <select id="tool" onchange="renderForm()"></select>
      <div class="desc" id="tool-desc"></div>
      <div id="fields"></div>
      <div class="row">
        <input type="checkbox" id="jsonOut">
        <label for="jsonOut" style="margin:0;">JSON output</label>
      </div>
      <button id="runBtn" onclick="runTool()">Run</button>
    </div>
    <div class="card">
      <div class="status" id="status">Ready.</div>
      <pre id="output">Output will appear here.</pre>
    </div>
  </section>

  <section id="view-threats" class="hidden">
    <div class="card">
      <h2>Threat reference (defender view)</h2>
      <p class="desc">The 12 most common attack types — what each is, how to
      <b>detect</b> it, and how to <b>defend</b> against it. Tap one to expand.</p>
      <div id="threats"></div>
    </div>
  </section>

  <section id="view-help" class="hidden">
    <div class="card">
      <h2>How to run</h2>
      <p class="desc">This page is served by the pentoolkit binary. Pick a tool above, fill in the target, and tap Run.</p>
      <h3>Recon workflow (in order)</h3>
      <p class="desc">Map the domain, enumerate hosts, scan ports, identify services, then inspect the web/TLS surface.</p>
      <ol class="desc">
        <li><b>dns</b> — what a domain points at</li>
        <li><b>subenum</b> — find subdomains</li>
        <li><b>portscan</b> — which ports are open</li>
        <li><b>banner</b> — identify a service on a port</li>
        <li><b>tlsinfo</b> — inspect the TLS certificate</li>
        <li><b>httpheaders</b> — audit web security headers</li>
        <li><b>httpprobe</b> — discover interesting paths</li>
      </ol>
      <h3>Run it from the command line too</h3>
      <span class="cmd">go build -o pentoolkit ./cmd/pentoolkit</span>
      <span class="cmd">./pentoolkit I need help</span>
      <span class="cmd">./pentoolkit dns -domain example.com</span>
      <h3>Start this web app</h3>
      <span class="cmd">./pentoolkit serve</span>
      <p class="desc">Then open <code>http://127.0.0.1:8787</code>. To use it from your phone on the
      same Wi-Fi, run <code>./pentoolkit serve -addr 0.0.0.0:8787</code> and browse to your
      computer's LAN IP on that port.</p>
      <h3>Notes</h3>
      <p class="desc">Add JSON output with the checkbox for machine-readable results. Works on Linux, macOS,
      and Windows — it's a single Go binary with no external dependencies.</p>
    </div>
  </section>
</main>

<script>
// Each tool's form fields. flag = CLI flag, def = default value.
const TOOLS = {
  dns:        { desc:"Look up a domain's addresses and mail/name servers.",
    fields:[{flag:"-domain", label:"Domain", ph:"example.com", req:true}] },
  subenum:    { desc:"Find extra sites hidden under a domain (mail., dev., ...).",
    fields:[{flag:"-domain", label:"Domain", ph:"example.com", req:true}] },
  portscan:   { desc:"See which ports (doors) are open on a machine.",
    fields:[{flag:"-host", label:"Host", ph:"example.com or 10.0.0.5", req:true},
            {flag:"-ports", label:"Ports", def:"1-1024", ph:"22,80,443 or 1-1024"}] },
  banner:     { desc:"Ask an open port what program and version is behind it.",
    fields:[{flag:"-host", label:"Host", ph:"example.com", req:true},
            {flag:"-port", label:"Port", type:"number", ph:"22", req:true}] },
  tlsinfo:    { desc:"Check a site's HTTPS certificate and encryption.",
    fields:[{flag:"-host", label:"Host", ph:"example.com", req:true},
            {flag:"-port", label:"Port", type:"number", def:"443"}] },
  httpheaders:{ desc:"Check whether a website's security settings are turned on.",
    fields:[{flag:"-url", label:"URL", ph:"https://example.com", req:true}] },
  httpprobe:  { desc:"Look for common hidden pages on a site (like /admin).",
    fields:[{flag:"-url", label:"URL", ph:"https://example.com", req:true}] },
};

function showTab(name) {
  ['tools','threats','help'].forEach(function(t){
    document.getElementById('view-'+t).classList.toggle('hidden', name!==t);
    document.getElementById('tab-'+t).classList.toggle('active', name===t);
  });
  if (name==='threats') loadThreats();
}

let threatsLoaded = false;
function loadThreats() {
  if (threatsLoaded) return;
  threatsLoaded = true;
  const box = document.getElementById('threats');
  fetch('/api/attacks').then(function(r){ return r.json(); }).then(function(list){
    box.innerHTML = '';
    list.forEach(function(a){
      const d = document.createElement('details');
      d.style.cssText = 'border:1px solid var(--border);border-radius:10px;padding:10px 12px;margin:8px 0;background:var(--panel2);';
      const s = document.createElement('summary');
      s.style.cssText = 'cursor:pointer;font-weight:600;';
      s.textContent = a.name;
      d.appendChild(s);
      const body = document.createElement('div');
      body.className = 'desc';
      body.innerHTML = '<p>'+esc(a.summary)+'</p>'
        + '<b>Detect</b><ul>' + a.detect.map(function(x){return '<li>'+esc(x)+'</li>';}).join('') + '</ul>'
        + '<b>Defend</b><ul>' + a.defend.map(function(x){return '<li>'+esc(x)+'</li>';}).join('') + '</ul>';
      d.appendChild(body);
      box.appendChild(d);
    });
  }).catch(function(e){ box.textContent = 'Failed to load: '+e; threatsLoaded=false; });
}

function esc(s){ const d=document.createElement('div'); d.textContent=s; return d.innerHTML; }

function initTools() {
  const sel = document.getElementById('tool');
  Object.keys(TOOLS).forEach(function(k){
    const o = document.createElement('option'); o.value=k; o.textContent=k; sel.appendChild(o);
  });
  renderForm();
}

function renderForm() {
  const key = document.getElementById('tool').value;
  const t = TOOLS[key];
  document.getElementById('tool-desc').textContent = t.desc;
  const box = document.getElementById('fields');
  box.innerHTML = '';
  t.fields.forEach(function(f, i){
    const lab = document.createElement('label'); lab.textContent = f.label + (f.req?' *':'');
    const inp = document.createElement('input');
    inp.type = f.type || 'text';
    inp.id = 'f'+i;
    inp.dataset.flag = f.flag;
    if (f.def) inp.value = f.def;
    if (f.ph) inp.placeholder = f.ph;
    box.appendChild(lab); box.appendChild(inp);
  });
}

function runTool() {
  const key = document.getElementById('tool').value;
  const inputs = document.querySelectorAll('#fields input');
  const args = [];
  let missing = false;
  const t = TOOLS[key];
  inputs.forEach(function(inp, i){
    const v = inp.value.trim();
    if (t.fields[i].req && !v) missing = true;
    if (v) { args.push(inp.dataset.flag); args.push(v); }
  });
  const status = document.getElementById('status');
  const out = document.getElementById('output');
  if (missing) { status.className='status err'; status.textContent='Please fill in the required (*) fields.'; return; }

  const btn = document.getElementById('runBtn');
  btn.disabled = true; status.className='status'; status.textContent='Running '+key+'...'; out.textContent='';

  fetch('/api/run', {
    method:'POST', headers:{'Content-Type':'application/json'},
    body: JSON.stringify({ tool:key, args:args, json:document.getElementById('jsonOut').checked })
  }).then(function(r){ return r.json(); }).then(function(res){
    out.textContent = res.output || '(no output)';
    if (res.ok) { status.className='status ok'; status.textContent='Done.'; }
    else { status.className='status err'; status.textContent = res.error || 'Command failed.'; }
  }).catch(function(e){
    status.className='status err'; status.textContent='Request failed: '+e;
  }).finally(function(){ btn.disabled=false; });
}

initTools();
</script>
</body>
</html>
`
