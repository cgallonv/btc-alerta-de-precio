# Script de desarrollo para BTC Price Alert en Windows
# Uso: .\scripts\dev.ps1 [comando]

param(
    [string]$Command = "help"
)

# Colores para output
$colors = @{
    Red = [ConsoleColor]::Red
    Green = [ConsoleColor]::Green
    Yellow = [ConsoleColor]::Yellow
    White = [ConsoleColor]::White
}

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor $colors.Green
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor $colors.Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor $colors.Red
}

# Función para limpiar archivos temporales
function Clean {
    Write-Info "Limpiando archivos temporales..."
    
    # Eliminar binarios
    if (Test-Path "btc-price-alert.exe") { Remove-Item "btc-price-alert.exe" -Force }
    if (Test-Path "btc-price-alert") { Remove-Item "btc-price-alert" -Force }
    
    # Eliminar archivos de base de datos
    Get-ChildItem -Path "." -Filter "*.db" | Remove-Item -Force
    Get-ChildItem -Path "." -Filter "*.sqlite*" | Remove-Item -Force
    
    # Eliminar logs
    Get-ChildItem -Path "." -Filter "*.log" | Remove-Item -Force
    if (Test-Path "logs") { Remove-Item "logs" -Recurse -Force }
    
    # Eliminar archivos temporales
    if (Test-Path "tmp") { Remove-Item "tmp" -Recurse -Force }
    if (Test-Path "temp") { Remove-Item "temp" -Recurse -Force }
    if (Test-Path ".tmp") { Remove-Item ".tmp" -Recurse -Force }
    
    Write-Info "Limpieza completada"
}

# Función para setup inicial
function Setup {
    Write-Info "Configurando entorno de desarrollo..."
    
    # Crear directorios necesarios
    @("logs", "data", "tmp") | ForEach-Object {
        if (!(Test-Path $_)) {
            New-Item -ItemType Directory -Path $_ | Out-Null
        }
    }
    
    # Copiar .env de ejemplo si no existe
    if (!(Test-Path ".env")) {
        Write-Info "Creando archivo .env desde env.example..."
        Copy-Item "env.example" ".env"
        Write-Warn "¡Recuerda configurar tus variables de entorno en .env!"
    }
    
    # Descargar dependencias
    Write-Info "Descargando dependencias..."
    go mod download
    go mod tidy
    
    Write-Info "Setup completado"
}

# Función para verificar antes de commit
function PreCommitCheck {
    Write-Info "Verificando archivos antes de commit..."
    
    # Verificar que no hay archivos grandes
    if ((Test-Path "btc-price-alert.exe") -or (Test-Path "btc-price-alert")) {
        Write-Error "¡Binario compilado detectado! Ejecuta 'Clean' primero"
        exit 1
    }
    
    if (Test-Path "alerts.db") {
        Write-Error "¡Base de datos detectada! Ejecuta 'Clean' primero"
        exit 1
    }
    
    if (Test-Path ".env") {
        Write-Warn "Archivo .env encontrado - asegúrate de que está en .gitignore"
    }
    
    # Ejecutar tests
    Write-Info "Ejecutando tests..."
    $testResult = go test ./...
    if ($LASTEXITCODE -eq 0) {
        Write-Info "✅ Todos los tests pasaron"
    } else {
        Write-Error "❌ Tests fallaron"
        exit 1
    }
    
    # Verificar formato
    Write-Info "Verificando formato del código..."
    $formatResult = go fmt ./...
    if ($formatResult) {
        Write-Error "❌ Código no formateado. Se ha formateado automáticamente."
    }
    
    Write-Info "✅ Verificación pre-commit completada"
}

# Función para desarrollo
function Dev {
    Write-Info "Iniciando en modo desarrollo..."
    Setup
    Set-Location -Path "$PSScriptRoot/.."  # Ensure we are in the project root
    go run main.go
}

# Función para build
function Build {
    Write-Info "Compilando aplicación..."
    Clean
    
    # Build para Windows
    go build -o btc-price-alert.exe .
    
    Write-Info "✅ Compilación completada: btc-price-alert.exe"
}

# Función para build multiplataforma
function BuildAll {
    Write-Info "Compilando para todas las plataformas..."
    Clean
    
    # Crear directorio de builds
    if (!(Test-Path "builds")) {
        New-Item -ItemType Directory -Path "builds" | Out-Null
    }
    
    # Linux
    $env:GOOS = "linux"; $env:GOARCH = "amd64"
    go build -o "builds/btc-price-alert-linux-amd64" .
    
    # Windows
    $env:GOOS = "windows"; $env:GOARCH = "amd64"
    go build -o "builds/btc-price-alert-windows-amd64.exe" .
    
    # macOS
    $env:GOOS = "darwin"; $env:GOARCH = "amd64"
    go build -o "builds/btc-price-alert-macos-amd64" .
    
    $env:GOOS = "darwin"; $env:GOARCH = "arm64"
    go build -o "builds/btc-price-alert-macos-arm64" .
    
    # Resetear variables de entorno
    Remove-Item Env:\GOOS
    Remove-Item Env:\GOARCH
    
    Write-Info "✅ Compilación completada en builds/"
}

# Función para commit seguro
function SafeCommit {
    param([string]$Message)
    
    if ([string]::IsNullOrWhiteSpace($Message)) {
        Write-Error "Uso: .\scripts\dev.ps1 safe_commit 'mensaje del commit'"
        exit 1
    }
    
    Write-Info "Realizando commit seguro..."
    
    # Limpiar primero
    Clean
    
    # Verificar antes de commit
    PreCommitCheck
    
    # Add solo archivos tracked y nuevos archivos fuente
    git add *.go go.mod go.sum
    git add internal/ config/ web/
    git add README.md Dockerfile docker-compose.yml Makefile env.example nginx.conf
    
    # Mostrar lo que se va a commitear
    Write-Info "Archivos a commitear:"
    git diff --cached --name-only
    
    # Commit
    git commit -m $Message
    
    Write-Info "✅ Commit realizado de forma segura"
}

# Función para mostrar ayuda
function Help {
    Write-Host "Script de desarrollo para BTC Price Alert (Windows)"
    Write-Host ""
    Write-Host "Comandos disponibles:"
    Write-Host "  setup           - Configurar entorno de desarrollo"
    Write-Host "  dev             - Iniciar en modo desarrollo"
    Write-Host "  clean           - Limpiar archivos temporales"
    Write-Host "  build           - Compilar aplicación"
    Write-Host "  build_all       - Compilar para todas las plataformas"
    Write-Host "  pre_commit      - Verificar antes de commit"
    Write-Host "  safe_commit 'msg' - Realizar commit seguro"
    Write-Host "  help            - Mostrar esta ayuda"
    Write-Host ""
    Write-Host "Ejemplos:"
    Write-Host "  .\scripts\dev.ps1 dev"
    Write-Host "  .\scripts\dev.ps1 safe_commit 'Agregar nueva funcionalidad'"
}

# Main
switch ($Command.ToLower()) {
    "setup" { Setup }
    "dev" { Dev }
    "clean" { Clean }
    "build" { Build }
    "build_all" { BuildAll }
    "pre_commit" { PreCommitCheck }
    "safe_commit" { 
        if ($args.Count -gt 0) {
            SafeCommit $args[0]
        } else {
            Write-Error "Falta el mensaje del commit"
            Help
        }
    }
    default { Help }
} 