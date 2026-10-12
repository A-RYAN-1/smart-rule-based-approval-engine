import axios, { AxiosError } from 'axios';

const api = axios.create({
    baseURL: import.meta.env.VITE_API_URL || '/api',
    withCredentials: true,
    headers: {
        'Content-Type': 'application/json',
    },
});

api.interceptors.response.use(
    (response) => response,
    (error: AxiosError) => {
        if (error.response) {
            console.error(`API Error (${error.config?.url}):`, error.response.data);
        }

        if (error.response?.status === 401) {
            console.error('Unauthorized access');
        }
        return Promise.reject(error);
    }
);

export default api;

/** A JSON object from the API, before it is mapped to an app type. */
export type RawRecord = Record<string, unknown>;

/** Paged list responses: { requests, total }. */
export interface RawPage {
    requests?: RawRecord[];
    total?: number;
}

/** /pending/all and /my-requests/all responses. */
export interface RawAllRequests {
    leave_request?: RawRecord[];
    expense_request?: RawRecord[];
    discount_request?: RawRecord[];
    total?: number;
}

/**
 * The backend wraps payloads as { success, message, data, error }; a few
 * older endpoints return the payload directly. unwrap returns the payload
 * either way.
 */
export function unwrap<T>(body: unknown): T {
    if (body && typeof body === 'object' && 'data' in body) {
        const data = (body as { data?: unknown }).data;
        if (data !== undefined && data !== null) return data as T;
    }
    return body as T;
}

/** The message the backend sent with an error response, or the fallback. */
export function apiErrorMessage(error: unknown, fallback: string): string {
    if (axios.isAxiosError(error)) {
        const body = error.response?.data as { message?: unknown; error?: unknown } | undefined;
        for (const v of [body?.message, body?.error]) {
            if (typeof v === 'string' && v.trim() !== '') return v;
        }
    }
    return fallback;
}
