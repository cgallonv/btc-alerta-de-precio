# Guía: Crear App en Meta Developer y Configurar WhatsApp Business

Esta guía detallada te ayudará a crear una aplicación en Meta Developer Portal y configurar WhatsApp Business API paso a paso.

## Paso 1: Acceder a Meta Developer Portal

1. Abre tu navegador y ve a [Meta Developer Portal](https://developers.facebook.com/)
2. Haz clic en "Log In" en la esquina superior derecha
3. Inicia sesión con tu cuenta de Facebook
   - Si no tienes una cuenta, deberás crear una
   - Se recomienda usar una cuenta empresarial

## Paso 2: Crear una Nueva Aplicación

1. En la página principal, haz clic en "My Apps" en la esquina superior derecha
2. Haz clic en el botón "Create App"
3. Selecciona el tipo de aplicación:
   - Elige "Business" como tipo de app
   - Esta opción es la más adecuada para WhatsApp Business API
4. Haz clic en "Next"

## Paso 3: Configuración Básica de la App

1. Completa el formulario de creación:
   - **App Name**: Nombre de tu aplicación (ej: "BTC Price Alert")
   - **App Contact Email**: Tu email de contacto
   - **Business Account**: Selecciona tu cuenta de negocio
     - Si no tienes una, deberás crear una en [Meta Business Suite](https://business.facebook.com)

2. Haz clic en "Create App"

## Paso 4: Agregar WhatsApp a tu App

1. En el dashboard de tu app, busca la sección "Add Products"
2. Encuentra "WhatsApp" en la lista de productos
3. Haz clic en el botón "Set Up" junto a WhatsApp
4. Espera a que se complete la configuración inicial

## Paso 5: Configurar WhatsApp Business

1. En el menú lateral, ve a "WhatsApp" → "Getting Started"
2. En la sección "From Phone Number":
   - Haz clic en "Add Phone Number"
   - Puedes usar un número de prueba para desarrollo
   - Para producción, necesitarás verificar un número real

3. Para número de prueba:
   - Selecciona "Test Number"
   - Se te asignará un número temporal
   - Guarda el "Phone Number ID" que se te proporciona

4. Para número real:
   - Selecciona "Register Phone Number"
   - Sigue el proceso de verificación
   - Necesitarás acceso al teléfono para códigos SMS/llamada

## Paso 6: Obtener Credenciales de API

1. Ve a "WhatsApp" → "API Setup"
2. Aquí encontrarás:
   - **Temporary Access Token**: Para pruebas
   - **Permanent Access Token**: Para producción
   - **Phone Number ID**: ID de tu número
   - **WhatsApp Business Account ID**: ID de tu cuenta

3. Guarda estos valores:
   ```env
   WHATSAPP_ACCESS_TOKEN=tu_token_aquí
   WHATSAPP_PHONE_NUMBER_ID=tu_phone_number_id
   WHATSAPP_BUSINESS_ACCOUNT_ID=tu_business_account_id
   ```

## Paso 7: Crear Plantillas de Mensaje

1. Ve a "WhatsApp" → "Message Templates"
2. Haz clic en "Create Template"
3. Configura la plantilla en español:
   - **Name**: btc_alert_es
   - **Category**: Alert Update
   - **Language**: Spanish
   - **Template Type**: Text Message
   - **Message**: 
     ```
     🚨 Alerta Bitcoin: {{1}}

     💰 Precio: {{2}}
     📊 Condición: {{3}}
     ⏰ Hora: {{4}}

     🤖 Enviado por BTC Price Alert
     ```
   - **Sample Values**:
     1. "Alerta Precio Alto"
     2. "$50,000.00"
     3. "Precio por encima de $50,000"
     4. "15:30:00 25/12/2023"

4. Repite para la plantilla en inglés:
   - **Name**: btc_alert_en
   - Misma estructura pero en inglés

## Paso 8: Verificación y Pruebas

1. **Verificar Configuración**:
   - Asegúrate de tener todos los valores necesarios
   - Verifica que las plantillas estén aprobadas

2. **Probar con Postman**:
   ```http
   POST https://graph.facebook.com/v17.0/{{Phone-Number-ID}}/messages
   Headers:
   - Authorization: Bearer {{Access-Token}}
   - Content-Type: application/json

   Body:
   {
     "messaging_product": "whatsapp",
     "to": "tu_numero_whatsapp",
     "type": "template",
     "template": {
       "name": "btc_alert_es",
       "language": {
         "code": "es"
       },
       "components": [
         {
           "type": "body",
           "parameters": [
             {"type": "text", "text": "Test Alert"},
             {"type": "text", "text": "$50,000.00"},
             {"type": "text", "text": "Test Condition"},
             {"type": "text", "text": "15:30:00"}
           ]
         }
       ]
     }
   }
   ```

## Paso 9: Integración con la Aplicación

1. Actualiza tu archivo `.env`:
   ```env
   ENABLE_WHATSAPP_NOTIFICATIONS=true
   WHATSAPP_ACCESS_TOKEN=tu_token
   WHATSAPP_PHONE_NUMBER_ID=tu_phone_number_id
   WHATSAPP_BUSINESS_ACCOUNT_ID=tu_business_account_id
   WHATSAPP_TEMPLATE_NAME_ES=btc_alert_es
   WHATSAPP_TEMPLATE_NAME_EN=btc_alert_en
   ```

2. Reinicia tu aplicación:
   ```bash
   go run main.go
   ```

## Solución de Problemas

### Error: Invalid Phone Number
- Asegúrate de usar el formato internacional correcto
- Ejemplo: 573001234567 (57 = Colombia)
- No incluyas el símbolo '+' ni espacios

### Error: Template Not Approved
- Las plantillas pueden tardar en ser aprobadas
- Asegúrate de seguir las políticas de WhatsApp
- No incluyas contenido promocional o sensible

### Error: Invalid Token
- Verifica que el token no haya expirado
- Regenera el token si es necesario
- Asegúrate de usar el token correcto (permanente vs temporal)

## Recursos Adicionales

- [Documentación Oficial de WhatsApp Business API](https://developers.facebook.com/docs/whatsapp)
- [Guía de Plantillas](https://developers.facebook.com/docs/whatsapp/message-templates/guidelines)
- [Políticas de WhatsApp Business](https://developers.facebook.com/docs/whatsapp/policies)
- [Foro de Desarrolladores](https://developers.facebook.com/community)

## Notas Importantes

1. **Ambiente de Pruebas**:
   - Los números de prueba son gratuitos
   - Perfectos para desarrollo y testing
   - Limitados en funcionalidad

2. **Ambiente de Producción**:
   - Requiere verificación de negocio
   - Puede tener costos asociados
   - Mayor límite de mensajes

3. **Mejores Prácticas**:
   - Mantén tus tokens seguros
   - No compartas credenciales
   - Usa variables de entorno
   - Implementa manejo de errores
   - Monitorea el uso y los costos 