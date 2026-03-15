# FortiLogin

FortiLogin is a small command-line tool for the NIT Kurukshetra firewall login flow.

It is meant to solve one specific problem cleanly:

- when your device is connected to the local network
- but the firewall is blocking internet access
- FortiLogin detects that state and logs you in automatically

It also gives you explicit commands for:

- `login`
- `logout`
- `update`
- `status`

It does not open a browser. It does not keep launching tabs. It is designed to sit quietly in the background.

## Quick Start

If you want the shortest working install path, use one of these.

Linux:

```bash
git clone https://github.com/shivamrkm/fortilogin
cd fortilogin
go build -o fortilogin ./cmd/fortilogin
sudo install -m 0755 fortilogin /usr/bin/fortilogin
fortilogin update
fortilogin daemon
```

Windows:

```powershell
git clone https://github.com/shivamrkm/fortilogin
cd fortilogin
go build -o fortilogin.exe .\cmd\fortilogin
New-Item -ItemType Directory -Force C:\Tools\FortiLogin | Out-Null
Copy-Item .\fortilogin.exe C:\Tools\FortiLogin\fortilogin.exe
cd C:\Tools\FortiLogin
.\fortilogin.exe update
.\fortilogin.exe daemon
```

## What It Does

FortiLogin can be used in two ways:

1. as a normal one-shot CLI command
2. as a background loop using `fortilogin daemon`

When the daemon is running, it checks connectivity every 30 seconds.

- If internet is already working, it does nothing.
- If the firewall is blocking internet, it attempts login.
- If saved credentials fail twice, it stops retrying and keeps warning until you update them.

## Commands

```bash
fortilogin daemon
fortilogin login
fortilogin logout
fortilogin update
fortilogin status
```

What they mean:

- `fortilogin daemon`: run the background auto-login loop
- `fortilogin login`: try one login immediately
- `fortilogin logout`: send the logout request
- `fortilogin update`: save or replace credentials
- `fortilogin status`: show current state

## Config Location

Saved credentials are stored here:

- Linux: `~/.config/fortilogin/config.json`
- Windows: `%AppData%\fortilogin\config.json`

On Linux, old config from this path is migrated automatically if present:

```text
~/.config/NitAgent/config.json
```

## Linux Setup

This section is written for someone starting from scratch on Linux.

### Linux Requirements

- Go 1.22 or newer
- `git`
- `systemd` if you want automatic background startup on boot

### Linux: Clone The Repository

```bash
git clone https://github.com/shivamrkm/fortilogin
cd fortilogin
```

Example:

```bash
git clone git@github.com:yourname/fortilogin.git
cd fortilogin
```

### Linux: Build The Binary

```bash
go build -o fortilogin ./cmd/fortilogin
```

That creates a local binary named `fortilogin` in the repo folder.

### Linux: Install The Binary

```bash
sudo install -m 0755 fortilogin /usr/bin/fortilogin
```

Verify installation:

```bash
fortilogin status
```

### Linux: Save Credentials

Run:

```bash
fortilogin update
```

You will be prompted for:

- roll number / username
- password

### Linux: Test It Manually

Check status:

```bash
fortilogin status
```

Try one immediate login:

```bash
fortilogin login
```

Try logout:

```bash
fortilogin logout
```

### Linux: Run The Daemon In The Terminal

If you want to see it work before setting up `systemd`, run:

```bash
fortilogin daemon
```

That keeps it running in the current terminal window until you stop it with `Ctrl+C`.

### Linux: Run It Automatically With systemd

If you want FortiLogin to start automatically on boot, use the included systemd service template:

```text
packaging/systemd/fortilogin.service
```

Replace `__FORTILOGIN_USER__` with your actual Linux username.

Examples:

- on a laptop user account, that might be `shivamrkm`
- on a server, that might be `hpc`

Run the following commands one by one:

