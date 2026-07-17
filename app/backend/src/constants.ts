import {User} from './types';

export const DEV_USER: User = {
    sub: 'dev',
    email: 'dev@localhost',
    name: 'Dev User',
    groups: ['noodles-org:admin'],
    role: 'admin',
};
