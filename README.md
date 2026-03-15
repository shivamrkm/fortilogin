# FortiLogin

FortiLogin is a small command-line tool for the NIT Kurukshetra firewall login flow.

Its job is simple:

- detect when your device has network access but the firewall is blocking internet
- log you in automatically in the background
- let you manually `login`, `logout`, `update`, or check `status`
- avoid browser popups and browser auto-opening

This repository supports both Linux and Windows.

## What It Does

FortiLogin runs in one of two ways:

- as a one-shot command, like `fortilogin login`
- as a background loop, using `fortilogin daemon`

When the daemon is running, it checks connectivity every 30 seconds.

If internet is already working, it does nothing.

If internet is blocked by the firewall, it attempts login using the saved credentials.

If credentials fail twice, it stops retrying and keeps warning until you run:

```bash
fortilogin update
```

## Commands

```bash
fortilogin daemon
fortilogin login
fortilogin logout
fortilogin update
fortilogin status
```

What each command does:

- `fortilogin daemon`: background auto-login loop
- `fortilogin login`: try one immediate login
- `fortilogin logout`: send the logout request and print a short success/failure message
- `fortilogin update`: prompt for new credentials and save them
- `fortilogin status`: show credential and network state

## Config Location

FortiLogin stores credentials here:

- Linux: `~/.config/fortilogin/config.json`
- Windows: `%AppData%\fortilogin\config.json`

It also migrates old Linux config from:

```text
~/.config/NitAgent/config.json
```

## Install On Linux

There are two practical Linux paths:

1. download a release binary
2. build it yourself from source

### Linux: Option 1, Use GitHub Release Binary

Download the Linux release asset, then install it:

```bash
chmod +x fortilogin-linux-amd64
sudo install -m 0755 fortilogin-linux-amd64 /usr/local/bin/fortilogin
```

Check that it works:

```bash
fortilogin status
```

### Linux: Option 2, Build From Source

Requirements:

- Go 1.22 or newer

Clone the repo and build:

```bash
git clone <your-repo-url>
cd fortilogin
go build -o fortilogin ./cmd/fortilogin
sudo install -m 0755 fortilogin /usr/local/bin/fortilogin
```

Check that it works:

```bash
fortilogin status
```

### Linux: First-Time Setup

Run this once:

```bash
fortilogin update
```

It will ask for:

- roll number / username
- password

Then verify:

```bash
fortilogin status
```

### Linux: Run It Manually

For a single login attempt:

```bash
fortilogin login
```

To keep it running in the foreground:

```bash
fortilogin daemon
```

### Linux: Run It As A systemd Service

If you want this to start automatically on boot, use the provided systemd template:

```text
packaging/systemd/fortilogin.service
```

Replace `__FORTILOGIN_USER__` with your Linux username, then install it:

```bash
sed 's/__FORTILOGIN_USER__/yourusername/g' packaging/systemd/fortilogin.service | sudo tee /etc/systemd/system/fortilogin.service >/dev/null
sudo systemctl daemon-reload
sudo systemctl enable --now fortilogin.service
sudo systemctl status fortilogin.service
```

Useful service commands:

```bash
sudo systemctl status fortilogin.service
sudo systemctl restart fortilogin.service
sudo systemctl stop fortilogin.service
sudo journalctl -u fortilogin.service -f
```

## Install On Windows

Windows support is CLI-based. There is no Windows service in this repo.

The normal Windows flow is:

1. get `fortilogin.exe`
2. save credentials once with `update`
3. either run `daemon` manually or add it to Startup so it launches after login

You can either download the Windows binary from GitHub Releases or compile it yourself.

### Windows: Option 1, Use GitHub Release Binary

Download:

- `fortilogin-windows-amd64.exe`

Place it somewhere stable, for example:

```text
C:\Tools\FortiLogin\fortilogin.exe
```

Open PowerShell in that folder and run:

```powershell
.\fortilogin.exe update
```

Then test it:

```powershell
.\fortilogin.exe status
.\fortilogin.exe login
```

### Windows: Option 2, Build From Source

Requirements:

- Go 1.22 or newer
- Git

Clone and build:

```powershell
git clone <your-repo-url>
cd fortilogin
go build -o fortilogin.exe .\cmd\fortilogin
```

Then run:

```powershell
.\fortilogin.exe update
.\fortilogin.exe status
```

### Windows: Run It Manually

To start it in the current terminal:

```powershell
.\fortilogin.exe daemon
```

This keeps running until you close the terminal.

### Windows: Start It Automatically At Login

The repository includes a helper script:

```text
packaging/windows/install-startup.ps1
```

This creates a shortcut in your Startup folder so Windows launches:

```text
fortilogin.exe daemon
```

Steps:

1. Keep `fortilogin.exe` in a permanent location, for example:

```text
C:\Tools\FortiLogin\fortilogin.exe
```

2. Open PowerShell in the repo folder and run:

```powershell
powershell -ExecutionPolicy Bypass -File .\packaging\windows\install-startup.ps1 -BinaryPath C:\Tools\FortiLogin\fortilogin.exe
```

3. Log out and log back in, or check that the shortcut exists in:

```text
%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup
```

If someone wants to do it manually instead of using the script, they can create a shortcut to:

```text
fortilogin.exe
```

with arguments:

```text
daemon
```

and place that shortcut in the same Startup folder.

## GitHub Releases

This repo is configured to publish release artifacts when you push a tag like:

```bash
v0.1.0
```

The GitHub Actions workflow publishes:

- `fortilogin-linux-amd64`
- `fortilogin-windows-amd64.exe`
- `fortilogin_<version>_amd64.deb`

To create a release:

```bash
git add .
git commit -m "Prepare release"
git push origin main
git tag v0.1.0
git push origin v0.1.0
```

## Build Release Files Locally

Build Linux binary:

```bash
go build -o fortilogin ./cmd/fortilogin
```

Build Windows binary:

```bash
GOOS=windows GOARCH=amd64 go build -o fortilogin.exe ./cmd/fortilogin
```

Build Debian package:

```bash
chmod +x scripts/build-deb.sh
./scripts/build-deb.sh v0.1.0
```

That produces:

```text
dist/deb/fortilogin_0.1.0_amd64.deb
```

The `.deb` installs:

- `/usr/bin/fortilogin`
- `/usr/share/fortilogin/fortilogin.service.example`

It does not auto-install a systemd service because the username in the service file must be set correctly first.

## Important Behavior Notes

- This tool is specific to the observed NIT Kurukshetra firewall behavior.
- `logout` uses the fixed logout token that was observed to work globally.
- The tool does not open a browser.
- The daemon checks every 30 seconds.
- If credentials fail twice, the daemon stops retrying and keeps warning until credentials are updated.
- If the firewall behavior changes, this tool will need changes too.
