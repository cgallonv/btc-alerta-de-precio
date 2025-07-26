# 🚨 Bitcoin Price Alert

Una aplicación completa en Go para monitorear el precio de Bitcoin y recibir alertas personalizadas mediante email y notificaciones de escritorio, con **actualización automática en tiempo real**.

## ✨ Características

- 📊 **Monitoreo en tiempo real** del precio de Bitcoin con **triple redundancia**
- 🚨 **Alertas personalizables**: precio por encima/debajo de un valor o cambio porcentual
- 📧 **Notificaciones por email** con diseño HTML atractivo
- 💻 **Notificaciones de escritorio** (macOS, Linux, Windows)
- 🌐 **Interfaz web moderna** con **actualización automática cada 15s**
- 📈 **Historial de precios** con gráficos interactivos
- 🎨 **Animaciones visuales** para cambios de precio y estados de conexión
- 🔄 **Triple redundancia de APIs**: **Binance** (principal) → CoinDesk → CoinGecko
- 🌐 **Indicadores de conexión** en tiempo real
- ⚡ **Sin refrescar página** - Todo se actualiza automáticamente
- 🐳 **Docker ready** para despliegue fácil en la nube
- 💾 **Base de datos SQLite** liviana y confiable

## 🚀 Instalación y Uso

### Prerrequisitos

- Go 1.20 o superior
- Git

### Instalación Local

1. **Clonar el repositorio:**
```bash
git clone <tu-repo>
cd btc-alerta-de-precio
```

2. **Instalar dependencias:**
```bash
go mod tidy
```

3. **Configurar variables de entorno:**
```bash
cp env.example .env
# Editar .env con tus configuraciones
```

4. **Ejecutar la aplicación:**
```bash
go run main.go
```

5. **Abrir en el navegador:**
```
http://localhost:8080
```

**�� ¡La interfaz se actualiza automáticamente cada 15 segundos!** No necesitas refrescar la página.

### Configuración de Email

Para habilitar notificaciones por email, configura las siguientes variables en tu archivo `.env`:

```env
# Gmail (recomendado)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=tu-email@gmail.com
SMTP_PASSWORD=tu-app-password  # No tu contraseña normal
FROM_EMAIL=tu-email@gmail.com
```

**Nota importante:** Para Gmail, necesitas generar una "App Password" en tu cuenta de Google con 2FA habilitado.

### Variables de Entorno

| Variable | Descripción | Valor por defecto |
|----------|-------------|-------------------|
| `PORT` | Puerto del servidor web | `8080` |
| `ENVIRONMENT` | Entorno (development/production) | `development` |
| `DATABASE_PATH` | Ruta de la base de datos SQLite | `alerts.db` |
| `CHECK_INTERVAL` | Intervalo de verificación del backend | `30s` |
| `SMTP_HOST` | Servidor SMTP para emails | `smtp.gmail.com` |
| `SMTP_PORT` | Puerto SMTP | `587` |
| `SMTP_USERNAME` | Usuario SMTP | - |
| `SMTP_PASSWORD` | Contraseña SMTP | - |
| `FROM_EMAIL` | Email remitente | - |

## 🔄 Fuentes de Datos de Bitcoin

La aplicación utiliza **triple redundancia** para máxima confiabilidad:

### **1. 🥇 Binance API (Principal)**
- **URL**: `https://api.binance.com/api/v3/ticker/price?symbol=BTCUSDT`
- **Ventajas**: Datos más actualizados y confiables del exchange líder mundial
- **Rate Limit**: Muy generoso (1200 requests/min)

### **2. 🥈 CoinDesk API (Respaldo Primario)**
- **URL**: `https://api.coindesk.com/v1/bpi/currentprice.json`
- **Ventajas**: API pública gratuita, muy estable
- **Se usa cuando**: Binance falla

### **3. 🥉 CoinGecko API (Respaldo Secundario)**
- **URL**: `https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd`
- **Ventajas**: También proporciona datos históricos
- **Se usa cuando**: Binance y CoinDesk fallan

## 🎨 Nuevas Características Visuales

