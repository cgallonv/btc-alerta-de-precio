// Variables globales
let priceChart;
let currentPrice = 0;
let webPushSupported = false;
let webPushSubscription = null;
let vapidPublicKey = null;

// Inicializar la aplicación
document.addEventListener('DOMContentLoaded', function() {
    initializeApp();
    setupEventListeners();
    startPriceUpdates();
});

function initializeApp() {
    // Inicializar elemento de cambio de precio
    const priceChangeElement = document.getElementById('priceChange');
    if (priceChangeElement) {
        priceChangeElement.textContent = '0.00%';
        priceChangeElement.className = 'price-change neutral';
    }
    
    // Inicializar Web Push notifications
    initializeWebPush();
    
    loadCurrentPrice();
    loadAlerts();
    loadPriceHistory();
}

function setupEventListeners() {
    // Form de alertas
    document.getElementById('alertForm').addEventListener('submit', createAlert);
    
    // Cambio de tipo de alerta
    document.getElementById('alertType').addEventListener('change', function() {
        toggleAlertFields(this.value);
    });
}

function toggleAlertFields(alertType) {
    const priceGroup = document.getElementById('priceGroup');
    const percentageGroup = document.getElementById('percentageGroup');
    
    if (alertType === 'change') {
        priceGroup.style.display = 'none';
        percentageGroup.style.display = 'block';
        document.getElementById('targetPrice').required = false;
        document.getElementById('percentage').required = true;
    } else {
        priceGroup.style.display = 'block';
        percentageGroup.style.display = 'none';
        document.getElementById('targetPrice').required = true;
        document.getElementById('percentage').required = false;
    }
}

// Funciones de API
async function apiCall(endpoint, options = {}) {
    const connectionIndicator = document.getElementById('connectionIndicator');
    
    try {
        // Mostrar estado de carga
        if (connectionIndicator) {
            connectionIndicator.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Actualizando...';
            connectionIndicator.className = 'badge bg-warning';
        }
        
        const response = await fetch(`/api/v1${endpoint}`, {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        });
        
        const data = await response.json();
        
        if (!data.success) {
            throw new Error(data.error || 'API Error');
        }
        
        // Mostrar estado de éxito
        if (connectionIndicator) {
            connectionIndicator.innerHTML = '<i class="fas fa-wifi"></i> Conectado';
            connectionIndicator.className = 'badge bg-success online';
        }
        
        return data;
    } catch (error) {
        console.error('API Call Error:', error);
        
        // Mostrar estado de error
        if (connectionIndicator) {
            connectionIndicator.innerHTML = '<i class="fas fa-exclamation-triangle"></i> Error';
            connectionIndicator.className = 'badge bg-danger offline';
            
            // Restaurar estado después de 5 segundos
            setTimeout(() => {
                connectionIndicator.innerHTML = '<i class="fas fa-wifi"></i> Conectado';
                connectionIndicator.className = 'badge bg-success online';
            }, 5000);
        }
        
        // Solo mostrar notificación para errores críticos, no para actualizaciones automáticas
        if (!endpoint.includes('/price') && !endpoint.includes('/stats')) {
            showNotification('Error: ' + error.message, 'danger');
        }
        
        throw error;
    }
}

