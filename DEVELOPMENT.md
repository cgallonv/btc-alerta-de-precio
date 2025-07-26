# 🛠️ Guía de Desarrollo - BTC Price Alert

Esta guía explica cómo trabajar con el proyecto usando la nueva arquitectura de desarrollo que **previene automáticamente** problemas con archivos grandes en Git.

## 🚀 Configuración Inicial

### **Primera vez (Mac/Linux):**
```bash
# 1. Clonar el repositorio
git clone https://github.com/cgallonv/btc-alerta-de-precio.git
cd btc-alerta-de-precio

# 2. Ejecutar instalación automática
./scripts/install.sh
```

### **Primera vez (Windows):**
```powershell
# 1. Clonar el repositorio
git clone https://github.com/cgallonv/btc-alerta-de-precio.git
cd btc-alerta-de-precio

# 2. Configurar entorno
.\scripts\dev.ps1 setup
```

## 🔧 Comandos de Desarrollo

### **🐧 Linux/Mac:**

```bash
# Desarrollo diario
./scripts/dev.sh dev                    # Ejecutar en modo desarrollo
./scripts/dev.sh clean                  # Limpiar archivos temporales
./scripts/dev.sh build                  # Compilar aplicación

# Commits seguros
./scripts/dev.sh pre_commit             # Verificar antes de commit
./scripts/dev.sh safe_commit "mensaje"  # Commit automático seguro

# Build multiplataforma
./scripts/dev.sh build_all              # Compilar para todos los OS
```

### **🪟 Windows:**

```powershell
# Desarrollo diario
.\scripts\dev.ps1 dev                    # Ejecutar en modo desarrollo
.\scripts\dev.ps1 clean                  # Limpiar archivos temporales
.\scripts\dev.ps1 build                  # Compilar aplicación

# Commits seguros
.\scripts\dev.ps1 pre_commit             # Verificar antes de commit
.\scripts\dev.ps1 safe_commit "mensaje"  # Commit automático seguro

# Build multiplataforma
.\scripts\dev.ps1 build_all              # Compilar para todos los OS
```

## 🛡️ Protecciones Automáticas

### **📋 Git Hooks (Pre-commit)**
Se ejecuta **automáticamente** antes de cada commit:

- ✅ **Bloquea archivos peligrosos:** `.env`, `alerts.db`, `btc-price-alert`
- ✅ **Detecta archivos grandes:** >1MB
- ✅ **Verifica formato:** código Go bien formateado
- ✅ **Valida dependencias:** `go.mod` íntegro

### **🚫 Archivos Automáticamente Ignorados:**
```
# Binarios y ejecutables
btc-price-alert*
*.exe, *.dll, *.so

# Bases de datos
alerts.db, *.sqlite*

# Configuración local
.env*

# Logs y temporales
*.log, logs/, tmp/

# Builds
builds/
```

## 🔄 Workflow Recomendado

### **1. Desarrollo Normal:**
```bash
# Iniciar desarrollo
./scripts/dev.sh dev

# Tu código aquí...

# Antes de commit (automático con hooks)
./scripts/dev.sh clean
```

### **2. Commit Seguro:**
```bash
# Opción A: Commit manual (con protección automática)
git add .
git commit -m "feat: nueva funcionalidad"

# Opción B: Commit completamente seguro
./scripts/dev.sh safe_commit "feat: nueva funcionalidad"
```

### **3. Push sin Problemas:**
```bash
git push origin main   # ¡Sin archivos grandes!
```

## 📁 Estructura del Proyecto

```
btc-alerta-de-precio/
├── scripts/              # Scripts de automatización
│   ├── dev.sh           # Script para Linux/Mac
│   ├── dev.ps1          # Script para Windows
│   └── install.sh       # Instalación automática
├── .githooks/           # Git hooks automáticos
│   └── pre-commit       # Verificaciones pre-commit
├── builds/              # Builds multiplataforma (ignorado)
├── logs/                # Logs de la aplicación (ignorado)
├── data/                # Data files (ignorado)
├── tmp/                 # Archivos temporales (ignorado)
└── ...                  # Resto del proyecto
```

## 🐛 Solución de Problemas

### **❌ "Binary was compiled with CGO_ENABLED=0"**
```bash
# Ya solucionado: Usamos driver SQLite puro Go
go mod tidy  # Debe descargar github.com/glebarez/sqlite
```

### **❌ "Pre-commit hook failed"**
```bash
./scripts/dev.sh clean     # Limpiar archivos problemáticos
./scripts/dev.sh pre_commit # Ver qué está fallando
```

### **❌ "RPC failed; HTTP 400" en Git Push**
```bash
# Ya no debería pasar, pero si ocurre:
./scripts/dev.sh clean
git status  # Verificar que no hay archivos grandes
```

### **❌ "Permission denied" en Linux/Mac**
```bash
chmod +x scripts/*.sh
chmod +x .githooks/pre-commit
```

## 🆘 Comandos de Emergencia

### **🧹 Limpieza Total:**
```bash
# Linux/Mac
./scripts/dev.sh clean
rm -rf builds/ logs/ data/ tmp/

# Windows
.\scripts\dev.ps1 clean
Remove-Item builds, logs, data, tmp -Recurse -Force
```

### **🔄 Reset de Git Hooks:**
```bash
git config core.hooksPath .githooks
chmod +x .githooks/pre-commit
```

### **📦 Reset de Dependencias:**
```bash
go clean -modcache
go mod download
go mod tidy
```

## 🎯 Beneficios de Esta Arquitectura

✅ **Sin más errores de push:** Archivos grandes bloqueados automáticamente  
✅ **Desarrollo más rápido:** Scripts automatizan tareas repetitivas  
✅ **Multiplataforma:** Funciona igual en Windows, Mac y Linux  
✅ **Detección temprana:** Problemas detectados antes del commit  
✅ **Limpieza automática:** No más archivos basura en el repo  
✅ **Builds organizados:** Compilaciones en carpeta separada  

## 📞 Ayuda

Si tienes problemas:
1. Ejecuta `./scripts/dev.sh help` (o `.\scripts\dev.ps1 help`)
2. Revisa los logs en `logs/`
3. Verifica que tengas Go 1.20+ y Git instalados 