<div align="center">

<br/>

```
 ▄▄▄· .▄▄ · ▄▄▄▄▄▄▄▄   ▄▄▄·     ·▄▄▄▄  ▄▄▄▄·
▐█ ▀█ ▐█ ▀. •██  ▀▄ █·▐█ ▀█     ██▪ ██ ▐█ ▀█▪
▄█▀▀█ ▄▀▀▀█▄ ▐█.▪▐▀▀▄ ▄█▀▀█     ▐█· ▐█▌▐█▀▀█▄
▐█ ▪▐▌▐█▄▪▐█ ▐█▌·▐█•█▌▐█ ▪▐▌    ██. ██ ██▄▪▐█
 ▀  ▀  ▀▀▀▀  ▀▀▀ .▀  ▀ ▀  ▀     ▀▀▀▀▀• ·▀▀▀▀
```

### AstraDB — Standalone Database Engine

**Fast. Lightweight. Zero-dependency database built for modern developers.**

[![Platform](https://img.shields.io/badge/platform-Windows%20|%20Linux%20|%20macOS-0a0a0a?style=flat-square\&labelColor=1a1a2e)](https://github.com/AnshulKhichi11/AstraDB/releases)
[![Language](https://img.shields.io/badge/built%20with-Go-00ADD8?style=flat-square\&labelColor=1a1a2e)](https://go.dev)
[![npm](https://img.shields.io/npm/v/astradb?style=flat-square\&labelColor=1a1a2e\&color=22c55e)](https://www.npmjs.com/package/astradb)
[![License](https://img.shields.io/badge/license-MIT-22c55e?style=flat-square\&labelColor=1a1a2e)](LICENSE)

<br/>

[Quick Start](#-quick-start) · [Install via npm](#-install-via-npm) · [Insert Data](#-insert-data) · [Architecture](#-architecture) · [Roadmap](#-roadmap)

</div>

---

# ✦ Overview

AstraDB is a **standalone database engine** designed to run instantly — no installation, no external dependencies, and no runtime setup.

It provides:

* Embedded-style simplicity
* Server-based flexibility
* Modern document storage
* High-performance Go backend

AstraDB runs as a single binary or via npm, making it ideal for:

* Local development
* Desktop applications
* Internal tools
* Lightweight production systems
* Custom database-powered applications

---

# ⚡ Install via npm (Recommended)

Install globally:

```bash
npm install -g astradb
```

Start AstraDB:

```bash
astradb start
```

AstraDB will start at:

```
http://localhost:8080
```

Verify:

```
http://localhost:8080/health
```

Response:

```json
{ "status": "ok" }
```

---

# 📥 Install via Binary

Download from Releases:

https://github.com/AnshulKhichi11/AstraDB/releases

Run:

Windows:

```powershell
.\astradb-windows-x64.exe
```

Linux/macOS:

```bash
chmod +x astradb-linux-x64
./astradb-linux-x64
```

---

# 📦 Insert Data

Create database, collection, and insert document:

```bash
curl -X POST http://localhost:8080/api/insert \
-H "Content-Type: application/json" \
-d "{\"db\":\"myapp\",\"collection\":\"users\",\"data\":{\"name\":\"John\",\"role\":\"developer\"}}"
```

Query data:

```bash
curl -X POST http://localhost:8080/api/query \
-H "Content-Type: application/json" \
-d "{\"db\":\"myapp\",\"collection\":\"users\",\"filter\":{}}"
```

---

# 📂 Storage Architecture

AstraDB stores data in:

```
data/
├── databases/
│   └── <database>/
│       └── collections/
│           └── <collection>/
│               ├── data.db
│               └── segments/
├── wal/
└── metadata/
```

Features:

* Segment-based storage engine
* Write-Ahead Logging (WAL)
* Crash recovery
* Persistent storage
* Automatic checkpointing

---

# ⚙ Core Features

* Document-based storage
* Zero-dependency binary
* WAL-based durability
* High-performance Go engine
* Index support
* REST API access
* Cross-platform

---

# 🧩 Upcoming Products

## AstraForge — Database Studio (Coming Soon)

AstraForge is a desktop application for AstraDB, similar to MongoDB Compass.

Features:

* Visual database explorer
* Document viewer and editor
* Query builder
* Index management
* Stats and monitoring

Built using:

* Electron
* React
* Modern UI architecture

---

## AstraCloud — Managed AstraDB (Coming Soon)

AstraCloud will allow you to:

* Store data in the cloud
* Access from anywhere
* Connect multiple applications
* Manage databases visually

---

# 🧠 Architecture

Backend:

* Go storage engine
* Segment-based database
* WAL crash recovery

Client ecosystem:

* CLI (astradb)
* npm package
* REST API
* AstraForge desktop studio (coming soon)

---

# 🚀 Project Status

Current version: v1.x
Status: Active development

Used internally for upcoming Astra ecosystem projects.

---

# 🛣 Roadmap

Planned features:

* AstraForge Desktop Studio
* AstraCloud managed hosting
* Schema validation
* Advanced indexing
* Replication
* Performance optimizations
* SDK support

---

# 📄 License

MIT License

---

<div align="center">

Built with Go
Part of the Astra ecosystem

</div>
