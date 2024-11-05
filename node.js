require('dotenv').config();
const TelegramBot = require('node-telegram-bot-api');
const axios = require('axios');
const qs = require('querystring');
const cheerio = require('cheerio');
const fs = require('fs');
const { wrapper } = require('axios-cookiejar-support');
const { CookieJar } = require('tough-cookie');
const iconv = require('iconv-lite');
const path = require('path');
const Transmission = require('transmission');

// Загрузка переменных из конфигурационного файла
const bot = new TelegramBot(process.env.TG_TOKEN, { polling: true });

// Логирование событий
function logEvent(message) {
    const time = new Date().toLocaleTimeString();
    console.log(`[${time}] ${message}`);
}

// Поддержка cookies
const jar = new CookieJar();
const client = wrapper(axios.create({ jar }));

// Функция логирования в Кинозал
async function kinozalLogin() {
    const urlLogin = `https://${process.env.KZ_ADDR}/takelogin.php`;
    const loginData = qs.stringify({
        username: process.env.KZ_USER,
        password: process.env.KZ_PASS
    });
    const headers = {
        'Content-Type': 'application/x-www-form-urlencoded',
        'Referer': `${process.env.KZ_ADDR}/`
    };

    try {
        await client.post(urlLogin, loginData, { headers });
        logEvent('Successfully logged in to Kinozal.');
    } catch (error) {
        logEvent('Error logging in to Kinozal: ' + error.message);
    }
}

// Получение хэша и списка файлов раздачи
async function getFilesAndHash(kzId) {
    const urlHash = `https://${process.env.KZ_ADDR}/get_srv_details.php?id=${kzId}&action=2`;
    await kinozalLogin();
    try {
        const response = await client.get(urlHash, {
            responseType: 'arraybuffer',
            withCredentials: true
        });

        const contentType = response.headers['content-type'];
        console.log(`[DEBUG] Content-Type for getFilesAndHash: ${contentType}`);
        
        // Декодирование данных с использованием iconv
        let decodedData;
        if (contentType && contentType.includes('windows-1251')) {
            decodedData = iconv.decode(Buffer.from(response.data), 'win1251');
        } else {
            // Пробуем альтернативную кодировку, если тип не указан
            decodedData = iconv.decode(Buffer.from(response.data), 'utf-8');
        }

        console.log(`[DEBUG] Response for getFilesAndHash (first 500 chars): ${decodedData.slice(0, 500)}`);

        // Проверяем на наличие ожидаемых данных
        if (!decodedData.includes("Инфо хеш")) {
            console.error('[ERROR] Expected data (Инфо хеш) not found for ID', kzId);
            return null;
        }

        return decodedData;
    } catch (error) {
        logEvent('Error fetching files and hash.');
        console.error('[ERROR]', error);
        return null;
    }
}

// Скачивание торрент-файла с использованием cookies
async function downloadTorrent(kzId, kzName, chatId) {
    const urlDownload = `https://dl.${process.env.KZ_ADDR}/download.php?id=${kzId}`;
    const filePath = path.join(__dirname, 'torrent_files', `${kzId}.torrent`);

    await kinozalLogin();  // Выполняем авторизацию перед скачиванием

    try {
        const response = await client.get(urlDownload, {
            responseType: 'arraybuffer',
            headers: {
                'Referer': `${process.env.KZ_ADDR}/`,
                'User-Agent': 'Mozilla/5.0'
            }
        });

        // Проверяем, не скачалась ли страница логина
        if (response.headers['content-type'] !== 'application/x-bittorrent') {
            bot.sendMessage(chatId, `Ошибка при загрузке торрент-файла ${kzName}. Возможно, требуется повторная авторизация.`);
            console.log('[ERROR] Скачалась страница логина вместо торрент-файла.');
            return;
        }

        fs.writeFileSync(filePath, response.data);
        bot.sendMessage(chatId, `Торрент-файл ${kzName} загружен успешно.`);
        console.log(`[INFO] Torrent ${kzName} downloaded successfully.`);
    } catch (error) {
        bot.sendMessage(chatId, `Ошибка при загрузке торрент-файла ${kzName}.`);
        console.error('[ERROR] Ошибка при загрузке торрент-файла:', error.message);
    }
}

const sessionData = {};

// Функция для получения HTML и вызова основной функции парсинга
async function fetchData(url, typeChat) {
    try {
        const response = await axios.get(url);
        const html = response.data;
        return await readHtml(html, url, typeChat);
    } catch (error) {
        console.error(`[ERROR] Ошибка при получении данных: ${error.message}`);
        throw error;
    }
}

