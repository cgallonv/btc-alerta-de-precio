# 📈📉 Alertas de Porcentaje Direccionales

## 🎯 Nueva Funcionalidad: Porcentajes Negativos

Ahora puedes crear alertas específicas para **subidas** o **bajadas** de precio usando porcentajes positivos y negativos.

## 🔢 Cómo Funciona

### ✅ **Porcentajes Positivos** (Solo Subidas)
```
Porcentaje: +3.0%
Triggerea cuando: Precio sube 3% o más
NO triggerea cuando: Precio baja (cualquier cantidad)
```

### ✅ **Porcentajes Negativos** (Solo Bajadas)  
```
Porcentaje: -3.0%
Triggerea cuando: Precio baja 3% o más  
NO triggerea cuando: Precio sube (cualquier cantidad)
```

## 🎮 Ejemplos Prácticos

### Escenario: BTC está a $50,000

| Alerta | Nuevo Precio | ¿Triggerea? | Motivo |
|--------|-------------|-------------|---------|
| `+5%` | $52,500 (+5%) | ✅ SÍ | Subió exactamente 5% |
| `+5%` | $47,500 (-5%) | ❌ NO | Bajó 5%, pero esperamos subida |
| `-5%` | $47,500 (-5%) | ✅ SÍ | Bajó exactamente 5% |
| `-5%` | $52,500 (+5%) | ❌ NO | Subió 5%, pero esperamos bajada |
| `+3%` | $51,000 (+2%) | ❌ NO | Solo subió 2%, necesitamos 3%+ |
| `-3%` | $49,000 (-2%) | ❌ NO | Solo bajó 2%, necesitamos 3%+ |

## 🚀 Casos de Uso

### 📈 **Bull Market Alerts (Porcentajes Positivos)**
```
+2%  → Momentum inicial detectado
+5%  → Movimiento fuerte alcista
+10% → Breakout significativo
```

### 📉 **Bear Market Alerts (Porcentajes Negativos)**
```
-2%  → Posible corrección iniciando
-5%  → Caída importante detectada
-10% → Crash o pánico en el mercado
```

### 🛡️ **Risk Management**
```
-3%  → Stop-loss suave
-7%  → Stop-loss agresivo
-15% → Alerta de emergencia
```

## 🎯 Configuración en la Web

1. **Selecciona** "Cambio de porcentaje" como tipo de alerta
2. **Ingresa** el porcentaje:
   - `3` para alertas de subida de 3%+
   - `-3` para alertas de bajada de 3%+
3. **Configura** tus notificaciones (Email, Telegram, Web Push)

## 🧪 Ejemplos de API

### Crear Alerta de Subida (+5%)
```json
POST /api/v1/alerts
{
  "name": "BTC Bull Alert 🚀",
  "type": "change",
  "percentage": 5.0,
  "email": "trader@example.com",
  "enable_email": true
}
```

### Crear Alerta de Bajada (-5%)
```json
POST /api/v1/alerts
{
  "name": "BTC Bear Alert 🐻",
  "type": "change", 
  "percentage": -5.0,
  "email": "trader@example.com",
  "enable_email": true
}
```

## ⚠️ Notas Importantes

- **Porcentaje 0**: No válido, nunca triggerea
- **Alertas One-Shot**: Solo se disparan una vez, luego se desactivan
- **Precisión**: Usa el precio anterior vs actual para calcular el cambio
- **Fuente**: Basado en datos de Binance (más preciso para porcentajes)

## 🔄 Migración Automática

Las alertas existentes mantienen su comportamiento:
- Si tenías `5%` antes, ahora funciona como `+5%` (solo subidas)
- No necesitas reconfigurar nada

¡Ahora tienes control total sobre la dirección de tus alertas! 🎯 