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

        if (!responseData.role) {
            throw new Error('Invalid login response: Missing role');
        }

        // If backend returns user object inside data, use it.
        // Otherwise, construct from responseData
        const user: User = {
            id: responseData.id || 0,
            name: responseData.name || email.split('@')[0],
            email: responseData.email || email,
            role: responseData.role.toLowerCase() as any,
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
        const response = await api.get<any>('/api/me');
        const data = response.data.data || response.data;
        return {
            id: data.id,
            name: data.name,
            email: data.email,
            role: data.role.toLowerCase() as any,
        };
    }
};
