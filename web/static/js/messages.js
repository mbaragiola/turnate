// Message handling JavaScript for Turnate
class MessageManager {
    constructor(app) {
        this.app = app;
        this.setupEventListeners();
        this.setupKeyboardShortcuts();
    }
    
    setupEventListeners() {
        // Message input focus/blur
        $('#messageInput').on('focus', () => {
            $('#messageInputContainer').addClass('focused');
        }).on('blur', () => {
            $('#messageInputContainer').removeClass('focused');
        });
        
        // Auto-resize message input
        $('#messageInput').on('input', (e) => {
            this.adjustInputHeight(e.target);
        });
        
        // Emoji shortcuts
        $('#messageInput').on('keyup', (e) => {
            this.handleEmojiShortcuts(e.target);
        });
    }
    
    setupKeyboardShortcuts() {
        // Ctrl/Cmd + Enter to send message
        $('#messageInput').on('keydown', (e) => {
            if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
                e.preventDefault();
                this.app.sendMessage();
            }
        });
        
        // Escape to cancel reply
        $(document).on('keydown', (e) => {
            if (e.key === 'Escape' && this.app.replyingTo) {
                this.app.cancelReply();
            }
        });
    }
    
    adjustInputHeight(textarea) {
        // Reset height to recalculate
        textarea.style.height = 'auto';
        
        // Set new height based on scroll height
        const maxHeight = 120; // Max 5 lines approximately
        const newHeight = Math.min(textarea.scrollHeight, maxHeight);
        
        textarea.style.height = newHeight + 'px';
        
        // Add scrollbar if content exceeds max height
        textarea.style.overflowY = textarea.scrollHeight > maxHeight ? 'scroll' : 'hidden';
    }
    
    handleEmojiShortcuts(input) {
        const cursorPos = input.selectionStart;
        const textBefore = input.value.substring(0, cursorPos);
        const textAfter = input.value.substring(cursorPos);
        
        // Check for emoji shortcuts at the end of current word
        const lastWord = textBefore.split(/\s/).pop();
        const emoji = this.getEmojiForShortcut(lastWord);
        
        if (emoji && textBefore.endsWith(lastWord)) {
            const newTextBefore = textBefore.substring(0, textBefore.length - lastWord.length) + emoji + ' ';
            input.value = newTextBefore + textAfter;
            input.selectionStart = input.selectionEnd = newTextBefore.length;
        }
    }
    
    getEmojiForShortcut(shortcut) {
        const emojiMap = {
            ':smile:': '😊',
            ':sad:': '😢',
            ':laugh:': '😂',
            ':love:': '❤️',
            ':fire:': '🔥',
            ':thumbsup:': '👍',
            ':thumbsdown:': '👎',
            ':party:': '🎉',
            ':rocket:': '🚀',
            ':wave:': '👋',
            ':check:': '✅',
            ':cross:': '❌',
            ':warning:': '⚠️',
            ':info:': 'ℹ️',
            ':question:': '❓',
            ':light:': '💡',
            ':coffee:': '☕',
            ':pizza:': '🍕',
            ':beer:': '🍺',
            ':music:': '🎵',
            ':camera:': '📸',
            ':phone:': '📱',
            ':computer:': '💻',
            ':email:': '📧',
            ':calendar:': '📅',
            ':clock:': '⏰',
            ':star:': '⭐',
            ':sun:': '☀️',
            ':moon:': '🌙',
            ':cloud:': '☁️',
            ':rain:': '🌧️',
            ':snow:': '❄️'
        };
        
        return emojiMap[shortcut.toLowerCase()] || null;
    }
    
    formatMessage(content) {
        // Enhanced message formatting
        return content
            // Basic emoji replacements
            .replace(/:\)/g, '😊')
            .replace(/:\(/g, '😢')
            .replace(/:D/g, '😃')
            .replace(/;\)/g, '😉')
            .replace(/<3/g, '❤️')
            .replace(/:\|/g, '😐')
            .replace(/:P/g, '😛')
            .replace(/:O/g, '😮')
            
            // Convert @mentions (basic implementation)
            .replace(/@([a-zA-Z0-9_]+)/g, '<span class="mention">@$1</span>')
            
            // Convert #channels
            .replace(/#([a-zA-Z0-9\-_]+)/g, '<span class="channel-mention">#$1</span>')
            
            // Convert URLs to links
            .replace(/(https?:\/\/[^\s]+)/g, '<a href="$1" target="_blank" rel="noopener">$1</a>')
            
            // Convert newlines to breaks
            .replace(/\n/g, '<br>');
    }
    
    createMessageHTML(message, isThread = false) {
        const messageTime = new Date(message.created_at);
        const timeString = this.formatMessageTime(messageTime);
        const isCurrentUser = message.user_id === this.app.currentUser.id;
        const formattedContent = this.formatMessage(message.content);
        
        const messageClass = isThread ? 'message thread-message' : 'message';
        const threadIndicator = message.thread_id ? '<i class="bi bi-reply text-muted me-1"></i>' : '';
        
        return `
            <div class="${messageClass}" data-message-id="${message.id}" data-user-id="${message.user_id}">
                <div class="message-header">
                    <span class="message-author ${isCurrentUser ? 'current-user' : ''}">${message.display_name || message.username}</span>
                    <span class="message-time" title="${messageTime.toLocaleString()}">${timeString}</span>
                    ${threadIndicator}
                </div>
                <div class="message-content">${formattedContent}</div>
                <div class="message-actions">
                    <button class="btn btn-sm btn-link p-0 me-2 reply-btn" title="Reply to this message">
                        <i class="bi bi-reply"></i>
                    </button>
                    ${message.reply_count > 0 ? `
                        <button class="btn btn-sm btn-link p-0 me-2 thread-btn" title="View thread">
                            <i class="bi bi-chat-dots"></i>
                            <span class="reply-count">${message.reply_count}</span>
                        </button>
                    ` : ''}
                    ${isCurrentUser ? `
                        <button class="btn btn-sm btn-link p-0 text-muted edit-btn" title="Edit message">
                            <i class="bi bi-pencil"></i>
                        </button>
                    ` : ''}
                </div>
            </div>
        `;
    }
    
    formatMessageTime(date) {
        const now = new Date();
        const diffInSeconds = Math.floor((now - date) / 1000);
        
        if (diffInSeconds < 60) {
            return 'just now';
        }
        
        if (diffInSeconds < 3600) {
            const minutes = Math.floor(diffInSeconds / 60);
            return `${minutes}m ago`;
        }
        
        if (diffInSeconds < 86400) {
            const hours = Math.floor(diffInSeconds / 3600);
            return `${hours}h ago`;
        }
        
        const days = Math.floor(diffInSeconds / 86400);
        if (days < 7) {
            return `${days}d ago`;
        }
        
        // For older messages, show date
        const isThisYear = date.getFullYear() === now.getFullYear();
        return isThisYear 
            ? date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
            : date.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
    }
    
    showTypingIndicator(user) {
        const existingIndicator = $('#typingIndicator');
        if (existingIndicator.length > 0) {
            existingIndicator.remove();
        }
        
        const indicator = $(`
            <div id="typingIndicator" class="typing-indicator p-2">
                <small class="text-muted">
                    <span class="spinner-grow spinner-grow-sm me-2"></span>
                    ${user.display_name || user.username} is typing...
                </small>
            </div>
        `);
        
        $('#messagesList').append(indicator);
        this.app.scrollToBottom();
        
        // Remove after 5 seconds
        setTimeout(() => {
            $('#typingIndicator').remove();
        }, 5000);
    }
    
    addMessageToDOM(message, prepend = false) {
        const messageHTML = this.createMessageHTML(message);
        const messageElement = $(messageHTML);
        
        // Add event listeners
        this.attachMessageEventListeners(messageElement);
        
        // Add to DOM
        if (prepend) {
            $('#messagesList').prepend(messageElement);
        } else {
            $('#messagesList').append(messageElement);
            this.app.scrollToBottom();
        }
        
        // Remove empty state if it exists
        $('.empty-state').remove();
        
        return messageElement;
    }
    
    attachMessageEventListeners(messageElement) {
        // Reply button
        messageElement.find('.reply-btn').on('click', () => {
            const messageId = messageElement.data('message-id');
            const userId = messageElement.data('user-id');
            const username = messageElement.find('.message-author').text();
            
            this.app.startReply({
                id: messageId,
                user_id: userId,
                username: username,
                display_name: username
            });
        });
        
        // Thread button
        messageElement.find('.thread-btn').on('click', () => {
            const messageId = messageElement.data('message-id');
            this.showThreadModal(messageId);
        });
        
        // Edit button
        messageElement.find('.edit-btn').on('click', () => {
            const messageId = messageElement.data('message-id');
            this.editMessage(messageId, messageElement);
        });
    }
    
    async showThreadModal(messageId) {
        try {
            const response = await this.app.makeRequest(
                `/api/v1/channels/${this.app.currentChannel.id}/messages/${messageId}/replies`
            );
            
            if (response.replies) {
                this.createThreadModal(messageId, response.replies);
            }
        } catch (error) {
            console.error('Failed to load thread:', error);
            this.app.showError('Failed to load thread');
        }
    }
    
    createThreadModal(threadId, replies) {
        const repliesHTML = replies.map(reply => this.createMessageHTML(reply, true)).join('');
        
        const modal = $(`
            <div class="modal fade" id="threadModal" tabindex="-1">
                <div class="modal-dialog modal-lg">
                    <div class="modal-content">
                        <div class="modal-header">
                            <h5 class="modal-title">
                                <i class="bi bi-chat-dots"></i> Thread Replies (${replies.length})
                            </h5>
                            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                        </div>
                        <div class="modal-body" style="max-height: 60vh; overflow-y: auto;">
                            <div class="thread-messages">
                                ${repliesHTML}
                            </div>
                        </div>
                        <div class="modal-footer">
                            <form id="threadReplyForm" class="d-flex w-100">
                                <input type="text" class="form-control me-2" id="threadReplyInput" 
                                       placeholder="Reply to thread..." required>
                                <button type="submit" class="btn btn-primary">
                                    <i class="bi bi-send"></i>
                                </button>
                            </form>
                        </div>
                    </div>
                </div>
            </div>
        `);
        
        $('body').append(modal);
        $('#threadModal').modal('show');
        
        // Handle thread reply
        $('#threadReplyForm').on('submit', async (e) => {
            e.preventDefault();
            const content = $('#threadReplyInput').val().trim();
            
            if (!content) return;
            
            try {
                const response = await this.app.makeRequest(
                    `/api/v1/channels/${this.app.currentChannel.id}/messages`,
                    'POST',
                    { content: content, thread_id: threadId }
                );
                
                if (response.message) {
                    const replyHTML = this.createMessageHTML(response.message, true);
                    $('.thread-messages').append(replyHTML);
                    $('#threadReplyInput').val('');
                    
                    // Update reply count in main message
                    const mainMessage = $(`.message[data-message-id="${threadId}"]`);
                    const replyBtn = mainMessage.find('.thread-btn .reply-count');
                    const currentCount = parseInt(replyBtn.text()) || 0;
                    replyBtn.text(currentCount + 1);
                }
            } catch (error) {
                console.error('Failed to send thread reply:', error);
            }
        });
        
        // Clean up when modal closes
        $('#threadModal').on('hidden.bs.modal', () => {
            $('#threadModal').remove();
        });
    }
    
    editMessage(messageId, messageElement) {
        const currentContent = messageElement.find('.message-content').text();
        
        const editForm = $(`
            <form class="edit-message-form">
                <div class="input-group">
                    <input type="text" class="form-control" value="${currentContent}">
                    <button type="submit" class="btn btn-sm btn-primary">Save</button>
                    <button type="button" class="btn btn-sm btn-secondary cancel-edit">Cancel</button>
                </div>
            </form>
        `);
        
        messageElement.find('.message-content').hide();
        messageElement.find('.message-actions').hide();
        messageElement.append(editForm);
        
        editForm.find('input').focus();
        
        // Handle edit submission
        editForm.on('submit', (e) => {
            e.preventDefault();
            // For now, just cancel edit (edit functionality would need backend support)
            this.cancelEdit(messageElement);
        });
        
        // Handle edit cancellation
        editForm.find('.cancel-edit').on('click', () => {
            this.cancelEdit(messageElement);
        });
    }
    
    cancelEdit(messageElement) {
        messageElement.find('.edit-message-form').remove();
        messageElement.find('.message-content, .message-actions').show();
    }
}

// Initialize message manager when app is ready
$(document).ready(() => {
    setTimeout(() => {
        if (window.turnateApp) {
            window.messageManager = new MessageManager(window.turnateApp);
        }
    }, 100);
});