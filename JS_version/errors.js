class BotError extends Error {
    constructor(message, code, details = {}) {
        super(message);
        this.name = 'BotError';
        this.code = code;
        this.details = details;
        Error.captureStackTrace(this, this.constructor);
    }
}

class KinozalError extends Error {
    constructor(message, details = {}) {
        super(message);
        this.name = 'KinozalError';
        this.details = details;
        Error.captureStackTrace(this, this.constructor);
    }
}

class TransmissionError extends Error {
    constructor(message, details = {}) {
        super(message);
        this.name = 'TransmissionError';
        this.details = details;
        Error.captureStackTrace(this, this.constructor);
    }
}

module.exports = {
    BotError,
    KinozalError,
    TransmissionError
};