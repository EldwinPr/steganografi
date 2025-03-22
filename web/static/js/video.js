document.addEventListener("DOMContentLoaded", function() {
    // Main tab switching
    const tabButtons = document.querySelectorAll(".tab-btn");
    const tabContents = document.querySelectorAll(".tab-content");
    
    tabButtons.forEach(button => {
        button.addEventListener("click", () => {
            // Remove active class from all buttons and contents
            tabButtons.forEach(btn => btn.classList.remove("active"));
            tabContents.forEach(content => content.classList.remove("active"));
            
            // Add active class to clicked button and corresponding content
            button.classList.add("active");
            const tabId = button.getAttribute("data-tab");
            document.getElementById(`${tabId}-tab`).classList.add("active");
        });
    });
    
    // Sub-tab switching
    const subTabButtons = document.querySelectorAll(".sub-tab-btn");
    
    subTabButtons.forEach(button => {
        button.addEventListener("click", () => {
            // Find parent tab content
            const parentTab = button.closest(".tab-content");
            
            // Remove active class from all sub-tab buttons and contents within this parent
            parentTab.querySelectorAll(".sub-tab-btn").forEach(btn => btn.classList.remove("active"));
            parentTab.querySelectorAll(".sub-tab-content").forEach(content => content.classList.remove("active"));
            
            // Add active class to clicked button and corresponding content
            button.classList.add("active");
            const subTabId = button.getAttribute("data-subtab");
            document.getElementById(`${subTabId}-tab`).classList.add("active");
        });
    });
    
    // Setup video previews
    function setupVideoPreview(inputId, previewId) {
        const input = document.getElementById(inputId);
        const preview = document.getElementById(previewId);
        
        if (input && preview) {
            input.addEventListener("change", function() {
                if (this.files && this.files[0]) {
                    const file = this.files[0];
                    
                    // Check if it's a video file
                    if (file.type.startsWith('video/') || file.name.endsWith('.avi')) {
                        preview.innerHTML = `
                            <div class="image-preview">
                                <video controls width="100%" height="auto">
                                    <source src="${URL.createObjectURL(file)}" type="video/x-msvideo">
                                    Your browser does not support the video tag.
                                </video>
                            </div>
                            <div class="file-info">
                                <strong>File:</strong> ${file.name}<br>
                                <strong>Size:</strong> ${formatFileSize(file.size)}
                            </div>
                        `;
                    } else {
                        preview.innerHTML = "Selected file is not a valid video format. Please use AVI format.";
                    }
                } else {
                    preview.innerHTML = "No video selected";
                }
            });
        }
    }
    
    // Setup all video previews
    setupVideoPreview("encode-text-video", "encode-text-preview");
    setupVideoPreview("decode-text-video", "decode-text-preview");
    
    // Format file size helper function
    function formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }
    
    // Form submission setup
    function setupFormSubmission(formId, endpoint, responseHandler) {
        const form = document.getElementById(formId);
        
        if (form) {
            form.addEventListener("submit", function(e) {
                e.preventDefault();
                const formData = new FormData(this);
                
                // Find the parent tab and sub-tab to determine the correct result element
                const parentTab = form.closest(".tab-content");
                const subTab = form.closest(".sub-tab-content");
                
                // Get the result element
                const resultId = subTab.id.replace("-tab", "-result");
                const resultContent = document.querySelector(`#${resultId} .result-content`);
                
                // Show loading state
                resultContent.innerHTML = "<p>Processing... Please wait.</p>";
                
                fetch(endpoint, {
                    method: "POST",
                    body: formData
                })
                .then(response => {
                    if (response.ok) {
                        if (endpoint.includes("/encode/")) {
                            return response.blob();
                        } else {
                            return response.json();
                        }
                    } else {
                        return response.json().then(data => {
                            throw new Error(data.message || "Operation failed");
                        });
                    }
                })
                .then(data => responseHandler(data, resultContent))
                .catch(error => {
                    resultContent.innerHTML = `<p class="error">Error: ${error.message}</p>`;
                    console.error("Form submission error:", error);
                });
            });
        } else {
            console.warn(`Form with ID ${formId} not found`);
        }
    }
    
    // Setup form submissions for video steganography
    setupFormSubmission("encode-text-form", "/api/video/encode/text", handleEncodeTextResponse);
    setupFormSubmission("decode-text-form", "/api/video/decode/text", handleDecodeTextResponse);
    
    // Response handlers
    function handleEncodeTextResponse(blob, resultContent) {
        const url = URL.createObjectURL(blob);
        resultContent.innerHTML = `
            <p>Message encoded successfully!</p>
            <div class="image-preview">
                <video controls width="100%" height="auto">
                    <source src="${url}" type="video/x-msvideo">
                    Your browser does not support the video tag.
                </video>
            </div>
            <a href="${url}" download="stego_video.avi" class="download-btn">Download Video</a>
        `;
    }
    
    function handleDecodeTextResponse(data, resultContent) {
        resultContent.innerHTML = `
            <p>Message decoded successfully!</p>
            <div class="message-box">${data.data.message}</div>
        `;
    }
});
