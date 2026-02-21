<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>AstraDB — README</title>
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@300;400;500;700&family=Syne:wght@400;600;700;800&display=swap" rel="stylesheet">
<style>
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

  :root {
    --bg: #080b14;
    --surface: #0d1120;
    --surface2: #111827;
    --border: #1e2d4a;
    --accent: #4f8ef7;
    --accent2: #7c3aed;
    --accent3: #06b6d4;
    --green: #22c55e;
    --text: #e2e8f0;
    --muted: #64748b;
    --code-bg: #0a0f1e;
  }

  html { scroll-behavior: smooth; }

  body {
    background: var(--bg);
    color: var(--text);
    font-family: 'Syne', sans-serif;
    line-height: 1.7;
    min-height: 100vh;
    overflow-x: hidden;
  }

  /* Starfield background */
  body::before {
    content: '';
    position: fixed;
    inset: 0;
    background-image:
      radial-gradient(1px 1px at 20% 30%, rgba(79,142,247,0.4) 0%, transparent 100%),
      radial-gradient(1px 1px at 80% 10%, rgba(124,58,237,0.3) 0%, transparent 100%),
      radial-gradient(1px 1px at 50% 80%, rgba(6,182,212,0.3) 0%, transparent 100%),
      radial-gradient(1px 1px at 10% 60%, rgba(79,142,247,0.2) 0%, transparent 100%),
      radial-gradient(1px 1px at 90% 70%, rgba(124,58,237,0.2) 0%, transparent 100%);
    pointer-events: none;
    z-index: 0;
  }

  .wrapper {
    position: relative;
    z-index: 1;
    max-width: 860px;
    margin: 0 auto;
    padding: 0 24px 80px;
  }

  /* HERO */
  .hero {
    text-align: center;
    padding: 72px 0 56px;
    position: relative;
  }

  .hero::after {
    content: '';
    position: absolute;
    bottom: 0;
    left: 50%;
    transform: translateX(-50%);
    width: 1px;
    height: 40px;
    background: linear-gradient(to bottom, var(--border), transparent);
  }

  .logo-wrap {
    display: inline-block;
    position: relative;
    margin-bottom: 32px;
  }

  .logo-glow {
    position: absolute;
    inset: -20px;
    background: radial-gradient(ellipse at center, rgba(79,142,247,0.15) 0%, transparent 70%);
    pointer-events: none;
  }

  .logo {
    font-family: 'JetBrains Mono', monospace;
    font-size: clamp(9px, 1.8vw, 14px);
    font-weight: 500;
    letter-spacing: 0.05em;
    line-height: 1.3;
    white-space: pre;
    background: linear-gradient(135deg, #4f8ef7 0%, #7c3aed 50%, #06b6d4 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    filter: drop-shadow(0 0 20px rgba(79,142,247,0.4));
  }

  .tagline {
    font-size: 1.05rem;
    color: var(--muted);
    font-weight: 400;
    letter-spacing: 0.02em;
    margin-bottom: 28px;
  }

  .tagline span {
    color: var(--accent);
    font-weight: 600;
  }

  /* Badges */
  .badges {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    justify-content: center;
    margin-bottom: 36px;
  }

  .badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 5px 12px;
    border-radius: 6px;
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.72rem;
    font-weight: 500;
    border: 1px solid var(--border);
    background: var(--surface);
    letter-spacing: 0.04em;
  }

  .badge-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
  }

  /* Nav links */
  .nav {
    display: flex;
    flex-wrap: wrap;
    gap: 6px 24px;
    justify-content: center;
  }

  .nav a {
    color: var(--muted);
    text-decoration: none;
    font-size: 0.82rem;
    font-family: 'JetBrains Mono', monospace;
    letter-spacing: 0.05em;
    transition: color 0.2s;
    position: relative;
  }

  .nav a::after {
    content: '';
    position: absolute;
    bottom: -2px;
    left: 0;
    right: 0;
    height: 1px;
    background: var(--accent);
    transform: scaleX(0);
    transition: transform 0.2s;
  }

  .nav a:hover { color: var(--accent); }
  .nav a:hover::after { transform: scaleX(1); }

  .nav-sep { color: var(--border); }

  /* Divider */
  .divider {
    height: 1px;
    background: linear-gradient(to right, transparent, var(--border), transparent);
    margin: 48px 0;
  }

  /* Sections */
  .section { margin-bottom: 48px; }

  .section-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 20px;
  }

  .section-icon {
    font-size: 1.1rem;
  }

  h2 {
    font-family: 'Syne', sans-serif;
    font-size: 1.15rem;
    font-weight: 700;
    color: var(--text);
    letter-spacing: 0.03em;
    text-transform: uppercase;
  }

  .section-line {
    flex: 1;
    height: 1px;
    background: linear-gradient(to right, var(--border), transparent);
  }

  p {
    color: #94a3b8;
    font-size: 0.95rem;
    font-weight: 400;
    line-height: 1.8;
  }

  p strong, .highlight {
    color: var(--text);
    font-weight: 600;
  }

  /* Download table */
  .dl-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
    margin-top: 16px;
  }

  .dl-card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 10px;
    padding: 16px 18px;
    display: flex;
    align-items: center;
    gap: 14px;
    text-decoration: none;
    transition: border-color 0.2s, background 0.2s;
    cursor: pointer;
  }

  .dl-card:hover {
    border-color: var(--accent);
    background: rgba(79,142,247,0.05);
  }

  .dl-os-icon { font-size: 1.6rem; line-height: 1; }

  .dl-info { display: flex; flex-direction: column; gap: 2px; }

  .dl-os {
    font-family: 'Syne', sans-serif;
    font-size: 0.82rem;
    font-weight: 700;
    color: var(--text);
    letter-spacing: 0.03em;
  }

  .dl-file {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.7rem;
    color: var(--accent);
    letter-spacing: 0.03em;
  }

  .dl-link {
    margin-left: auto;
    font-size: 0.75rem;
    color: var(--muted);
    font-family: 'JetBrains Mono', monospace;
  }

  /* Steps */
  .steps { display: flex; flex-direction: column; gap: 24px; margin-top: 8px; }

  .step {
    display: flex;
    gap: 16px;
  }

  .step-num {
    flex-shrink: 0;
    width: 28px;
    height: 28px;
    border-radius: 50%;
    background: linear-gradient(135deg, var(--accent2), var(--accent));
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.72rem;
    font-weight: 700;
    font-family: 'JetBrains Mono', monospace;
    color: white;
    margin-top: 2px;
  }

  .step-content { flex: 1; }

  .step-title {
    font-size: 0.88rem;
    font-weight: 700;
    color: var(--text);
    margin-bottom: 8px;
    letter-spacing: 0.02em;
  }

  /* OS tabs */
  .os-tabs { display: flex; gap: 8px; margin-bottom: 10px; }

  .os-tab {
    padding: 4px 12px;
    border-radius: 5px;
    font-size: 0.72rem;
    font-family: 'JetBrains Mono', monospace;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--muted);
    cursor: pointer;
    transition: all 0.15s;
  }

  .os-tab.active {
    background: var(--accent);
    border-color: var(--accent);
    color: white;
  }

  /* Code blocks */
  .code-block {
    background: var(--code-bg);
    border: 1px solid var(--border);
    border-radius: 8px;
    overflow: hidden;
  }

  .code-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 14px;
    border-bottom: 1px solid var(--border);
  }

  .code-lang {
    font-size: 0.68rem;
    font-family: 'JetBrains Mono', monospace;
    color: var(--muted);
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  .code-dots { display: flex; gap: 5px; }
  .code-dot { width: 8px; height: 8px; border-radius: 50%; }
  .dot-r { background: #ef4444; }
  .dot-y { background: #eab308; }
  .dot-g { background: #22c55e; }

  pre {
    padding: 14px 16px;
    overflow-x: auto;
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.82rem;
    line-height: 1.6;
    color: #c9d8ef;
  }

  .cmd { color: #7dd3fc; }
  .flag { color: #a78bfa; }
  .path { color: var(--accent3); }
  .comment { color: var(--muted); font-style: italic; }
  .str { color: #86efac; }

  /* Verify card */
  .verify-card {
    display: flex;
    align-items: center;
    gap: 16px;
    background: rgba(34,197,94,0.06);
    border: 1px solid rgba(34,197,94,0.2);
    border-radius: 10px;
    padding: 16px 20px;
    margin-top: 12px;
  }

  .verify-icon { font-size: 1.5rem; }

  .verify-text { flex: 1; }

  .verify-url {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.8rem;
    color: var(--accent3);
    margin-bottom: 4px;
  }

  .verify-resp {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.78rem;
    color: var(--green);
  }

  /* Data tree */
  .data-tree {
    background: var(--code-bg);
    border: 1px solid var(--border);
    border-radius: 10px;
    padding: 20px 24px;
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.82rem;
    line-height: 1.9;
    margin-top: 14px;
  }

  .tree-root { color: var(--accent); font-weight: 700; }
  .tree-dir { color: #7dd3fc; }
  .tree-desc { color: var(--muted); font-size: 0.75rem; margin-left: 4px; }
  .tree-branch { color: var(--border); }

  /* Info callout */
  .callout {
    display: flex;
    gap: 12px;
    background: rgba(79,142,247,0.07);
    border: 1px solid rgba(79,142,247,0.2);
    border-left: 3px solid var(--accent);
    border-radius: 0 8px 8px 0;
    padding: 14px 16px;
    margin-top: 14px;
    font-size: 0.88rem;
    color: #94a3b8;
  }

  .callout-icon { flex-shrink: 0; color: var(--accent); }

  /* Feature pills */
  .feature-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
    margin-top: 14px;
  }

  .feature-pill {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 12px 14px;
    font-size: 0.83rem;
  }

  .pill-check {
    color: var(--green);
    font-size: 0.9rem;
    margin-top: 1px;
    flex-shrink: 0;
  }

  .pill-title {
    font-weight: 700;
    color: var(--text);
    font-size: 0.82rem;
    margin-bottom: 2px;
  }

  .pill-desc { color: var(--muted); font-size: 0.76rem; }

  /* Update steps */
  .update-steps {
    display: flex;
    gap: 0;
    margin-top: 14px;
  }

  .update-step {
    flex: 1;
    text-align: center;
    padding: 18px 12px;
    background: var(--surface);
    border: 1px solid var(--border);
    position: relative;
  }

  .update-step:first-child { border-radius: 8px 0 0 8px; }
  .update-step:last-child { border-radius: 0 8px 8px 0; }

  .update-step:not(:last-child)::after {
    content: '→';
    position: absolute;
    right: -10px;
    top: 50%;
    transform: translateY(-50%);
    color: var(--accent);
    font-size: 0.9rem;
    z-index: 1;
    background: var(--surface);
    padding: 2px 4px;
  }

  .update-num {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.65rem;
    color: var(--accent);
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    margin-bottom: 4px;
  }

  .update-action {
    font-size: 0.78rem;
    font-weight: 600;
    color: var(--text);
  }

  /* Footer */
  .footer {
    text-align: center;
    padding-top: 48px;
    border-top: 1px solid var(--border);
    color: var(--muted);
    font-size: 0.78rem;
    font-family: 'JetBrains Mono', monospace;
    letter-spacing: 0.04em;
  }

  .footer a { color: var(--accent); text-decoration: none; }
  .footer a:hover { text-decoration: underline; }

  .footer-sep { margin: 0 10px; opacity: 0.4; }

  @media (max-width: 600px) {
    .dl-grid { grid-template-columns: 1fr; }
    .feature-grid { grid-template-columns: 1fr; }
    .update-steps { flex-direction: column; }
    .update-step:first-child { border-radius: 8px 8px 0 0; }
    .update-step:last-child { border-radius: 0 0 8px 8px; }
    .update-step:not(:last-child)::after { content: '↓'; right: 50%; bottom: -10px; top: auto; transform: translateX(50%); }
    .logo { font-size: 7px; }
  }
</style>
</head>
<body>
<div class="wrapper">

  <!-- HERO -->
  <div class="hero">
    <div class="logo-wrap">
      <div class="logo-glow"></div>
      <div class="logo"> ▄▄▄· .▄▄ · ▄▄▄▄▄▄▄▄   ▄▄▄·     ·▄▄▄▄  ▄▄▄▄·
▐█ ▀█ ▐█ ▀. •██  ▀▄ █·▐█ ▀█     ██▪ ██ ▐█ ▀█▪
▄█▀▀█ ▄▀▀▀█▄ ▐█.▪▐▀▀▄ ▄█▀▀█     ▐█· ▐█▌▐█▀▀█▄
▐█ ▪▐▌▐█▄▪▐█ ▐█▌·▐█•█▌▐█ ▪▐▌    ██. ██ ██▄▪▐█
 ▀  ▀  ▀▀▀▀  ▀▀▀ .▀  ▀ ▀  ▀     ▀▀▀▀▀• ·▀▀▀▀</div>
    </div>

    <p class="tagline">A <span>blazing-fast</span>, standalone database engine — no dependencies, no setup.</p>

    <div class="badges">
      <span class="badge"><span class="badge-dot" style="background:#4f8ef7"></span>Windows · Linux · macOS</span>
      <span class="badge"><span class="badge-dot" style="background:#00ADD8"></span>Built with Go</span>
      <span class="badge"><span class="badge-dot" style="background:#22c55e"></span>MIT License</span>
      <span class="badge"><span class="badge-dot" style="background:#7c3aed"></span>Active</span>
    </div>

    <nav class="nav">
      <a href="#download">Download</a>
      <span class="nav-sep">·</span>
      <a href="#quickstart">Quick Start</a>
      <span class="nav-sep">·</span>
      <a href="#data">Data Storage</a>
      <span class="nav-sep">·</span>
      <a href="#background">Background</a>
      <span class="nav-sep">·</span>
      <a href="#updating">Updating</a>
    </nav>
  </div>

  <div class="divider"></div>

  <!-- WHAT IS -->
  <div class="section">
    <div class="section-header">
      <span class="section-icon">✦</span>
      <h2>What is AstraDB?</h2>
      <div class="section-line"></div>
    </div>
    <p>AstraDB is a <strong>standalone database engine</strong> distributed as a single pre-built binary. Drop it in a folder, run it, and you have a fully operational database server — no installers, no runtime dependencies, no Go environment required.</p>
  </div>

  <!-- DOWNLOAD -->
  <div class="section" id="download">
    <div class="section-header">
      <span class="section-icon">⬇</span>
      <h2>Download</h2>
      <div class="section-line"></div>
    </div>
    <p>Head to the <a href="https://github.com/AnshulKhichi11/AstraDB/releases" style="color:var(--accent);text-decoration:none;font-weight:600">Releases page</a> and grab the binary for your platform:</p>

    <div class="dl-grid">
      <div class="dl-card">
        <div class="dl-os-icon">🪟</div>
        <div class="dl-info">
          <div class="dl-os">Windows x64</div>
          <div class="dl-file">astradb-windows-x64.exe</div>
        </div>
        <div class="dl-link">↓ .exe</div>
      </div>
      <div class="dl-card">
        <div class="dl-os-icon">🐧</div>
        <div class="dl-info">
          <div class="dl-os">Linux x64</div>
          <div class="dl-file">astradb-linux-x64</div>
        </div>
        <div class="dl-link">↓ bin</div>
      </div>
      <div class="dl-card">
        <div class="dl-os-icon">🍎</div>
        <div class="dl-info">
          <div class="dl-os">macOS Intel</div>
          <div class="dl-file">astradb-darwin-x64</div>
        </div>
        <div class="dl-link">↓ bin</div>
      </div>
      <div class="dl-card">
        <div class="dl-os-icon">🍎</div>
        <div class="dl-info">
          <div class="dl-os">macOS Apple Silicon</div>
          <div class="dl-file">astradb-darwin-arm64</div>
        </div>
        <div class="dl-link">↓ bin</div>
      </div>
    </div>
  </div>

  <!-- QUICK START -->
  <div class="section" id="quickstart">
    <div class="section-header">
      <span class="section-icon">⚡</span>
      <h2>Quick Start</h2>
      <div class="section-line"></div>
    </div>

    <div class="steps">
      <div class="step">
        <div class="step-num">1</div>
        <div class="step-content">
          <div class="step-title">Windows — Open PowerShell in your download folder</div>
          <div class="code-block">
            <div class="code-header">
              <div class="code-dots"><div class="code-dot dot-r"></div><div class="code-dot dot-y"></div><div class="code-dot dot-g"></div></div>
              <span class="code-lang">powershell</span>
            </div>
            <pre><span class="cmd">.\astradb-windows-x64.exe</span></pre>
          </div>
        </div>
      </div>

      <div class="step">
        <div class="step-num">2</div>
        <div class="step-content">
          <div class="step-title">Linux & macOS — Make executable and run</div>
          <div class="code-block">
            <div class="code-header">
              <div class="code-dots"><div class="code-dot dot-r"></div><div class="code-dot dot-y"></div><div class="code-dot dot-g"></div></div>
              <span class="code-lang">bash</span>
            </div>
            <pre><span class="cmd">chmod</span> <span class="flag">+x</span> <span class="path">astradb-linux-x64</span>
<span class="path">./astradb-linux-x64</span></pre>
          </div>
        </div>
      </div>

      <div class="step">
        <div class="step-num">3</div>
        <div class="step-content">
          <div class="step-title">Verify the server is running</div>
          <div class="verify-card">
            <div class="verify-icon">🌐</div>
            <div class="verify-text">
              <div class="verify-url">http://localhost:8080/health</div>
              <div class="verify-resp">→ {"status":"ok"} &nbsp;✓ You're live!</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- DATA -->
  <div class="section" id="data">
    <div class="section-header">
      <span class="section-icon">📂</span>
      <h2>Data Storage</h2>
      <div class="section-line"></div>
    </div>
    <p>AstraDB automatically creates a <code style="background:var(--code-bg);padding:2px 6px;border-radius:4px;font-family:'JetBrains Mono',monospace;font-size:0.82rem;color:var(--accent3)">data/</code> directory on first run:</p>

    <div class="data-tree">
      <div><span class="tree-root">data/</span></div>
      <div><span class="tree-branch">├──</span> <span class="tree-dir">databases/</span><span class="tree-desc">← your databases</span></div>
      <div><span class="tree-branch">├──</span> <span class="tree-dir">collections/</span><span class="tree-desc">← collection data</span></div>
      <div><span class="tree-branch">├──</span> <span class="tree-dir">wal/</span><span class="tree-desc">← write-ahead log (crash recovery)</span></div>
      <div><span class="tree-branch">└──</span> <span class="tree-dir">metadata/</span><span class="tree-desc">← internal metadata</span></div>
    </div>

    <div class="callout">
      <span class="callout-icon">ℹ</span>
      <span><strong style="color:var(--text)">Your data is always safe.</strong> Updating AstraDB only replaces the binary — the <code style="font-family:'JetBrains Mono',monospace;font-size:0.82rem">data/</code> folder is never modified.</span>
    </div>
  </div>

  <!-- STOP & BACKGROUND -->
  <div class="section" id="background">
    <div class="section-header">
      <span class="section-icon">🔁</span>
      <h2>Stop & Background</h2>
      <div class="section-line"></div>
    </div>

    <p style="margin-bottom:14px">To stop AstraDB, press <code style="background:var(--code-bg);padding:2px 8px;border-radius:4px;font-family:'JetBrains Mono',monospace;font-size:0.82rem;color:#f87171;border:1px solid var(--border)">Ctrl + C</code> in the terminal.</p>

    <p style="margin-bottom:12px">To keep AstraDB running after the terminal closes:</p>

    <div class="code-block" style="margin-bottom:10px">
      <div class="code-header">
        <div class="code-dots"><div class="code-dot dot-r"></div><div class="code-dot dot-y"></div><div class="code-dot dot-g"></div></div>
        <span class="code-lang">Windows</span>
      </div>
      <pre><span class="cmd">start</span> <span class="path">astradb-windows-x64.exe</span></pre>
    </div>

    <div class="code-block">
      <div class="code-header">
        <div class="code-dots"><div class="code-dot dot-r"></div><div class="code-dot dot-y"></div><div class="code-dot dot-g"></div></div>
        <span class="code-lang">Linux / macOS</span>
      </div>
      <pre><span class="cmd">nohup</span> <span class="path">./astradb-linux-x64</span> <span class="flag">&</span></pre>
    </div>
  </div>

  <!-- UPDATING -->
  <div class="section" id="updating">
    <div class="section-header">
      <span class="section-icon">🔄</span>
      <h2>Updating</h2>
      <div class="section-line"></div>
    </div>

    <div class="update-steps">
      <div class="update-step">
        <div class="update-num">Step 01</div>
        <div class="update-action">Download new binary</div>
      </div>
      <div class="update-step">
        <div class="update-num">Step 02</div>
        <div class="update-action">Replace old binary</div>
      </div>
      <div class="update-step">
        <div class="update-num">Step 03</div>
        <div class="update-action">Restart AstraDB</div>
      </div>
    </div>

    <div class="callout" style="margin-top:14px">
      <span class="callout-icon">✓</span>
      <span>Your data in <code style="font-family:'JetBrains Mono',monospace;font-size:0.82rem">data/</code> is preserved automatically across all updates.</span>
    </div>
  </div>

  <!-- NO SETUP -->
  <div class="section">
    <div class="section-header">
      <span class="section-icon">🧩</span>
      <h2>No Setup Required</h2>
      <div class="section-line"></div>
    </div>

    <div class="feature-grid">
      <div class="feature-pill">
        <span class="pill-check">✓</span>
        <div>
          <div class="pill-title">No Installer</div>
          <div class="pill-desc">Just download and run</div>
        </div>
      </div>
      <div class="feature-pill">
        <span class="pill-check">✓</span>
        <div>
          <div class="pill-title">No Dependencies</div>
          <div class="pill-desc">Single static binary</div>
        </div>
      </div>
      <div class="feature-pill">
        <span class="pill-check">✓</span>
        <div>
          <div class="pill-title">No Go Required</div>
          <div class="pill-desc">Pre-compiled for your platform</div>
        </div>
      </div>
      <div class="feature-pill">
        <span class="pill-check">✓</span>
        <div>
          <div class="pill-title">No Config Needed</div>
          <div class="pill-desc">Sensible defaults out of the box</div>
        </div>
      </div>
    </div>
  </div>

  <!-- FOOTER -->
  <div class="footer">
    <p>Built with ♥ using Go
      <span class="footer-sep">·</span>
      <a href="#">MIT License</a>
      <span class="footer-sep">·</span>
      <a href="https://github.com/AnshulKhichi11/AstraDB/issues">Report an Issue</a>
      <span class="footer-sep">·</span>
      <a href="https://github.com/AnshulKhichi11/AstraDB/releases">Releases</a>
    </p>
  </div>

</div>
</body>
</html>
