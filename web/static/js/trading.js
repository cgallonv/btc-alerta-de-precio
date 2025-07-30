// Trading page functionality
document.addEventListener('DOMContentLoaded', function() {
    // Initialize tooltips
    var tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'))
    var tooltipList = tooltipTriggerList.map(function (tooltipTriggerEl) {
        return new bootstrap.Tooltip(tooltipTriggerEl)
    });

    // Initialize TradingView widget with better configuration
    new TradingView.widget({
        "autosize": true,
        "symbol": "BINANCE:BTCUSDT",
        "interval": "60",
        "timezone": "exchange",
        "theme": "light",
        "style": "1",
        "locale": "es",
        "toolbar_bg": "#f1f3f6",
        "enable_publishing": false,
        "withdateranges": true,
        "hide_side_toolbar": false,
        "allow_symbol_change": false,
        "details": true,
        "hotlist": true,
        "calendar": true,
        "studies": [
            "MASimple@tv-basicstudies",
            "MAExp@tv-basicstudies",
            "RSI@tv-basicstudies",
            "MACD@tv-basicstudies"
        ],
        "studies_overrides": {
            "moving average exponential.length": 21,
            "moving average exponential.plottype": "line",
            "moving average exponential.color": "#2962FF"
        },
        "container_id": "tradingChart",
        "show_popup_button": true,
        "popup_width": "1000",
        "popup_height": "650",
        "hide_volume": false
    });

    // Handle timeframe buttons
    document.querySelectorAll('.timeframe-buttons .btn').forEach(button => {
        button.addEventListener('click', function() {
            // Remove active class from all buttons
            document.querySelectorAll('.timeframe-buttons .btn').forEach(btn => {
                btn.classList.remove('active');
            });
            // Add active class to clicked button
            this.classList.add('active');
            // TODO: Update chart timeframe
        });
    });

    // Handle risk level changes
    document.getElementById('riskLevel').addEventListener('change', function() {
        updateRiskLevel(this.value);
    });

    // Handle strategy type changes
    document.getElementById('strategyType').addEventListener('change', function() {
        updateStrategySettings(this.value);
    });

    // Handle indicator toggles
    document.querySelectorAll('.indicator-toggle').forEach(checkbox => {
        checkbox.addEventListener('change', function() {
            toggleIndicator(this.id, this.checked);
        });
    });

    // Initialize position size calculator
    initializePositionCalculator();
});

function updateRiskLevel(level) {
    const badge = document.getElementById('riskLevelBadge');
    badge.className = 'risk-level ' + level;
    badge.textContent = level.charAt(0).toUpperCase() + level.slice(1) + ' Risk';
}

function updateStrategySettings(type) {
    // TODO: Update strategy settings based on type
    console.log('Strategy type changed to:', type);
}

function toggleIndicator(indicatorId, enabled) {
    // TODO: Toggle indicator on chart
    console.log('Indicator', indicatorId, enabled ? 'enabled' : 'disabled');
}

function initializePositionCalculator() {
    const positionSize = document.getElementById('positionSize');
    const stopLoss = document.getElementById('stopLoss');
    const takeProfit = document.getElementById('takeProfit');

    [positionSize, stopLoss, takeProfit].forEach(input => {
        input.addEventListener('input', calculateRisk);
    });
}

function calculateRisk() {
    // TODO: Implement position risk calculation
    console.log('Calculating position risk...');
}

// Handle buy/sell buttons
document.getElementById('buyBtn')?.addEventListener('click', () => {
    openPosition('long');
});

document.getElementById('sellBtn')?.addEventListener('click', () => {
    openPosition('short');
});

function openPosition(type) {
    const size = document.getElementById('positionSize').value;
    const stopLoss = document.getElementById('stopLoss').value;
    const takeProfit = document.getElementById('takeProfit').value;

    // TODO: Implement position opening logic
    console.log('Opening', type, 'position:', {
        size,
        stopLoss,
        takeProfit
    });
}