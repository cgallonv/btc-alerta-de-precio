# 🚨 Bitcoin Price Alert

Una aplicación completa en Go para monitorear el precio de Bitcoin y recibir alertas personalizadas mediante email, con **actualización automática en tiempo real**.

## ✨ Características

- 📊 **Monitoreo en tiempo real** del precio de Bitcoin
- 🚨 **Alertas personalizables**: precio por encima/debajo de un valor o **cambio porcentual positivo/negativo**
- 📧 **Notificaciones por email** y Telegram
- 🌐 **Interfaz web moderna** con actualización automática cada 15s
- 📈 **Historial de precios** con gráficos interactivos
- 🔄 **Triple redundancia de APIs**: Binance → CoinDesk → CoinGecko
- 🐳 **Docker ready** para despliegue fácil

## 🎯 Nueva Funcionalidad: Porcentajes Negativos

Ahora puedes crear alertas específicas para **subidas** o **bajadas** de precio usando porcentajes positivos y negativos.

### 🔢 Cómo Funciona

#### ✅ **Porcentajes Positivos** (Solo Subidas)
```
Porcentaje: +3.0%
Triggerea cuando: Precio sube 3% o más
NO triggerea cuando: Precio baja (cualquier cantidad)
```

#### ✅ **Porcentajes Negativos** (Solo Bajadas)  
```
Porcentaje: -3.0%
Triggerea cuando: Precio baja 3% o más  
NO triggerea cuando: Precio sube (cualquier cantidad)
```

### 📋 Ejemplos de Uso

| Tipo de Alerta | Valor | Cuándo se Activa |
|----------------|-------|------------------|
| **Subida** | `+5%` | Solo cuando BTC sube 5% o más |
| **Bajada** | `-3%` | Solo cuando BTC baja 3% o más |
| **Precio fijo** | `$50,000` | Cuando BTC alcanza exactamente $50,000 |

## 🚀 Instalación Rápida

### Prerrequisitos
- Go 1.20+ 
- Git

### 1. Clonar y Configurar
```bash
git clone <tu-repo>
cd btc-alerta-de-precio
cp env.example .env
# Editar .env con tu configuración de email
```

### 2. Ejecutar
```bash
go mod tidy
go run main.go
```

### 3. Usar
```
http://localhost:8080
```

### 📧 Configuración de Email (Gmail)

En tu archivo `.env`:
```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=tu-email@gmail.com
SMTP_PASSWORD=tu-app-password  # App Password de Google, no tu contraseña normal
FROM_EMAIL=tu-email@gmail.com
```