// Cargar precio actual
async function loadCurrentPrice() {
    try {
        // Mostrar indicador de carga
        const priceElement = document.getElementById('currentPrice');
        const updateElement = document.getElementById('lastUpdate');
        
        // Añadir efecto de pulsación para mostrar actualización
        priceElement.style.opacity = '0.7';
        
        const response = await apiCall('/price');
        currentPrice = response.data.price;
        
        // Detectar cambio de precio para animación
        const oldPriceText = priceElement.textContent || '$0.00';
        const oldPrice = parseFloat(oldPriceText.replace(/[$,]/g, '')) || 0;
        const newPrice = currentPrice;
        const priceChanged = Math.abs(oldPrice - newPrice) > 0.01; // Cambio mínimo de $0.01
        
        // Actualizar precio con formato
        const formattedPrice = new Intl.NumberFormat('es-ES', {
            style: 'currency',
            currency: 'USD'
        }).format(currentPrice);
        
        priceElement.textContent = formattedPrice;
        
        // Obtener porcentaje de cambio desde la API
        updatePriceChangeFromAPI();
        
        // Mostrar fuente de datos
        const sourceText = response.data.source ? ` (${response.data.source})` : '';
        updateElement.textContent = 
            new Date(response.data.timestamp).toLocaleTimeString('es-ES') + sourceText;
        
        // Debug log
        console.log(`Precio actualizado: $${oldPrice} -> $${newPrice} (cambió: ${priceChanged})`);
        
        // Animación de cambio de precio
        if (priceChanged && !isNaN(oldPrice)) {
            if (newPrice > oldPrice) {
                priceElement.classList.add('text-success');
                priceElement.classList.remove('text-danger');
                showPriceAnimation('↗', 'success');
            } else if (newPrice < oldPrice) {
                priceElement.classList.add('text-danger');
                priceElement.classList.remove('text-success');
                showPriceAnimation('↘', 'danger');
            }
            
            // Remover colores después de 3 segundos
            setTimeout(() => {
                priceElement.classList.remove('text-success', 'text-danger');
            }, 3000);
        }
        
        // Restaurar opacidad
        priceElement.style.opacity = '1';
        
    } catch (error) {
        console.error('Error loading current price:', error);
        const updateElement = document.getElementById('lastUpdate');
        const priceElement = document.getElementById('currentPrice');
        
        if (updateElement) {
            updateElement.textContent = 'Error de conexión - ' + new Date().toLocaleTimeString('es-ES');
        }
        
        // Restaurar opacidad en caso de error
        if (priceElement) {
            priceElement.style.opacity = '1';
        }
    }
}

// Actualizar porcentaje de cambio desde la API
async function updatePriceChangeFromAPI() {
    try {
        const response = await apiCall('/price/percentage');
        const percentage = response.data.percentage;
        
        const priceChangeElement = document.getElementById('priceChange');
        if (!priceChangeElement) return;
        
        // Formatear el porcentaje
        const formattedPercentage = percentage > 0 
            ? `+${percentage.toFixed(2)}%` 
            : `${percentage.toFixed(2)}%`;
        
        priceChangeElement.textContent = formattedPercentage;
        
        // Aplicar clase CSS según el cambio
        priceChangeElement.className = 'price-change';
        if (percentage > 0) {
            priceChangeElement.classList.add('positive');
        } else if (percentage < 0) {
            priceChangeElement.classList.add('negative');
        } else {
            priceChangeElement.classList.add('neutral');
        }
        
        console.log(`Porcentaje de cambio actualizado desde API: ${formattedPercentage}`);
        
    } catch (error) {
        console.error('Error loading percentage change:', error);
        // Si falla la API del porcentaje, mantener el valor actual
    }
}

// Actualizar estadísticas del dashboard
async function updateStats() {
    try {
        const response = await apiCall('/stats');
        const stats = response.data;
        
        // Actualizar solo los números que pueden cambiar
        if (stats.current_price !== undefined) {
            // El precio ya se actualiza en loadCurrentPrice()
        }
        
        // Actualizar contador de alertas activas si cambió
        const activeAlertsElements = document.querySelectorAll('.display-6');
        if (activeAlertsElements.length > 0 && stats.active_alerts !== undefined) {
            // El segundo .display-6 es el de alertas activas
            const activeAlertsElement = activeAlertsElements[0]; // Primer elemento = alertas activas
            if (activeAlertsElement && activeAlertsElement.textContent !== stats.active_alerts.toString()) {
                activeAlertsElement.textContent = stats.active_alerts;
                // Pequeña animación para indicar cambio
                activeAlertsElement.style.transform = 'scale(1.1)';
                setTimeout(() => {
                    activeAlertsElement.style.transform = 'scale(1)';
                }, 200);
            }
        }
        
    } catch (error) {
        console.error('Error updating stats:', error);
    }
}

// Mostrar animación de cambio de precio
function showPriceAnimation(arrow, type) {
    const priceElement = document.getElementById('currentPrice');
    
    // Crear elemento de animación
    const animation = document.createElement('span');
    animation.textContent = arrow;
    animation.className = `price-change-animation text-${type}`;
    animation.style.cssText = `
        position: absolute;
        font-size: 1.5rem;
        font-weight: bold;
        opacity: 1;
        transform: translateY(0px);
        transition: all 1s ease-out;
        margin-left: 10px;
        z-index: 10;
    `;
    
    // Agregar al contenedor del precio
    priceElement.parentElement.style.position = 'relative';
    priceElement.parentElement.appendChild(animation);
    
    // Animar hacia arriba y desvanecer
    setTimeout(() => {
        animation.style.transform = 'translateY(-20px)';
        animation.style.opacity = '0';
    }, 100);
    
    // Remover elemento después de la animación
    setTimeout(() => {
        if (animation.parentElement) {
            animation.parentElement.removeChild(animation);
        }
    }, 1200);
}

