const winston = require('winston');
const path = require('path');

// Define log levels and colors
const levels = {
    error: 0,
    warn: 1,
    info: 2,
    debug: 3
};

const colors = {
    error: 'red',
    warn: 'yellow',
    info: 'green',
    debug: 'blue'
};

// Add colors to Winston
winston.addColors(colors);

// Create format for console output
const consoleFormat = winston.format.combine(
    winston.format.timestamp({ format: 'YYYY-MM-DD HH:mm:ss' }),
    winston.format.colorize({ all: true }),
    winston.format.printf(
        info => `${info.timestamp} ${info.level}: ${info.message}${info.details ? ' ' + JSON.stringify(info.details, null, 2) : ''}`
    )
);

// Create format for file output (without colors)
const fileFormat = winston.format.combine(
    winston.format.timestamp({ format: 'YYYY-MM-DD HH:mm:ss' }),
    winston.format.printf(
        info => `${info.timestamp} ${info.level}: ${info.message}${info.details ? ' ' + JSON.stringify(info.details, null, 2) : ''}`
    )
);

// Create the logger instance
const logger = winston.createLogger({
    levels,
    level: process.env.LOG_LEVEL || 'info', // Can be controlled via environment variable
    transports: [
        // Console transport for development
        new winston.transports.Console({
            format: consoleFormat
        }),
        // File transport for errors
        new winston.transports.File({
            filename: path.join('logs', 'error.log'),
            level: 'error',
            format: fileFormat,
            maxsize: 5242880, // 5MB
            maxFiles: 5
        }),
        // File transport for all logs
        new winston.transports.File({
            filename: path.join('logs', 'combined.log'),
            format: fileFormat,
            maxsize: 5242880, // 5MB
            maxFiles: 5
        })
    ],
    // Handle exceptions and rejections
    exceptionHandlers: [
        new winston.transports.File({
            filename: path.join('logs', 'exceptions.log'),
            format: fileFormat,
            maxsize: 5242880, // 5MB
            maxFiles: 5
        })
    ],
    rejectionHandlers: [
        new winston.transports.File({
            filename: path.join('logs', 'rejections.log'),
            format: fileFormat,
            maxsize: 5242880, // 5MB
            maxFiles: 5
        })
    ]
});

// Add request logging method
logger.logRequest = (req, details = {}) => {
    logger.info(`${req.method} ${req.path}`, {
        details: {
            ...details,
            query: req.query,
            params: req.params,
            ip: req.ip
        }
    });
};


module.exports = {
    logger
};