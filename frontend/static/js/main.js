// Frontend Interaction and Form Handling
document.addEventListener('DOMContentLoaded', () => {
    // Form Validation
    const validateForm = (form) => {
        const inputs = form.querySelectorAll('input, textarea');
        let isValid = true;

        inputs.forEach(input => {
            if (input.hasAttribute('required') && !input.value.trim()) {
                input.classList.add('is-invalid');
                isValid = false;
            } else {
                input.classList.remove('is-invalid');
            }
        });

        return isValid;
    };

    // AJAX Form Submission
    const setupAjaxForms = () => {
        const forms = document.querySelectorAll('form[data-ajax]');
        
        forms.forEach(form => {
            form.addEventListener('submit', async (e) => {
                e.preventDefault();
                
                if (!validateForm(form)) return;

                const submitButton = form.querySelector('button[type="submit"]');
                submitButton.disabled = true;
                submitButton.innerHTML = 'Processing...';

                try {
                    const formData = new FormData(form);
                    const response = await fetch(form.action, {
                        method: form.method,
                        body: formData,
                        headers: {
                            'Accept': 'application/json'
                        }
                    });

                    const result = await response.json();

                    // Handle different response types
                    if (response.ok) {
                        showAlert('Success', result.message || 'Operation completed successfully');
                        
                        // Redirect or update UI based on form type
                        if (form.id === 'login-form') {
                            localStorage.setItem('token', result.token);
                            window.location.href = '/dashboard';
                        } else if (form.id === 'register-form') {
                            window.location.href = '/login';
                        }
                    } else {
                        showAlert('Error', result.error || 'An error occurred');
                    }
                } catch (error) {
                    showAlert('Error', 'Network error. Please try again.');
                    console.error('Submission error:', error);
                } finally {
                    submitButton.disabled = false;
                    submitButton.innerHTML = 'Submit';
                }
            });
        });
    };

    // Alert Messaging
    const showAlert = (type, message) => {
        const alertContainer = document.getElementById('alert-container');
        if (!alertContainer) return;

        alertContainer.innerHTML = `
            <div class="alert alert-${type.toLowerCase()}">
                ${message}
            </div>
        `;

        // Auto-dismiss after 5 seconds
        setTimeout(() => {
            alertContainer.innerHTML = '';
        }, 5000);
    };

    // Theme Toggle (Optional)
    const setupThemeToggle = () => {
        const themeToggle = document.getElementById('theme-toggle');
        if (themeToggle) {
            themeToggle.addEventListener('click', () => {
                document.body.classList.toggle('dark-mode');
                localStorage.setItem('theme', 
                    document.body.classList.contains('dark-mode') ? 'dark' : 'light'
                );
            });
        }
    };

    // Initialize Functions
    setupAjaxForms();
    setupThemeToggle();
});