// Cargar historial de precios
async function loadPriceHistory() {
    try {
        const response = await apiCall('/price/history?limit=24');
        updatePriceChart(response.data);
    } catch (error) {
        console.error('Error loading price history:', error);
    }
}

// Actualizar gráfico de precios
function updatePriceChart(priceData) {
    const ctx = document.getElementById('priceChart').getContext('2d');
    
    if (priceChart) {
        priceChart.destroy();
    }
    
    const labels = priceData.map(item => 
        new Date(item.timestamp).toLocaleTimeString('es-ES', {
            hour: '2-digit',
            minute: '2-digit'
        })
    ).reverse();
    
    const prices = priceData.map(item => item.price).reverse();
    
    priceChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: 'Precio BTC (USD)',
                data: prices,
                borderColor: '#f7931a',
                backgroundColor: 'rgba(247, 147, 26, 0.1)',
                borderWidth: 2,
                fill: true,
                tension: 0.1
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    beginAtZero: false,
                    ticks: {
                        callback: function(value) {
                            return '$' + value.toLocaleString();
                        }
                    }
                }
            },
            plugins: {
                legend: {
                    display: false
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            return '$' + context.parsed.y.toLocaleString();
                        }
                    }
                }
            }
        }
    });
}

// Cargar alertas
async function loadAlerts() {
    try {
        const response = await apiCall('/alerts');
        displayAlerts(response.data);
    } catch (error) {
        console.error('Error loading alerts:', error);
        document.getElementById('alertsList').innerHTML = 
            '<div class="text-center text-danger">Error cargando alertas</div>';
    }
}

// Mostrar alertas
function displayAlerts(alerts) {
    const container = document.getElementById('alertsList');
    
    if (alerts.length === 0) {
        container.innerHTML = `
            <div class="text-center text-muted">
                <i class="fas fa-bell-slash fa-2x mb-2"></i>
                <p>No tienes alertas configuradas</p>
                <small>Crea tu primera alerta en el panel de la izquierda</small>
            </div>
        `;
        return;
    }
    
    container.innerHTML = alerts.map(alert => `
        <div class="card alert-card mb-3">
            <div class="card-body">
                <div class="d-flex justify-content-between align-items-start">
                    <div>
                        <h6 class="card-title">
                            <i class="fas fa-bell"></i> ${alert.name}
                            <span class="badge ${
                                alert.last_triggered ? 'bg-warning' : 
                                alert.is_active ? 'bg-success' : 'bg-secondary'
                            } ms-2">
                                ${
                                    alert.last_triggered ? 'Disparada' : 
                                    alert.is_active ? 'Activa' : 'Inactiva'
                                }
                            </span>
                        </h6>
                        <p class="card-text text-muted mb-1">
                            ${getAlertDescription(alert)}
                        </p>
                        <small class="text-muted">
                            <i class="fas fa-envelope"></i> ${alert.email}
                            ${alert.last_triggered ? 
                                `• Última activación: ${new Date(alert.last_triggered).toLocaleString('es-ES')}` : 
                                '• Nunca activada'
                            }
                        </small>
                        <div class="mt-2">
                            <small class="text-muted">Notificaciones: </small>
                            ${alert.enable_email ? 
                                '<span class="badge bg-primary me-1"><i class="fas fa-envelope"></i> Email</span>' : ''
                            }
                            ${alert.enable_telegram ? 
                                '<span class="badge bg-info me-1"><i class="fab fa-telegram"></i> Telegram</span>' : ''
                            }
                            ${alert.enable_web_push ? 
                                '<span class="badge bg-warning me-1"><i class="fas fa-globe"></i> Web Push</span>' : ''
                            }
                        </div>
                        ${alert.trigger_count > 0 ? 
                            `<br><small class="text-info">Activada ${alert.trigger_count} vez${alert.trigger_count > 1 ? 'es' : ''}</small>` : 
                            ''
                        }
                    </div>
                    <div class="btn-group-vertical btn-group-sm">
                        <button class="btn btn-outline-primary" onclick="testAlert(${alert.id})" title="Probar">
                            <i class="fas fa-vial"></i>
                        </button>
                        <button class="btn btn-outline-${alert.is_active ? 'warning' : 'success'}" 
                                onclick="toggleAlert(${alert.id})" 
                                title="${alert.is_active ? 'Desactivar' : 'Activar'}">
                            <i class="fas fa-power-off"></i>
                        </button>
                        ${alert.last_triggered ? 
                            `<button class="btn btn-outline-warning" onclick="resetAlert(${alert.id})" title="Resetear">
                                <i class="fas fa-redo"></i>
                            </button>` : ''
                        }
                        <button class="btn btn-outline-info" onclick="editAlert(${alert.id})" title="Editar">
                            <i class="fas fa-edit"></i>
                        </button>
                        <button class="btn btn-outline-danger" onclick="deleteAlert(${alert.id})" title="Eliminar">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </div>
            </div>
        </div>
    `).join('');
}

