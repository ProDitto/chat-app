import { create } from 'zustand';
import { jwtDecode } from 'jwt-decode';
import { type User } from '../types/user';
import { loginUser, refreshToken as refreshApi, getUserProfile } from '../api/auth';
import { api } from '../api/client';
import { toast } from '../hooks/use-toast';

interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  initializeAuth: () => void;
  refreshAccessToken: () => Promise<string | null>;
  setUser: (user: User) => void; // Added for updating profile pic etc.
}

interface DecodedToken {
  user_id: string;
  username: string;
  exp: number;
  // Add other claims like profilePictureUrl if present in JWT
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  accessToken: null,
  refreshToken: null,
  isAuthenticated: false,

  setUser: (user) => set({ user }),

  login: async (email, password) => {
    const { accessToken, refreshToken, user: userData } = await loginUser({ email, password });
    
    localStorage.setItem('accessToken', accessToken);
    localStorage.setItem('refreshToken', refreshToken);
    api.defaults.headers.common['Authorization'] = `Bearer ${accessToken}`;

    set({ user: userData, accessToken, refreshToken, isAuthenticated: true });
  },

  logout: () => {
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
    delete api.defaults.headers.common['Authorization'];
    set({ user: null, accessToken: null, refreshToken: null, isAuthenticated: false });
  },
  
  initializeAuth: async () => {
    const accessToken = localStorage.getItem('accessToken');
    const refreshToken = localStorage.getItem('refreshToken');

    if (accessToken && refreshToken) {
      try {
        const decoded: DecodedToken = jwtDecode(accessToken);
        if (decoded.exp * 1000 > Date.now()) {
          // Token is valid, fetch user profile to ensure up-to-date data
          api.defaults.headers.common['Authorization'] = `Bearer ${accessToken}`;
          const userProfile = await getUserProfile(); // Fetch full user profile
          set({ user: userProfile, accessToken, refreshToken, isAuthenticated: true });
        } else {
          // Token expired, try refreshing
          await get().refreshAccessToken();
        }
      } catch (e) {
        console.error("Failed to initialize auth or refresh token:", e);
        get().logout();
      }
    }
  },

  refreshAccessToken: async () => {
    const currentRefreshToken = get().refreshToken;
    if (!currentRefreshToken) {
      get().logout();
      return null;
    }
    try {
      const { accessToken, refreshToken: newRefreshToken } = await refreshApi(currentRefreshToken);
      
      localStorage.setItem('accessToken', accessToken);
      localStorage.setItem('refreshToken', newRefreshToken);
      api.defaults.headers.common['Authorization'] = `Bearer ${accessToken}`;

      // Re-fetch user profile after token refresh to get updated details (e.g., profile pic)
      const userProfile = await getUserProfile(); 

      set({ user: userProfile, accessToken, refreshToken: newRefreshToken, isAuthenticated: true });
      return accessToken;
    } catch (error) {
      toast({ title: 'Session Expired', description: 'Please log in again.', variant: 'destructive' });
      get().logout();
      return null;
    }
  },
}));
