# Kinozal Bot 🤖

Telegram bot for downloading torrents from Kinozal.tv with automatic integration with Transmission.

## Features

- 🔍 Search torrents on Kinozal
- 📥 Automatic torrent file downloading
- 🚀 Transmission integration
- 👥 User access control system
- 📂 Organized downloads (movies, TV shows, audiobooks)
- 🔐 Secure credentials storage
- 🐳 Docker support

## Installation on Synology
Is [here] (SYNOLOGY.md)

## Prerequisites

- Docker & Docker Compose
- Kinozal.tv account
- Telegram Bot Token

## Configuration & Installation

1. Create a directory for your project:
```bash
mkdir kinozal-bot && cd kinozal-bot
mkdir -p downloads/{torrents,films,series,audiobooks}
```

2. Create `docker-compose.yml`:

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
      - BOT_ALLOWED_USERS=user_id_1,user_id_2,user_id_3
      
      # Kinozal Configuration
      - KINOZAL_ADDRESS=kinozal.tv
      - KINOZAL_USERNAME=your_username
      - KINOZAL_PASSWORD=your_password
      
      # Transmission Configuration
      - TRANSMISSION_HOST=transmission
      - TRANSMISSION_PORT=9091
      - TRANSMISSION_USERNAME=your_transmission_username
      - TRANSMISSION_PASSWORD=your_transmission_password
      
      # Paths Configuration
      - TORRENTS_PATH=/downloads/torrents
      - FILMS_PATH=/downloads/films
      - SERIES_PATH=/downloads/series
      - AUDIOBOOKS_PATH=/downloads/audiobooks
    volumes:
      - ./downloads:/downloads
    depends_on:
      - transmission
    networks:
      - torrent-network

  transmission:
    image: linuxserver/transmission
    container_name: transmission
    environment:
      - PUID=1000
      - PGID=1000
      - TZ=Europe/London
      - USER=your_transmission_username
      - PASS=your_transmission_password
    volumes:
      - ./downloads:/downloads
    ports:
      - "9091:9091"
      - "51413:51413"
      - "51413:51413/udp"
    restart: unless-stopped
    networks:
      - torrent-network

networks:
  torrent-network:
    driver: bridge
```

3. Get your Telegram Bot Token:
   - Start chat with @BotFather in Telegram
   - Send `/newbot` command
   - Follow the instructions
   - Copy the token

4. Get your Telegram ID:
   - Start chat with @userinfobot in Telegram
   - It will show your user ID
   - Use this as BOT_ADMIN_ID

5. Start the services:
```bash
docker-compose up -d
```

## Usage

### Basic Commands

- `/start` - Start the bot
- `/help` - Show help
- `/find <query>` - Search for torrents

### Admin Commands

- `/adduser <id>` - Add a user
- `/removeuser <id>` - Remove a user
- `/listusers` - List allowed users

### Search Examples

- Movies: `/find Matrix 1999`
- TV Shows: `/find Game of Thrones S01`
- With quality: `/find Dune 2021 4K`

## Directory Structure

```
kinozal-bot/
├── docker-compose.yml
└── downloads/
    ├── torrents/
    ├── films/
    ├── series/
    └── audiobooks/
```

## Updating

To update to the latest version:

```bash
docker-compose pull
docker-compose up -d
```

## Troubleshooting

### View Logs
```bash
# View bot logs
docker logs kinozal-bot

# View Transmission logs
docker logs transmission

# Follow logs in real-time
docker logs -f kinozal-bot
```

### Common Issues

1. Bot can't connect to Transmission:
   - Check if both containers are running: `docker-compose ps`
   - Verify Transmission credentials in docker-compose.yml
   - Check if containers are in the same network

2. Download issues:
   - Check volume permissions: `ls -l downloads/`
   - Verify Transmission settings in web interface (localhost:9091)
   - Check container logs

3. Kinozal login fails:
   - Verify Kinozal credentials in docker-compose.yml
   - Check if Kinozal.tv is accessible
   - View logs for detailed error messages

## Security Recommendations

- Use strong passwords
- Don't expose Transmission port 9091 to the internet
- Regularly update Docker images
- Use user access control feature
- Keep docker-compose.yml secure (it contains sensitive data)

## Support

If you encounter any issues:

1. Check the [Troubleshooting](#troubleshooting) section
2. Look through the logs
3. Create an issue in the repository

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

MIT License. See LICENSE file for details.