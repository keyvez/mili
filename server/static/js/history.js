function toCamelCase(str) {
    return str.toLowerCase()
        .replace(/[-_\s]+(.)/g, (_, c) => c.toUpperCase())
        .replace(/^(.)/, c => c.toLowerCase());
}

function fromCamelCase(str) {
    return str
        .replace(/([a-z])([A-Z])/g, '$1 $2')
        .replace(/^(.)/, c => c.toUpperCase());
}

function updateActiveCategory(category) {
    const cards = document.querySelectorAll('.category-card');
    cards.forEach(card => {
        if (card.dataset.category === category) {
            card.classList.add('bg-blue-600', 'text-white', 'shadow-lg', 'scale-105');
            card.classList.remove('bg-white', 'opacity-60');
        } else {
            card.classList.remove('bg-blue-600', 'text-white', 'shadow-lg', 'scale-105');
            card.classList.add('bg-white', 'opacity-60');
        }
    });
}

document.addEventListener('DOMContentLoaded', function() {
    document.body.addEventListener('htmx:afterRequest', function(event) {
        if (event.detail.requestConfig.path.startsWith('/resources')) {
            const urlParams = new URLSearchParams(event.detail.requestConfig.path.split('?')[1]);
            const category = urlParams.get('category');
            if (category) {
                const camelCaseCategory = toCamelCase(category);
                history.pushState({ category }, '', `#${camelCaseCategory}`);
                updateActiveCategory(category);
            }
        }
    });

    // Handle back/forward navigation
    window.addEventListener('popstate', function(event) {
        if (event.state && event.state.category) {
            const category = fromCamelCase(event.state.category);
            htmx.ajax('POST', `/resources?category=${encodeURIComponent(category)}`, {
                target: '#resources-container',
                indicator: '#loading'
            });
            updateActiveCategory(category);
        } else {
            // Reset all cards if no category is selected
            const cards = document.querySelectorAll('.category-card');
            cards.forEach(card => {
                card.classList.remove('bg-blue-600', 'text-white', 'shadow-lg', 'scale-105', 'opacity-60');
                card.classList.add('bg-white');
            });
        }
    });

    // Handle initial load with hash
    const hash = window.location.hash.slice(1);
    if (hash) {
        const category = fromCamelCase(hash);
        htmx.ajax('POST', `/resources?category=${encodeURIComponent(category)}`, {
            target: '#resources-container',
            indicator: '#loading'
        });
        updateActiveCategory(category);
    }
});
