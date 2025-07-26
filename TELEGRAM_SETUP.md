# 🤖 Configuración del Bot de Telegram

Esta guía te explica cómo configurar las notificaciones de Telegram para recibir alertas de Bitcoin en tu celular Samsung.

## 📱 **1. Crear Bot de Telegram**

### **Paso 1: Buscar BotFather**
1. Abrir Telegram en tu celular
2. Buscar: `@BotFather` 
3. Iniciar conversación

### **Paso 2: Crear el bot**
```
Enviar: /newbot
```

BotFather te preguntará:
1. **Nombre del bot** (ejemplo: `Bitcoin Price Alert Bot`)
2. **Username del bot** (ejemplo: `btc_price_alert_cgallon_bot`)

### **Paso 3: Obtener TOKEN**
BotFather te dará un mensaje como:
```
Done! Congratulations on your new bot. You have a token to access the HTTP API:

1234567890:ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghi

Keep your token secure and store it safely...
```

**⚠️ IMPORTANTE:** Guarda este TOKEN, lo necesitas para la configuración.

## 🆔 **2. Obtener tu Chat ID**

### **Método 1: Usando la API**
1. Envía **cualquier mensaje** a tu bot (ejemplo: "Hola")
2. Ve a esta URL en tu navegador:
   ```
   https://api.telegram.org/bot<TU_TOKEN>/getUpdates
   ```
   (Reemplaza `<TU_TOKEN>` con el token que te dio BotFather)

3. Busca algo como:
   ```json
   "chat": {
     "id": 123456789,
     "first_name": "Tu Nombre",
     "type": "private"
   }
   ```

4. **Copia el número del `id`** (ejemplo: `123456789`)

### **Método 2: Usando un bot helper**
1. Buscar `@userinfobot` en Telegram
2. Enviar `/start`
3. Te dará tu Chat ID directamente

## ⚙️ **3. Configurar la Aplicación**

### **Editar archivo .env**
```bash
# En Windows
notepad .env

# Agregar estas líneas:
ENABLE_TELEGRAM_NOTIFICATIONS=true
TELEGRAM_BOT_TOKEN=1234567890:ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghi
TELEGRAM_CHAT_ID=123456789
```

**Ejemplo completo del .env:**
```env
# Configuración de notificaciones
ENABLE_DESKTOP_NOTIFICATIONS=true
ENABLE_EMAIL_NOTIFICATIONS=true
ENABLE_TELEGRAM_NOTIFICATIONS=true

# Telegram Bot
TELEGRAM_BOT_TOKEN=1234567890:ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghi
TELEGRAM_CHAT_ID=123456789

# Email (si quieres también email)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=tu-email@gmail.com
SMTP_PASSWORD=tu-contraseña-de-app
FROM_EMAIL=tu-email@gmail.com
```

## 🧪 **4. Probar la Configuración**

### **Ejecutar la aplicación:**
```powershell
# Windows
.\scripts\dev.ps1 dev

# Linux/Mac
./scripts/dev.sh dev
```

### **Buscar en los logs:**
Deberías ver algo como:
```
🧪 Probando todas las notificaciones...
📱 Probando notificación de Telegram...
📱 Enviando notificación de prueba a Telegram...
📱 Notificación de Telegram enviada exitosamente
✅ Telegram enviado correctamente
```

### **En tu celular:**
Deberías recibir un mensaje como:
```
🚨 BITCOIN ALERT 🚨

💰 Precio: $50,000.00
📊 Condición: Precio por encima de $49,000
⏰ Hora: 14:30:25 26/01/2025

🤖 Enviado por BTC Price Alert
```

## 🔧 **5. Configurar Notificaciones del Celular**

### **En Samsung:**
1. **Telegram → Configuración → Notificaciones**
2. Activar **"Notificaciones"**
3. Activar **"Sonido"** y **"Vibración"**
4. En **Android:** Configuración → Apps → Telegram → Notificaciones → Activar todo

### **Para que sean más visibles:**
1. **Telegram → Configuración → Notificaciones → Privados**
2. Activar **"Vista previa"**
3. Configurar **"Importancia"** como **Alta**

## ❌ **6. Solución de Problemas**

### **Error: "telegram bot token o chat ID no configurados"**
- Revisa que el TOKEN y CHAT_ID estén correctos en el `.env`
- Reinicia la aplicación después de cambiar el `.env`

### **Error: "telegram API error: status 400"**
- El TOKEN está incorrecto
- Ve a @BotFather y regenera el token

### **Error: "telegram API error: status 403"**
- Necesitas enviar un mensaje al bot primero
- Busca tu bot por username y envía `/start`

### **No recibo mensajes:**
- Verifica que las notificaciones de Telegram estén activadas
- Revisa que el CHAT_ID sea correcto
- Asegúrate de haber iniciado conversación con el bot

### **Probar manualmente:**
Puedes probar enviando un mensaje directo:
```bash
curl -X POST "https://api.telegram.org/bot<TU_TOKEN>/sendMessage" \
     -H "Content-Type: application/json" \
     -d '{"chat_id": "<TU_CHAT_ID>", "text": "Hola desde la terminal!"}'
```

## 🎯 **7. Ejemplos de Mensajes**

### **Alerta de precio por encima:**
```
🚨 BITCOIN ALERT 🚨

💰 Precio: $52,340.67
📊 Condición: Precio por encima de $52,000
⏰ Hora: 15:45:22 26/01/2025

🤖 Enviado por BTC Price Alert
```

### **Alerta de precio por debajo:**
```
🚨 BITCOIN ALERT 🚨

💰 Precio: $49,850.23
📊 Condición: Precio por debajo de $50,000
⏰ Hora: 09:15:44 26/01/2025

🤖 Enviado por BTC Price Alert
```

## ✅ **¡Listo!**

Ahora recibirás alertas de Bitcoin por:
- 📧 **Email** (si está configurado)
- 🖥️ **Notificaciones de escritorio** (en tu PC)
- 📱 **Telegram** (en tu celular Samsung)

### **Para crear alertas:**
1. Ve a: `http://localhost:8080`
2. Crea tus alertas de precio
3. ¡Espera las notificaciones!

---

**💡 Tip:** Telegram es gratis, instantáneo y funciona perfectamente. ¡Es la mejor opción para notificaciones móviles! 