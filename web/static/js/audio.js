document.addEventListener('DOMContentLoaded', function() {
    // Tab switching functionality
    const tabs = document.querySelectorAll('.tab-btn');
    const tabContents = document.querySelectorAll('.tab-content');
    
    tabs.forEach(tab => {
        tab.addEventListener('click', function() {
            // Remove active class from all tabs and contents
            tabs.forEach(t => t.classList.remove('active'));
            tabContents.forEach(content => content.classList.remove('active'));
            
            // Add active class to clicked tab and corresponding content
            this.classList.add('active');
            const targetId = this.getAttribute('data-tab') + '-tab';
            document.getElementById(targetId).classList.add('active');
        });
    });
    
    // Sub-tab switching functionality
    const subTabs = document.querySelectorAll('.sub-tab-btn');
    const subTabContents = document.querySelectorAll('.sub-tab-content');
    
    subTabs.forEach(subTab => {
        subTab.addEventListener('click', function() {
            // Get parent tab content
            const parentTabContent = this.closest('.tab-content');
            
            // Remove active class from all sub-tabs and contents within this parent
            const siblingSubTabs = parentTabContent.querySelectorAll('.sub-tab-btn');
            siblingSubTabs.forEach(t => t.classList.remove('active'));
            
            const siblingSubContents = parentTabContent.querySelectorAll('.sub-tab-content');
            siblingSubContents.forEach(content => content.classList.remove('active'));
            
            // Add active class to clicked sub-tab and corresponding content
            this.classList.add('active');
            const targetId = this.getAttribute('data-subtab') + '-tab';
            document.getElementById(targetId).classList.add('active');
        });
    });
    
    // File input change handler for capacity calculation
    const fileInputs = document.querySelectorAll('input[type="file"]');
    fileInputs.forEach(input => {
        input.addEventListener('change', function(e) {
            const fileLabel = this.nextElementSibling;
            if (fileLabel && fileLabel.classList.contains('file-name')) {
                if (this.files.length > 0) {
                    fileLabel.textContent = this.files[0].name;
                    
                    // Calculate capacity if this is an encode form
                    if (this.form.id.includes('encode')) {
                        calculateCapacity(this.files[0]);
                    }
                } else {
                    fileLabel.textContent = 'No file selected';
                }
            }
        });
    });
    
    // Function to calculate and display capacity
    function calculateCapacity(file) {
        // You can implement capacity calculation logic here
        // For now, we'll just show a placeholder message
        const capacityInfo = document.querySelector('.capacity-info');
        if (capacityInfo) {
            capacityInfo.textContent = `File selected: ${file.name} (${(file.size / 1024).toFixed(2)} KB)`;
        }
    }
    
    // Handle encode form submission
    const encodeTextForm = document.getElementById("encode-text-audio-form");
    if (encodeTextForm) {
        encodeTextForm.addEventListener("submit", function(e) {
            e.preventDefault();
            
            const formData = new FormData(this);
            const audioFile = formData.get("audio");
            
            // Validate file is a WAV
            if (!audioFile.name.toLowerCase().endsWith('.wav')) {
                alert("Only WAV files are supported. Please select a valid WAV file.");
                return;
            }
            
            // Check if seed is empty, if so set it to -1
            const seedInput = formData.get("seed");
            if (!seedInput || seedInput.trim() === "") {
                formData.set("seed", "-1");
            }
            
            // Remove bits per sample parameter if our API doesn't use it
            if (formData.has("bitsPerSample")) {
                formData.delete("bitsPerSample");
            }
            
            // Show loading indicator
            const loadingIndicator = document.createElement("div");
            loadingIndicator.className = "loading-indicator";
            loadingIndicator.innerHTML = "<p>Processing... Please wait</p>";
            
            // Clear any previous results
            const resultContainer = document.querySelector('.encode-result-container');
            if (resultContainer) {
                resultContainer.innerHTML = '';
                resultContainer.appendChild(loadingIndicator);
            } else {
                const oldResult = encodeTextForm.nextElementSibling;
                if (oldResult && (oldResult.className === "encode-result" || oldResult.className === "loading-indicator")) {
                    oldResult.remove();
                }
                encodeTextForm.after(loadingIndicator);
            }
            
            // Use fetch with blob response type for direct file download
            fetch(`/api/audio/encode/text`, {
                method: "POST",
                body: formData
            })
            .then(response => {
                if (!response.ok) {
                    if (response.headers.get("Content-Type")?.includes("application/json")) {
                        return response.json().then(errorData => {
                            throw new Error(errorData.message || "Server returned an error");
                        });
                    }
                    throw new Error(`Server error: ${response.status}`);
                }
                return response.blob(); // Get response as blob for download
            })
            .then(blob => {
                // Remove loading indicator
                const loadingIndicator = document.querySelector('.loading-indicator');
                if (loadingIndicator) {
                    loadingIndicator.remove();
                }
                
                // Create blob URL for the audio file
                const blobUrl = URL.createObjectURL(blob);
                
                // Create result container
                const resultDiv = document.createElement("div");
                resultDiv.className = "encode-result";
                
                // Create success message
                const successMsg = document.createElement("h3");
                successMsg.className = "success-message";
                successMsg.innerText = "Encoding Successful!";
                resultDiv.appendChild(successMsg);
                
                // Create audio container
                const audioContainer = document.createElement("div");
                audioContainer.className = "audio-container";
                
                // Create audio player
                const audioPlayer = document.createElement("audio");
                audioPlayer.controls = true;
                audioPlayer.className = "audio-player";
                audioPlayer.src = blobUrl;
                audioContainer.appendChild(audioPlayer);
                
                // Create download button
                const downloadLink = document.createElement("a");
                downloadLink.href = blobUrl;
                downloadLink.download = "encoded-audio.wav";
                downloadLink.className = "download-btn";
                downloadLink.innerText = "Download Encoded Audio";
                audioContainer.appendChild(downloadLink);
                
                // Add audio container to result div
                resultDiv.appendChild(audioContainer);
                
                // Add result to page
                const resultContainer = document.querySelector('.encode-result-container');
                if (resultContainer) {
                    resultContainer.appendChild(resultDiv);
                } else {
                    encodeTextForm.after(resultDiv);
                }
            })
            .catch(error => {
                // Remove loading indicator
                const loadingIndicator = document.querySelector('.loading-indicator');
                if (loadingIndicator) {
                    loadingIndicator.remove();
                }
                
                // Show error message
                const errorDiv = document.createElement("div");
                errorDiv.className = "error-message";
                errorDiv.innerText = `Error: ${error.message}`;
                
                const resultContainer = document.querySelector('.encode-result-container');
                if (resultContainer) {
                    resultContainer.appendChild(errorDiv);
                } else {
                    encodeTextForm.after(errorDiv);
                }
            });
        });
    }
    
    // Handle decode form submission
    const decodeTextForm = document.getElementById("decode-text-audio-form");
    if (decodeTextForm) {
        decodeTextForm.addEventListener("submit", function(e) {
            e.preventDefault();
            
            const formData = new FormData(this);
            const audioFile = formData.get("audio");
            
            // Validate file is a WAV
            if (!audioFile.name.toLowerCase().endsWith('.wav')) {
                alert("Only WAV files are supported. Please select a valid WAV file.");
                return;
            }
            
            // Check if seed is empty, if so set it to -1
            const seedInput = formData.get("seed");
            if (!seedInput || seedInput.trim() === "") {
                formData.set("seed", "-1");
            }
            
            // Remove bits per sample parameter if our API doesn't use it
            if (formData.has("bitsPerSample")) {
                formData.delete("bitsPerSample");
            }
            
            // Show loading indicator
            const loadingIndicator = document.createElement("div");
            loadingIndicator.className = "loading-indicator";
            loadingIndicator.innerHTML = "<p>Processing... Please wait</p>";
            
            // Clear any previous results
            const resultContainer = document.querySelector('.decode-result-container');
            if (resultContainer) {
                resultContainer.innerHTML = '';
                resultContainer.appendChild(loadingIndicator);
            } else {
                const oldResult = decodeTextForm.nextElementSibling;
                if (oldResult && (oldResult.className === "decode-result" || oldResult.className === "loading-indicator")) {
                    oldResult.remove();
                }
                decodeTextForm.after(loadingIndicator);
            }
            
            fetch(`/api/audio/decode/text`, {
                method: "POST",
                body: formData
            })
            .then(response => {
                if (!response.ok) {
                    return response.json().then(errorData => {
                        throw new Error(errorData.message || "Server returned an error");
                    });
                }
                return response.json();
            })
            .then(data => {
                // Remove loading indicator
                const loadingIndicator = document.querySelector('.loading-indicator');
                if (loadingIndicator) {
                    loadingIndicator.remove();
                }
                
                // Create result container
                const resultDiv = document.createElement("div");
                resultDiv.className = "decode-result";
                
                // Add decoded message
                const successMsg = document.createElement("h3");
                successMsg.className = "success-message";
                successMsg.innerText = "Decoding Successful!";
                resultDiv.appendChild(successMsg);
                
                const messageBox = document.createElement("div");
                messageBox.className = "message-box";
                messageBox.innerText = data.data.message;
                resultDiv.appendChild(messageBox);
                
                // Add copy button
                const copyBtn = document.createElement("button");
                copyBtn.className = "copy-btn";
                copyBtn.innerText = "Copy Message";
                copyBtn.addEventListener("click", function() {
                    navigator.clipboard.writeText(data.data.message)
                        .then(() => {
                            copyBtn.innerText = "Copied!";
                            setTimeout(() => {
                                copyBtn.innerText = "Copy Message";
                            }, 2000);
                        });
                });
                resultDiv.appendChild(copyBtn);
                
                // Add result to page
                const resultContainer = document.querySelector('.decode-result-container');
                if (resultContainer) {
                    resultContainer.appendChild(resultDiv);
                } else {
                    decodeTextForm.after(resultDiv);
                }
            })
            .catch(error => {
                // Remove loading indicator
                const loadingIndicator = document.querySelector('.loading-indicator');
                if (loadingIndicator) {
                    loadingIndicator.remove();
                }
                
                // Show error message
                const errorDiv = document.createElement("div");
                errorDiv.className = "error-message";
                errorDiv.innerText = `Error: ${error.message}`;
                
                const resultContainer = document.querySelector('.decode-result-container');
                if (resultContainer) {
                    resultContainer.appendChild(errorDiv);
                } else {
                    decodeTextForm.after(errorDiv);
                }
            });
        });
    }
    
    // Initialize - activate first tab by default
    if (tabs.length > 0) {
        tabs[0].click();
    }
});
