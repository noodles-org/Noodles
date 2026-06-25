import dotenv from 'dotenv';
import { resolve } from 'path';

if (process.env.NODE_ENV !== 'production') {
    dotenv.config({ path: resolve(__dirname, '../../.env') });
}