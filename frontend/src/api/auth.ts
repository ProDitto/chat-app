import { api } from './client';
import type { AuthResponse, LoginCredentials, RegisterData } from '../types/auth';
import type { User } from '../types/user';

export const loginUser = async (credentials: LoginCredentials): Promise<AuthResponse & { user: User }> => {
 const response = await api.post('/login', credentials);
 return response.data;
};

export const registerUser = async (data: RegisterData): Promise<User> => {
 const response = await api.post('/register', data);
 return response.data;
};

export const refreshToken = async (token: string): Promise<AuthResponse> => {
 const response = await api.post('/refresh', { refresh_token: token });
 return response.data;
};

export const verifyEmail = async (token: string): Promise<void> => {
 await api.post('/verify-email', { token });
};

export const requestPasswordReset = async (email: string): Promise<void> => {
 await api.post('/request-password-reset', { email });
};

export const resetPassword = async (email: string, otp: string, newPassword: string): Promise<void> => {
 await api.post('/reset-password', { email, otp, new_password: newPassword });
};

export const getUserProfile = async (): Promise<User> => {
 const response = await api.get('/me');
 return response.data;
};