### **Actualización Automática en Tiempo Real**
- ⚡ **Precio cada 15 segundos** - Sin refrescar página
- 📊 **Estadísticas dinámicas** - Contadores actualizados
- 📈 **Historial cada 2 minutos** - Gráfico siempre actualizado
- 🚨 **Alertas cada 30 segundos** - Lista sincronizada

### **Animaciones Inteligentes**
- 🟢 **Verde con ↗** cuando el precio sube
- 🔴 **Rojo con ↘** cuando el precio baja
- ⭕ **Efecto de pulsación** durante actualizaciones
- 🔄 **Indicador de conexión** en tiempo real

### **Indicadores de Estado**
- ✅ **Conectado** - Verde con icono WiFi
- ⚠️ **Actualizando** - Amarillo con spinner
- ❌ **Error** - Rojo con advertencia
- 📡 **Fuente de datos** mostrada (Binance/CoinDesk/CoinGecko)

## 🐳 Despliegue con Docker

### Build y Run Local

```bash
# Construir imagen
docker build -t btc-price-alert .

# Ejecutar contenedor
docker run -d \
  --name btc-alerts \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -e SMTP_USERNAME=tu-email@gmail.com \
  -e SMTP_PASSWORD=tu-app-password \
  -e FROM_EMAIL=tu-email@gmail.com \
  btc-price-alert
```

### Docker Compose

```bash
# Iniciar servicios
docker-compose up -d

# Ver logs
docker-compose logs -f

# Detener servicios
docker-compose down
```

### Despliegue en la Nube

#### Heroku

```bash
# Instalar Heroku CLI y login
heroku create tu-app-btc-alerts

# Configurar variables de entorno
heroku config:set SMTP_USERNAME=tu-email@gmail.com
heroku config:set SMTP_PASSWORD=tu-app-password
heroku config:set FROM_EMAIL=tu-email@gmail.com

# Desplegar
git push heroku main
```

#### DigitalOcean App Platform

1. Fork este repositorio
2. Conecta tu cuenta de GitHub a DigitalOcean
3. Crea una nueva App desde tu repositorio
4. Configura las variables de entorno
5. ¡Despliega!

## 🎯 Uso de la Aplicación

### 1. Crear Alertas

1. Abre la interfaz web en `http://localhost:8080`
2. Completa el formulario "Nueva Alerta":
   - **Nombre**: Identifica tu alerta
   - **Tipo**: Elige entre:
     - **Precio por encima de**: Se activa cuando BTC supera un precio
     - **Precio por debajo de**: Se activa cuando BTC cae por debajo de un precio
     - **Cambio porcentual**: Se activa con cambios significativos
   - **Email**: Para recibir notificaciones
3. Haz clic en "Crear Alerta"

### 2. Gestionar Alertas

- **👁️ Ver todas las alertas**: En el panel principal (se actualiza cada 30s)
- **🧪 Probar alerta**: Botón azul para enviar notificación de prueba
- **⏸️ Activar/Desactivar**: Botón amarillo/verde
- **✏️ Editar**: Botón azul (próximamente)
- **🗑️ Eliminar**: Botón rojo

### 3. Monitorear Precios en Tiempo Real

- **💰 Precio actual**: Se actualiza automáticamente cada 15 segundos
- **📈 Gráfico**: Historial actualizado cada 2 minutos
- **📊 Estadísticas**: Dashboard con métricas en tiempo real
- **🔔 Indicador de conexión**: Esquina superior derecha

## ⚙️ Configuración de Intervalos

### **Backend (Sistema Crítico)**
- **Intervalo**: 30 segundos (configurable con `CHECK_INTERVAL`)
- **Propósito**: Monitoreo automático, verificación de alertas, guardado en BD
- **Funciona**: Siempre, independiente de usuarios conectados

### **Frontend (Interfaz de Usuario)**
- **Precio**: 15 segundos (actualización visual)
- **Historial**: 2 minutos (gráfico)
- **Alertas**: 30 segundos (sincronización)
- **Funciona**: Solo cuando hay navegador abierto

## 🛠️ API REST

