import axios from 'axios';
import { LoginRequest, SignupRequest, AuthResponse, ApiResponse } from '@/types/auth';
import { Message, SendMessageRequest, GetMessagesResponse } from '@/types/message';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('auth_token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export const authApi = {
  login: async (data: LoginRequest): Promise<AuthResponse> => {
    const response = await api.post<ApiResponse<AuthResponse>>('/auth/login', data);
    return response.data.data!;
  },

  signup: async (data: SignupRequest): Promise<AuthResponse> => {
    const response = await api.post<ApiResponse<AuthResponse>>('/auth/signup', data);
    return response.data.data!;
  },

  getMe: async (): Promise<AuthResponse['user']> => {
    const response = await api.get<ApiResponse<AuthResponse['user']>>('/auth/me');
    return response.data.data!;
  },

  refreshToken: async (): Promise<AuthResponse> => {
    const response = await api.post<ApiResponse<AuthResponse>>('/auth/refresh');
    return response.data.data!;
  },
};

export const messageApi = {
  getMessages: async (channel: string, page = 1, limit = 50): Promise<Message[]> => {
    const response = await api.get<ApiResponse<GetMessagesResponse>>(
      `/messages?channel=${channel}&page=${page}&limit=${limit}`
    );
    return response.data.data?.messages || [];
  },

  sendMessage: async (data: SendMessageRequest): Promise<Message> => {
    const response = await api.post<ApiResponse<Message>>('/messages', data);
    return response.data.data!;
  },
};

export default api;