import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';

import { notifyAuthExpired, notifyAuthRefresh } from '@/lib/auth-events.ts';

// HTTP 428, see server/middleware/jwt.go (PasswordChangeRequiredStatus).
const PASSWORD_CHANGE_REQUIRED = 428;
import { getBaseUrl } from '@/lib/service.ts';

type Response = {
  code: number;
  msg: string;
  data: any;
};

class Http {
  private instance: AxiosInstance;

  constructor() {
    const baseURL = getBaseUrl('http');
    const withCredentials = (import.meta.env.VITE_WITH_CREDENTIALS as string) !== 'false';

    this.instance = axios.create({
      baseURL,
      withCredentials,
      timeout: 60 * 1000
    });

    this.setInterceptors();
  }

  private setInterceptors() {
    this.instance.interceptors.request.use((config) => {
      if (config.headers) {
        config.headers.Accept = 'application/json';
      }

      return config;
    });

    this.instance.interceptors.response.use(
      (response) => {
        return response.data;
      },
      (error) => {
        console.log(error);
        const code = error.response?.status;
        if (code === 401) {
          notifyAuthExpired();
        } else if (code === PASSWORD_CHANGE_REQUIRED) {
          // The server refuses everything but the password page while the
          // account still uses the factory password.
          if (!window.location.hash.startsWith('#/auth/password')) {
            window.location.hash = '#/auth/password';
          }
        } else if (code === 403) {
          notifyAuthRefresh();
        }
        return Promise.reject(error);
      }
    );
  }

  public get(url: string, params?: any): Promise<Response> {
    return this.instance.request({
      method: 'get',
      url,
      params
    });
  }

  public post(url: string, data?: any, config?: AxiosRequestConfig): Promise<Response> {
    return this.instance.request({
      method: 'post',
      url,
      data,
      ...config
    });
  }

  public delete(url: string, data?: any): Promise<Response> {
    return this.instance.request({
      method: 'delete',
      url,
      data
    });
  }

  public request(config: AxiosRequestConfig): Promise<Response> {
    return this.instance.request(config);
  }
}

export const http = new Http();
