#!/bin/bash

# Script de desarrollo para BTC Price Alert
# Uso: ./scripts/dev.sh [comando]

set -e

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Función para logging
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Función para limpiar archivos temporales
clean() {
    log_info "Limpiando archivos temporales..."
    
    # Eliminar binarios
    rm -f btc-price-alert btc-price-alert.exe
    
    # Eliminar archivos de base de datos de desarrollo
    rm -f alerts.db *.sqlite *.sqlite3
    
    # Eliminar logs
    rm -f *.log
    rm -rf logs/
    
    # Eliminar archivos temporales
    rm -rf tmp/ temp/ .tmp/
    
    log_info "Limpieza completada"
}

# Función para setup inicial
setup() {
    log_info "Configurando entorno de desarrollo..."
    
    # Crear directorios necesarios
    mkdir -p logs data tmp
    
    # Copiar .env de ejemplo si no existe
    if [ ! -f .env ]; then
        log_info "Creando archivo .env desde env.example..."
        cp env.example .env
        log_warn "¡Recuerda configurar tus variables de entorno en .env!"
    fi
    
    # Descargar dependencias
    log_info "Descargando dependencias..."
    go mod download
    go mod tidy
    
    log_info "Setup completado"
}

# Función para verificar antes de commit
pre_commit_check() {
    log_info "Verificando archivos antes de commit..."
    
    # Verificar que no hay archivos grandes
    if [ -f btc-price-alert ] || [ -f btc-price-alert.exe ]; then
        log_error "¡Binario compilado detectado! Ejecuta 'clean' primero"
        exit 1
    fi
    
    if [ -f alerts.db ]; then
        log_error "¡Base de datos detectada! Ejecuta 'clean' primero"
        exit 1
    fi
    
    if [ -f .env ]; then
        log_warn "Archivo .env encontrado - asegúrate de que está en .gitignore"
    fi
    
    # Ejecutar tests
    log_info "Ejecutando tests..."
    if go test ./...; then
        log_info "✅ Todos los tests pasaron"
    else
        log_error "❌ Tests fallaron"
        exit 1
    fi
    
    # Verificar formato
    log_info "Verificando formato del código..."
    if [ -n "$(gofmt -l .)" ]; then
        log_error "❌ Código no formateado. Ejecuta 'go fmt ./...'"
        exit 1
    fi
    
    log_info "✅ Verificación pre-commit completada"
}

# Función para desarrollo
dev() {
    log_info "Iniciando en modo desarrollo..."
    setup
    go run main.go
}

# Función para build
build() {
    log_info "Compilando aplicación..."
    clean
    
    # Build para el SO actual
    go build -o btc-price-alert .
    
    log_info "✅ Compilación completada: btc-price-alert"
}

# Función para build multiplataforma
build_all() {
    log_info "Compilando para todas las plataformas..."
    clean
    
    # Crear directorio de builds
    mkdir -p builds/
    
    # Linux
    GOOS=linux GOARCH=amd64 go build -o builds/btc-price-alert-linux-amd64 .
    
    # Windows
    GOOS=windows GOARCH=amd64 go build -o builds/btc-price-alert-windows-amd64.exe .
    
    # macOS
    GOOS=darwin GOARCH=amd64 go build -o builds/btc-price-alert-macos-amd64 .
    GOOS=darwin GOARCH=arm64 go build -o builds/btc-price-alert-macos-arm64 .
    
    log_info "✅ Compilación completada en builds/"
}

# Función para commit seguro
safe_commit() {
    if [ -z "$1" ]; then
        log_error "Uso: $0 safe_commit \"mensaje del commit\""
        exit 1
    fi
    
    log_info "Realizando commit seguro..."
    
    # Limpiar primero
    clean
    
    # Verificar antes de commit
    pre_commit_check
    
    # Add solo archivos tracked y nuevos archivos fuente
    git add *.go go.mod go.sum
    git add internal/ config/ web/
    git add README.md Dockerfile docker-compose.yml Makefile env.example nginx.conf
    
    # Mostrar lo que se va a commitear
    log_info "Archivos a commitear:"
    git diff --cached --name-only
    
    # Commit
    git commit -m "$1"
    
    log_info "✅ Commit realizado de forma segura"
}

# Función para mostrar ayuda
help() {
    echo "Script de desarrollo para BTC Price Alert"
    echo ""
    echo "Comandos disponibles:"
    echo "  setup           - Configurar entorno de desarrollo"
    echo "  dev             - Iniciar en modo desarrollo"
    echo "  clean           - Limpiar archivos temporales"
    echo "  build           - Compilar aplicación"
    echo "  build_all       - Compilar para todas las plataformas"
    echo "  pre_commit      - Verificar antes de commit"
    echo "  safe_commit \"msg\" - Realizar commit seguro"
    echo "  help            - Mostrar esta ayuda"
}

# Main
case "${1:-help}" in
    setup)
        setup
        ;;
    dev)
        dev
        ;;
    clean)
        clean
        ;;
    build)
        build
        ;;
    build_all)
        build_all
        ;;
    pre_commit)
        pre_commit_check
        ;;
    safe_commit)
        safe_commit "$2"
        ;;
    help|*)
        help
        ;;
esac 