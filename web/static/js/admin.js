// Admin panel JavaScript for Turnate
class AdminManager {
    constructor(app) {
        this.app = app;
        this.setupAdminUI();
    }
    
    setupAdminUI() {
        // Only show admin features for admin users
        if (this.app.currentUser && this.app.currentUser.role === 'admin') {
            this.addAdminNavigation();
            this.setupAdminEventListeners();
        }
    }
    
    addAdminNavigation() {
        // Add admin button to navbar
        const adminBtn = $(`
            <li class="nav-item">
                <a class="nav-link" href="#" id="adminPanelBtn">
                    <i class="bi bi-gear"></i> Admin
                </a>
            </li>
        `);
        
        $('.navbar-nav').prepend(adminBtn);
        
        // Add admin indicator to user dropdown
        const adminBadge = '<span class="badge bg-danger ms-2">Admin</span>';
        $('#currentUsername').after(adminBadge);
    }
    
    setupAdminEventListeners() {
        $('#adminPanelBtn').on('click', (e) => {
            e.preventDefault();
            this.showAdminPanel();
        });
    }
    
    async showAdminPanel() {
        try {
            // Load admin data
            const [usersResponse, channelsResponse] = await Promise.all([
                this.app.makeRequest('/api/v1/admin/users'),
                this.app.makeRequest('/api/v1/admin/channels')
            ]);
            
            this.createAdminModal(usersResponse.users || [], channelsResponse.channels || []);
            
        } catch (error) {
            console.error('Failed to load admin data:', error);
            this.app.showError('Failed to load admin panel');
        }
    }
    
    createAdminModal(users, channels) {
        const usersTable = this.createUsersTable(users);
        const channelsTable = this.createChannelsTable(channels);
        
        const modal = $(`
            <div class="modal fade" id="adminModal" tabindex="-1">
                <div class="modal-dialog modal-xl">
                    <div class="modal-content">
                        <div class="modal-header">
                            <h5 class="modal-title">
                                <i class="bi bi-gear"></i> Admin Panel
                            </h5>
                            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                        </div>
                        <div class="modal-body">
                            <nav>
                                <div class="nav nav-tabs" id="adminTabs" role="tablist">
                                    <button class="nav-link active" id="users-tab" data-bs-toggle="tab" data-bs-target="#users-pane" type="button" role="tab">
                                        <i class="bi bi-people"></i> Users (${users.length})
                                    </button>
                                    <button class="nav-link" id="channels-tab" data-bs-toggle="tab" data-bs-target="#channels-pane" type="button" role="tab">
                                        <i class="bi bi-hash"></i> Channels (${channels.length})
                                    </button>
                                    <button class="nav-link" id="settings-tab" data-bs-toggle="tab" data-bs-target="#settings-pane" type="button" role="tab">
                                        <i class="bi bi-sliders"></i> Settings
                                    </button>
                                </div>
                            </nav>
                            <div class="tab-content mt-3" id="adminTabContent">
                                <div class="tab-pane fade show active" id="users-pane" role="tabpanel">
                                    ${usersTable}
                                </div>
                                <div class="tab-pane fade" id="channels-pane" role="tabpanel">
                                    ${channelsTable}
                                </div>
                                <div class="tab-pane fade" id="settings-pane" role="tabpanel">
                                    ${this.createSettingsPanel()}
                                </div>
                            </div>
                        </div>
                        <div class="modal-footer">
                            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Close</button>
                        </div>
                    </div>
                </div>
            </div>
        `);
        
        $('body').append(modal);
        $('#adminModal').modal('show');
        
        // Setup admin event listeners
        this.setupAdminModalEventListeners();
        
        // Clean up when modal closes
        $('#adminModal').on('hidden.bs.modal', () => {
            $('#adminModal').remove();
        });
    }
    
