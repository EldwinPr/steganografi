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
            
            // Make sure the correct method content is shown
            updateActiveMethodContent(parentTab);
        });
    });

    // Method tab switching
    const methodTabButtons = document.querySelectorAll(".method-tab-btn");
    
    methodTabButtons.forEach(button => {
        button.addEventListener("click", () => {
            // Find the closest tabs container
            const tabsContainer = button.closest(".tabs, .method-tabs");
            // Find all method buttons in this container and remove active class
            tabsContainer.querySelectorAll(".method-tab-btn").forEach(btn => 
                btn.classList.remove("active"));
            // Add active class to clicked button
            button.classList.add("active");
            
            // Find parent tab content
            const parentTab = button.closest(".tab-content");
            updateActiveMethodContent(parentTab);
        });
    });
    
    function updateActiveMethodContent(parentTab) {
        // Find active sub-tab content
        const activeSubTab = parentTab.querySelector(".sub-tab-content.active");
        if (!activeSubTab) return;
        
        // Get active method
        const activeMethod = parentTab.querySelector(".method-tab-btn.active")?.getAttribute("data-method");
        if (!activeMethod) return;
        
        // Get sub-tab ID
        const subTabId = activeSubTab.id.replace("-tab", "");
        
        // Hide all method contents in this sub-tab
        activeSubTab.querySelectorAll(".method-content").forEach(content => 
            content.classList.remove("active"));
        
        // Show the correct method content
        const methodContentId = `${subTabId}-${activeMethod}`;
        const methodContent = document.getElementById(methodContentId);
        if (methodContent) {
            methodContent.classList.add("active");
        }
    }
    
    // Setup image previews
    function setupImagePreview(inputId, previewId) {
        const input = document.getElementById(inputId);
        const preview = document.getElementById(previewId);
        
        if (input && preview) {
            input.addEventListener("change", function() {
                if (this.files && this.files[0]) {
                    const reader = new FileReader();
                    
                    reader.onload = function(e) {
                        preview.innerHTML = `<img src="${e.target.result}" alt="Preview">`;
                    };
                    
                    reader.readAsDataURL(this.files[0]);
                } else {
                    preview.innerHTML = "No image selected";
                }
            });
        }
    }

    // Setup all image previews
    setupImagePreview("encode-text-lsb-image", "encode-text-lsb-preview");
    setupImagePreview("encode-text-bpcs-image", "encode-text-bpcs-preview");
    setupImagePreview("encode-file-lsb-image", "encode-file-lsb-image-preview");
    setupImagePreview("encode-file-bpcs-image", "encode-file-bpcs-image-preview");
    setupImagePreview("decode-text-lsb-image", "decode-text-lsb-preview");
    setupImagePreview("decode-text-bpcs-image", "decode-text-bpcs-preview");
    setupImagePreview("decode-file-lsb-image", "decode-file-lsb-preview");
    setupImagePreview("decode-file-bpcs-image", "decode-file-bpcs-preview");
    
    // Helper function for file info display
    function setupFileInfo(inputId, infoId) {
        const fileInput = document.getElementById(inputId);
        const fileInfo = document.getElementById(infoId);
        
        if (fileInput && fileInfo) {
            fileInput.addEventListener("change", function() {
                if (this.files && this.files[0]) {
                    const file = this.files[0];
                    const fileSize = formatFileSize(file.size);
                    fileInfo.innerHTML = `
                        <strong>File:</strong> ${file.name}<br>
                        <strong>Type:</strong> ${file.type || "Unknown"}<br>
                        <strong>Size:</strong> ${fileSize}
                    `;
                } else {
                    fileInfo.innerHTML = "No file selected";
                }
            });
        }
    }
    
    // Setup file info displays
    setupFileInfo("encode-file-lsb-file", "encode-file-lsb-info");
    setupFileInfo("encode-file-bpcs-file", "encode-file-bpcs-info");
    
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
    
    // Setup form submissions for LSB and BPCS methods
    setupFormSubmission("encode-text-lsb-form", "/api/lsb/encode/text", handleEncodeTextResponse);
    setupFormSubmission("encode-file-lsb-form", "/api/lsb/encode/file", handleEncodeFileResponse);
    setupFormSubmission("decode-text-lsb-form", "/api/lsb/decode/text", handleDecodeTextResponse);
    setupFormSubmission("decode-file-lsb-form", "/api/lsb/decode/file", handleDecodeFileResponse);
    
    setupFormSubmission("encode-text-bpcs-form", "/api/bpcs/encode/text", handleEncodeTextResponse);
    setupFormSubmission("encode-file-bpcs-form", "/api/bpcs/encode/file", handleEncodeFileResponse);
    setupFormSubmission("decode-text-bpcs-form", "/api/bpcs/decode/text", handleDecodeTextResponse);
    setupFormSubmission("decode-file-bpcs-form", "/api/bpcs/decode/file", handleDecodeFileResponse);
    
    // Response handlers
    function handleEncodeTextResponse(blob, resultContent) {
        // Create download link for the encoded image
        const url = URL.createObjectURL(blob);
        
        resultContent.innerHTML = `
            <p>Message encoded successfully!</p>
            <div class="image-preview">
                <img src="${url}" alt="Encoded image">
            </div>
            <a href="${url}" download="stego_image.png" class="btn" style="margin-top: 15px;">
                Download Encoded Image
            </a>
        `;
    }
    
    function handleEncodeFileResponse(blob, resultContent) {
        // Create download link for the encoded image
        const url = URL.createObjectURL(blob);
        
        resultContent.innerHTML = `
            <p>File hidden successfully!</p>
            <div class="image-preview">
                <img src="${url}" alt="Encoded image">
            </div>
            <a href="${url}" download="stego_image.png" class="btn" style="margin-top: 15px;">
                Download Encoded Image
            </a>
        `;
    }
    
    function handleDecodeTextResponse(data, resultContent) {
        if (data.success) {
            resultContent.innerHTML = `
                <p>Message decoded successfully:</p>
                <div class="message-box">${data.data.message}</div>
            `;
        } else {
            throw new Error(data.message);
        }
    }
    
    function handleDecodeFileResponse(data, resultContent) {
        if (data.success) {
            const fileData = data.data.fileData;
            const fileName = data.data.fileName;
            const fileExt = data.data.fileExt.toLowerCase();
            const fileSize = formatFileSize(data.data.fileSize);
            
            // Convert base64 to blob
            const byteCharacters = atob(fileData);
            const byteArrays = [];
            
            for (let offset = 0; offset < byteCharacters.length; offset += 512) {
                const slice = byteCharacters.slice(offset, offset + 512);
                
                const byteNumbers = new Array(slice.length);
                for (let i = 0; i < slice.length; i++) {
                    byteNumbers[i] = slice.charCodeAt(i);
                }
                
                const byteArray = new Uint8Array(byteNumbers);
                byteArrays.push(byteArray);
            }
            
            const blob = new Blob(byteArrays);
            const url = URL.createObjectURL(blob);
            
            // Determine file type and create appropriate preview
            let filePreview = '';
            
            // Image preview
            if (['.jpg', '.jpeg', '.png', '.gif', '.bmp', '.webp'].includes(fileExt)) {
                filePreview = `<img src="${url}" alt="${fileName}" class="file-preview-image">`;
            }
            // PDF preview
            else if (fileExt === '.pdf') {
                filePreview = `<iframe src="${url}" class="file-preview-pdf"></iframe>`;
            }
            // Audio preview
            else if (['.mp3', '.wav', '.ogg'].includes(fileExt)) {
                filePreview = `<audio controls class="file-preview-audio"><source src="${url}"></audio>`;
            }
            // Video preview
            else if (['.mp4', '.webm', '.ogg'].includes(fileExt)) {
                filePreview = `<video controls class="file-preview-video"><source src="${url}"></video>`;
            }
            // Text preview (for txt, html, css, js, etc.)
            else if (['.txt', '.html', '.css', '.js', '.json', '.xml', '.md'].includes(fileExt)) {
                // For text files, read and display content
                const reader = new FileReader();
                reader.onload = function(e) {
                    document.querySelector('.file-preview-text').textContent = e.target.result;
                };
                reader.readAsText(blob);
                filePreview = `<div class="file-preview-text">Loading text content...</div>`;
            }
            // No preview for other file types
            else {
                filePreview = `<p>No preview available for this file type.</p>`;
            }
            
            resultContent.innerHTML = `
                <p>File extracted successfully!</p>
                <div class="file-info">
                    <strong>File:</strong> ${fileName}<br>
                    <strong>Type:</strong> ${fileExt}<br>
                    <strong>Size:</strong> ${fileSize}
                </div>
                <a href="${url}" download="${fileName}" class="download-btn">
                    Download File
                </a>
                <div class="file-preview">
                    ${filePreview}
                </div>
            `;
        } else {
            throw new Error(data.message);
        }
    }
    
    function formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }
    
    // Initialize the correct method content for each tab
    document.querySelectorAll(".tab-content").forEach(updateActiveMethodContent);
});