function getAlertDescription(alert) {
    switch (alert.type) {
        case 'above':
            return `Precio por encima de $${alert.target_price.toLocaleString()}`;
        case 'below':
            return `Precio por debajo de $${alert.target_price.toLocaleString()}`;
        case 'change':
            if (alert.percentage > 0) {
                return `Subida de ${alert.percentage}% o más`;
            } else if (alert.percentage < 0) {
                return `Bajada de ${Math.abs(alert.percentage)}% o más`;
            } else {
                return `Cambio de ${alert.percentage}% en el precio`;
            }
        default:
            return 'Tipo de alerta desconocido';
    }
}

// Crear nueva alerta
async function createAlert(event) {
    event.preventDefault();
    
    const formData = new FormData(event.target);
    const alertData = {
        name: document.getElementById('alertName').value,
        type: document.getElementById('alertType').value,
        email: document.getElementById('alertEmail').value,
        enable_email: document.getElementById('enableEmail').checked,
        enable_telegram: document.getElementById('enableTelegram').checked,
        enable_web_push: document.getElementById('enableWebPush').checked,
        is_active: true
    };
    
    if (alertData.type === 'change') {
        alertData.percentage = parseFloat(document.getElementById('percentage').value);
    } else {
        alertData.target_price = parseFloat(document.getElementById('targetPrice').value);
    }
    
    try {
        await apiCall('/alerts', {
            method: 'POST',
            body: JSON.stringify(alertData)
        });
        
        showNotification('Alerta creada exitosamente', 'success');
        document.getElementById('alertForm').reset();
        loadAlerts();
    } catch (error) {
        console.error('Error creating alert:', error);
    }
}

// Probar alerta
async function testAlert(alertId) {
    try {
        await apiCall(`/alerts/${alertId}/test`, { method: 'POST' });
        showNotification('Notificación de prueba enviada', 'info');
    } catch (error) {
        console.error('Error testing alert:', error);
    }
}

// Alternar estado de alerta
async function toggleAlert(alertId) {
    try {
        await apiCall(`/alerts/${alertId}/toggle`, { method: 'POST' });
        showNotification('Estado de alerta actualizado', 'success');
        loadAlerts();
    } catch (error) {
        console.error('Error toggling alert:', error);
    }
}

// Resetear alerta para que pueda dispararse de nuevo
async function resetAlert(alertId) {
    if (confirm('¿Estás seguro de que quieres resetear esta alerta? Podrá dispararse de nuevo cuando se cumpla la condición.')) {
        try {
            await apiCall(`/alerts/${alertId}/reset`, { method: 'POST' });
            showNotification('🔄 Alerta reseteada exitosamente', 'success');
            loadAlerts();
        } catch (error) {
            console.error('Error resetting alert:', error);
            showNotification('Error al resetear la alerta', 'error');
        }
    }
}

// Editar alerta - Abrir modal
async function editAlert(alertId) {
    try {
        // Obtener datos de la alerta
        const response = await apiCall(`/alerts/${alertId}`);
        const alert = response.data;
        
        // Configurar el modal según el tipo de alerta
        document.getElementById('editAlertId').value = alertId;
        document.getElementById('editAlertType').value = alert.type;
        
        const editValueLabel = document.getElementById('editValueLabel');
        const editValueHelp = document.getElementById('editValueHelp');
        const editValueInput = document.getElementById('editValue');
        
        if (alert.type === 'above' || alert.type === 'below') {
            editValueLabel.textContent = 'Precio Objetivo ($)';
            editValueHelp.textContent = 'Ingresa el nuevo precio objetivo en dólares';
            editValueInput.value = alert.target_price;
            editValueInput.step = '0.01';
            editValueInput.min = '0';
        } else if (alert.type === 'change') {
            editValueLabel.textContent = 'Porcentaje de Cambio (%)';
            editValueHelp.textContent = 'Ingresa el nuevo porcentaje de cambio';
            editValueInput.value = alert.percentage;
            editValueInput.step = '0.1';
            editValueInput.min = '0.1';
            editValueInput.max = '100';
        }
        
        // Mostrar modal
        const modal = new bootstrap.Modal(document.getElementById('editAlertModal'));
        modal.show();
        
    } catch (error) {
        console.error('Error loading alert for editing:', error);
        showNotification('Error al cargar la alerta', 'error');
    }
}

