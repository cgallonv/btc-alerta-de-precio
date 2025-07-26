#!/bin/bash

# Script de instalación para BTC Price Alert
# Configura el entorno de desarrollo automáticamente

set -e

echo "🚀 Configurando entorno de desarrollo para BTC Price Alert..."

# Verificar dependencias
echo "📋 Verificando dependencias..."

# Verificar Go
if ! command -v go &> /dev/null; then
    echo "❌ Go no está instalado. Por favor instala Go 1.20+ primero."
    exit 1
fi

echo "✅ Go $(go version | cut -d' ' -f3) encontrado"

# Verificar Git
if ! command -v git &> /dev/null; then
    echo "❌ Git no está instalado."
    exit 1
fi

echo "✅ Git encontrado"

# Hacer scripts ejecutables
echo "🔧 Configurando permisos de scripts..."
chmod +x scripts/dev.sh
chmod +x scripts/install.sh
chmod +x .githooks/pre-commit

# Configurar Git hooks
echo "🎣 Configurando Git hooks..."
git config core.hooksPath .githooks

# Crear directorios necesarios
echo "📁 Creando directorios..."
mkdir -p logs data tmp builds .githooks

# Configurar archivo .env si no existe
if [ ! -f .env ]; then
    echo "⚙️ Creando archivo .env desde template..."
    cp env.example .env
    echo "📝 Archivo .env creado. ¡Recuerda configurar tus variables!"
fi

# Descargar dependencias
echo "📦 Descargando dependencias Go..."
go mod download
go mod tidy

# Verificar que todo funciona
echo "🧪 Ejecutando tests..."
if go test ./...; then
    echo "✅ Tests pasaron correctamente"
else
    echo "⚠️ Algunos tests fallaron, pero la instalación continúa"
fi

# Información final
echo ""
echo "🎉 ¡Instalación completada!"
echo ""
echo "📖 Comandos disponibles:"
echo "  ./scripts/dev.sh dev          - Ejecutar en modo desarrollo"
echo "  ./scripts/dev.sh clean        - Limpiar archivos temporales"
echo "  ./scripts/dev.sh build        - Compilar aplicación"
echo "  ./scripts/dev.sh safe_commit  - Hacer commit seguro"
echo ""
echo "🔧 Git hooks configurados:"
echo "  Pre-commit: Verifica archivos antes de cada commit"
echo ""
echo "📝 Próximos pasos:"
echo "  1. Configura tu archivo .env"
echo "  2. Ejecuta: ./scripts/dev.sh dev"
echo "" 