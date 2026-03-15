param(
    [string]$BinaryPath = "$PSScriptRoot\..\..\dist\fortilogin-windows-amd64.exe"
)

$ResolvedBinary = (Resolve-Path $BinaryPath).Path
$StartupDir = [Environment]::GetFolderPath("Startup")
$ShortcutPath = Join-Path $StartupDir "FortiLogin.lnk"

$WshShell = New-Object -ComObject WScript.Shell
$Shortcut = $WshShell.CreateShortcut($ShortcutPath)
$Shortcut.TargetPath = $ResolvedBinary
$Shortcut.Arguments = "daemon"
$Shortcut.WorkingDirectory = Split-Path $ResolvedBinary
$Shortcut.WindowStyle = 7
$Shortcut.Description = "Start FortiLogin in background at user logon"
$Shortcut.Save()

Write-Host "Startup shortcut created at $ShortcutPath"