// Основная функция для парсинга данных раздачи
async function readHtml(html, a, typeChat) {
    const $ = cheerio.load(html);

    // Извлечение ID раздачи
    const idMatch = a.match(/id=(\d+)/);
    const idKz = idMatch ? idMatch[1] : null;
    if (!idKz) {
        console.error(`[ERROR] ID не найден в переданной ссылке: ${a}`);
        throw new Error('ID не найден в переданной ссылке.');
    }

    // Извлечение названия
    let name = $('title').text().replace(/[`_|"&;]|quot/g, '').replace(/ё/g, 'е').split('/')[0].trim();

    // Извлечение жанра
    let genre = $('a.lnks_tobrs').first().text().replace(/ё/g, 'е').trim();

    // Извлечение размера
    let size = $('li:contains("Вес")').find('.floatright').text().trim();
    if (!size) size = "Размер не найден";

    // Извлечение количества раздающих
    let seedersText = $('a[onclick*="Раздают"]').text().trim();
    let seeders = seedersText ? seedersText.split(' ').pop() : "Нет данных";

    // Вызов функции для получения хеша
    const filesAndHash = await getFilesAndHash(idKz);
    if (!filesAndHash) {
        console.error(`[ERROR] Не удалось получить данные хеша для ID ${idKz}`);
        throw new Error('Ошибка при получении данных хеша.');
    }

    const infoHashMatch = filesAndHash.match(/Инфо хеш: ([a-fA-F0-9]+)/);
    const infoHash = infoHashMatch ? infoHashMatch[1] : null;
    if (!infoHash) {
        console.error(`[ERROR] Инфо хеш не найден в данных для ID ${idKz}`);
        throw new Error('Инфо хеш не найден в данных.');
    }

    // Формирование строки с данными
    let data = `*${name}*\n\n*Жанр:* ${genre}\n*Размер:* ${size}\n*Раздают:* ${seeders}\n*Инфо хеш:* \`${infoHash}\`\n`;
    return data;
}

// Функция поиска и отображения результатов
async function handleFindKinozal(chatId, query) {
    const urlSearch = `https://${process.env.KZ_ADDR}/browse.php?s=${encodeURI(query)}`;
    await kinozalLogin();
    try {
        const response = await client.get(urlSearch, {
            responseType: 'arraybuffer',
            withCredentials: true
        });

        const decodedData = iconv.decode(Buffer.from(response.data), 'win1251');
        const $ = cheerio.load(decodedData);

        const results = [];
        $('div.bx2_0 table.t_peer.w100p tbody tr.bg').each((_, element) => {
            const titleElement = $(element).find('td.nam a');
            const sizeElement = $(element).find('td.s').eq(1); // Индекс 1 для второго `td` с размером
            const seedsElement = $(element).find('td.sl_s'); // Колонка с сидерами

            let title = titleElement.text().trim();
            const link = titleElement.attr('href');
            const size = sizeElement.text().trim();
            const seeds = parseInt(seedsElement.text().trim(), 10);

            const kzIdMatch = link ? link.match(/id=(\d+)/) : null;
            const kzId = kzIdMatch ? kzIdMatch[1] : null;

            // Обрезаем название до 30 символов и добавляем "..." в конце, если оно длиннее
            if (title.length > 30) {
                title = title.slice(0, 30) + '...';
            }

            if (title && kzId) {
                results.push({ title, kzId, size, seeds });
            }
        });

        // Сортировка по числу сидов по убыванию
        results.sort((a, b) => b.seeds - a.seeds);

        if (results.length > 0) {
            const inlineKeyboard = results.slice(0, 5).map((result, index) => {
                // Сохранение в sessionData для дальнейшей обработки при загрузке
                sessionData[`${chatId}_${index}`] = { 
                    kzId: result.kzId, 
                    title: result.title,
                    size: result.size,
                    seeds: result.seeds 
                };
                
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
        logEvent('Error finding Kinozal content.');
        bot.sendMessage(chatId, 'Произошла ошибка при поиске на Кинозале.');
    }
}

// Обработка запроса на подробную информацию о раздаче и добавление кнопки "Скачать"
bot.on('callback_query', async (callbackQuery) => {
    const chatId = callbackQuery.message.chat.id;
    const data = callbackQuery.data;

    if (data.startsWith('download_')) {
        const [, chatId, index] = data.split('_');
        const sessionKey = `${chatId}_${index}`;
        const { kzId, title } = sessionData[sessionKey] || {};

        if (kzId && title) {
            const urlDetails = `https://${process.env.KZ_ADDR}/details.php?id=${kzId}`;
            try {
                const response = await client.get(urlDetails, { responseType: 'arraybuffer', withCredentials: true });
                const decodedData = iconv.decode(Buffer.from(response.data), 'win1251');
                const message = await readHtml(decodedData, urlDetails, 'Group');

                const downloadCallbackData = `startdownload_${kzId}`;
                console.log(downloadCallbackData)
                const downloadButton = {
                    reply_markup: {
                        inline_keyboard: [
                            [{ text: 'Скачать', callback_data: downloadCallbackData }]
                        ]
                    }
                };

                bot.sendMessage(chatId, message, { parse_mode: 'Markdown', ...downloadButton });
                delete sessionData[sessionKey];
            } catch (error) {
                bot.sendMessage(chatId, `Ошибка при получении данных для раздачи ${title}.`);
                console.error('[ERROR]', error);
            }
        } else {
            bot.sendMessage(chatId, `Ошибка: данные для скачивания не найдены.`);
        }
    }

    // Обработка нажатия кнопки "Скачать"
    if (data.startsWith('startdownload_')) {
        const [, kzId] = data.split('_');
        const kzName = `Раздача-${kzId}`;
        
        console.log(kzId, kzName, chatId)

        // Скачиваем торрент-файл
        await downloadTorrent(kzId, kzName, chatId);

        // Создаем кнопки выбора папки для загрузки в Transmission
        const keyboard = downloadFolders.map(folder => [{
            text: folder.name,
            callback_data: `selectfolder_${kzId}_${kzName}_${folder.path}`
        }]);

        bot.sendMessage(chatId, 'Выберите папку для загрузки:', {
            reply_markup: {
                inline_keyboard: keyboard
            }
        });
    }

    // Обработка выбора папки
    if (data.startsWith('selectfolder_')) {
        console.log(data)
        const [, kzId, kzName, ...folderPathParts] = data.split('_');
        const folderPath = folderPathParts.join('_');

        const torrentFilePath = path.join(__dirname, 'torrent_files', `${kzId}.torrent`);
        console.log(torrentFilePath)
        if (!fs.existsSync(torrentFilePath)) {
            bot.sendMessage(chatId, `Торрент-файл не найден. Пожалуйста, скачайте его заново.`);
            return;
        }

        // Добавляем торрент в Transmission
        transmission.addFile(torrentFilePath, { 'download-dir': folderPath }, (err, result) => {
            if (err) {
                bot.sendMessage(chatId, `Ошибка при добавлении в Transmission: ${err.message}`);
                logEvent(`Error adding torrent to Transmission: ${err.message}`);
            } else {
                bot.sendMessage(chatId, `Торрент ${kzName} добавлен в Transmission и будет загружен в папку "${folderPath}".`);
                logEvent(`Torrent ${kzName} added to Transmission.`);
            }
        });
    }
});


// Команда /find для поиска на Кинозале
bot.onText(/\/find (.+)/, (msg, match) => {
    const chatId = msg.chat.id;
    const query = match[1];
    handleFindKinozal(chatId, query);
});

// Команда для загрузки торрент-файла по ID раздачи
bot.onText(/\/download (\d+) (.+)/, (msg, match) => {
    const chatId = msg.chat.id;
    const kzId = match[1];
    const kzName = match[2];
    downloadTorrent(kzId, kzName, chatId);
});

// Подключение к Transmission
const transmission = new Transmission({
    host: process.env.TRANS_ADDR,
    port: 9091,
    username: process.env.TRANS_USER,
    password: process.env.TRANS_PASS
});

// Список доступных папок для загрузки
const downloadFolders = [
    { name: "Фильмы", path: process.env.FILMS_FOLDER },
    { name: "Сериалы", path: process.env.SERIES_FOLDER },
    { name: "Сериалы", path: process.env.AUDIOBOOKS_FOLDER }
];

// Команда для добавления в Transmission с выбором папки
bot.onText(/\/add_transmission (\d+) (.+)/, async (msg, match) => {
    const chatId = msg.chat.id;
    const kzId = match[1];
    const kzName = match[2];

    const keyboard = downloadFolders.map(folder => [{
        text: folder.name,
        callback_data: `/select_folder ${kzId} ${kzName} ${folder.path}`
    }]);

    bot.sendMessage(chatId, 'Выберите папку для загрузки:', {
        reply_markup: {
            inline_keyboard: keyboard
        }
    });
});

console.log('Bot is running...');