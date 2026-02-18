## 📥 Download & Installation

AstraDB is distributed as pre-built binaries. You do NOT need Go installed to run AstraDB.

### Step 1: Download Binary

Go to the official Releases page:

🔗 https://github.com/AnshulKhichi11/AstraDB/releases

Download the appropriate binary for your operating system:

| OS | File |
|---|---|
| Windows (x64) | astradb-windows-x64.exe |
| Linux (x64) | astradb-linux-x64 |
| Mac Intel | astradb-darwin-x64 |
| Mac Apple Silicon (M1/M2/M3) | astradb-darwin-arm64 |

---

### Step 2: Run AstraDB

#### Windows
Open Command Prompt or PowerShell in the download folder:

```powershell
.\astradb-windows-x64.exe
Linux / Mac
Make executable and run:

chmod +x astradb-linux-x64
./astradb-linux-x64
Step 3: Verify Server
Open your browser:

http://localhost:8080/health
Expected response:

{"status":"ok"}
📂 Default Data Directory
AstraDB automatically creates a data directory in the same folder:

data/
├── databases/
├── collections/
├── wal/
└── metadata/
This directory stores all database files and logs.

🛑 Stop Server
Press:

Ctrl + C
⚙️ Run in Background (Optional)
Windows:

start astradb-windows-x64.exe
Linux / Mac:

nohup ./astradb-linux-x64 &
🔄 Updating AstraDB
Download the latest binary from Releases

Replace old binary

Restart AstraDB

Your data remains safe in the data/ folder.

🧩 No Installation Required
AstraDB runs as a standalone executable:

No installer needed

No dependencies required

No Go installation required


---

```md
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20Mac-blue)
![Language](https://img.shields.io/badge/language-Go-00ADD8)
![License](https://img.shields.io/badge/license-MIT-green)
![Status](https://img.shields.io/badge/status-Active-success)
