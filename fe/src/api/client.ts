import axios, { AxiosInstance, InternalAxiosRequestConfig } from 'axios';

class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: import.meta.env.VITE_API_URL || '',
      timeout: 30000,
      withCredentials: true, // Important for cookies - backend middleware handles refresh automatically
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Request interceptor
    this.client.interceptors.request.use(
      (config: InternalAxiosRequestConfig) => {
        // Add Accept-Language header from localStorage
        const uiLanguage = localStorage.getItem('ui_language') || 'en';
        config.headers['Accept-Language'] = uiLanguage;
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // Response interceptor
    // The backend auth middleware automatically handles token refresh
    // If we get a 401, it means both access and refresh tokens are expired/invalid
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          const currentPath = window.location.pathname;
          const publicPaths = ['/login', '/auth', '/inactive', '/unauthorized'];
          const isPublicPage = publicPaths.some(path => currentPath.startsWith(path));

          if (!isPublicPage) {
            // Clear user data and redirect to login
            localStorage.removeItem('user');
            window.location.href = '/login';
          }
        }
        return Promise.reject(error);
      }
    );
  }

  getInstance(): AxiosInstance {
    return this.client;
  }
}

export const apiClient = new ApiClient().getInstance();
