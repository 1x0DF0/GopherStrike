# Building GopherStrike

## Quick Start

1. **Build the binary:**
   ```bash
   ./build.sh
   ```

2. **Install to your system:**
   ```bash
   ./install.sh
   ```

## Manual Build

If you prefer to build manually:

```bash
go build -o GopherStrike main.go
```

## Running GopherStrike

After installation, you can run GopherStrike from anywhere:
```bash
gopherstrike
```

Or run directly without installation:
```bash
./GopherStrike
```

## Port Scanner Usage

The port scanner now handles privilege escalation automatically:

- **With sudo:** `sudo gopherstrike` - Will run directly without re-escalating
- **Without sudo:** `gopherstrike` - Will prompt for admin password when needed

### Ctrl+C Behavior

- **In main menu:** Exits the program
- **In any tool:** Returns to main menu gracefully

## Troubleshooting

If the port scanner fails with permission errors:
1. Make sure you have sudo/admin privileges
2. On Linux/macOS, the tool will attempt to use pkexec or sudo
3. On macOS, it may prompt for your password via GUI

## Dependencies

The port scanner requires Python dependencies:
```bash
sudo pip3 install python-nmap scapy
```