```bash
sudo systemctl unmask fortilogin.service 2>/dev/null || true
sed 's/__FORTILOGIN_USER__/yourusername/g' packaging/systemd/fortilogin.service | sudo tee /etc/systemd/system/fortilogin.service >/dev/null
sudo systemctl daemon-reload
sudo systemctl enable --now fortilogin.service
```

Verify it:

```bash
sudo systemctl status fortilogin.service
```

Watch logs:

```bash
sudo journalctl -u fortilogin.service -f
```

Useful service commands:

```bash
sudo systemctl restart fortilogin.service
sudo systemctl stop fortilogin.service
sudo systemctl start fortilogin.service
```

### Linux: Full Quick Path

If someone just wants the full Linux flow in order:

```bash
git clone https://github.com/shivamrkm/fortilogin
cd fortilogin
go build -o fortilogin ./cmd/fortilogin
sudo install -m 0755 fortilogin /usr/bin/fortilogin
fortilogin update
fortilogin status
fortilogin daemon
```

If they want background startup on boot too:

```bash
sudo systemctl unmask fortilogin.service 2>/dev/null || true
sed 's/__FORTILOGIN_USER__/yourusername/g' packaging/systemd/fortilogin.service | sudo tee /etc/systemd/system/fortilogin.service >/dev/null
sudo systemctl daemon-reload
sudo systemctl enable --now fortilogin.service
sudo systemctl status fortilogin.service
```

Example for a machine where the login user is `hpc`:

```bash
sudo systemctl unmask fortilogin.service 2>/dev/null || true
sed 's/__FORTILOGIN_USER__/hpc/g' packaging/systemd/fortilogin.service | sudo tee /etc/systemd/system/fortilogin.service >/dev/null
sudo systemctl daemon-reload
sudo systemctl enable --now fortilogin.service
sudo systemctl status fortilogin.service --no-pager
```

## Windows Setup

This section is written for someone starting from scratch on Windows.

Windows support is CLI-based. There is no Windows Service in this repository.

The intended Windows setup is:

1. clone the repo
2. build `fortilogin.exe`
3. save credentials once
4. test it from PowerShell or Command Prompt
5. add it to Startup so it runs automatically when the user signs in

### Windows Requirements

- Go 1.22 or newer
- Git
- PowerShell

### Windows: Clone The Repository

Open PowerShell and run:

```powershell
git clone https://github.com/shivamrkm/fortilogin
cd fortilogin
```

Example:

```powershell
git clone git@github.com:yourname/fortilogin.git
cd fortilogin
```

### Windows: Build The Executable

In PowerShell:

```powershell
go build -o fortilogin.exe .\cmd\fortilogin
```

That creates `fortilogin.exe` in the repo folder.

### Windows: Move It To A Stable Location

Do not leave the executable in a temporary folder if you plan to auto-start it.

A good place is something like:

```text
C:\Tools\FortiLogin\fortilogin.exe
```

Example:

```powershell
New-Item -ItemType Directory -Force C:\Tools\FortiLogin | Out-Null
Copy-Item .\fortilogin.exe C:\Tools\FortiLogin\fortilogin.exe
```

### Windows: Save Credentials

Open PowerShell in the folder where `fortilogin.exe` is stored:

```powershell
cd C:\Tools\FortiLogin
.\fortilogin.exe update
```

It will ask for:

- roll number / username
- password

### Windows: Test It Manually

In PowerShell:

```powershell
.\fortilogin.exe status
.\fortilogin.exe login
.\fortilogin.exe logout
```

### Windows: Run It In A PowerShell Window

To keep it running in the current PowerShell session:

```powershell
.\fortilogin.exe daemon
```

It will keep running until you close the window or press `Ctrl+C`.

### Windows: Start It Automatically When You Sign In

The repository includes a PowerShell helper:

```text
packaging/windows/install-startup.ps1
```

This helper creates a shortcut in the current user's Startup folder so Windows runs:

```text
fortilogin.exe daemon
```

Example:

