(function() {
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
    });
})(); 