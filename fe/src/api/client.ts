import axios, { AxiosInstance, InternalAxiosRequestConfig, AxiosResponse, AxiosError } from 'axios';
import { enqueueSnackbar } from 'notistack';
import i18n from '../i18n';

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
        const interfaceLanguage = localStorage.getItem('interface_language') || 'en';
        config.headers['Accept-Language'] = interfaceLanguage;
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
      (response: AxiosResponse) => {
        // Show success notification for POST, PUT, PATCH, DELETE requests
        const method = response.config.method?.toUpperCase();
        if (method && ['POST', 'PUT', 'PATCH', 'DELETE'].includes(method)) {
          const message = response.data?.message || i18n.t('notification.operationSuccessful');
          enqueueSnackbar(message, {
            variant: 'success',
            persist: false,
          });
        }
        return response;
      },
      (error: AxiosError<any>) => {
        const currentPath = window.location.pathname;
        const publicPaths = ['/login', '/auth', '/inactive', '/unauthorized'];
        const isPublicPage = publicPaths.some(path => currentPath.startsWith(path));

        // Handle 401 - Unauthorized
        if (error.response?.status === 401) {
          if (!isPublicPage) {
            // Clear user data and redirect to login (no notification for auth errors)
            localStorage.removeItem('user');
            window.location.href = '/login';
          }
        }
        // Handle 403 - Forbidden with account deactivation
        else if (error.response?.status === 403) {
          const errorCode = error.response?.data?.error_code;

          // Check if it's an account deactivation error
          if (errorCode === 'ACCOUNT_DEACTIVATED' && !isPublicPage) {
            // Redirect to inactive page without showing notification
            window.location.href = '/inactive';
          } else {
            // For other 403 errors (insufficient permissions), show error notification
            const errorMessage = error.response?.data?.error ||
                                'You do not have permission to perform this action.';
            enqueueSnackbar(errorMessage, {
              variant: 'error',
              persist: false,
            });
          }
        }
        else if (error.response?.status !== 404) {
          // Show error notification for all other errors except 404
          const errorMessage = error.response?.data?.error ||
                              error.response?.data?.message ||
                              'An error occurred. Please try again.';
          enqueueSnackbar(errorMessage, {
            variant: 'error',
            persist: false,
          });
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