**⚠️ Importante**: Para Gmail necesitas una [App Password](https://support.google.com/accounts/answer/185833) con 2FA habilitado.

## 🎮 Uso

### Crear Alertas
1. Ve a `http://localhost:8080`
2. Completa el formulario:
   - **Nombre**: Identifica tu alerta
   - **Tipo**: 
     - `Precio por encima de`: Alerta cuando BTC > valor
     - `Precio por debajo de`: Alerta cuando BTC < valor
     - `Cambio porcentual`: Alerta por cambios +/- (ej: +5%, -3%)
   - **Email**: Para recibir notificaciones
3. Haz clic en "Crear Alerta"

### Gestionar Alertas
- **Ver todas**: Panel principal (actualizado cada 30s)
- **Probar**: Botón azul (envía notificación de prueba)
- **Activar/Desactivar**: Botón amarillo/verde
- **Eliminar**: Botón rojo

## 🐳 Docker

### Ejecutar con Docker
```bash
# Build y run
docker build -t btc-price-alert .
docker run -d --name btc-alerts -p 8080:8080 btc-price-alert

# Con docker-compose
docker-compose up -d
```

### Variables de Entorno Docker
```bash
docker run -d \
  --name btc-alerts \
  -p 8080:8080 \
  -e SMTP_USERNAME=tu-email@gmail.com \
  -e SMTP_PASSWORD=tu-app-password \
  -e FROM_EMAIL=tu-email@gmail.com \
  btc-price-alert
```

## 🛠️ API REST

### Endpoints Principales
```bash
GET  /api/v1/price              # Precio actual
GET  /api/v1/price/history      # Historial de precios
GET  /api/v1/alerts             # Listar alertas
POST /api/v1/alerts             # Crear alerta
PUT  /api/v1/alerts/{id}        # Actualizar alerta
DELETE /api/v1/alerts/{id}      # Eliminar alerta
POST /api/v1/alerts/{id}/toggle # Activar/desactivar
POST /api/v1/alerts/{id}/test   # Probar alerta
GET  /api/v1/stats              # Estadísticas
GET  /api/v1/health             # Health check
```

### Ejemplo: Crear Alerta
```bash
curl -X POST http://localhost:8080/api/v1/alerts \
  -H "Content-Type: application/json" \
  -d '{
    "name": "BTC a la luna 🚀",
    "type": "above",
    "target_price": 100000,
    "email": "tu-email@gmail.com",
    "enable_email": true
  }'
```

## 📈 Datos Históricos de Binance

### Cargar Datos Históricos

La aplicación incluye un script para cargar datos históricos de precios de Bitcoin desde Binance:

```bash
go run scripts/backfill_historical_data.go
```

### Características

- 📊 Carga datos de los últimos 60 días
- ⏱️ Intervalos de 1 minuto para máxima precisión
- 🔄 Manejo automático de límites de rate de la API
- 💾 Almacenamiento en la base de datos local
- 🔍 Datos completos incluyendo:
  - Precio de apertura/cierre
  - Máximos y mínimos
  - Volumen de trading
  - Número de trades

### Configuración

1. Asegúrate de tener las credenciales de Binance en tu archivo `.env`:
```env
BINANCE_API_KEY=tu_api_key
BINANCE_API_SECRET=tu_api_secret
DATABASE_PATH=btc_market_data_prod.db
```

2. Ejecuta el script:
```bash
go run scripts/backfill_historical_data.go
```

### Detalles Técnicos

- **Intervalo**: 1 minuto
- **Puntos de datos**: ~86,400 (60 días × 24 horas × 60 minutos)
- **Chunks**: Datos obtenidos en bloques de 24 horas
- **Rate Limiting**: Espera automática entre chunks
- **Manejo de errores**: Continúa con el siguiente chunk si hay errores
- **Compatibilidad**: Usa el mismo esquema de base de datos que la aplicación principal

### Uso de los Datos

Los datos históricos se pueden usar para:
- 📊 Análisis de tendencias
- 📈 Gráficos detallados
- 🔍 Backtesting de estrategias
- 📉 Análisis de volatilidad

## 🔄 Actualizaciones Automáticas

- **💰 Precio**: Cada 15 segundos (frontend)
- **📊 Historial**: Cada 2 minutos (gráfico)
- **🚨 Alertas**: Cada 30 segundos (backend)
- **📈 Estadísticas**: En tiempo real

## 📝 Roadmap

### ✅ Completado
- [x] Alertas con porcentajes negativos
- [x] Triple redundancia de APIs
- [x] Interfaz responsive
- [x] Clean Architecture & Testing

### 🚀 Próximamente
- [ ] Editar alertas desde la web
- [ ] Notificaciones Web Push
- [ ] Bot de Telegram
- [ ] Soporte multi-criptomoneda
- [ ] Autenticación de usuarios
- [ ] Alertas de análisis técnico

## 📚 Documentación Adicional

- **[DEVELOPMENT.md](DEVELOPMENT.md)** - Guía completa de desarrollo, arquitectura técnica, y troubleshooting
- **[TELEGRAM_SETUP.md](TELEGRAM_SETUP.md)** - Configuración paso a paso de notificaciones Telegram

## 🆘 Soporte

1. **Revisa la documentación** en [DEVELOPMENT.md](DEVELOPMENT.md)
2. **Busca en Issues** existentes  
3. **Crea un nuevo Issue** con detalles del problema

## 🤝 Contribuir

1. Fork el proyecto
2. Crea una rama (`git checkout -b feature/MiFeature`)
3. Commit tus cambios (`git commit -m 'Add MiFeature'`)
4. Push (`git push origin feature/MiFeature`)
5. Abre un Pull Request

## 📄 Licencia

MIT License - Ver [LICENSE](LICENSE) para detalles.

---

**⚠️ Disclaimer**: Solo para fines informativos. No constituye asesoramiento financiero.

## Notificaciones Soportadas

El sistema soporta múltiples canales de notificación:

1. **Email**: Notificaciones por correo electrónico (requiere configuración SMTP)
2. **Telegram**: Alertas vía bot de Telegram (ver [TELEGRAM_SETUP.md](TELEGRAM_SETUP.md))
3. **Web Push**: Notificaciones en el navegador (Chrome)
4. **WhatsApp**: Mensajes vía WhatsApp Business API (ver [WHATSAPP_SETUP.md](WHATSAPP_SETUP.md))

### Configuración de Notificaciones

Cada tipo de notificación requiere su propia configuración:

- **Email**: Configura las variables SMTP en `.env`
- **Telegram**: Sigue la guía en [TELEGRAM_SETUP.md](TELEGRAM_SETUP.md)
- **Web Push**: Se configura automáticamente al activar en el navegador
- **WhatsApp**: 
  - Guía rápida en [WHATSAPP_SETUP.md](WHATSAPP_SETUP.md)
  - Guía detallada de Meta Developer en [docs/META_APP_SETUP.md](docs/META_APP_SETUP.md)
