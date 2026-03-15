# FortiLogin

FortiLogin is a small Go CLI and Linux daemon for the NIT Kurukshetra firewall login flow.

It is built for a very specific workflow:

- auto-detect when internet access is blocked by the firewall
- auto-login in the background
- provide explicit `login`, `logout`, `update`, and `status` commands
- avoid browser popups and repeated browser opening
- stop retrying after repeated bad credentials and keep warning until credentials are updated

## Features

- `fortilogin daemon` checks connectivity every 30 seconds
- `fortilogin login` performs a one-shot login
- `fortilogin logout` uses the fixed logout token and prints a short success message
- `fortilogin update` updates saved credentials
- `fortilogin status` shows current state
- stores config in `~/.config/fortilogin/config.json`
- migrates old config from `~/.config/NitAgent/config.json`

## Commands

```bash
fortilogin daemon
fortilogin login
fortilogin logout
fortilogin update
fortilogin status
```

## Build

```bash
go build -o fortilogin ./cmd/fortilogin
```

## Install Binary Manually

```bash
sudo install -m 0755 fortilogin /usr/local/bin/fortilogin
```

## Systemd Service

The repository includes a systemd service template at `packaging/systemd/fortilogin.service`.

Before installing it, replace `__FORTILOGIN_USER__` with the actual Linux username that should own the saved config and run the daemon.

Example:

```bash
sed 's/__FORTILOGIN_USER__/yourusername/g' packaging/systemd/fortilogin.service | sudo tee /etc/systemd/system/fortilogin.service >/dev/null
sudo systemctl daemon-reload
sudo systemctl enable --now fortilogin.service
sudo systemctl status fortilogin.service
```

## GitHub Release Plan

This repo is set up to publish two release artifacts from GitHub Actions when you push a tag like `v0.1.0`:

- `fortilogin-linux-amd64`
- `fortilogin_<version>_amd64.deb`

Tag and push:

```bash
git tag v0.1.0
git push origin main --tags
```

## Build a .deb Locally

```bash
chmod +x scripts/build-deb.sh
./scripts/build-deb.sh v0.1.0
```

Output:

```bash
dist/deb/fortilogin_0.1.0_amd64.deb
```

The `.deb` installs:

- `/usr/bin/fortilogin`
- `/usr/share/fortilogin/fortilogin.service.example`

It does not auto-install a system service because the service file needs the correct Linux username filled in first.

## Notes

- This tool is network-specific and tailored to the observed NIT KKR firewall behavior.
- The logout implementation assumes the fixed logout token remains valid for all users as observed.
- If the firewall changes its login or logout flow, this tool will need updates.
