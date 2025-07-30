@echo off
echo Opening Bitcoin Price Alert development environment...
echo Working directory: E:\cgallon\proyectos\btc-alerta-de-precio

:: Verificar si el directorio existe
if not exist "E:\cgallon\proyectos\btc-alerta-de-precio" (
    echo Error: Project directory not found!
    echo Expected: E:\cgallon\proyectos\btc-alerta-de-precio
    pause
    exit /b 1
)

:: Abrir PowerShell en la ruta del proyecto con un título personalizado y comandos disponibles
start powershell -NoExit -Command ^
    "$host.UI.RawUI.WindowTitle = 'BTC Price Alert Dev Environment'; ^
    Set-Location 'E:\cgallon\proyectos\btc-alerta-de-precio'; ^
    Write-Host 'Bitcoin Price Alert Development Environment' -ForegroundColor Green; ^
    Write-Host 'Current Directory:' -NoNewline; ^
    Write-Host ' E:\cgallon\proyectos\btc-alerta-de-precio' -ForegroundColor Cyan; ^
    Write-Host ''; ^
    Write-Host 'Available commands:' -ForegroundColor Yellow; ^
    Write-Host '  .\scripts\dev.ps1 dev         - Run in development mode' -ForegroundColor White; ^
    Write-Host '  .\scripts\dev.ps1 setup       - Configure development environment' -ForegroundColor White; ^
    Write-Host '  .\scripts\dev.ps1 clean       - Clean temporary files' -ForegroundColor White; ^
    Write-Host '  .\scripts\dev.ps1 build       - Build application' -ForegroundColor White; ^
    Write-Host '  .\scripts\dev.ps1 help        - Show all commands' -ForegroundColor White; ^
    Write-Host ''; ^
    Write-Host 'Go version:' -ForegroundColor Yellow; ^
    go version; ^
    Write-Host ''; ^
    Write-Host 'Environment ready!' -ForegroundColor Green;"