La aplicación expone una API REST completa:

### Endpoints Principales

```bash
# Obtener precio actual (datos de Binance)
GET /api/v1/price

# Historial de precios
GET /api/v1/price/history?limit=100

# Listar alertas
GET /api/v1/alerts

# Crear alerta
POST /api/v1/alerts

# Obtener alerta específica
GET /api/v1/alerts/{id}

# Actualizar alerta
PUT /api/v1/alerts/{id}

# Eliminar alerta
DELETE /api/v1/alerts/{id}

# Activar/desactivar alerta
POST /api/v1/alerts/{id}/toggle

# Probar alerta
POST /api/v1/alerts/{id}/test

# Estadísticas en tiempo real
GET /api/v1/stats

# Health check
GET /api/v1/health
```

### Ejemplo de Uso con cURL

```bash
# Crear una nueva alerta
curl -X POST http://localhost:8080/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '{
    "name": "BTC a la luna 🚀",
    "type": "above",
    "target_price": 100000,
    "email": "tu-email@gmail.com",
    "enable_email": true,
    "enable_desktop": true
  }'

# Obtener precio actual (desde Binance)
curl http://localhost:8080/api/v1/price
```

## 🏗️ Arquitectura

```
btc-alerta-de-precio/
├── main.go                 # Punto de entrada
├── config/                 # Configuración y variables de entorno
├── internal/
│   ├── api/               # Handlers HTTP y rutas
│   ├── alerts/            # Lógica de alertas y monitoreo
│   ├── bitcoin/           # Cliente APIs (Binance→CoinDesk→CoinGecko)
│   ├── notifications/     # Sistema de notificaciones (email + desktop)
│   └── storage/           # Base de datos SQLite y modelos
├── web/
│   ├── templates/         # Templates HTML con efectos visuales
│   └── static/           # CSS, JS con animaciones en tiempo real
├── docker/               # Archivos Docker y docker-compose
└── docs/                # Documentación adicional
```

## 🧪 Testing

```bash
# Ejecutar tests
go test ./...

# Test con coverage
go test -cover ./...

# Test de integración
go test -tags=integration ./...

# Usar Makefile para tareas comunes
make help          # Ver comandos disponibles
make dev           # Ejecutar en desarrollo
make docker-build  # Construir imagen Docker
make test-api      # Probar endpoints de API
```

## 📝 Roadmap v2.0

- [ ] ✏️ Edición de alertas desde la interfaz web
- [ ] 🔔 Notificaciones push para navegadores (Web Push)
- [ ] 📱 Webhooks para integraciones externas
- [ ] 🏦 Soporte para múltiples criptomonedas (ETH, ADA, etc.)
- [ ] 📊 Métricas y análisis técnicos avanzados
- [ ] 🔐 Autenticación y múltiples usuarios
- [ ] 📲 Telegram Bot integration
- [ ] 🎨 Temas personalizables (dark mode)
- [ ] 📈 Alertas de análisis técnico (RSI, MACD, etc.)

## 🤝 Contribuir

1. Fork el proyecto
2. Crea una rama para tu feature (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request

## 📄 Licencia

Este proyecto está bajo la Licencia MIT. Ver el archivo `LICENSE` para más detalles.

## 🆘 Soporte

Si tienes problemas o preguntas:

1. **Revisa la documentación** en este README
2. **Busca en Issues** existentes
3. **Crea un nuevo Issue** con detalles del problema
4. **Consulta los logs** de la aplicación para debugging

## 🙏 Agradecimientos

- **Binance API** - Fuente principal de datos de precio más confiable
- **CoinDesk API** - Datos de precios como respaldo
- **CoinGecko API** - Datos históricos y respaldo secundario
- **Gin Framework** - Framework web rápido para Go
- **GORM** - ORM elegante para Go
- **Bootstrap 5** - Framework CSS moderno
- **Chart.js** - Gráficos interactivos y responsivos

---

**⚠️ Disclaimer**: Esta aplicación es solo para fines informativos. Las fluctuaciones de precios de criptomonedas son altamente volátiles. No constituye asesoramiento financiero.
