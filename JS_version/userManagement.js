const { logger } = require('./logger');

function setupUserManagement(bot, config, logger) {
    // Add user command
    bot.onText(/\/adduser (\d+)/, async (msg, match) => {
        const adminId = msg.from.id;
        const chatId = msg.chat.id;
        
        if (!config.isAdmin(adminId)) {
            await bot.sendMessage(chatId, 'Эта команда доступна только администратору.');
            return;
        }

        const userToAdd = parseInt(match[1], 10);
        if (!config.bot.allowedUsers.includes(userToAdd)) {
            config.bot.allowedUsers.push(userToAdd);
            
            logger.info('User added to allowed list', {
                adminId,
                addedUserId: userToAdd
            });
            
            await bot.sendMessage(chatId, `Пользователь ID:${userToAdd} добавлен в список разрешенных.`);
            
            // Try to notify the added user if possible
            try {
                await bot.sendMessage(
                    userToAdd,
                    'Вам предоставлен доступ к боту. Используйте /start для начала работы.'
                );
            } catch (error) {
                logger.warn('Could not notify new user', {
                    userId: userToAdd,
                    error: error.message
                });
            }
        } else {
            await bot.sendMessage(chatId, 'Этот пользователь уже имеет доступ.');
        }
    });

    // Remove user command
    bot.onText(/\/removeuser (\d+)/, async (msg, match) => {
        const adminId = msg.from.id;
        const chatId = msg.chat.id;
        
        if (!config.isAdmin(adminId)) {
            await bot.sendMessage(chatId, 'Эта команда доступна только администратору.');
            return;
        }

        const userToRemove = parseInt(match[1], 10);
        const index = config.bot.allowedUsers.indexOf(userToRemove);
        
        if (index > -1) {
            config.bot.allowedUsers.splice(index, 1);
            
            logger.info('User removed from allowed list', {
                adminId,
                removedUserId: userToRemove
            });
            
            await bot.sendMessage(chatId, `Пользователь ID:${userToRemove} удален из списка разрешенных.`);
            
            // Try to notify the removed user
            try {
                await bot.sendMessage(
                    userToRemove,
                    'Ваш доступ к боту был отозван. Обратитесь к администратору для получения информации.'
                );
            } catch (error) {
                logger.warn('Could not notify removed user', {
                    userId: userToRemove,
                    error: error.message
                });
            }
        } else {
            await bot.sendMessage(chatId, 'Этот пользователь не найден в списке разрешенных.');
        }
    });

    // List allowed users command
    bot.onText(/\/listusers/, async (msg) => {
        const adminId = msg.from.id;
        const chatId = msg.chat.id;
        
        if (!config.isAdmin(adminId)) {
            await bot.sendMessage(chatId, 'Эта команда доступна только администратору.');
            return;
        }

        if (config.bot.allowedUsers.length === 0) {
            await bot.sendMessage(chatId, 
                '📋 Список разрешенных пользователей пуст\n\n' +
                'Используйте /adduser для добавления пользователей'
            );
            return;
        }

        const userList = config.bot.allowedUsers
            .map(id => `• ${id}`)
            .join('\n');
            
        await bot.sendMessage(
            chatId,
            '📋 *Список разрешенных пользователей:*\n\n' +
            `${userList}\n\n` +
            `👑 *Администратор:* ${config.bot.adminId}`,
            { parse_mode: 'Markdown' }
        );
    });
}

module.exports = { setupUserManagement };