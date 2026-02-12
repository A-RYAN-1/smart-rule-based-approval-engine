import api from '@/lib/api';
import { User } from '@/types';

interface LoginResponse {
    message: string;
    user: User;
}

interface RegisterResponse {
    message: string;
    user: User;
}

export const authService = {
    async login(email: string, password: string): Promise<User> {
        const response = await api.post<any>('/auth/login', { email, password });
        const responseData = response.data.data || response.data;

        // Check for user object in response (standardized backend)
        const userData = responseData.user || responseData;

        if (!userData.role) {
            throw new Error('Invalid login response: Missing role');
        }

        const user: User = {
            id: userData.id || 0,
            name: userData.name || email.split('@')[0],
            email: userData.email || email,
            role: userData.role.toLowerCase() as any,
        };

        return user;
    },

    async register(name: string, email: string, password: string): Promise<User> {
        const response = await api.post<RegisterResponse>('/auth/register', { name, email, password });
        return response.data.user;
    },

    async logout(): Promise<void> {
        await api.post('/auth/logout');
    },

    async getUserInfo(): Promise<User> {
        const response = await api.get<any>('/me');
        const data = response.data.data || response.data;
        return {
            id: data.id,
            name: data.name,
            email: data.email,
            role: data.role.toLowerCase() as any,
        };
    }
};
