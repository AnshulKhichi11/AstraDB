<div align="center">

<br/>

```
 ▄▄▄· .▄▄ · ▄▄▄▄▄▄▄▄   ▄▄▄·     ·▄▄▄▄  ▄▄▄▄·
▐█ ▀█ ▐█ ▀. •██  ▀▄ █·▐█ ▀█     ██▪ ██ ▐█ ▀█▪
▄█▀▀█ ▄▀▀▀█▄ ▐█.▪▐▀▀▄ ▄█▀▀█     ▐█· ▐█▌▐█▀▀█▄
▐█ ▪▐▌▐█▄▪▐█ ▐█▌·▐█•█▌▐█ ▪▐▌    ██. ██ ██▄▪▐█
 ▀  ▀  ▀▀▀▀  ▀▀▀ .▀  ▀ ▀  ▀     ▀▀▀▀▀• ·▀▀▀▀
```

# AstraDB

### ⚡ Lightning-Fast Standalone Database Engine

<p align="center">

![Go](https://img.shields.io/badge/Built%20with-Go-00ADD8?style=for-the-badge)
![npm](https://img.shields.io/npm/v/astradb?style=for-the-badge\&color=7c3aed)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-black?style=for-the-badge)
![Storage](https://img.shields.io/badge/storage-Segment%20Engine-06b6d4?style=for-the-badge)
![WAL](https://img.shields.io/badge/durability-WAL-22c55e?style=for-the-badge)
![License](https://img.shields.io/badge/license-MIT-22c55e?style=for-the-badge)

</p>

<br/>

```bash
npm install -g astradb
astradb start
```

Database ready at:

```
http://localhost:8080
```

</div>

---

# What is AstraDB

AstraDB is a modern standalone database engine designed for speed, simplicity, and zero friction.

No Docker
No installers
No dependencies
No configuration

Just run and build.

---

# Core Principles

```
Speed        → Go-based storage engine
Simplicity   → Single command startup
Durability   → WAL crash recovery
Portability  → Single binary distribution
Control      → Full local ownership
```

---

# Architecture Diagram

```
                ┌─────────────────────┐
                │     Applications    │
                │                     │
                │  Web Apps          │
                │  Desktop Apps     │
                │  Backend APIs     │
                └─────────┬──────────┘
                          │ REST API
                          ▼
                ┌─────────────────────┐
                │     AstraDB API    │
                │   (HTTP Server)   │
                └─────────┬──────────┘
                          │
                          ▼
                ┌─────────────────────┐
                │    Query Engine     │
                └─────────┬──────────┘
                          │
          ┌───────────────┼───────────────┐
          ▼               ▼               ▼
  ┌────────────┐ ┌────────────┐ ┌────────────┐
  │  Segments  │ │   Indexes  │ │     WAL    │
  │  Storage   │ │            │ │ Crash Log  │
  └────────────┘ └────────────┘ └────────────┘
          │
          ▼
    ┌──────────────┐
    │ Disk Storage │
    └──────────────┘
```

---

# Astra Ecosystem Diagram

```
              ┌─────────────────────┐
              │     AstraCloud      │
              │  Managed Database   │
              └─────────▲──────────┘
                        │
                        │
┌──────────────┐        │        ┌──────────────┐
│ AstraForge   │◄───────┼───────►│   AstraDB    │
│ Desktop App  │        │        │ Database     │
└──────────────┘        │        └──────────────┘
                        │
                        ▼
               ┌──────────────────┐
               │  Local Storage   │
               └──────────────────┘
```

---

# Feature Overview

### Storage Engine

Segment-based architecture
Optimized disk writes
Fast append operations

### Durability

Write-Ahead Logging
Crash recovery
Automatic checkpointing

### Performance

Go concurrency model
Minimal memory footprint
Fast startup

### Developer Experience

npm install support
REST API access
CLI interface

---

# Quick Example

Insert document:

```bash
curl -X POST http://localhost:8080/api/insert \
-H "Content-Type: application/json" \
-d "{\"db\":\"app\",\"collection\":\"users\",\"data\":{\"name\":\"Astra\",\"role\":\"developer\"}}"
```

Query:

```bash
curl -X POST http://localhost:8080/api/query \
-H "Content-Type: application/json" \
-d "{\"db\":\"app\",\"collection\":\"users\",\"filter\":{}}"
```

---

# Storage Layout

```
data/
├── databases/
│   └── app/
│       └── users/
│           ├── data.db
│           └── segments/
├── wal/
└── metadata/
```

---

# Installation Methods

### npm (Recommended)

```
npm install -g astradb
astradb start
```

### Binary

Download from GitHub Releases and run.

---

# Ecosystem Components

### AstraDB

Core database engine

### AstraForge

Visual database studio (Desktop App)

Features:

Database explorer
Query editor
Document viewer

Coming soon.

### AstraCloud

Managed cloud database

Features:

Cloud access
Remote connections
Managed infrastructure

Coming soon.

---

# Performance Model

```
Write → WAL → Segment → Disk
Read  → Index → Segment → Result
```

Optimized for durability and speed.

---

# Roadmap

```
AstraForge Studio
AstraCloud Platform
Schema Validation
Replication
Advanced Indexing
SDK Support
```

---

# Technology Stack

```
Backend      → Go
Desktop App  → Electron + React
CLI          → Node.js
Storage      → Custom Segment Engine
Durability   → WAL System
```

---

# Philosophy

AstraDB is not just a database.

It is the foundation for the Astra ecosystem.

Built for developers who value speed, simplicity, and control.

---

<div align="center">

Install AstraDB and experience instant database power.

npm install -g astradb

</div>
