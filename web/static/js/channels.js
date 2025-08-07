// Channel management JavaScript for Turnate
class ChannelManager {
    constructor(app) {
        this.app = app;
        this.setupEventListeners();
    }
    
    setupEventListeners() {
        // Create channel button
        $('#createChannelBtn').on('click', () => this.showCreateChannelModal());
        
        // Create channel form submission
        $('#confirmCreateChannel').on('click', () => this.handleCreateChannel());
        
        // Channel form validation
        $('#channelName').on('input', (e) => this.validateChannelName(e.target));
    }
    
    showCreateChannelModal() {
        $('#createChannelModal').modal('show');
        $('#channelName').val('');
        $('#channelDescription').val('');
        $('#channelType').val('public');
        this.clearCreateChannelErrors();
    }
    
    async handleCreateChannel() {
        const name = $('#channelName').val().trim();
        const description = $('#channelDescription').val().trim();
        const type = $('#channelType').val();
        
        if (!name) {
            this.showCreateChannelError('Channel name is required');
            return;
        }
        
        if (name.length < 2) {
            this.showCreateChannelError('Channel name must be at least 2 characters');
            return;
        }
        
        if (!this.isValidChannelName(name)) {
            this.showCreateChannelError('Channel name can only contain letters, numbers, and hyphens');
            return;
        }
        
        try {
            this.setCreateChannelLoading(true);
            
            const response = await this.app.makeRequest('/api/v1/channels', 'POST', {
                name: name,
                description: description,
                type: type
            });
            
            if (response.channel) {
                this.showCreateChannelSuccess(response.message || 'Channel created successfully! 🎉');
                
                // Reload channels
                await this.app.loadChannels();
                
                // Select the new channel
                const newChannel = {
                    ...response.channel,
                    is_member: true
                };
                this.app.selectChannel(newChannel);
                
                // Close modal
                setTimeout(() => {
                    $('#createChannelModal').modal('hide');
                }, 1000);
            }
            
        } catch (error) {
            console.error('Failed to create channel:', error);
            this.showCreateChannelError(error.message || 'Failed to create channel');
        } finally {
            this.setCreateChannelLoading(false);
        }
    }
    
    validateChannelName(input) {
        let value = input.value.toLowerCase();
        
        // Remove invalid characters and replace spaces with hyphens
        value = value.replace(/[^a-z0-9\-\s]/g, '').replace(/\s+/g, '-');
        
        // Remove multiple consecutive hyphens
        value = value.replace(/-+/g, '-');
        
        // Remove hyphens from start and end
        value = value.replace(/^-+|-+$/g, '');
        
        input.value = value;
        
        // Update character count or validation message
        const isValid = this.isValidChannelName(value) && value.length >= 2;
        input.classList.toggle('is-valid', isValid && value.length > 0);
        input.classList.toggle('is-invalid', !isValid && value.length > 0);
    }
    
    isValidChannelName(name) {
        const channelNameRegex = /^[a-z0-9][a-z0-9\-]*[a-z0-9]$|^[a-z0-9]$/;
        return channelNameRegex.test(name) && name.length <= 50;
    }
    
    setCreateChannelLoading(isLoading) {
        const submitBtn = $('#confirmCreateChannel');
        
        if (isLoading) {
            submitBtn.prop('disabled', true)
                     .html('<span class="spinner-border spinner-border-sm me-2"></span>Creating...');
        } else {
            submitBtn.prop('disabled', false)
                     .html('<i class="bi bi-plus-lg"></i> Create Channel');
        }
    }
    
    showCreateChannelError(message) {
        // Remove any existing alerts
        $('.create-channel-alert').remove();
        
        const alert = $(`
            <div class="alert alert-danger create-channel-alert mt-2" role="alert">
                <i class="bi bi-exclamation-triangle"></i> ${message}
            </div>
        `);
        
        $('#createChannelForm').after(alert);
    }
    
    showCreateChannelSuccess(message) {
        // Remove any existing alerts
        $('.create-channel-alert').remove();
        
        const alert = $(`
            <div class="alert alert-success create-channel-alert mt-2" role="alert">
                <i class="bi bi-check-circle"></i> ${message}
            </div>
        `);
        
        $('#createChannelForm').after(alert);
    }
    
    clearCreateChannelErrors() {
        $('.create-channel-alert').remove();
        $('#channelName').removeClass('is-valid is-invalid');
    }
}

// Initialize channel manager when app is ready
$(document).ready(() => {
    setTimeout(() => {
        if (window.turnateApp) {
            window.channelManager = new ChannelManager(window.turnateApp);
        }
    }, 100);
});