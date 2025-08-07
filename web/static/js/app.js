// Main application JavaScript for Turnate
class TurnateApp {
    constructor() {
        this.currentUser = null;
        this.currentChannel = null;
        this.currentToken = localStorage.getItem('turnate_token');
        this.replyingTo = null;
        this.pollingInterval = null;
        
        this.init();
    }
    
    init() {
        console.log('🚀 Initializing Turnate...');
        
        // Check if user is already logged in
        if (this.currentToken) {
            this.loadUserProfile();
        } else {
            this.showAuthModal();
        }
        
        this.setupEventListeners();
    }
    
    setupEventListeners() {
        // Logout
        $('#logoutBtn').on('click', () => this.logout());
        
        // Message form
        $('#messageForm').on('submit', (e) => {
            e.preventDefault();
            this.sendMessage();
        });
        
        // Cancel reply
        $('#cancelReply').on('click', () => this.cancelReply());
        
        // Profile button
        $('#profileBtn').on('click', () => this.showProfile());
        
        // Join channel
        $('#joinChannelBtn').on('click', () => this.joinCurrentChannel());
        
        // Channel members
        $('#channelMembersBtn').on('click', () => this.showChannelMembers());
    }
    
    showAuthModal() {
        $('#authModal').modal('show');
        $('#app').addClass('d-none');
    }
    
    hideAuthModal() {
        $('#authModal').modal('hide');
        $('#app').removeClass('d-none');
    }
    
    async loadUserProfile() {
        try {
            const response = await this.makeRequest('/api/v1/users/me');
            if (response.user) {
                this.currentUser = response.user;
                this.updateUserUI();
                this.hideAuthModal();
                await this.loadChannels();
                this.startPolling();
            } else {
                throw new Error('Invalid user data');
            }
        } catch (error) {
            console.error('Failed to load user profile:', error);
            this.logout();
        }
    }
    
    updateUserUI() {
        $('#currentUsername').text(this.currentUser.display_name || this.currentUser.username);
    }
    
    async loadChannels() {
        try {
            const response = await this.makeRequest('/api/v1/channels');
            if (response.channels) {
                this.displayChannels(response.channels);
                // Auto-select general channel
                const generalChannel = response.channels.find(c => c.name === 'general');
                if (generalChannel) {
                    this.selectChannel(generalChannel);
                }
            }
        } catch (error) {
            console.error('Failed to load channels:', error);
            this.showError('Failed to load channels');
        }
    }
    
    displayChannels(channels) {
        const channelList = $('#channelList');
        channelList.empty();
        
        channels.forEach(channel => {
            const channelItem = $(`
                <button class="channel-item" data-channel-id="${channel.id}">
                    <div class="channel-name">
                        <span>
                            ${channel.type === 'private' ? '🔒' : '#'} ${channel.name}
                            ${channel.is_member ? '' : ' <i class="bi bi-plus-circle text-muted"></i>'}
                        </span>
                        <small class="member-count">${channel.member_count}</small>
                    </div>
                    ${channel.description ? `<div class="channel-type">${channel.description}</div>` : ''}
                </button>
            `);
            
            channelItem.on('click', () => this.selectChannel(channel));
            channelList.append(channelItem);
        });
    }
    
    async selectChannel(channel) {
        if (this.currentChannel?.id === channel.id) return;
        
        try {
            // Update UI immediately
            $('.channel-item').removeClass('active');
            $(`.channel-item[data-channel-id="${channel.id}"]`).addClass('active');
            
            this.currentChannel = channel;
            this.updateChannelHeader(channel);
            
            // Load messages if user is a member
            if (channel.is_member) {
                await this.loadMessages(channel.id);
                $('#messageInputContainer').show();
                $('#joinChannelBtn').hide();
            } else {
                $('#messagesList').html(`
                    <div class="empty-state">
                        <i class="bi bi-lock"></i>
                        <p>Join this channel to see messages</p>
                    </div>
                `);
                $('#messageInputContainer').hide();
                $('#joinChannelBtn').show();
            }
        } catch (error) {
            console.error('Failed to select channel:', error);
        }
    }
    
    updateChannelHeader(channel) {
        $('#currentChannelName').text(`${channel.type === 'private' ? '🔒' : '#'} ${channel.name}`);
        $('#currentChannelDescription').text(channel.description || 'No description');
    }
    
    async loadMessages(channelId, offset = 0) {
        try {
            $('#loadingMessages').removeClass('d-none');
            const response = await this.makeRequest(`/api/v1/channels/${channelId}/messages?limit=50&offset=${offset}`);
            
            if (response.messages) {
                this.displayMessages(response.messages, offset === 0);
            }
        } catch (error) {
            console.error('Failed to load messages:', error);
            this.showError('Failed to load messages');
        } finally {
            $('#loadingMessages').addClass('d-none');
        }
    }
    
