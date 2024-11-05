const fs = require('fs').promises;
const path = require('path');
const { config } = require('./config');
const { logger } = require('./logger');
const { BotError } = require('./errors');

class FileUtils {
    /**
     * Save torrent file to disk
     * @param {string} kzId - Kinozal torrent ID
     * @param {Buffer} data - Torrent file data
     * @returns {Promise<string>} Path to saved torrent file
     */
    async saveTorrentFile(kzId, data) {
        try {
            const torrentPath = path.join(config.folders.torrents, `${kzId}.torrent`);
            await fs.writeFile(torrentPath, data);
            logger.debug(`Torrent file saved: ${torrentPath}`);
            return torrentPath;
        } catch (error) {
            throw new BotError(
                'Failed to save torrent file',
                'FILE_SYSTEM_ERROR',
                { kzId, originalError: error.message }
            );
        }
    }

    /**
     * Clean up torrent file after it's added to Transmission
     * @param {string} torrentPath - Path to torrent file
     */
    async cleanupTorrentFile(torrentPath) {
        try {
            await fs.unlink(torrentPath);
            logger.debug(`Torrent file cleaned up: ${torrentPath}`);
        } catch (error) {
            // Log but don't throw - cleanup failure shouldn't stop the process
            logger.warn('Failed to cleanup torrent file', {
                torrentPath,
                error: error.message
            });
        }
    }



    /**
     * Get file size
     * @param {string} filePath - File path
     * @returns {Promise<number>} File size in bytes
     */
    async getFileSize(filePath) {
        try {
            const stats = await fs.stat(filePath);
            return stats.size;
        } catch (error) {
            throw new BotError(
                'Failed to get file size',
                'FILE_SYSTEM_ERROR',
                { path: filePath, originalError: error.message }
            );
        }
    }
}

// Export singleton instance
module.exports = {
    fileUtils: new FileUtils()
};