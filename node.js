require('dotenv').config();
const TelegramBot = require('node-telegram-bot-api');
const axios = require('axios');
const qs = require('querystring');
const cheerio = require('cheerio');
const path = require('path');
const { wrapper } = require('axios-cookiejar-support');
const { CookieJar } = require('tough-cookie');
const iconv = require('iconv-lite');
const Transmission = require('transmission');

// Import your existing modules
const { config } = require('./config');
const { fileUtils } = require('./fileutils');
const { logger } = require('./logger');  // Make sure this is imported before other custom modules
const { BotError, KinozalError, TransmissionError } = require('./errors');
const { errorHandler } = require('./errorHandler');
const AccessMiddleware = require('./middleware');
const { setupUserManagement } = require('./userManagement');
const { setupBotCommands, handleStart, handleHelp } = require('./menu');

// Initialize bot with validated config
const bot = new TelegramBot(config.telegram.token, config.telegram.options);

// Initialize middleware with logger
const accessMiddleware = new AccessMiddleware(config, bot, logger);


// Initialize cookie jar and axios client
const jar = new CookieJar();
const client = wrapper(axios.create({ jar }));

// Initialize Transmission client
const transmission = new Transmission({
    host: config.transmission.host,
    port: config.transmission.port,
    username: config.transmission.auth.username,
    password: config.transmission.auth.password
});

// Session data storage
const sessionData = {};

// Available download folders
const downloadFolders = [
    { name: "Фильмы", path: config.folders.films },
    { name: "Сериалы", path: config.folders.series },
    { name: "Аудиокниги", path: config.folders.audiobooks }
];

/// Add this function to your node.js file or create a separate kinozal.js module

async function kinozalLogin() {
    const loginUrl = `https://${config.kinozal.address}${config.kinozal.endpoints.login}`;
    
    logger.debug('Starting login process', {
        url: loginUrl,
        username: config.kinozal.username ? '****' : undefined,
        hasPassword: !!config.kinozal.password
    });

    // Prepare login data
    const loginData = qs.stringify({
        username: config.kinozal.username,
        password: config.kinozal.password,
        returnto: '/',
        before: '//',
        auth_submit_login: 'submit'  // Added this field
    });

    try {
        // First, try to get the main page to ensure connectivity and get initial cookies
        logger.debug('Testing connection to main page');
        const testResponse = await client.get(`https://${config.kinozal.address}/`, {
            headers: {
                'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36',
                'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8',
                'Accept-Language': 'ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7',
                'Accept-Encoding': 'gzip, deflate, br',
                'Connection': 'keep-alive',
                'Upgrade-Insecure-Requests': '1'
            },
            maxRedirects: 5,
            validateStatus: (status) => status >= 200 && status < 500
        });

        logger.debug('Main page response:', {
            status: testResponse.status,
            headers: testResponse.headers,
            cookies: jar.toJSON().cookies
        });

        // Perform login request with additional headers
        logger.debug('Sending login request');
        const response = await client.post(loginUrl, loginData, {
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
                'Referer': `https://${config.kinozal.address}/`,
                'Origin': `https://${config.kinozal.address}`,
                'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36',
                'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8',
                'Accept-Language': 'ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7',
                'Accept-Encoding': 'gzip, deflate, br',
                'Connection': 'keep-alive',
                'Cache-Control': 'max-age=0',
                'Upgrade-Insecure-Requests': '1'
            },
            maxRedirects: 5,
            validateStatus: (status) => status >= 200 && status < 500,
            withCredentials: true
        });

        logger.debug('Login response:', {
            status: response.status,
            headers: response.headers,
            cookies: jar.toJSON().cookies,
            responseUrl: response.request?.res?.responseUrl
        });

        // Check cookies after login
        const cookies = jar.toJSON().cookies;
        const hasUidCookie = cookies.some(cookie => cookie.key === 'uid');
        const hasPassCookie = cookies.some(cookie => cookie.key === 'pass');

        logger.debug('Cookie check:', {
            hasUidCookie,
            hasPassCookie,
            totalCookies: cookies.length
        });

        if (hasUidCookie && hasPassCookie) {
            logger.info('Successfully logged in to Kinozal');
            return true;
        }

        // If we got here without getting cookies, login failed
        throw new KinozalError('Login failed - missing required cookies', {
            status: response.status,
            cookies: cookies.map(c => c.key)
        });

    } catch (error) {
        logger.error('Login error:', {
            error: error.message,
            stack: error.stack,
            response: {
                status: error.response?.status,
                statusText: error.response?.statusText,
                headers: error.response?.headers
            }
        });

        if (error instanceof KinozalError) {
            throw error;
        }

        if (error.response) {
            throw new KinozalError('Login request failed', {
                status: error.response.status,
                statusText: error.response.statusText,
                headers: error.response.headers
            });
        } else if (error.request) {
            throw new KinozalError('No response from Kinozal server', {
                error: error.message,
                request: error.request
            });
        } else {
            throw new KinozalError('Login request configuration error', {
                message: error.message
            });
        }
    }
}