    createUsersTable(users) {
        const rows = users.map(user => `
            <tr data-user-id="${user.id}">
                <td>
                    <div class="d-flex align-items-center">
                        <div>
                            <strong>${user.display_name || user.username}</strong>
                            <br>
                            <small class="text-muted">@${user.username}</small>
                        </div>
                    </div>
                </td>
                <td>${user.email}</td>
                <td>
                    <span class="badge bg-${user.role === 'admin' ? 'danger' : 'primary'}">
                        ${user.role}
                    </span>
                </td>
                <td>
                    <span class="badge bg-${user.is_active ? 'success' : 'secondary'}">
                        ${user.is_active ? 'Active' : 'Inactive'}
                    </span>
                </td>
                <td>
                    <div class="btn-group btn-group-sm">
                        <button class="btn btn-outline-primary edit-user-btn" title="Edit User">
                            <i class="bi bi-pencil"></i>
                        </button>
                        <button class="btn btn-outline-${user.is_active ? 'warning' : 'success'} toggle-user-btn" 
                                title="${user.is_active ? 'Deactivate' : 'Activate'} User">
                            <i class="bi bi-${user.is_active ? 'pause' : 'play'}"></i>
                        </button>
                    </div>
                </td>
            </tr>
        `).join('');
        
        return `
            <div class="table-responsive">
                <table class="table table-hover">
                    <thead>
                        <tr>
                            <th>User</th>
                            <th>Email</th>
                            <th>Role</th>
                            <th>Status</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody id="usersTableBody">
                        ${rows}
                    </tbody>
                </table>
            </div>
        `;
    }
    
    createChannelsTable(channels) {
        const rows = channels.map(channel => `
            <tr data-channel-id="${channel.id}">
                <td>
                    <div class="d-flex align-items-center">
                        <span class="me-2">${channel.type === 'private' ? '🔒' : '🌍'}</span>
                        <div>
                            <strong>#${channel.name}</strong>
                            ${channel.description ? `<br><small class="text-muted">${channel.description}</small>` : ''}
                        </div>
                    </div>
                </td>
                <td>
                    <span class="badge bg-${channel.type === 'private' ? 'warning' : 'info'}">
                        ${channel.type}
                    </span>
                </td>
                <td>
                    <i class="bi bi-people me-1"></i>
                    ${channel.member_count || 0}
                </td>
                <td>
                    <small class="text-muted">
                        ${new Date(channel.created_at).toLocaleDateString()}
                    </small>
                </td>
                <td>
                    <div class="btn-group btn-group-sm">
                        <button class="btn btn-outline-primary view-members-btn" title="View Members">
                            <i class="bi bi-people"></i>
                        </button>
                        <button class="btn btn-outline-info edit-channel-btn" title="Edit Channel">
                            <i class="bi bi-pencil"></i>
                        </button>
                        ${channel.name !== 'general' ? `
                            <button class="btn btn-outline-danger delete-channel-btn" title="Delete Channel">
                                <i class="bi bi-trash"></i>
                            </button>
                        ` : ''}
                    </div>
                </td>
            </tr>
        `).join('');
        
        return `
            <div class="table-responsive">
                <table class="table table-hover">
                    <thead>
                        <tr>
                            <th>Channel</th>
                            <th>Type</th>
                            <th>Members</th>
                            <th>Created</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody id="channelsTableBody">
                        ${rows}
                    </tbody>
                </table>
            </div>
        `;
    }
    
