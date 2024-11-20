class AccessMiddleware {
    constructor(config, bot, logger) {
        this.config = config;
        this.bot = bot;
        this.logger = logger;
    }

    /**
     * Check if user is allowed to use the bot
     * @param {number} userId - Telegram user ID
     * @param {number} chatId - Telegram chat ID
     * @returns {boolean} - Whether user is allowed
     */
    async checkAccess(userId, chatId) {
        const isAllowed = this.config.isAllowedUser(userId);
        
        if (!isAllowed) {
            this.logger.warn('Unauthorized access attempt', {
                userId,
                chatId
            });
            
            await this.bot.sendMessage(
                chatId,
                'У вас нет доступа к этому боту. Обратитесь к администратору.'
            );
            
            // Notify admin about unauthorized access attempt
            if (this.config.bot.adminId) {
                await this.bot.sendMessage(
                    this.config.bot.adminId,
                    `⚠️ Попытка несанкционированного доступа:\nUser ID: ${userId}\nChat ID: ${chatId}`
                );
            }
        }
        
        return isAllowed;
    }
}
module.exports = AccessMiddleware;
