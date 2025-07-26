// Variables globales
let priceChart;
let currentPrice = 0;

// Inicializar la aplicación
document.addEventListener('DOMContentLoaded', function() {
    initializeApp();
    setupEventListeners();
    startPriceUpdates();
});

function initializeApp() {
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
                            ${alert.enable_desktop ? 
                                '<span class="badge bg-success me-1"><i class="fas fa-desktop"></i> Desktop</span>' : ''
                            }
                            ${alert.enable_telegram ? 
                                '<span class="badge bg-info me-1"><i class="fab fa-telegram"></i> Telegram</span>' : ''
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
            return `Cambio de ${alert.percentage}% en el precio`;
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
        enable_desktop: document.getElementById('enableDesktop').checked,
        enable_telegram: document.getElementById('enableTelegram').checked,
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