# Configuración de WhatsApp Business API

Este documento explica cómo configurar las notificaciones de WhatsApp usando la API de WhatsApp Business Cloud.

## Índice
1. [Requisitos Previos](#requisitos-previos)
2. [Crear una Cuenta Meta Business](#crear-una-cuenta-meta-business)
3. [Configurar WhatsApp Business API](#configurar-whatsapp-business-api)
4. [Crear Plantillas de Mensaje](#crear-plantillas-de-mensaje)
5. [Configuración en la Aplicación](#configuración-en-la-aplicación)
6. [Pruebas y Verificación](#pruebas-y-verificación)
7. [Solución de Problemas](#solución-de-problemas)

## Requisitos Previos

1. Una cuenta de Facebook Business
2. Un número de teléfono empresarial (no puede ser un número personal)
3. Documentos de verificación de negocio
4. Acceso a Meta Developer Portal

## Crear una Cuenta Meta Business

1. Visita [Meta Business Suite](https://business.facebook.com/)
2. Haz clic en "Crear cuenta"
3. Sigue los pasos de verificación de negocio
4. Guarda tu Business Account ID

## Configurar WhatsApp Business API

### 1. Crear una App en Meta Developer Portal

1. Ve a [Meta Developer Portal](https://developers.facebook.com/)
2. Haz clic en "My Apps" → "Create App"
3. Selecciona "Business" como tipo de app
4. Completa la información básica
5. En "Add Products", selecciona "WhatsApp"

### 2. Configurar el Número de Teléfono

1. En el panel de WhatsApp, ve a "Getting Started"
2. Haz clic en "Add phone number"
3. Sigue el proceso de verificación
4. Guarda el Phone Number ID proporcionado

### 3. Obtener Access Token

1. Ve a "WhatsApp" → "API Setup"
2. Genera un Permanent Access Token
3. Guarda este token de forma segura

## Crear Plantillas de Mensaje

### Plantilla en Español (btc_alert_es)

1. Ve a "Message Templates"
2. Clic en "Create Template"
3. Configura:
   - **Nombre**: btc_alert_es
   - **Categoría**: Alert Update
   - **Idioma**: Spanish
   - **Mensaje**:
     ```
     🚨 Alerta Bitcoin: {{1}}

     💰 Precio: {{2}}
     📊 Condición: {{3}}
     ⏰ Hora: {{4}}

     🤖 Enviado por BTC Price Alert
     ```
   - **Variables**:
     1. Nombre de la alerta
     2. Precio actual
     3. Descripción de la condición
     4. Hora del evento

### Plantilla en Inglés (btc_alert_en)

1. Crea otra plantilla similar:
   - **Nombre**: btc_alert_en
   - **Mensaje**:
     ```
     🚨 Bitcoin Alert: {{1}}

     💰 Price: {{2}}
     📊 Condition: {{3}}
     ⏰ Time: {{4}}

     🤖 Sent by BTC Price Alert
     ```

## Configuración en la Aplicación

### 1. Variables de Entorno

Añade estas variables a tu archivo `.env`:

```env
# WhatsApp Business API Configuration
ENABLE_WHATSAPP_NOTIFICATIONS=true
WHATSAPP_ACCESS_TOKEN=your_access_token_here
WHATSAPP_PHONE_NUMBER_ID=your_phone_number_id
WHATSAPP_BUSINESS_ACCOUNT_ID=your_business_account_id
WHATSAPP_TEMPLATE_NAME_ES=btc_alert_es
WHATSAPP_TEMPLATE_NAME_EN=btc_alert_en
```

### 2. Valores Requeridos

- **WHATSAPP_ACCESS_TOKEN**: Token permanente generado en el paso anterior
- **WHATSAPP_PHONE_NUMBER_ID**: ID del número de teléfono de WhatsApp Business
- **WHATSAPP_BUSINESS_ACCOUNT_ID**: ID de tu cuenta de Business

## Pruebas y Verificación

1. **Verificar Configuración**:
   ```bash
   # Verifica que las variables de entorno estén cargadas
   go run main.go
   ```

2. **Probar Notificación**:
   - Crea una nueva alerta con WhatsApp habilitado
   - Usa el botón "Probar" para enviar una notificación de prueba
   - Verifica que la notificación llegue al número especificado

3. **Monitorear Logs**:
   - Revisa los logs del servidor para errores
   - Verifica las respuestas de la API de WhatsApp

## Solución de Problemas

### Errores Comunes

1. **Error 400: Invalid WhatsApp Number**
   - Asegúrate de usar el formato internacional correcto
   - Ejemplo: 573001234567 (57 = Colombia)

2. **Error 401: Invalid Token**
   - Verifica que el token de acceso sea válido
   - Regenera el token si es necesario

3. **Error 404: Template Not Found**
   - Verifica que los nombres de las plantillas coincidan
   - Asegúrate de que las plantillas estén aprobadas

4. **Error 429: Too Many Requests**
   - Respeta los límites de la API
   - Implementa rate limiting si es necesario

### Verificación de Estado

Para verificar el estado de una plantilla:

1. Ve a Meta Developer Portal
2. Navega a WhatsApp → Message Templates
3. Revisa el estado de aprobación

### Límites y Restricciones

- **Números de Teléfono**: Solo números verificados pueden recibir mensajes
- **Plantillas**: Deben ser aprobadas antes de su uso
- **Rate Limits**: Consulta los límites actuales en la documentación
- **Ventana de Mensajes**: 24 horas para respuestas fuera de plantillas

## Referencias

- [Documentación Oficial de WhatsApp Business API](https://developers.facebook.com/docs/whatsapp)
- [Guía de Plantillas de Mensaje](https://developers.facebook.com/docs/whatsapp/message-templates)
- [Políticas de WhatsApp Business](https://developers.facebook.com/docs/whatsapp/policies)
- [Límites y Restricciones](https://developers.facebook.com/docs/whatsapp/limits)

## Soporte

Para problemas técnicos:
1. Revisa los logs del servidor
2. Consulta el [Meta for Developers Forum](https://developers.facebook.com/community)
3. Contacta al [Soporte de Meta Business](https://business.facebook.com/help) 