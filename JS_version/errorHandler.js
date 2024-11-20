const { logger } = require('./logger');
const { BotError, KinozalError, TransmissionError } = require('./errors');

class ErrorHandler {
    constructor() {
        this.errorMessages = {
            // Kinozal-related errors
            LOGIN_ERROR: 'Ошибка авторизации на сайте. Попробуйте позже.',
            PARSE_ERROR: 'Ошибка обработки данных с сайта.',
            DOWNLOAD_ERROR: 'Ошибка при скачивании торрент-файла.',
            SEARCH_ERROR: 'Ошибка при поиске. Попробуйте изменить запрос.',
            
            // Transmission-related errors
            TRANSMISSION_CONNECTION_ERROR: 'Ошибка подключения к Transmission.',
            TRANSMISSION_ADD_ERROR: 'Ошибка при добавлении торрента в Transmission.',
            
            // File system errors
            FILE_SYSTEM_ERROR: 'Ошибка при работе с файловой системой.',
            
            // Session errors
            SESSION_ERROR: 'Ошибка сессии. Попробуйте выполнить поиск заново.',
            
            // Generic errors
            UNKNOWN_ERROR: 'Произошла неизвестная ошибка. Попробуйте позже.'
        };
    }

    /**
     * Handle different types of errors and send appropriate messages to the user
     * @param {Error} error - The error object
     * @param {TelegramBot} bot - The Telegram bot instance
     * @param {number} chatId - The chat ID where the error occurred
     */
    async handle(error, bot, chatId) {
        let userMessage;
        let logLevel = 'error';
        
        try {
            // Log the full error for debugging
            logger.debug('Full error object:', {
                name: error.name,
                message: error.message,
                stack: error.stack,
                details: error.details || {}
            });

            // Handle specific error types
            if (error instanceof KinozalError) {
                userMessage = this.handleKinozalError(error);
            } else if (error instanceof TransmissionError) {
                userMessage = this.handleTransmissionError(error);
            } else if (error instanceof BotError) {
                userMessage = this.handleBotError(error);
            } else {
                userMessage = this.errorMessages.UNKNOWN_ERROR;
                logLevel = 'error';
            }

            // Log the error with appropriate level
            logger[logLevel]('Error handled:', {
                type: error.constructor.name,
                message: error.message,
                details: error.details || {},
                userMessage
            });

            // Send error message to user if we have a valid bot and chatId
            if (bot && chatId) {
                await bot.sendMessage(chatId, userMessage, { parse_mode: 'Markdown' });
            }
        } catch (handlingError) {
            // Log error handler failures
            logger.error('Error in error handler:', {
                originalError: error,
                handlingError,
                chatId
            });

            // Attempt to send a generic error message to user
            try {
                if (bot && chatId) {
                    await bot.sendMessage(
                        chatId,
                        'Произошла критическая ошибка. Пожалуйста, попробуйте позже.'
                    );
                }
            } catch (messagingError) {
                logger.error('Failed to send error message to user:', messagingError);
            }
        }
    }

    /**
     * Handle Kinozal-specific errors
     * @param {KinozalError} error
     * @returns {string} User-friendly error message
     */
    handleKinozalError(error) {
        if (error.message.includes('Login failed')) {
            return this.errorMessages.LOGIN_ERROR;
        }
        if (error.message.includes('Failed to get files and hash')) {
            return this.errorMessages.PARSE_ERROR;
        }
        if (error.message.includes('Search failed')) {
            return this.errorMessages.SEARCH_ERROR;
        }
        return this.errorMessages.UNKNOWN_ERROR;
    }

    /**
     * Handle Transmission-specific errors
     * @param {TransmissionError} error
     * @returns {string} User-friendly error message
     */
    handleTransmissionError(error) {
        if (error.message.includes('Failed to add torrent')) {
            return this.errorMessages.TRANSMISSION_ADD_ERROR;
        }
        if (error.message.includes('connection')) {
            return this.errorMessages.TRANSMISSION_CONNECTION_ERROR;
        }
        return this.errorMessages.TRANSMISSION_ADD_ERROR;
    }

    /**
     * Handle bot-specific errors
     * @param {BotError} error
     * @returns {string} User-friendly error message
     */
    handleBotError(error) {
        // Map error codes to messages
        const messageMap = {
            'PARSE_ERROR': this.errorMessages.PARSE_ERROR,
            'DOWNLOAD_ERROR': this.errorMessages.DOWNLOAD_ERROR,
            'SESSION_ERROR': this.errorMessages.SESSION_ERROR
        };

        return messageMap[error.code] || this.errorMessages.UNKNOWN_ERROR;
    }
}

// Export singleton instance
module.exports = {
    errorHandler: new ErrorHandler()
};