    displayMessages(messages, clear = true) {
        const messagesList = $('#messagesList');
        
        if (clear) {
            messagesList.empty();
        }
        
        if (messages.length === 0 && clear) {
            messagesList.html(`
                <div class="empty-state">
                    <i class="bi bi-chat-dots"></i>
                    <p>No messages yet. Start the conversation! 💬</p>
                </div>
            `);
            return;
        }
        
        messages.forEach(message => {
            const messageEl = this.createMessageElement(message);
            if (clear) {
                messagesList.append(messageEl);
            } else {
                messagesList.prepend(messageEl);
            }
        });
        
        if (clear) {
            this.scrollToBottom();
        }
    }
    
    createMessageElement(message) {
        const messageTime = new Date(message.created_at).toLocaleString();
        const isCurrentUser = message.user_id === this.currentUser.id;
        
        const messageEl = $(`
            <div class="message" data-message-id="${message.id}">
                <div class="message-header">
                    <span class="message-author">${message.display_name || message.username}</span>
                    <span class="message-time">${messageTime}</span>
                </div>
                <div class="message-content">${this.formatMessageContent(message.content)}</div>
                <div class="message-actions">
                    <button class="btn btn-sm btn-link p-0 reply-btn" title="Reply">
                        <i class="bi bi-reply"></i>
                    </button>
                    ${message.reply_count > 0 ? `
                        <a href="#" class="reply-count" title="View replies">
                            💬 ${message.reply_count} ${message.reply_count === 1 ? 'reply' : 'replies'}
                        </a>
                    ` : ''}
                </div>
            </div>
        `);
        
        // Reply button
        messageEl.find('.reply-btn').on('click', () => this.startReply(message));
        
        // View replies
        messageEl.find('.reply-count').on('click', (e) => {
            e.preventDefault();
            this.toggleThreadReplies(message.id);
        });
        
        return messageEl;
    }
    
