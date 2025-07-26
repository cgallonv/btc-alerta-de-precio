// Service Worker para manejar Web Push Notifications
const CACHE_NAME = 'btc-price-alert-v1';

// Instalación del Service Worker
self.addEventListener('install', event => {
    console.log('🔧 Service Worker instalando...');
    
    event.waitUntil(
        caches.open(CACHE_NAME).then(cache => {
            console.log('📦 Cache abierto');
            return cache.addAll([
                '/',
                '/static/css/style.css',
                '/static/js/app.js',
                '/static/images/bitcoin-icon.png'
            ]);
        })
    );
});

// Activación del Service Worker
self.addEventListener('activate', event => {
    console.log('🚀 Service Worker activado');
    
    event.waitUntil(
        caches.keys().then(cacheNames => {
            return Promise.all(
                cacheNames.map(cacheName => {
                    if (cacheName !== CACHE_NAME) {
                        console.log('🗑️ Eliminando cache antiguo:', cacheName);
                        return caches.delete(cacheName);
                    }
                })
            );
        })
    );
});

// Manejar notificaciones push
self.addEventListener('push', event => {
    console.log('📨 Push notification recibida:', event);

    if (!event.data) {
        console.log('❌ No hay datos en la notificación push');
        return;
    }

    let notificationData;
    try {
        notificationData = event.data.json();
    } catch (e) {
        console.error('❌ Error parsing notification data:', e);
        return;
    }

    const title = notificationData.title || '🚨 Bitcoin Price Alert';
    const options = {
        body: notificationData.body || 'Nueva alerta de precio de Bitcoin',
        icon: notificationData.icon || '/static/images/bitcoin-icon.png',
        badge: notificationData.badge || '/static/images/bitcoin-badge.png',
        tag: 'bitcoin-price-alert',
        data: notificationData.data || {},
        actions: notificationData.actions || [
            {
                action: 'view',
                title: '👁️ Ver Dashboard'
            },
            {
                action: 'close',
                title: '❌ Cerrar'
            }
        ],
        requireInteraction: true,
        vibrate: [100, 50, 100]
    };

    event.waitUntil(
        self.registration.showNotification(title, options)
    );
});

// Manejar clicks en notificaciones
self.addEventListener('notificationclick', event => {
    console.log('👆 Click en notificación:', event);

    event.notification.close();

    if (event.action === 'view') {
        // Abrir o enfocar la aplicación
        event.waitUntil(
            clients.matchAll({ type: 'window' }).then(clientList => {
                // Si ya hay una ventana abierta, enfocarla
                for (const client of clientList) {
                    if (client.url === '/' || client.url.includes(self.location.origin)) {
                        return client.focus();
                    }
                }
                
                // Si no hay ventana abierta, abrir una nueva
                return clients.openWindow('/');
            })
        );
    } else if (event.action === 'close') {
        // Solo cerrar la notificación (ya se hizo arriba)
        console.log('🚫 Notificación cerrada por el usuario');
    } else {
        // Click en el cuerpo de la notificación (no en los botones)
        event.waitUntil(
            clients.openWindow('/')
        );
    }
});

// Manejar cierre de notificaciones
self.addEventListener('notificationclose', event => {
    console.log('🚫 Notificación cerrada:', event);
});

// Fetch event para cache (opcional)
self.addEventListener('fetch', event => {
    // Solo cachear recursos estáticos
    if (event.request.url.includes('/static/')) {
        event.respondWith(
            caches.match(event.request).then(response => {
                return response || fetch(event.request);
            })
        );
    }
}); 