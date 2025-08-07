// Authentication JavaScript for Turnate
class AuthManager {
    constructor(app) {
        this.app = app;
        this.setupEventListeners();
    }
    
    setupEventListeners() {
        // Toggle between login and register forms
        $('#showRegister').on('click', (e) => {
            e.preventDefault();
            this.showRegisterForm();
        });
        
        $('#showLogin').on('click', (e) => {
            e.preventDefault();
            this.showLoginForm();
        });
        
        // Form submissions
        $('#loginFormElement').on('submit', (e) => {
            e.preventDefault();
            this.handleLogin();
        });
        
        $('#registerFormElement').on('submit', (e) => {
            e.preventDefault();
            this.handleRegister();
        });
    }
    
    showLoginForm() {
        $('#registerForm').addClass('d-none');
        $('#loginForm').removeClass('d-none');
        this.clearMessages();
    }
    
    showRegisterForm() {
        $('#loginForm').addClass('d-none');
        $('#registerForm').removeClass('d-none');
        this.clearMessages();
    }
    
    async handleLogin() {
        const username = $('#loginUsername').val().trim();
        const password = $('#loginPassword').val();
        
        if (!username || !password) {
            this.showError('Please fill in all fields');
            return;
        }
        
        try {
            this.setLoading(true);
            this.clearMessages();
            
            const response = await fetch('/api/v1/auth/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    username: username,
                    password: password
                })
            });
            
            const result = await response.json();
            
            if (response.ok) {
                // Store token
                localStorage.setItem('turnate_token', result.token);
                this.app.currentToken = result.token;
                this.app.currentUser = result.user;
                
                this.showSuccess(result.message || 'Login successful! 🎉');
                
                // Close modal and load app
                setTimeout(() => {
                    this.app.updateUserUI();
                    this.app.hideAuthModal();
                    this.app.loadChannels();
                    this.app.startPolling();
                }, 1000);
                
            } else {
                this.showError(result.error || 'Login failed');
            }
            
        } catch (error) {
            console.error('Login error:', error);
            this.showError('Network error. Please check your connection.');
        } finally {
            this.setLoading(false);
        }
    }
    
    async handleRegister() {
        const username = $('#registerUsername').val().trim();
        const email = $('#registerEmail').val().trim();
        const displayName = $('#registerDisplayName').val().trim();
        const password = $('#registerPassword').val();
        
        if (!username || !email || !password) {
            this.showError('Please fill in all required fields');
            return;
        }
        
        if (password.length < 6) {
            this.showError('Password must be at least 6 characters long');
            return;
        }
        
        if (!this.isValidEmail(email)) {
            this.showError('Please enter a valid email address');
            return;
        }
        
        if (!this.isValidUsername(username)) {
            this.showError('Username can only contain letters, numbers, and underscores');
            return;
        }
        
        try {
            this.setLoading(true);
            this.clearMessages();
            
            const response = await fetch('/api/v1/auth/register', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    username: username,
                    email: email,
                    display_name: displayName || username,
                    password: password
                })
            });
            
            const result = await response.json();
            
            if (response.ok) {
                // Store token
                localStorage.setItem('turnate_token', result.token);
                this.app.currentToken = result.token;
                this.app.currentUser = result.user;
                
                this.showSuccess(result.message || 'Registration successful! 🎉');
                
                // Close modal and load app
                setTimeout(() => {
                    this.app.updateUserUI();
                    this.app.hideAuthModal();
                    this.app.loadChannels();
                    this.app.startPolling();
                }, 1000);
                
            } else {
                this.showError(result.error || 'Registration failed');
            }
            
        } catch (error) {
            console.error('Registration error:', error);
            this.showError('Network error. Please check your connection.');
        } finally {
            this.setLoading(false);
        }
    }
    
    setLoading(isLoading) {
        const loginBtn = $('#loginFormElement button[type="submit"]');
        const registerBtn = $('#registerFormElement button[type="submit"]');
        
        if (isLoading) {
            loginBtn.prop('disabled', true).html('<span class="spinner-border spinner-border-sm me-2"></span>Signing In...');
            registerBtn.prop('disabled', true).html('<span class="spinner-border spinner-border-sm me-2"></span>Creating Account...');
        } else {
            loginBtn.prop('disabled', false).html('<i class="bi bi-box-arrow-in-right"></i> Sign In');
            registerBtn.prop('disabled', false).html('<i class="bi bi-person-plus"></i> Create Account');
        }
    }
    
    showError(message) {
        $('#authSuccess').addClass('d-none');
        $('#authError').removeClass('d-none').text(message);
    }
    
    showSuccess(message) {
        $('#authError').addClass('d-none');
        $('#authSuccess').removeClass('d-none').text(message);
    }
    
    clearMessages() {
        $('#authError, #authSuccess').addClass('d-none');
    }
    
    isValidEmail(email) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
    }
    
    isValidUsername(username) {
        const usernameRegex = /^[a-zA-Z0-9_]+$/;
        return usernameRegex.test(username) && username.length >= 3;
    }
}

// Initialize auth manager when app is ready
$(document).ready(() => {
    // Wait for the main app to be initialized
    setTimeout(() => {
        if (window.turnateApp) {
            window.authManager = new AuthManager(window.turnateApp);
        }
    }, 100);
});