// Guardar cambios de la alerta editada
async function saveEditAlert() {
    const alertId = document.getElementById('editAlertId').value;
    const alertType = document.getElementById('editAlertType').value;
    const newValue = parseFloat(document.getElementById('editValue').value);
    
    if (!newValue || newValue <= 0) {
        showNotification('Por favor ingresa un valor válido', 'error');
        return;
    }
    
    try {
        const updateData = {};
        
        if (alertType === 'above' || alertType === 'below') {
            updateData.target_price = newValue;
        } else if (alertType === 'change') {
            updateData.percentage = newValue;
        }
        
        await apiCall(`/alerts/${alertId}`, {
            method: 'PUT',
            body: JSON.stringify(updateData)
        });
        
        // Cerrar modal
        const modal = bootstrap.Modal.getInstance(document.getElementById('editAlertModal'));
        modal.hide();
        
        // Mostrar mensaje de éxito
        showNotification('✅ Alerta actualizada exitosamente', 'success');
        
        // Recargar lista de alertas
        loadAlerts();
        
    } catch (error) {
        console.error('Error updating alert:', error);
        showNotification('Error al actualizar la alerta', 'error');
    }
}

// Eliminar alerta
async function deleteAlert(alertId) {
    if (!confirm('¿Estás seguro de que quieres eliminar esta alerta?')) {
        return;
    }
    
    try {
        await apiCall(`/alerts/${alertId}`, { method: 'DELETE' });
        showNotification('Alerta eliminada exitosamente', 'success');
        loadAlerts();
    } catch (error) {
        console.error('Error deleting alert:', error);
    }
}

// Mostrar notificaciones
function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.className = `alert alert-${type} alert-dismissible fade show position-fixed`;
    notification.style.cssText = 'top: 20px; right: 20px; z-index: 9999; min-width: 300px;';
    notification.innerHTML = `
        ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
    `;
    
    document.body.appendChild(notification);
    
    // Auto-dismiss después de 5 segundos
    setTimeout(() => {
        if (notification.parentNode) {
            notification.remove();
        }
    }, 5000);
}

// Actualizar precios periódicamente
function startPriceUpdates() {
    // Actualizar precio cada 15 segundos
    setInterval(() => {
        loadCurrentPrice();
        updateStats(); // También actualizar estadísticas
    }, 15000); // 15 segundos
    
    // Actualizar historial cada 2 minutos
    setInterval(() => {
        loadPriceHistory();
    }, 120000); // 2 minutos
    
    // Actualizar alertas cada 30 segundos
    setInterval(() => {
        loadAlerts();
    }, 30000); // 30 segundos
}

// ===========================================
// WEB PUSH NOTIFICATIONS
// ===========================================

// Inicializar Web Push notifications
async function initializeWebPush() {
    console.log('🔄 Inicializando Web Push notifications...');

    // Verificar soporte del navegador
    if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
        console.log('❌ Web Push no soportado por este navegador');
        webPushSupported = false;
        return;
    }

    webPushSupported = true;
    console.log('✅ Web Push soportado');

    try {
        // Registrar Service Worker
        const registration = await navigator.serviceWorker.register('/static/sw.js');
        console.log('✅ Service Worker registrado:', registration);

        // Obtener VAPID public key del servidor
        await loadVAPIDPublicKey();

        // Verificar subscripción existente
        await checkExistingSubscription(registration);

        // Mostrar botón de notificaciones
        updateWebPushUI();

    } catch (error) {
        console.error('❌ Error inicializando Web Push:', error);
        webPushSupported = false;
    }
}

// Cargar VAPID public key del servidor
async function loadVAPIDPublicKey() {
    try {
        const response = await apiCall('/webpush/vapid-public-key');
        vapidPublicKey = response.data.publicKey;
        console.log('✅ VAPID Public Key cargada');
    } catch (error) {
        console.error('❌ Error cargando VAPID key:', error);
        throw error;
    }
}