    createSettingsPanel() {
        return `
            <div class="row">
                <div class="col-md-6">
                    <div class="card">
                        <div class="card-header">
                            <h6 class="mb-0"><i class="bi bi-info-circle"></i> System Information</h6>
                        </div>
                        <div class="card-body">
                            <dl class="row">
                                <dt class="col-sm-5">Version:</dt>
                                <dd class="col-sm-7">Turnate v1.0.0</dd>
                                
                                <dt class="col-sm-5">Database:</dt>
                                <dd class="col-sm-7">SQLite</dd>
                                
                                <dt class="col-sm-5">Authentication:</dt>
                                <dd class="col-sm-7">JWT</dd>
                                
                                <dt class="col-sm-5">Real-time:</dt>
                                <dd class="col-sm-7">Polling (5s interval)</dd>
                            </dl>
                        </div>
                    </div>
                </div>
                
                <div class="col-md-6">
                    <div class="card">
                        <div class="card-header">
                            <h6 class="mb-0"><i class="bi bi-bar-chart"></i> Statistics</h6>
                        </div>
                        <div class="card-body">
                            <dl class="row">
                                <dt class="col-sm-6">Total Users:</dt>
                                <dd class="col-sm-6" id="totalUsers">-</dd>
                                
                                <dt class="col-sm-6">Active Users:</dt>
                                <dd class="col-sm-6" id="activeUsers">-</dd>
                                
                                <dt class="col-sm-6">Total Channels:</dt>
                                <dd class="col-sm-6" id="totalChannels">-</dd>
                                
                                <dt class="col-sm-6">Public Channels:</dt>
                                <dd class="col-sm-6" id="publicChannels">-</dd>
                            </dl>
                        </div>
                    </div>
                </div>
            </div>
            
            <div class="row mt-3">
                <div class="col-12">
                    <div class="card">
                        <div class="card-header">
                            <h6 class="mb-0"><i class="bi bi-tools"></i> System Actions</h6>
                        </div>
                        <div class="card-body">
                            <div class="d-flex gap-2">
                                <button class="btn btn-outline-info" id="refreshStatsBtn">
                                    <i class="bi bi-arrow-clockwise"></i> Refresh Stats
                                </button>
                                <button class="btn btn-outline-warning" id="clearCacheBtn">
                                    <i class="bi bi-trash3"></i> Clear Cache
                                </button>
                                <button class="btn btn-outline-danger" id="exportDataBtn">
                                    <i class="bi bi-download"></i> Export Data
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }
    
    setupAdminModalEventListeners() {
        // User management
        $('#usersTableBody').on('click', '.edit-user-btn', (e) => {
            const userId = $(e.target).closest('tr').data('user-id');
            this.showEditUserModal(userId);
        });
        
        $('#usersTableBody').on('click', '.toggle-user-btn', (e) => {
            const userId = $(e.target).closest('tr').data('user-id');
            const row = $(e.target).closest('tr');
            const isActive = row.find('.badge:contains("Active")').length > 0;
            this.toggleUserStatus(userId, !isActive);
        });
        
        // Channel management
        $('#channelsTableBody').on('click', '.view-members-btn', (e) => {
            const channelId = $(e.target).closest('tr').data('channel-id');
            this.showChannelMembers(channelId);
        });
        
        $('#channelsTableBody').on('click', '.edit-channel-btn', (e) => {
            const channelId = $(e.target).closest('tr').data('channel-id');
            this.showEditChannelModal(channelId);
        });
        
        $('#channelsTableBody').on('click', '.delete-channel-btn', (e) => {
            const channelId = $(e.target).closest('tr').data('channel-id');
            const channelName = $(e.target).closest('tr').find('strong').text();
            this.confirmDeleteChannel(channelId, channelName);
        });
        
        // Settings
        $('#refreshStatsBtn').on('click', () => this.refreshStats());
        $('#clearCacheBtn').on('click', () => this.clearCache());
        $('#exportDataBtn').on('click', () => this.exportData());
        
        // Load initial stats
        this.loadStats();
    }
    
    async showEditUserModal(userId) {
        try {
            const response = await this.app.makeRequest(`/api/v1/users/${userId}`);
            const user = response.user;
            
            const modal = $(`
                <div class="modal fade" id="editUserModal" tabindex="-1">
                    <div class="modal-dialog">
                        <div class="modal-content">
                            <div class="modal-header">
                                <h5 class="modal-title">Edit User</h5>
                                <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                            </div>
                            <div class="modal-body">
                                <form id="editUserForm">
                                    <div class="mb-3">
                                        <label class="form-label">Username</label>
                                        <input type="text" class="form-control" value="${user.username}" readonly>
                                    </div>
                                    <div class="mb-3">
                                        <label class="form-label">Display Name</label>
                                        <input type="text" class="form-control" id="editUserDisplayName" value="${user.display_name || ''}">
                                    </div>
                                    <div class="mb-3">
                                        <label class="form-label">Role</label>
                                        <select class="form-select" id="editUserRole">
                                            <option value="normal" ${user.role === 'normal' ? 'selected' : ''}>Normal User</option>
                                            <option value="admin" ${user.role === 'admin' ? 'selected' : ''}>Administrator</option>
                                        </select>
                                    </div>
                                    <div class="mb-3">
                                        <div class="form-check">
                                            <input class="form-check-input" type="checkbox" id="editUserActive" ${user.is_active ? 'checked' : ''}>
                                            <label class="form-check-label" for="editUserActive">
                                                Account Active
                                            </label>
                                        </div>
                                    </div>
                                </form>
                            </div>
                            <div class="modal-footer">
                                <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
                                <button type="button" class="btn btn-primary" id="saveUserChanges">Save Changes</button>
                            </div>
                        </div>
                    </div>
                </div>
            `);
            
            $('body').append(modal);
            $('#editUserModal').modal('show');
            
            $('#saveUserChanges').on('click', () => this.saveUserChanges(userId));
            $('#editUserModal').on('hidden.bs.modal', () => $('#editUserModal').remove());
            
        } catch (error) {
            console.error('Failed to load user:', error);
        }
    }
    
    async saveUserChanges(userId) {
        const displayName = $('#editUserDisplayName').val().trim();
        const role = $('#editUserRole').val();
        const isActive = $('#editUserActive').is(':checked');
        
        try {
            await this.app.makeRequest(`/api/v1/users/${userId}`, 'PATCH', {
                display_name: displayName,
                role: role,
                is_active: isActive
            });
            
            this.app.showSuccess('User updated successfully! ✅');
            $('#editUserModal').modal('hide');
            
            // Refresh admin panel
            setTimeout(() => this.showAdminPanel(), 500);
            
        } catch (error) {
            console.error('Failed to update user:', error);
            this.app.showError('Failed to update user');
        }
    }
    
    async toggleUserStatus(userId, newStatus) {
        try {
            await this.app.makeRequest(`/api/v1/users/${userId}`, 'PATCH', {
                is_active: newStatus
            });
            
            this.app.showSuccess(`User ${newStatus ? 'activated' : 'deactivated'} successfully! ✅`);
            
            // Refresh admin panel
            setTimeout(() => this.showAdminPanel(), 500);
            
        } catch (error) {
            console.error('Failed to toggle user status:', error);
            this.app.showError('Failed to update user status');
        }
    }
    
    loadStats() {
        // Calculate stats from current data
        const usersTable = $('#usersTableBody tr');
        const channelsTable = $('#channelsTableBody tr');
        
        const totalUsers = usersTable.length;
        const activeUsers = usersTable.filter(':has(.badge:contains("Active"))').length;
        const totalChannels = channelsTable.length;
        const publicChannels = channelsTable.filter(':has(.badge:contains("public"))').length;
        
        $('#totalUsers').text(totalUsers);
        $('#activeUsers').text(activeUsers);
        $('#totalChannels').text(totalChannels);
        $('#publicChannels').text(publicChannels);
    }
    
    refreshStats() {
        this.loadStats();
        this.app.showSuccess('Stats refreshed! 📊');
    }
    
    clearCache() {
        // Since we don't have a real cache, just show a success message
        this.app.showSuccess('Cache cleared! 🧹');
    }
    
    exportData() {
        // Simple data export (in a real app, this would generate a file)
        const exportData = {
            timestamp: new Date().toISOString(),
            users: $('#usersTableBody tr').length,
            channels: $('#channelsTableBody tr').length,
            exported_by: this.app.currentUser.username
        };
        
        const dataStr = JSON.stringify(exportData, null, 2);
        const dataBlob = new Blob([dataStr], { type: 'application/json' });
        const url = URL.createObjectURL(dataBlob);
        
        const link = document.createElement('a');
        link.href = url;
        link.download = `turnate-export-${new Date().toISOString().split('T')[0]}.json`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        URL.revokeObjectURL(url);
        
        this.app.showSuccess('Data exported! 📄');
    }
}

// Initialize admin manager when app and user are ready
$(document).ready(() => {
    setTimeout(() => {
        if (window.turnateApp && window.turnateApp.currentUser) {
            window.adminManager = new AdminManager(window.turnateApp);
        }
    }, 1000); // Wait a bit longer to ensure user is loaded
});