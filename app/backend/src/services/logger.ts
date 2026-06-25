import winston from 'winston';
import { config } from '../config';

export const logger = winston.createLogger({
    level: config.isProduction ? 'info' : 'debug',
    format: winston.format.combine(
        winston.format.timestamp(),
        winston.format.errors({ stack: true }),
        config.isProduction
            ? winston.format.json()
            : winston.format.combine(winston.format.colorize(), winston.format.simple()),
    ),
    defaultMeta: { service: 'noodles-dashboard' },
    transports: [new winston.transports.Console()],
});