# Getting Started with Turnate

Welcome to Turnate! This guide will help you get up and running quickly with your new chat platform.

## 🚀 Quick Start (5 minutes)

### 1. Download and Run
```bash
# Clone the repository
git clone <repository-url>
cd turnate

# Install dependencies  
go mod tidy

# Build and run
go build -o bin/turnate ./cmd/turnate
./bin/turnate
```

### 2. Open Your Browser
Navigate to `http://localhost:8080`

### 3. First Login
Use the default admin account:
- **Username**: `admin`
- **Email**: `admin@turnate.com`  
- **Password**: `admin123`

⚠️ **Important**: Change the admin password immediately after first login!

## 📝 Initial Setup

### Change Admin Password
1. Click on your profile dropdown (top right)
2. Select "Profile"
3. Update your display name and information
4. Change password through profile settings

### Create Your First Channel
1. Click the "+" button next to "Channels" 
2. Enter channel name (e.g., "dev-team")
3. Add description (optional)
4. Choose "Public" or "Private"
5. Click "Create Channel" 🎉

### Invite Team Members
Share the registration link: `http://localhost:8080`

New users can register with:
- Unique username
- Valid email address
- Secure password (min 6 characters)

## 💬 Using Turnate

### Sending Messages
1. Select a channel from the sidebar
2. Type your message in the bottom input field
3. Press Enter or click "Send" 📤

### Message Features
- **Emojis**: Type `:smile:` → 😊, `:rocket:` → 🚀
- **Threading**: Click reply button on any message 
- **Mentions**: Use `@username` to mention someone
- **Channels**: Reference channels with `#channel-name`

### Keyboard Shortcuts
- `Ctrl/Cmd + Enter` - Send message
- `Escape` - Cancel reply
- Emoji shortcuts (`:smile:`, `:fire:`, etc.)

## 👥 User Management

### User Roles
- **Normal Users**: Can join public channels, send messages, create channels
- **Admins**: All user permissions + user management, access to private channels

### Admin Features
Click "Admin" in the navbar (admins only) to access:
- **User Management**: View, edit, activate/deactivate users
- **Channel Management**: Manage all channels and memberships
- **System Settings**: View system stats and configuration

### Channel Management
- **Join Channels**: Click channels in sidebar, then "Join" button
- **Leave Channels**: Channel settings → "Leave Channel"
- **Create Private Channels**: Admins can create invite-only channels
- **View Members**: Click "Members" button in channel header

## 🛠️ Configuration

### Environment Variables
Create a `.env` file or export variables:

```bash
# Server Configuration
export PORT=8080                    # Server port
export DATABASE_URL=turnate.db      # Database file location
export JWT_SECRET=your-secret-key   # JWT signing secret

# Start server
./bin/turnate
```

### Custom Configuration
```bash
# Custom port
export PORT=3000
./bin/turnate

# Custom database location  
export DATABASE_URL=/data/my-chat.db
./bin/turnate

# Production mode
export GIN_MODE=release
./bin/turnate
```

## 🔒 Security Best Practices

### For Administrators
1. **Change default passwords** immediately
2. **Use strong JWT secrets** in production
3. **Enable HTTPS** for production deployments
4. **Regular backups** of the database
5. **Monitor user activity** through admin panel

### For Users
1. **Use strong passwords** (8+ characters, mixed case, numbers)
2. **Don't share login credentials**
3. **Log out from shared computers**
4. **Report suspicious activity** to admins

## 🎨 Customization

### UI Customization
Edit `/web/static/css/style.css` to customize:
- Colors and themes
- Font sizes and families
- Layout and spacing
- Button styles and icons

### Emoji Shortcuts
Default shortcuts available:
```
:smile: → 😊    :fire: → 🔥      :rocket: → 🚀
:heart: → ❤️    :check: → ✅     :cross: → ❌
:wave: → 👋     :party: → 🎉     :coffee: → ☕
```

Add more in `/web/static/js/messages.js`

## 📱 Mobile Usage

Turnate is fully responsive and works great on mobile:
- **Touch-friendly interface**
- **Mobile-optimized sidebar**
- **Responsive design** adapts to screen size
- **Fast loading** on mobile connections

## 🐛 Troubleshooting

### Common Issues

#### Can't Access Admin Panel
- Make sure you're logged in as admin user
- Check that your user role is "admin" in profile

#### Messages Not Loading
- Check internet connection
- Refresh the page (F5)
- Check browser console for errors

#### Can't Join Channel
- Make sure the channel is public
- Ask an admin to add you to private channels
- Check that you're logged in

#### Forgot Admin Password
Reset through database:
```bash
# Stop server
# Delete admin user from database or reset entire DB
rm turnate.db
# Restart server (will recreate with default admin)
./bin/turnate
```

### Getting Help

1. **Check Logs**: Look at server console output
2. **Browser Console**: Press F12, check for JavaScript errors
3. **Network Tab**: Check for failed API requests
4. **Admin Panel**: Check system status and stats

## 🚀 Next Steps

### Production Deployment
1. **Set strong JWT secret**: `export JWT_SECRET=$(openssl rand -base64 32)`
2. **Use reverse proxy**: Nginx or Apache for HTTPS and performance
3. **Enable monitoring**: Setup log monitoring and health checks
4. **Backup strategy**: Automated database backups
5. **Security hardening**: Firewall, fail2ban, regular updates

### Advanced Features
- **Custom themes**: Modify CSS for branded look
- **Integration**: Use API endpoints for external integrations
- **Monitoring**: Set up application monitoring
- **Scaling**: Deploy multiple instances with load balancer

### API Integration
Use the REST API for:
- **Bots and integrations**
- **Mobile apps**
- **External services**
- **Automation scripts**

See [API Documentation](API.md) for detailed endpoints.

## 📚 Additional Resources

- **API Documentation**: [docs/API.md](API.md)
- **Deployment Guide**: [docs/DEPLOYMENT.md](DEPLOYMENT.md)  
- **Source Code**: Browse the codebase for customizations
- **Tests**: Run `go test ./tests/unit/...` to see examples

## 💡 Tips & Tricks

### For Team Admins
- Create channels for different topics (general, dev, random)
- Set clear channel descriptions
- Use private channels for sensitive discussions
- Regularly review user access and permissions

### For Users  
- Use threads to keep conversations organized
- Set up channels for different projects
- Use emoji reactions to acknowledge messages
- Search through message history efficiently

### Performance Tips
- Keep channels focused and organized
- Archive old channels when not needed
- Use threads for long discussions
- Regular cleanup of old messages if needed

## 🎯 Use Cases

### Small Team Chat
- Replace email for internal communication
- Organize by projects/departments
- Quick file sharing and discussions

### Development Teams
- Code reviews and technical discussions
- Deploy notifications and alerts
- Documentation and knowledge sharing

### Community Groups
- Event planning and coordination
- Announcements and updates
- Member discussions and support

### Customer Support
- Internal support team coordination
- Escalation channels
- Knowledge base discussions

---

**Welcome to Turnate! Happy chatting! 💬🚀**

Need help? Check our documentation or create an issue on GitHub.