function processLink(link) {
    if (link && link.includes('@') && !link.startsWith('mailto:')) {
        return 'mailto:' + link;
    }
    return link;
}

async function handleBackToTabs() {
    const sheetID = "1L0dQpfj3c86mXRjADRrLshUCZrFzA3vcM_TfYxITjmc";
    const loadingSpinner = document.getElementById('loading-spinner');
    
    const showLoading = () => {
        if (loadingSpinner) loadingSpinner.classList.remove('hidden');
    };
    
    const hideLoading = () => {
        if (loadingSpinner) loadingSpinner.classList.add('hidden');
    };

    try {
        showLoading();
        const tabsResponse = await fetch(`/api/sheet-tabs/${sheetID}`);
        if (!tabsResponse.ok) {
            throw new Error('Failed to fetch tabs');
        }
        const tabs = await tabsResponse.json();
        
        // Get both containers
        const tabsContainer = document.getElementById('sheet-tabs');
        const sheetView = document.getElementById('sheet-view');
        const sheetData = document.getElementById('sheet-data');
        
        // Render and show the tabs
        const response = await fetch('/api/render/sheet-tabs', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                tabs: tabs,
                sheetID: sheetID, // Make sure to use the same property name as expected
            }),
        });

        if (!response.ok) {
            throw new Error('Failed to render tabs');
        }

        const html = await response.text();
        
        // Update the DOM
        if (tabsContainer) {
            tabsContainer.innerHTML = html;
            tabsContainer.classList.remove('hidden');
        }
        
        // Hide and clear sheet view
        if (sheetView) {
            sheetView.classList.add('hidden');
            sheetView.innerHTML = '';
        }
        
        // Hide and clear sheet data if it exists
        if (sheetData) {
            sheetData.classList.add('hidden');
            sheetData.innerHTML = '';
        }
    } catch (error) {
        console.error('Error:', error);
        alert('Failed to load tabs. Please try again.');
    } finally {
        hideLoading();
    }
}

function handleCopyClick(element) {
    copyToClipboard(element.dataset.code);
}

function handleCopyButtonClick(element, event) {
    event.stopPropagation();
    copyToClipboard(element.dataset.code);
}

function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(() => {
        // Show a temporary success message
        const toast = document.createElement('div');
        toast.className = 'fixed bottom-4 right-4 bg-green-500 text-white px-4 py-2 rounded shadow-lg transition-opacity duration-500';
        toast.textContent = 'Copied to clipboard!';
        document.body.appendChild(toast);
        
        // Remove the toast after 2 seconds
        setTimeout(() => {
            toast.style.opacity = '0';
            setTimeout(() => toast.remove(), 500);
        }, 2000);
    }).catch(err => {
        console.error('Failed to copy:', err);
    });
}