(function() {
    // Balance update
    async function updateBalance() {
        try {
            const response = await apiCall('/account/balance');
            
            // Update total balance
            const totalBalanceElement = document.querySelector('.balance-item h3');
            if (totalBalanceElement) {
                totalBalanceElement.textContent = formatCurrency(response.data.total_balance);
            }

            // Update available balance
            const availableBalanceElement = document.querySelector('.balance-item h3.text-success');
            if (availableBalanceElement) {
                availableBalanceElement.textContent = formatCurrency(response.data.available_balance);
            }

            // Update last updated timestamp
            const lastUpdatedElement = document.querySelector('.balance-summary .text-muted + span');
            if (lastUpdatedElement) {
                lastUpdatedElement.textContent = new Date(response.data.last_updated).toLocaleString();
            }

            // Update assets table
            updateAssetsTable(response.data.assets);
        } catch (error) {
            console.error('Error updating balance:', error);
            showNotification('Error updating balance information', 'danger');
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
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: 'USD'
        }).format(value);
    }

    function formatNumber(value) {
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