    formatMessageContent(content) {
        // Simple emoji conversion and link detection
        return content
            .replace(/:\)/g, '😊')
            .replace(/:\(/g, '😞')
            .replace(/:D/g, '😃')
            .replace(/;\)/g, '😉')
            .replace(/<3/g, '❤️')
            .replace(/\n/g, '<br>');
    }
    
    async sendMessage() {
        const content = $('#messageInput').val().trim();
        if (!content || !this.currentChannel) return;
        
        try {
            const payload = {
                content: content,
                thread_id: this.replyingTo?.id || null
            };
            
            const response = await this.makeRequest(
                `/api/v1/channels/${this.currentChannel.id}/messages`,
                'POST',
                payload
            );
            
            if (response.message) {
                $('#messageInput').val('');
                this.cancelReply();
                
                // If replying to a thread, don't add to main messages
                if (!this.replyingTo) {
                    this.displayMessages([response.message], false);
                }
            }
        } catch (error) {
            console.error('Failed to send message:', error);
            this.showError('Failed to send message');
        }
    }
    
    startReply(message) {
        this.replyingTo = message;
        $('#replyToUser').text(message.display_name || message.username);
        $('#replyingTo').removeClass('d-none');
        $('#messageInput').focus();
    }
    
    cancelReply() {
        this.replyingTo = null;
        $('#replyingTo').addClass('d-none');
    }
    
    async toggleThreadReplies(messageId) {
        const messageEl = $(`.message[data-message-id="${messageId}"]`);
        let repliesContainer = messageEl.find('.thread-replies');
        
        if (repliesContainer.length > 0) {
            repliesContainer.remove();
            return;
        }
        
        try {
            const response = await this.makeRequest(
                `/api/v1/channels/${this.currentChannel.id}/messages/${messageId}/replies`
            );
            
            if (response.replies && response.replies.length > 0) {
                repliesContainer = $('<div class="thread-replies"></div>');
                
                response.replies.forEach(reply => {
                    const replyEl = this.createMessageElement(reply);
                    replyEl.addClass('thread-message');
                    repliesContainer.append(replyEl);
                });
                
                messageEl.append(repliesContainer);
            }
        } catch (error) {
            console.error('Failed to load thread replies:', error);
        }
    }
    
    async joinCurrentChannel() {
        if (!this.currentChannel) return;
        
        try {
            await this.makeRequest(`/api/v1/channels/${this.currentChannel.id}/join`, 'POST');
            this.showSuccess('Joined channel successfully! 🎉');
            
            // Reload channels to update membership status
            await this.loadChannels();
        } catch (error) {
            console.error('Failed to join channel:', error);
            this.showError('Failed to join channel');
        }
    }
    
    async showChannelMembers() {
        if (!this.currentChannel) return;
        
        try {
            const response = await this.makeRequest(`/api/v1/channels/${this.currentChannel.id}/members`);
            
            if (response.members) {
                const membersList = response.members.map(member => 
                    `<li class="list-group-item d-flex justify-content-between align-items-center">
                        ${member.display_name || member.username}
                        <span class="badge bg-${member.role === 'admin' ? 'danger' : 'primary'} rounded-pill">
                            ${member.role}
                        </span>
                    </li>`
                ).join('');
                
                const modal = `
                    <div class="modal fade" id="membersModal" tabindex="-1">
                        <div class="modal-dialog">
                            <div class="modal-content">
                                <div class="modal-header">
                                    <h5 class="modal-title">Channel Members (${response.members.length})</h5>
                                    <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                                </div>
                                <div class="modal-body">
                                    <ul class="list-group">${membersList}</ul>
                                </div>
                            </div>
                        </div>
                    </div>
                `;
                
                $('body').append(modal);
                $('#membersModal').modal('show');
                $('#membersModal').on('hidden.bs.modal', () => $('#membersModal').remove());
            }
        } catch (error) {
            console.error('Failed to load channel members:', error);
        }
    }
    
    showProfile() {
        const modal = `
            <div class="modal fade" id="profileModal" tabindex="-1">
                <div class="modal-dialog">
                    <div class="modal-content">
                        <div class="modal-header">
                            <h5 class="modal-title">Your Profile</h5>
                            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                        </div>
                        <div class="modal-body">
                            <div class="mb-3">
                                <label class="form-label">Username</label>
                                <input type="text" class="form-control" value="${this.currentUser.username}" readonly>
                            </div>
                            <div class="mb-3">
                                <label class="form-label">Email</label>
                                <input type="email" class="form-control" value="${this.currentUser.email}" readonly>
                            </div>
                            <div class="mb-3">
                                <label class="form-label">Display Name</label>
                                <input type="text" class="form-control" id="profileDisplayName" 
                                       value="${this.currentUser.display_name || ''}">
                            </div>
                            <div class="mb-3">
                                <label class="form-label">Role</label>
                                <input type="text" class="form-control" value="${this.currentUser.role}" readonly>
                            </div>
                        </div>
                        <div class="modal-footer">
                            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Close</button>
                            <button type="button" class="btn btn-primary" id="updateProfileBtn">Update</button>
                        </div>
                    </div>
                </div>
            </div>
        `;
        
        $('body').append(modal);
        $('#profileModal').modal('show');
        $('#profileModal').on('hidden.bs.modal', () => $('#profileModal').remove());
        
        $('#updateProfileBtn').on('click', () => this.updateProfile());
    }
    
    async updateProfile() {
        const displayName = $('#profileDisplayName').val().trim();
        
        try {
            const response = await this.makeRequest(
                `/api/v1/users/${this.currentUser.id}`,
                'PATCH',
                { display_name: displayName }
            );
            
            if (response.user) {
                this.currentUser = response.user;
                this.updateUserUI();
                this.showSuccess('Profile updated successfully! ✅');
                $('#profileModal').modal('hide');
            }
        } catch (error) {
            console.error('Failed to update profile:', error);
            this.showError('Failed to update profile');
        }
    }
    
    startPolling() {
        if (this.pollingInterval) {
            clearInterval(this.pollingInterval);
        }
        
        // Poll for new messages every 5 seconds
        this.pollingInterval = setInterval(() => {
            if (this.currentChannel && this.currentChannel.is_member) {
                this.loadMessages(this.currentChannel.id);
            }
        }, 5000);
    }
    
    logout() {
        localStorage.removeItem('turnate_token');
        this.currentToken = null;
        this.currentUser = null;
        this.currentChannel = null;
        
        if (this.pollingInterval) {
            clearInterval(this.pollingInterval);
            this.pollingInterval = null;
        }
        
        this.showAuthModal();
        this.showSuccess('Logged out successfully! 👋');
    }
    
    async makeRequest(url, method = 'GET', data = null) {
        const options = {
            method: method,
            headers: {
                'Content-Type': 'application/json'
            }
        };
        
        if (this.currentToken) {
            options.headers['Authorization'] = `Bearer ${this.currentToken}`;
        }
        
        if (data) {
            options.body = JSON.stringify(data);
        }
        
        const response = await fetch(url, options);
        const result = await response.json();
        
        if (!response.ok) {
            throw new Error(result.error || 'Request failed');
        }
        
        return result;
    }
    
    scrollToBottom() {
        const container = $('#messagesContainer');
        container.scrollTop(container[0].scrollHeight);
    }
    
    showError(message) {
        // Use Bootstrap toast or simple alert
        alert('Error: ' + message);
    }
    
    showSuccess(message) {
        // Use Bootstrap toast or simple alert
        console.log('Success: ' + message);
    }
}

// Initialize the app when DOM is ready
$(document).ready(() => {
    window.turnateApp = new TurnateApp();
});