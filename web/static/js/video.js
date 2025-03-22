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
                            <video controls width="100%" height="auto">
                                <source src="${URL.createObjectURL(file)}" type="video/x-msvideo">
                                Your browser does not support the video tag.
                            </video>
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
    setupVideoPreview("encode-file-video", "encode-file-video-preview");
    setupVideoPreview("decode-text-video", "decode-text-preview");
    setupVideoPreview("decode-file-video", "decode-file-preview");
    
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
    
    // Format file size helper function
    function formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }
    
    // Setup file info displays
    setupFileInfo("encode-file-file", "encode-file-info");
    
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
    setupFormSubmission("encode-file-form", "/api/video/encode/file", handleEncodeFileResponse);
    setupFormSubmission("decode-text-form", "/api/video/decode/text", handleDecodeTextResponse);
    setupFormSubmission("decode-file-form", "/api/video/decode/file", handleDecodeFileResponse);
    
    // Response handlers
    function handleEncodeTextResponse(blob, resultContent) {
        // Create download link for the encoded video
        const url = URL.createObjectURL(blob);
        
        resultContent.innerHTML = `
            <p>Message encoded successfully!</p>
            <div class="video-preview">
                <video controls width="100%" height="auto">
                    <source src="${url}" type="video/x-msvideo">
                    Your browser does not support the video tag.
                </video>
            </div>
            <a href="${url}" download="stego_video.avi" class="btn" style="margin-top: 15px;">
                Download Encoded Video
            </a>
        `;
    }
    
    function handleEncodeFileResponse(blob, resultContent) {
        // Create download link for the encoded video
        const url = URL.createObjectURL(blob);
        
        resultContent.innerHTML = `
            <p>File hidden successfully!</p>
            <div class="video-preview">
                <video controls width="100%" height="auto">
                    <source src="${url}" type="video/x-msvideo">
                    Your browser does not support the video tag.
                </video>
            </div>
            <a href="${url}" download="stego_video.avi" class="btn" style="margin-top: 15px;">
                Download Encoded Video
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
            const fileType = data.data.fileType || "";
            const fileUrl = `data:${fileType};base64,${fileData}`;
            
            let filePreview = "";
            
            // Create appropriate preview based on file type
            if (fileType.startsWith("image/")) {
                filePreview = `<img src="${fileUrl}" alt="${fileName}" class="file-preview-image">`;
            } else if (fileType.startsWith("audio/")) {
                filePreview = `<audio controls class="file-preview-audio"><source src="${fileUrl}" type="${fileType}">Your browser does not support audio playback.</audio>`;
            } else if (fileType.startsWith("video/")) {
                filePreview = `<video controls class="file-preview-video"><source src="${fileUrl}" type="${fileType}">Your browser does not support video playback.</video>`;
            } else if (fileType === "application/pdf") {
                filePreview = `<iframe src="${fileUrl}" class="file-preview-pdf"></iframe>`;
            } else if (fileType.startsWith("text/") || fileType === "application/json") {
                // For text files, we'll need to decode and display
                const textContent = atob(fileData);
                filePreview = `<pre class="file-preview-text">${textContent}</pre>`;
            } else {
                filePreview = `<p>Preview not available for this file type.</p>`;
            }
            
            resultContent.innerHTML = `
                <p>File extracted successfully!</p>
                <div class="file-info">
                    <strong>File:</strong> ${fileName}<br>
                    <strong>Type:</strong> ${fileType || "Unknown"}<br>
                </div>
                <div class="file-preview">
                    ${filePreview}
                </div>
                <a href="${fileUrl}" download="${fileName}" class="btn" style="margin-top: 15px;">
                    Download Extracted File
                </a>
            `;
        } else {
            throw new Error(data.message);
        }
    }
});