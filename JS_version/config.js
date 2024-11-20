const path = require('path');
const { BotError } = require('./errors');

class Config {
    constructor() {
        this.telegram = {
            token: process.env.TG_TOKEN,
            options: {
                polling: true
            }
        };

        this.kinozal = {
            address: process.env.KZ_ADDR || 'kinozal.tv',
            username: process.env.KZ_USER,
            password: process.env.KZ_PASS,
            endpoints: {
                login: '/takelogin.php',
                hash: '/get_srv_details.php',
                details: '/details.php',
                search: '/browse.php'
            }
        };

        this.transmission = {
            host: process.env.TRANS_ADDR || 'localhost',
            port: parseInt(process.env.TRANS_PORT, 10) || 9091,
            auth: {
                username: process.env.TRANS_USER,
                password: process.env.TRANS_PASS
            }
        };

        this.folders = {
            // Default paths can be overridden by environment variables
            torrents: path.join(process.cwd(), 'torrents'),
            films: process.env.FILMS_FOLDER || path.join(process.cwd(), 'downloads', 'films'),
            series: process.env.SERIES_FOLDER || path.join(process.cwd(), 'downloads', 'series'),
            audiobooks: process.env.AUDIOBOOKS_FOLDER || path.join(process.cwd(), 'downloads', 'audiobooks')
        };
        this.bot = {
            // Admin can add/remove users
            adminId: process.env.BOT_ADMIN_ID,
            // List of allowed user IDs (comma-separated in env)
            allowedUsers: process.env.BOT_ALLOWED_USERS 
                ? process.env.BOT_ALLOWED_USERS.split(',').map(id => parseInt(id.trim(), 10))
                : []
        };
    }

    /**
     * Validate configuration
     * @throws {BotError} If configuration is invalid
     */
    validate() {
        // Validate Telegram configuration
        if (!this.telegram.token) {
            throw new BotError(
                'Telegram bot token is not configured',
                'CONFIG_ERROR',
                { setting: 'TELEGRAM_BOT_TOKEN' }
            );
        }

        // Validate Kinozal configuration
        if (!this.kinozal.username || !this.kinozal.password) {
            throw new BotError(
                'Kinozal credentials are not configured',
                'CONFIG_ERROR',
                { setting: 'KINOZAL_USERNAME/KINOZAL_PASSWORD' }
            );
        }

        // Validate Transmission configuration if auth is used
        if ((this.transmission.auth.username && !this.transmission.auth.password) ||
            (!this.transmission.auth.username && this.transmission.auth.password)) {
            throw new BotError(
                'Transmission credentials are incomplete',
                'CONFIG_ERROR',
                { setting: 'TRANSMISSION_USERNAME/TRANSMISSION_PASSWORD' }
            );
        }
    }

    /**
     * Get full path for a file in the torrents directory
     * @param {string} filename - Name of the file
     * @returns {string} Full path
     */
    getTorrentPath(filename) {
        return path.join(this.folders.torrents, filename);
    }

    /**
     * Get download path based on content type
     * @param {string} type - Content type (films, series, audiobooks)
     * @returns {string} Download path
     */
    getDownloadPath(type) {
        const validTypes = ['films', 'series', 'audiobooks'];
        if (!validTypes.includes(type)) {
            throw new BotError(
                'Invalid content type',
                'CONFIG_ERROR',
                { type, validTypes }
            );
        }
        return this.folders[type];
    }
    isAdmin(userId) {
        return userId.toString() === this.bot.adminId;
    }

    isAllowedUser(userId) {
        return this.isAdmin(userId) || this.bot.allowedUsers.includes(userId);
    }
}

// Export singleton instance
module.exports = {
    config: new Config()
};