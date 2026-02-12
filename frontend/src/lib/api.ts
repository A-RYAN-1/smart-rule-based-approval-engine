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
