<div align="center">

<br/>

```id="astraheader"
 ▄▄▄· .▄▄ · ▄▄▄▄▄▄▄▄   ▄▄▄·     ·▄▄▄▄  ▄▄▄▄·
▐█ ▀█ ▐█ ▀. •██  ▀▄ █·▐█ ▀█     ██▪ ██ ▐█ ▀█▪
▄█▀▀█ ▄▀▀▀█▄ ▐█.▪▐▀▀▄ ▄█▀▀█     ▐█· ▐█▌▐█▀▀█▄
▐█ ▪▐▌▐█▄▪▐█ ▐█▌·▐█•█▌▐█ ▪▐▌    ██. ██ ██▄▪▐█
 ▀  ▀  ▀▀▀▀  ▀▀▀ .▀  ▀ ▀  ▀     ▀▀▀▀▀• ·▀▀▀▀
```

# AstraDB

### The Future of Lightweight Databases

**Blazing fast. Zero setup. Built for developers who move fast.**

<br/>

<p align="center">

[![npm](https://img.shields.io/npm/v/astradb?style=for-the-badge\&color=7c3aed)](https://www.npmjs.com/package/astradb)
[![Downloads](https://img.shields.io/npm/dt/astradb?style=for-the-badge\&color=06b6d4)](https://www.npmjs.com/package/astradb)
[![Platform](https://img.shields.io/badge/platform-Windows%20|%20Linux%20|%20macOS-black?style=for-the-badge)](https://github.com/AnshulKhichi11/AstraDB/releases)
[![Go](https://img.shields.io/badge/built%20with-Go-00ADD8?style=for-the-badge)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-22c55e?style=for-the-badge)](LICENSE)

</p>

<br/>

```bash id="installblock"
npm install -g astradb
astradb start
```

Your database is now live at:

```
http://localhost:8080
```

No configuration. No setup. No dependencies.

</div>

---

# Built for Modern Developers

AstraDB reimagines what a database should feel like.

Not heavy. Not complex. Not slow.

Just fast, simple, and powerful.

Run instantly. Store instantly. Build instantly.

Whether you're building a startup, desktop app, internal tool, or backend service — AstraDB is designed to stay out of your way.

---

# Why AstraDB Exists

Modern databases are powerful — but often too complex for everyday development.

You shouldn't need:

• Docker
• Complex installers
• External runtimes
• Heavy infrastructure

AstraDB removes all friction.

One command. One binary. Fully operational database.

---

# Experience Instant Setup

Install globally:

```bash id="installcmd"
npm install -g astradb
```

Start instantly:

```bash id="startcmd"
astradb start
```

Verify:

```
http://localhost:8080/health
```

Response:

```json id="healthresp"
{
  "status": "ok"
}
```

You're ready to build.

---

# Store Your First Document

```bash id="insertcmd"
curl -X POST http://localhost:8080/api/insert \
-H "Content-Type: application/json" \
-d "{\"db\":\"app\",\"collection\":\"users\",\"data\":{\"name\":\"Astra\",\"role\":\"developer\"}}"
```

Query instantly:

```bash id="querycmd"
curl -X POST http://localhost:8080/api/query \
-H "Content-Type: application/json" \
-d "{\"db\":\"app\",\"collection\":\"users\",\"filter\":{}}"
```

Simple. Fast. Reliable.

---

# Designed for Performance

AstraDB uses a modern architecture optimized for speed and durability.

```
data/
├── databases/
├── collections/
├── wal/
└── metadata/
```

Core technologies include:

Segment-based storage engine
Write-Ahead Logging (WAL)
Crash recovery system
Automatic checkpointing
Persistent disk storage

Your data remains safe and consistent at all times.

---

# Core Features

Ultra fast startup time
Zero dependency runtime
Single binary database engine
Persistent storage with crash recovery
REST API interface
Cross platform support
Indexing support
npm CLI support

Built in Go for maximum performance.

---

# Meet the Astra Ecosystem

AstraDB is just the beginning.

We're building a complete database platform.

---

## AstraForge

### Visual Database Studio

A powerful desktop interface for AstraDB.

Explore, query, and manage your database visually.

Features:

Visual database explorer
Document editor
Query builder
Index manager
Database monitoring

Built using Electron and React.

Launching soon.

---

## AstraCloud

### Managed Cloud Database

The cloud version of AstraDB.

Access your database from anywhere.

Features:

Cloud storage
Remote access
Multi-device sync
Managed infrastructure

Launching soon.

---

# Architecture

AstraDB is powered by a high-performance Go storage engine.

Core layers include:

Storage Engine (Go)
Segment Manager
WAL Crash Recovery
REST API Server
CLI Interface
Desktop Studio (AstraForge)
Cloud Platform (AstraCloud)

---

# Built for Builders

AstraDB is designed for developers who want speed and simplicity.

Not configuration.

Not complexity.

Just build.

---

# Roadmap

Coming soon:

AstraForge Desktop Studio
AstraCloud platform
Schema validation
Advanced indexing
Replication
SDK support
Performance optimizations

---

# Install Now

```bash id="finalinstall"
npm install -g astradb
astradb start
```

Your database is ready in seconds.

---

<div align="center">

Built with Go
Part of the Astra ecosystem

The foundation of future Astra products

</div>
