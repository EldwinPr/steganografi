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
    const subTabContents = document.querySelectorAll(".sub-tab-content");
    
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
    
    // Image previews
    setupImagePreview("encode-text-image", "encode-text-preview");
    setupImagePreview("encode-file-image", "encode-file-image-preview");
    setupImagePreview("decode-text-image", "decode-text-preview");
    setupImagePreview("decode-file-image", "decode-file-preview");
    
    // File info display
    const fileInput = document.getElementById("encode-file-file");
    const fileInfo = document.getElementById("encode-file-info");
    
    if (fileInput) {
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
    
    // Form submissions
    setupFormSubmission("encode-text-form", "/api/encode/text", handleEncodeTextResponse);
    setupFormSubmission("encode-file-form", "/api/encode/file", handleEncodeFileResponse);
    setupFormSubmission("decode-text-form", "/api/decode/text", handleDecodeTextResponse);
    setupFormSubmission("decode-file-form", "/api/decode/file", handleDecodeFileResponse);
    
    // Helper functions
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
    
    function setupFormSubmission(formId, endpoint, responseHandler) {
        const form = document.getElementById(formId);
        
        if (form) {
            form.addEventListener("submit", function(e) {
                e.preventDefault();
                const formData = new FormData(this);
                
                // Get the result element
                const resultId = formId.replace("form", "result");
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
                });
            });
        }
    }
    
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
});