```powershell
cd <path-to-the-repo>
powershell -ExecutionPolicy Bypass -File .\packaging\windows\install-startup.ps1 -BinaryPath C:\Tools\FortiLogin\fortilogin.exe
```

After that:

- sign out and sign back in, or
- open the Startup folder and verify the shortcut exists

Startup folder location:

```text
%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup
```

### Windows: Full Quick Path

If someone wants the full Windows flow in order:

```powershell
git clone https://github.com/shivamrkm/fortilogin
cd fortilogin
go build -o fortilogin.exe .\cmd\fortilogin
New-Item -ItemType Directory -Force C:\Tools\FortiLogin | Out-Null
Copy-Item .\fortilogin.exe C:\Tools\FortiLogin\fortilogin.exe
cd C:\Tools\FortiLogin
.\fortilogin.exe update
.\fortilogin.exe status
.\fortilogin.exe daemon
```

If they want automatic startup too:

```powershell
cd <path-to-the-repo>
powershell -ExecutionPolicy Bypass -File .\packaging\windows\install-startup.ps1 -BinaryPath C:\Tools\FortiLogin\fortilogin.exe
```

## GitHub Releases

Right now, the main documented installation path is:

1. clone the repo
2. build the binary
3. install or place it properly
4. run the daemon

The repository is also set up to publish release artifacts from GitHub Actions when you push a version tag like:

```bash
v0.1.0
```

Release artifacts configured in the workflow:

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

If you have not created a tagged release yet, those binary assets will not appear on GitHub yet. That is why this README does not treat GitHub Releases as the main install method.

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

It does not auto-install a systemd service because the username in the service file must be customized first.

## Troubleshooting

### Linux: `fortilogin.service is masked`

Run:

```bash
sudo systemctl unmask fortilogin.service
sudo systemctl daemon-reload
```

Then install the unit again and enable it.

### Linux: `status=217/USER`

This usually means the `User=` or `Group=` in the service file is wrong.

Check the installed unit:

```bash
sudo systemctl cat fortilogin.service
```

Make sure it contains the correct Linux username, then reinstall the unit with the right replacement value and run:

```bash
sudo systemctl daemon-reload
sudo systemctl restart fortilogin.service
```

### Linux: Service starts but does nothing

Check logs:

```bash
sudo journalctl -u fortilogin.service -f
```

Check the CLI directly:

```bash
fortilogin status
fortilogin login
```

### Linux: Binary not found by systemd

If you followed the README build/install path, the binary should be here:

```text
/usr/local/bin/fortilogin
```

Verify it:

```bash
ls -l /usr/local/bin/fortilogin
```

### Windows: Startup helper ran but FortiLogin does not start on sign-in

Check that the shortcut exists in:

```text
%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup
```

Also make sure the binary still exists at the exact path you gave to `install-startup.ps1`.

## Default Behavior

FortiLogin behaves like this by default:

- `fortilogin daemon` checks connectivity every 30 seconds
- if internet is already working, it does nothing
- if the firewall is blocking internet, it attempts login
- it does not open a browser
- `fortilogin logout` uses the fixed logout token behavior currently built into the tool
- after 2 failed credential attempts, the daemon stops retrying and keeps warning until `fortilogin update` is run

## Recommended Linux Install Flow

If you are installing on Linux for actual daily use, this is the recommended order:

1. Clone the repo.
2. Build the binary.
3. Install it to `/usr/bin/fortilogin`.
4. Run `fortilogin update`.
5. Run `fortilogin status`.
6. Test `fortilogin login`.
7. Install the `systemd` unit.
8. Enable and start the service.

That order avoids most setup mistakes.

## Notes

- This tool is specific to the observed NIT Kurukshetra firewall behavior.
- `logout` uses the fixed logout token that was observed to work globally.
- The tool does not open a browser.
- The daemon checks every 30 seconds.
- If credentials fail twice, the daemon stops retrying and keeps warning until credentials are updated.
- If the firewall flow changes, the tool will need changes.
