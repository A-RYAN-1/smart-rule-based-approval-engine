import api, { unwrap } from '@/lib/api';
import { User, UserRole } from '@/types';

/** The user object the backend returns (role in upper case). */
interface RawUser {
    id?: number;
    name?: string;
    email?: string;
    role?: string;
}

interface LoginResponse {
    message: string;
    user: User;
}

interface RegisterResponse {
    message: string;
    user: User;
}

/** A public demo account, served only when the backend runs in DEMO_MODE. */
export interface DemoLogin {
    role: string;
    email: string;
    password: string;
}

export const authService = {
    async login(email: string, password: string): Promise<User> {
        const response = await api.post<unknown>('/auth/login', { email, password });
        const responseData = unwrap<RawUser & { user?: RawUser }>(response.data);

        // Check for user object in response (standardized backend)
        const userData: RawUser = responseData.user || responseData;

        if (!userData.role) {
            throw new Error('Invalid login response: Missing role');
        }

        const user: User = {
            id: userData.id || 0,
            name: userData.name || email.split('@')[0],
            email: userData.email || email,
            role: userData.role.toLowerCase() as UserRole,
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

    /** The demo accounts, or an empty list when demo mode is off. */
    async demoLogins(): Promise<DemoLogin[]> {
        try {
            const response = await api.get<{ logins?: DemoLogin[] }>('/demo-logins');
            return response.data.logins ?? [];
        } catch {
            return [];
        }
    },

    async getUserInfo(): Promise<User> {
        const response = await api.get<unknown>('/me');
        const data = unwrap<Required<RawUser>>(response.data);
        return {
            id: data.id,
            name: data.name,
            email: data.email,
            role: data.role.toLowerCase() as UserRole,
        };
    }
};