// Get files and hash information
async function getFilesAndHash(kzId) {
    const urlHash = `https://${config.kinozal.address}${config.kinozal.endpoints.hash}?id=${kzId}&action=2`;
    
    try {
        await kinozalLogin();
        const response = await client.get(urlHash, {
            responseType: 'arraybuffer',
            withCredentials: true
        });

        const contentType = response.headers['content-type'];
        logger.debug('Content-Type for getFilesAndHash:', contentType);
        
        const decodedData = contentType?.includes('windows-1251') 
            ? iconv.decode(Buffer.from(response.data), 'win1251')
            : iconv.decode(Buffer.from(response.data), 'utf-8');

        logger.debug('Response data preview:', decodedData.slice(0, 500));

        if (!decodedData.includes("Инфо хеш")) {
            throw new KinozalError('Hash info not found', { kzId });
        }

        return decodedData;
    } catch (error) {
        throw new KinozalError('Failed to get files and hash', {
            kzId,
            originalError: error.message
        });
    }
}

// Download torrent file
async function downloadTorrent(kzId, kzName, chatId) {
    const urlDownload = `https://dl.${config.kinozal.address}/download.php?id=${kzId}`;
    
    try {
        await kinozalLogin();

        const response = await client.get(urlDownload, {
            responseType: 'arraybuffer',
            headers: {
                'Referer': `${config.kinozal.address}/`,
                'User-Agent': 'Mozilla/5.0'
            }
        });

        if (response.headers['content-type'] !== 'application/x-bittorrent') {
            throw new KinozalError('Received non-torrent response', {
                contentType: response.headers['content-type'],
                kzId,
                kzName
            });
        }

        const torrentPath = await fileUtils.saveTorrentFile(kzId, response.data);
        
        bot.sendMessage(chatId, `Торрент-файл ${kzName} загружен успешно.`);
        logger.info(`Torrent ${kzName} downloaded successfully.`);
        
        return torrentPath;
    } catch (error) {
        throw new BotError(
            'Failed to download torrent',
            'DOWNLOAD_ERROR',
            { kzId, kzName, originalError: error.message }
        );
    }
}

// Add torrent to Transmission
async function addToTransmission(torrentPath, downloadPath, chatId, kzName) {
    try {
        return new Promise((resolve, reject) => {
            transmission.addFile(torrentPath, { 'download-dir': downloadPath }, async (err, result) => {
                try {
                    await fileUtils.cleanupTorrentFile(torrentPath);

                    if (err) {
                        reject(new TransmissionError('Failed to add torrent to Transmission', {
                            originalError: err.message
                        }));
                        return;
                    }

                    bot.sendMessage(chatId, 
                        `Торрент ${kzName} добавлен в Transmission и будет загружен в папку "${downloadPath}".`
                    );
                    logger.info(`Torrent ${kzName} added to Transmission.`);
                    resolve(result);
                } catch (cleanupError) {
                    logger.error('Error during torrent cleanup:', cleanupError);
                    resolve(result);
                }
            });
        });
    } catch (error) {
        throw new TransmissionError('Failed to process torrent', {
            torrentPath,
            downloadPath,
            originalError: error.message
        });
    }
}