// Verificar subscripción existente
async function checkExistingSubscription(registration) {
    try {
        webPushSubscription = await registration.pushManager.getSubscription();
        if (webPushSubscription) {
            console.log('✅ Subscripción Web Push existente encontrada');
        } else {
            console.log('ℹ️ No hay subscripción Web Push existente');
        }
    } catch (error) {
        console.error('❌ Error verificando subscripción:', error);
    }
}

// Solicitar permisos y suscribirse a Web Push
async function subscribeToWebPush() {
    if (!webPushSupported) {
        showNotification('Web Push no soportado por este navegador', 'error');
        return false;
    }

    try {
        // Solicitar permiso de notificaciones
        const permission = await Notification.requestPermission();
        if (permission !== 'granted') {
            showNotification('Permisos de notificación denegados', 'error');
            return false;
        }

        // Obtener registration del Service Worker
        const registration = await navigator.serviceWorker.ready;

        // Crear subscripción
        const subscription = await registration.pushManager.subscribe({
            userVisibleOnly: true,
            applicationServerKey: urlBase64ToUint8Array(vapidPublicKey)
        });

        // Enviar subscripción al servidor
        const response = await apiCall('/webpush/subscribe', {
            method: 'POST',
            body: JSON.stringify({
                endpoint: subscription.endpoint,
                p256dh: arrayBufferToBase64(subscription.getKey('p256dh')),
                auth: arrayBufferToBase64(subscription.getKey('auth')),
                user_id: 'anonymous' // TODO: Implementar usuarios
            })
        });

        if (response.success) {
            webPushSubscription = subscription;
            showNotification('✅ Notificaciones Web Push activadas', 'success');
            updateWebPushUI();
            return true;
        } else {
            throw new Error(response.error || 'Error del servidor');
        }

    } catch (error) {
        console.error('❌ Error suscribiendo a Web Push:', error);
        showNotification('Error activando notificaciones: ' + error.message, 'error');
        return false;
    }
}

// Cancelar subscripción a Web Push
async function unsubscribeFromWebPush() {
    if (!webPushSubscription) {
        showNotification('No hay subscripción activa', 'warning');
        return false;
    }

    try {
        // Cancelar subscripción en el navegador
        await webPushSubscription.unsubscribe();

        // Notificar al servidor
        await apiCall('/webpush/unsubscribe', {
            method: 'DELETE',
            body: JSON.stringify({
                endpoint: webPushSubscription.endpoint
            })
        });

        webPushSubscription = null;
        showNotification('Notificaciones Web Push desactivadas', 'info');
        updateWebPushUI();
        return true;

    } catch (error) {
        console.error('❌ Error cancelando subscripción Web Push:', error);
        showNotification('Error desactivando notificaciones: ' + error.message, 'error');
        return false;
    }
}

// Actualizar interfaz de usuario para Web Push
function updateWebPushUI() {
    const webPushButton = document.getElementById('webPushToggle');
    if (!webPushButton) return;

    if (!webPushSupported) {
        webPushButton.style.display = 'none';
        return;
    }

    webPushButton.style.display = 'block';
    
    if (webPushSubscription) {
        webPushButton.textContent = '🔕 Desactivar Web Push';
        webPushButton.className = 'btn btn-warning btn-sm';
        webPushButton.onclick = unsubscribeFromWebPush;
    } else {
        webPushButton.textContent = '🔔 Activar Web Push';
        webPushButton.className = 'btn btn-success btn-sm';
        webPushButton.onclick = subscribeToWebPush;
    }
}

// Funciones auxiliares
function urlBase64ToUint8Array(base64String) {
    const padding = '='.repeat((4 - base64String.length % 4) % 4);
    const base64 = (base64String + padding)
        .replace(/-/g, '+')
        .replace(/_/g, '/');

    const rawData = window.atob(base64);
    const outputArray = new Uint8Array(rawData.length);

    for (let i = 0; i < rawData.length; ++i) {
        outputArray[i] = rawData.charCodeAt(i);
    }
    return outputArray;
}

function arrayBufferToBase64(buffer) {
    const bytes = new Uint8Array(buffer);
    let binary = '';
    for (let i = 0; i < bytes.byteLength; i++) {
        binary += String.fromCharCode(bytes[i]);
    }
    return window.btoa(binary);
} 