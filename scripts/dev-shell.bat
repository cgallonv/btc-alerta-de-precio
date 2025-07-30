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

:: Crear un archivo temporal con el script de PowerShell
echo $host.UI.RawUI.WindowTitle = 'BTC Price Alert Dev Environment' > "%TEMP%\dev-env.ps1"
echo Set-Location 'E:\cgallon\proyectos\btc-alerta-de-precio' >> "%TEMP%\dev-env.ps1"
echo Write-Host 'Bitcoin Price Alert Development Environment' -ForegroundColor Green >> "%TEMP%\dev-env.ps1"
echo Write-Host 'Current Directory:' -NoNewline >> "%TEMP%\dev-env.ps1"
echo Write-Host ' E:\cgallon\proyectos\btc-alerta-de-precio' -ForegroundColor Cyan >> "%TEMP%\dev-env.ps1"
echo Write-Host '' >> "%TEMP%\dev-env.ps1"
echo Write-Host 'Available commands:' -ForegroundColor Yellow >> "%TEMP%\dev-env.ps1"
echo Write-Host '  .\scripts\dev.ps1 dev         - Run in development mode' -ForegroundColor White >> "%TEMP%\dev-env.ps1"
echo Write-Host '  .\scripts\dev.ps1 setup       - Configure development environment' -ForegroundColor White >> "%TEMP%\dev-env.ps1"
echo Write-Host '  .\scripts\dev.ps1 clean       - Clean temporary files' -ForegroundColor White >> "%TEMP%\dev-env.ps1"
echo Write-Host '  .\scripts\dev.ps1 build       - Build application' -ForegroundColor White >> "%TEMP%\dev-env.ps1"
echo Write-Host '  .\scripts\dev.ps1 help        - Show all commands' -ForegroundColor White >> "%TEMP%\dev-env.ps1"
echo Write-Host '' >> "%TEMP%\dev-env.ps1"
echo Write-Host 'Go version:' -ForegroundColor Yellow >> "%TEMP%\dev-env.ps1"
echo go version >> "%TEMP%\dev-env.ps1"
echo Write-Host '' >> "%TEMP%\dev-env.ps1"
echo Write-Host 'Environment ready!' -ForegroundColor Green >> "%TEMP%\dev-env.ps1"

:: Ejecutar PowerShell con el script temporal
start powershell -NoExit -File "%TEMP%\dev-env.ps1"

:: Limpiar el archivo temporal (opcional)
del "%TEMP%\dev-env.ps1"