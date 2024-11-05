# Kinozal Bot: Synology NAS Quick Start Guide

Quick installation guide for setting up Kinozal Bot on Synology NAS using Docker.

## Prerequisites

- Synology NAS with Docker package installed
- Transmission installed from Package Center
- Telegram account

## Installation Steps

### 1. Prepare Directory Structure

1. Open **File Station**
2. Navigate to your preferred location (e.g., `docker/kinozal-bot`)
3. Create the following folders:
   - `downloads/torrents`
   - `downloads/films`
   - `downloads/series`
   - `downloads/audiobooks`

### 2. Get Telegram Bot Token

1. Open Telegram
2. Find `@BotFather`
3. Send `/newbot`
4. Follow instructions
5. Save the token for later

### 3. Get Your Telegram ID

1. Find `@userinfobot` in Telegram
2. Send any message
3. Save your ID for later

### 4. Setup in Container Manager

1. Open **Container Manager**
2. Go to "Project" tab
3. Click "Create"
4. Click "Add container"
5. Create new compose file:

```yaml
version: '3.8'

services:
  kinozal-bot:
    image: dzarlax/kinozalbot:latest
    container_name: kinozal-bot
    restart: unless-stopped
    environment:
      # Telegram Configuration
      - TELEGRAM_BOT_TOKEN=your_telegram_bot_token
      - BOT_ADMIN_ID=your_telegram_id
      - BOT_ALLOWED_USERS=your_telegram_id
      
      # Kinozal Configuration
      - KINOZAL_ADDRESS=kinozal.tv
      - KINOZAL_USERNAME=your_kinozal_username
      - KINOZAL_PASSWORD=your_kinozal_password
      
      # Transmission Configuration
      - TRANSMISSION_HOST=localhost
      - TRANSMISSION_PORT=9091
      - TRANSMISSION_USERNAME=admin
      - TRANSMISSION_PASSWORD=your_transmission_password
      
      # Paths Configuration (use your Transmission paths or Plex folders)
      - TORRENTS_PATH=/volume1/downloads/transmission/torrents
      - FILMS_PATH=/volume1/downloads/transmission/films
      - SERIES_PATH=/volume1/downloads/transmission/tv
      - AUDIOBOOKS_PATH=/volume1/downloads/transmission/audio
    network_mode: host
```

### 5. Configure Transmission

1. Open **Package Center**
2. Find Transmission
3. Open its settings:
   - Set username to "admin"
   - Set your password
   - Note these credentials for docker-compose.yml
   - Configure download directories to match your paths

### 6. Start the Bot

1. In Container Manager, click "Apply"
2. Wait for container to start
3. Check logs for any errors

## Testing

1. Open Telegram
2. Find your bot
3. Send `/start`
4. Try searching: `/find Matrix 1999`
5. Select quality and download location
6. Check if torrent appears in Transmission

## Troubleshooting

### Common Issues

1. **Bot doesn't respond:**
   - Check Container Manager logs
   - Verify Telegram token
   - Ensure container is running

2. **Can't connect to Transmission:**
   - Verify Transmission is running in Package Center
   - Check credentials in compose file
   - Ensure ports aren't blocked

3. **Download issues:**
   - Check folder permissions
   - Verify paths in compose file
   - Ensure enough storage space

### Checking Logs

1. Open **Container Manager**
2. Select kinozal-bot container
3. Click "Log" button
4. Look for error messages

### Folder Permissions

If you have permission issues:
1. Open **Control Panel**
2. Go to "Shared Folders"
3. Ensure docker has access to your download folders

## Tips

### Security
- Keep your docker-compose.yml secure (contains passwords)
- Use strong passwords for Transmission
- Regularly update the container image

### Performance
- Place download folders on appropriate volume
- Consider using SSD for torrents folder
- Monitor resource usage in Resource Monitor

### Maintenance

Update container:
```bash
# In Container Manager:
1. Select the project
2. Click "Pull" to get latest image
3. Click "Apply" to restart with new image
```

## Quick Commands Reference

- `/start` - Initialize bot
- `/help` - Show commands
- `/find <query>` - Search torrents
- `/adduser <id>` - Add user (admin only)
- `/listusers` - Show allowed users (admin only)

## Support

If you encounter issues:
1. Check Container Manager logs
2. Verify all credentials and paths
3. Ensure Transmission is running
4. Create issue on GitHub if needed

## Additional Notes

### Paths
Default Synology paths:
- `/volume1/` - First storage volume
- `/volume2/` - Second storage volume (if exists)

### Network
Using `network_mode: host` because:
- Simplifies connection to Transmission
- Avoids potential DNS issues
- Works better with Synology's network stack

### Storage
Consider:
- Using different volumes for different content
- Setting up Synology RAID for data protection
- Regular backup of configuration

### Updates
To update:
1. In Container Manager
2. Select kinozal-bot project
3. Click "Pull"
4. Click "Apply"