// Parse HTML and extract torrent information
async function readHtml(html, url, typeChat) {
    const $ = cheerio.load(html);

    const idMatch = url.match(/id=(\d+)/);
    const idKz = idMatch ? idMatch[1] : null;
    if (!idKz) {
        throw new BotError('Failed to extract ID from URL', 'PARSE_ERROR', { url });
    }

    let name = $('title').text().replace(/[`_|"&;]|quot/g, '').replace(/ё/g, 'е').split('/')[0].trim();
    let genre = $('a.lnks_tobrs').first().text().replace(/ё/g, 'е').trim();
    let size = $('li:contains("Вес")').find('.floatright').text().trim() || "Размер не найден";
    let seedersText = $('a[onclick*="Раздают"]').text().trim();
    let seeders = seedersText ? seedersText.split(' ').pop() : "Нет данных";

    const filesAndHash = await getFilesAndHash(idKz);
    const infoHashMatch = filesAndHash.match(/Инфо хеш: ([a-fA-F0-9]+)/);
    const infoHash = infoHashMatch ? infoHashMatch[1] : null;

    if (!infoHash) {
        throw new BotError('Hash not found in response', 'PARSE_ERROR', { idKz });
    }

    return `*${name}*\n\n*Жанр:* ${genre}\n*Размер:* ${size}\n*Раздают:* ${seeders}\n*Инфо хеш:* \`${infoHash}\`\n`;
}

// Handle search functionality
async function handleFindKinozal(chatId, query) {
    const urlSearch = `https://${config.kinozal.address}${config.kinozal.endpoints.search}?s=${encodeURI(query)}`;
    
    try {
        await kinozalLogin();
        const response = await client.get(urlSearch, {
            responseType: 'arraybuffer',
            withCredentials: true
        });

        const decodedData = iconv.decode(Buffer.from(response.data), 'win1251');
        const $ = cheerio.load(decodedData);

        const results = [];
        $('div.bx2_0 table.t_peer.w100p tbody tr.bg').each((_, element) => {
            const titleElement = $(element).find('td.nam a');
            const sizeElement = $(element).find('td.s').eq(1);
            const seedsElement = $(element).find('td.sl_s');

            let title = titleElement.text().trim();
            const link = titleElement.attr('href');
            const size = sizeElement.text().trim();
            const seeds = parseInt(seedsElement.text().trim(), 10);

            const kzIdMatch = link ? link.match(/id=(\d+)/) : null;
            const kzId = kzIdMatch ? kzIdMatch[1] : null;

            if (title.length > 30) {
                title = title.slice(0, 30) + '...';
            }

            if (title && kzId) {
                results.push({ title, kzId, size, seeds });
            }
        });

        results.sort((a, b) => b.seeds - a.seeds);

        if (results.length > 0) {
            const inlineKeyboard = results.slice(0, 5).map((result, index) => {
                sessionData[`${chatId}_${index}`] = result;
                return [{
                    text: `${result.title} (${result.size}, сидов: ${result.seeds})`,
                    callback_data: `download_${chatId}_${index}`
                }];
            });

            bot.sendMessage(chatId, `Результаты по запросу "${query}":`, {
                reply_markup: {
                    inline_keyboard: inlineKeyboard
                }
            });
        } else {
            bot.sendMessage(chatId, `По запросу "${query}" ничего не найдено.`);
        }
    } catch (error) {
        throw new KinozalError('Search failed', {
            query,
            originalError: error.message
        });
    }
}

// Bot command handlers
bot.onText(/\/find (.+)/, async (msg, match) => {
    const chatId = msg.chat.id;
    const userId = msg.from.id;
    
    try {
        // Check access before processing command
        if (!await accessMiddleware.checkAccess(userId, chatId)) {
            return;
        }
        
        await handleFindKinozal(chatId, match[1]);
    } catch (error) {
        await errorHandler.handle(error, bot, chatId);
    }
});

// Callback query handler
bot.on('callback_query', async (callbackQuery) => {
    const chatId = callbackQuery.message.chat.id;
    const userId = callbackQuery.from.id;
    const data = callbackQuery.data;

    try {
        if (!await accessMiddleware.checkAccess(userId, chatId)) {
            return;
        }
        if (data.startsWith('download_')) {
            const [, chatId, index] = data.split('_');
            const sessionKey = `${chatId}_${index}`;
            const { kzId, title } = sessionData[sessionKey] || {};

            if (!kzId || !title) {
                throw new BotError('Invalid session data', 'SESSION_ERROR', { sessionKey });
            }

            const urlDetails = `https://${config.kinozal.address}${config.kinozal.endpoints.details}?id=${kzId}`;
            const response = await client.get(urlDetails, {
                responseType: 'arraybuffer',
                withCredentials: true
            });

            const decodedData = iconv.decode(Buffer.from(response.data), 'win1251');
            const message = await readHtml(decodedData, urlDetails, 'Group');

            const downloadCallbackData = `startdownload_${kzId}`;
            const downloadButton = {
                reply_markup: {
                    inline_keyboard: [[{ text: 'Скачать', callback_data: downloadCallbackData }]]
                }
            };

            await bot.sendMessage(chatId, message, {
                parse_mode: 'Markdown',
                ...downloadButton
            });
            delete sessionData[sessionKey];
        }

        else if (data.startsWith('startdownload_')) {
            const [, kzId] = data.split('_');
            const kzName = `Раздача-${kzId}`;
            
            const torrentPath = await downloadTorrent(kzId, kzName, chatId);
            const keyboard = downloadFolders.map(folder => [{
                text: folder.name,
                callback_data: `selectfolder_${kzId}_${kzName}_${folder.path}`
            }]);

            await bot.sendMessage(chatId, 'Выберите папку для загрузки:', {
                reply_markup: {
                    inline_keyboard: keyboard
                }
            });
        }

        else if (data.startsWith('selectfolder_')) {
            const [, kzId, kzName, ...folderPathParts] = data.split('_');
            const folderPath = folderPathParts.join('_');

            await addToTransmission(
                path.join(config.folders.torrents, `${kzId}.torrent`),
                folderPath,
                chatId,
                kzName
            );
        }
    } catch (error) {
        await errorHandler.handle(error, bot, chatId);
    }
});

// Startup validation
async function startup() {
    try {
        // Validate configuration
        config.validate();
        logger.info('Configuration validated successfully');

        // Setup bot commands and handlers
        await setupBotCommands(bot);
        handleStart(bot, config);
        handleHelp(bot, config);

        // Test Transmission connection
        await new Promise((resolve, reject) => {
            transmission.sessionStats((error, result) => {
                if (error) reject(error);
                else resolve(result);
            });
        });
        logger.info('Transmission connection tested successfully');

        // Log access control configuration
        logger.info('Access control configured:', {
            adminId: config.bot.adminId,
            allowedUsers: config.bot.allowedUsers
        });

        logger.info('Bot startup complete');
    } catch (error) {
        logger.error('Startup failed:', error);
        process.exit(1);
    }
}
setupUserManagement(bot, config, logger);
startup();

// Export for testing
module.exports = {
    bot,
    client,
    transmission,
    handleFindKinozal,
    downloadTorrent,
    addToTransmission
};