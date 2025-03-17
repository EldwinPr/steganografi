document.addEventListener("DOMContentLoaded", function() {
    // Tab switching
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
    
    // Image preview for encode
    const encodeImageInput = document.getElementById("encode-image");
    const encodePreview = document.getElementById("encode-preview");
    
    encodeImageInput.addEventListener("change", function() {
        displayImagePreview(this, encodePreview);
    });
    
    // Image preview for decode
    const decodeImageInput = document.getElementById("decode-image");
    const decodePreview = document.getElementById("decode-preview");
    
    decodeImageInput.addEventListener("change", function() {
        displayImagePreview(this, decodePreview);
    });
    
    // Encode form submission
    const encodeForm = document.getElementById("encode-form");
    const encodeResult = document.getElementById("encode-result");
    
    encodeForm.addEventListener("submit", function(e) {
        e.preventDefault();
        const formData = new FormData(this);
        
        // Show loading state
        encodeResult.querySelector(".result-content").innerHTML = 
            "<p>Encoding your message... Please wait.</p>";
        
        fetch("/api/encode", {
            method: "POST",
            body: formData
        })
        .then(response => {
            if (response.ok) {
                return response.blob();
            } else {
                return response.json().then(data => {
                    throw new Error(data.message || "Failed to encode message");
                });
            }
        })
        .then(blob => {
            // Create download link for the encoded image
            const url = URL.createObjectURL(blob);
            const resultContent = encodeResult.querySelector(".result-content");
            
            resultContent.innerHTML = `
                <p>Message encoded successfully!</p>
                <div class="image-preview">
                    <img src="${url}" alt="Encoded image">
                </div>
                <a href="${url}" download="stego_image.png" class="btn" style="margin-top: 15px;">
                    Download Encoded Image
                </a>
            `;
        })
        .catch(error => {
            encodeResult.querySelector(".result-content").innerHTML = 
                `<p class="error">Error: ${error.message}</p>`;
        });
    });
    
    // Decode form submission
    const decodeForm = document.getElementById("decode-form");
    const decodeResult = document.getElementById("decode-result");
    
    decodeForm.addEventListener("submit", function(e) {
        e.preventDefault();
        const formData = new FormData(this);
        
        // Show loading state
        decodeResult.querySelector(".result-content").innerHTML = 
            "<p>Decoding message... Please wait.</p>";
        
        fetch("/api/decode", {
            method: "POST",
            body: formData
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                decodeResult.querySelector(".result-content").innerHTML = `
                    <p>Message decoded successfully:</p>
                    <div class="message-box">${data.data.message}</div>
                `;
            } else {
                throw new Error(data.message);
            }
        })
        .catch(error => {
            decodeResult.querySelector(".result-content").innerHTML = 
                `<p class="error">Error: ${error.message}</p>`;
        });
    });
    
    // Helper function to display image preview
    function displayImagePreview(input, previewElement) {
        if (input.files && input.files[0]) {
            const reader = new FileReader();
            
            reader.onload = function(e) {
                previewElement.innerHTML = `<img src="${e.target.result}" alt="Preview">`;
            };
            
            reader.readAsDataURL(input.files[0]);
        } else {
            previewElement.innerHTML = "No image selected";
        }
    }
});
