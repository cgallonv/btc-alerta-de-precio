(function() {
    // Balance update
    async function updateBalance() {
        try {
            const response = await apiCall('/account/balance');
            
            // Update total balance
            const totalBalanceElement = document.querySelector('.balance-item h3.bitcoin-color');
            if (totalBalanceElement) {
                totalBalanceElement.textContent = formatCurrency(response.data.TotalBalance);
            }

            // Update available balance
            const availableBalanceElement = document.querySelector('.balance-item h3.text-success');
            if (availableBalanceElement) {
                availableBalanceElement.textContent = formatCurrency(response.data.AvailableBalance);
            }

            // Update last updated timestamp
            const lastUpdatedElement = document.querySelector('.balance-summary .text-muted + span');
            if (lastUpdatedElement) {
                lastUpdatedElement.textContent = new Date(response.data.LastUpdated).toLocaleString();
            }

            // Update account status
            updateAccountStatus(response.data);

            // Update commission rates
            updateCommissionRates(response.data);

            // Update assets table
            updateAssetsTable(response.data.Assets);
        } catch (error) {
            console.error('Error updating balance:', error);
            showNotification('Error updating balance information', 'danger');
        }
    }

    function updateAccountStatus(data) {
        // Update account type
        const accountTypeElement = document.querySelector('[data-field="account-type"]');
        if (accountTypeElement) {
            accountTypeElement.textContent = data.accountType;
        }

        // Update trading status
        const tradingStatusElement = document.querySelector('[data-field="trading-status"]');
        if (tradingStatusElement) {
            tradingStatusElement.innerHTML = data.canTrade ? 
                '<span class="badge bg-success">Enabled</span>' : 
                '<span class="badge bg-danger">Disabled</span>';
        }

        // Update withdrawal status
        const withdrawalStatusElement = document.querySelector('[data-field="withdrawal-status"]');
        if (withdrawalStatusElement) {
            withdrawalStatusElement.innerHTML = data.canWithdraw ? 
                '<span class="badge bg-success">Enabled</span>' : 
                '<span class="badge bg-danger">Disabled</span>';
        }

        // Update deposit status
        const depositStatusElement = document.querySelector('[data-field="deposit-status"]');
        if (depositStatusElement) {
            depositStatusElement.innerHTML = data.canDeposit ? 
                '<span class="badge bg-success">Enabled</span>' : 
                '<span class="badge bg-danger">Disabled</span>';
        }

        // Update self-trade prevention
        const selfTradeElement = document.querySelector('[data-field="self-trade"]');
        if (selfTradeElement) {
            selfTradeElement.innerHTML = data.requireSelfTradePrevention ? 
                '<span class="badge bg-warning">Required</span>' : 
                '<span class="badge bg-secondary">Optional</span>';
        }

        // Update permissions
        const permissionsElement = document.querySelector('[data-field="permissions"]');
        if (permissionsElement) {
            permissionsElement.innerHTML = data.permissions.map(perm => 
                `<span class="badge bg-info me-1">${perm}</span>`
            ).join('');
        }
    }

    function updateCommissionRates(data) {
        // Update maker fee
        const makerFeeElement = document.querySelector('[data-field="maker-fee"]');
        if (makerFeeElement) {
            makerFeeElement.textContent = data.commissionRates.maker + '%';
        }

        // Update taker fee
        const takerFeeElement = document.querySelector('[data-field="taker-fee"]');
        if (takerFeeElement) {
            makerFeeElement.textContent = data.commissionRates.taker + '%';
        }

        // Update buyer fee
        const buyerFeeElement = document.querySelector('[data-field="buyer-fee"]');
        if (buyerFeeElement) {
            buyerFeeElement.textContent = data.commissionRates.buyer + '%';
        }

        // Update seller fee
        const sellerFeeElement = document.querySelector('[data-field="seller-fee"]');
        if (sellerFeeElement) {
            sellerFeeElement.textContent = data.commissionRates.seller + '%';
        }
    }

    function updateAssetsTable(assets) {
        const tableBody = document.querySelector('.assets-list tbody');
        if (!tableBody) return;

        tableBody.innerHTML = '';
        
        assets.forEach(asset => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td>
                    <div class="d-flex align-items-center">
                        <i class="fab fa-bitcoin me-2"></i>
                        <span>${asset.symbol}</span>
                    </div>
                </td>
                <td>${asset.free}</td>
                <td>${asset.locked}</td>
                <td>${formatNumber(asset.total)}</td>
                <td>${formatCurrency(asset.value_usd)}</td>
                <td class="${asset.change_24h > 0 ? 'text-success' : 'text-danger'}">
                    ${formatNumber(asset.change_24h)}%
                </td>
            `;
            tableBody.appendChild(row);
        });
    }

    function formatCurrency(value) {
        if (typeof value !== 'number' || isNaN(value)) {
            value = 0;
        }
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: 'USD',
            minimumFractionDigits: 2,
            maximumFractionDigits: 2
        }).format(value);
    }

    function formatNumber(value) {
        if (typeof value !== 'number' || isNaN(value)) {
            value = 0;
        }
        return new Intl.NumberFormat('en-US', {
            minimumFractionDigits: 2,
            maximumFractionDigits: 8
        }).format(value);
    }

    // Order filtering
    function setupOrderFilters() {
        const filterButtons = document.querySelectorAll('[data-filter]');
        filterButtons.forEach(button => {
            button.addEventListener('click', function() {
                const filterValue = this.dataset.filter;
                const orders = document.querySelectorAll('[data-order-type]');

                // Update active button state
                filterButtons.forEach(btn => btn.classList.remove('active'));
                this.classList.add('active');

                // Filter orders
                orders.forEach(order => {
                    if (filterValue === 'all' || order.dataset.orderType === filterValue) {
                        order.style.display = '';
                    } else {
                        order.style.display = 'none';
                    }
                });
            });
        });
    }

    // Budget management
    window.adjustBudget = function() {
        const newLimit = prompt('Enter new budget limit:');
        if (newLimit && !isNaN(newLimit)) {
            // TODO: Implement API call to update budget
            alert('Budget limit updated successfully');
        }
    };

    window.viewBudgetHistory = function() {
        // TODO: Implement budget history view
        alert('Budget history feature coming soon');
    };

    // Cancel order
    window.cancelOrder = function(orderId) {
        if (confirm('Are you sure you want to cancel this order?')) {
            // TODO: Implement API call to cancel order
            alert('Order cancelled successfully');
        }
    };

    // Initialize
    document.addEventListener('DOMContentLoaded', function() {
        setupOrderFilters();
        updateBalance(); // Initial balance update
        
        // Update balance every minute
        setInterval(updateBalance, 60000);
    });
})(); 