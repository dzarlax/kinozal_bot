const { logger } = require('./logger')
async function setupBotCommands(bot) {
    try {
        // Define commands for regular users
        const userCommands = [
            { command: 'start', description: 'Запустить бота и получить информацию' },
            { command: 'find', description: 'Поиск торрента (например: /find Матрица)' },
            { command: 'help', description: 'Показать справку по использованию' },
        ];

        // Additional commands for admin
        const adminCommands = [
            ...userCommands,
            { command: 'adduser', description: 'Добавить пользователя (ID)' },
            { command: 'removeuser', description: 'Удалить пользователя (ID)' },
            { command: 'listusers', description: 'Список разрешенных пользователей' },
        ];

        // Set up command menu for regular users
        await bot.setMyCommands(userCommands);
        logger.info('Bot commands menu set up successfully');
    } catch (error) {
        logger.error('Failed to set up bot commands menu:', error);
    }
}

// Handle /start command
async function handleStart(bot, config) {
    bot.onText(/\/start/, async (msg) => {
        const chatId = msg.chat.id;
        const userId = msg.from.id;
        const username = msg.from.username || 'пользователь';

        try {
            // Check if user has access
            if (!config.isAllowedUser(userId)) {
                const message = `Привет, ${username}!\n\n` +
                    'К сожалению, у вас нет доступа к этому боту.\n' +
                    'Пожалуйста, обратитесь к администратору для получения доступа.\n\n' +
                    `Ваш ID: ${userId}`;
                
                await bot.sendMessage(chatId, message);
                
                // Notify admin
                if (config.bot.adminId) {
                    await bot.sendMessage(
                        config.bot.adminId,
                        `🆕 Новая попытка доступа:\nПользователь: @${username}\nID: ${userId}`
                    );
                }
                return;
            }

            // Welcome message for authorized users
            const welcomeMessage = 
                `Привет, ${username}! 👋\n\n` +
                '🤖 Я бот для скачивания торрентов с Kinozal.\n\n' +
                '*Основные команды:*\n' +
                '▫️ /find - поиск торрентов\n' +
                '  Пример: `/find Матрица`\n' +
                '▫️ /help - справка по использованию\n\n' +
                '*Как пользоваться:*\n' +
                '1. Используйте команду /find для поиска\n' +
                '2. Выберите нужный торрент из результатов\n' +
                '3. Нажмите кнопку "Скачать"\n' +
                '4. Выберите папку для загрузки\n\n' +
                '*Примечание:*\n' +
                '• Торренты автоматически добавляются в Transmission\n' +
                '• Для скачивания требуется активное подключение к Transmission';

            await bot.sendMessage(chatId, welcomeMessage, {
                parse_mode: 'Markdown',
                reply_markup: {
                    inline_keyboard: [[
                        { text: 'Начать поиск', callback_data: 'show_search_help' }
                    ]]
                }
            });
        } catch (error) {
            await errorHandler.handle(error, bot, chatId);
        }
    });

    // Handle "Начать поиск" button click
    bot.on('callback_query', async (callbackQuery) => {
        if (callbackQuery.data === 'show_search_help') {
            const chatId = callbackQuery.message.chat.id;
            
            const searchHelp = 
                '*🔍 Как искать торренты:*\n\n' +
                'Используйте команду /find следующим образом:\n\n' +
                '• Фильмы: `/find Название фильма год`\n' +
                '  Пример: `/find Матрица 1999`\n\n' +
                '• Сериалы: `/find Название сериала сезон`\n' +
                '  Пример: `/find Игра престолов 1 сезон`\n\n' +
                '• Точный поиск: добавляйте год или качество\n' +
                '  Пример: `/find Дюна 2021 4K`\n\n' +
                '*💡 Советы:*\n' +
                '• Используйте русские или английские названия\n' +
                '• Добавляйте год для точности поиска\n' +
                '• Для сериалов указывайте номер сезона';

            await bot.sendMessage(chatId, searchHelp, {
                parse_mode: 'Markdown'
            });
        }
    });
}

// Handle /help command
async function handleHelp(bot, config) {
    bot.onText(/\/help/, async (msg) => {
        const chatId = msg.chat.id;
        const userId = msg.from.id;

        try {
            if (!config.isAllowedUser(userId)) {
                return; // Access middleware will handle unauthorized users
            }

            let helpMessage = 
                '*📖 Справка по командам:*\n\n' +
                '🔍 *Поиск торрентов*\n' +
                '`/find` - поиск с указанием названия\n' +
                'Примеры:\n' +
                '• `/find Матрица 1999`\n' +
                '• `/find Игра престолов 1 сезон`\n\n' +
                '📂 *Папки для загрузки:*\n' +
                '• Фильмы - для фильмов\n' +
                '• Сериалы - для сериалов\n' +
                '• Аудиокниги - для аудиокниг\n\n' +
                '⚙️ *Дополнительно:*\n' +
                '• Торренты автоматически добавляются в Transmission\n' +
                '• Файлы сохраняются в выбранную папку\n' +
                '• Результаты сортируются по количеству сидов';

            // Add admin commands to help if user is admin
            if (config.isAdmin(userId)) {
                helpMessage += '\n\n👑 *Команды администратора:*\n' +
                    '`/adduser` - добавить пользователя по ID\n' +
                    '`/removeuser` - удалить пользователя по ID\n' +
                    '`/listusers` - список разрешенных пользователей\n\n' +
                    '*Примеры:*\n' +
                    '• `/adduser 123456789`\n' +
                    '• `/removeuser 123456789`';
            }

            await bot.sendMessage(chatId, helpMessage, {
                parse_mode: 'Markdown',
                disable_web_page_preview: true
            });
        } catch (error) {
            await errorHandler.handle(error, bot, chatId);
        }
    });
}

// Export all menu-related functions
module.exports = {
    setupBotCommands,
    handleStart,
    handleHelp
};