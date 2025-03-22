document.addEventListener('DOMContentLoaded', () => {
// Tab Management
const setupTabs = (tabSelector, contentSelector) => {
    const tabs = document.querySelectorAll(tabSelector);
    const contents = document.querySelectorAll(contentSelector);

    tabs.forEach(tab => {
        tab.addEventListener('click', () => {
            // Remove active states
            tabs.forEach(t => t.classList.remove('active'));
            contents.forEach(c => c.classList.remove('active'));

            // Add active state to clicked tab
            tab.classList.add('active');
            document.getElementById(`${tab.dataset.tab}-tab`).classList.add('active');
        });
    });
};

// Setup main tabs and sub-tabs
setupTabs('.tab-btn', '.tab-content');
setupTabs('.sub-tab-btn', '.sub-tab-content');

// Form Submission Handler
const handleFormSubmission = (formId, endpoint) => {
    const form = document.getElementById(formId);
    const resultContent = form.closest('.sub-tab-content').querySelector('.result-content');

    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(form);

        // Show loading state
        resultContent.innerHTML = '<p>Processing... Please wait.</p>';

        try {
            const response = await fetch(endpoint, {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || 'Operation failed');
            }

            // Handle different response types based on endpoint
            const data = endpoint.includes('/encode/') 
                ? await response.blob() 
                : await response.json();

            // Process response
            if (endpoint.includes('/encode/')) {
                const url = URL.createObjectURL(data);
                resultContent.innerHTML = `
                    <p>Message encoded successfully!</p>
                    <a href="${url}" download="stego_video.avi" class="download-btn">
                        Download Video
                    </a>
                `;
            } else {
                resultContent.innerHTML = `
                    <p>Message decoded successfully!</p>
                    <div class="message-box">${data.data.message}</div>
                `;
            }
        } catch (error) {
            console.error('Form submission error:', error);
            resultContent.innerHTML = `<p class="error">Error: ${error.message}</p>`;
        }
    });
};

// Setup form submissions
handleFormSubmission('encode-text-form', '/api/video/encode/text');
handleFormSubmission('decode-text-form', '/api/video/decode/text');
});