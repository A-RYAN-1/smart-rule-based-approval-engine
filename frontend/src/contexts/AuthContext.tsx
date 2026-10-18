import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { User } from '@/types';
import { authService } from '@/services/auth.service';
import { toast } from 'sonner';
import axios from 'axios';

const USER_KEY = 'approval_engine_user';

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (name: string, email: string, password: string) => Promise<void>;
  logout: () => void;
  isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  // The session cookie is httpOnly, so the stored user is the only hint that
  // this browser has signed in before. Without it there is no session to
  // check, and calling /api/me would only log a 401 on every first visit.
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(() => localStorage.getItem(USER_KEY) !== null);
  const queryClient = useQueryClient();

  useEffect(() => {
    const storedUser = localStorage.getItem(USER_KEY);
    if (storedUser === null) return;

    const checkAuth = async () => {
      try {
        const fullUser = await authService.getUserInfo();
        setUser(fullUser);
        localStorage.setItem(USER_KEY, JSON.stringify(fullUser));
      } catch (e) {
        if (axios.isAxiosError(e) && e.response?.status === 401) {
          // The session expired or was revoked: forget the stored user.
          localStorage.removeItem(USER_KEY);
          return;
        }
        // Offline or the API is down: keep the stored user until it answers.
        try {
          setUser(JSON.parse(storedUser));
        } catch {
          localStorage.removeItem(USER_KEY);
        }
      } finally {
        setIsLoading(false);
      }
    };

    checkAuth();
  }, []);

  const login = async (email: string, password: string) => {
    setIsLoading(true);
    try {
      // Critical: clear ALL cached queries BEFORE login to prevent data leakage
      await queryClient.cancelQueries();
      queryClient.clear();
      
      const user = await authService.login(email, password);
      setUser(user);
      localStorage.setItem(USER_KEY, JSON.stringify(user));
      toast.success('Logged in successfully');
    } catch (error) {
      console.error('Login error:', error);
      toast.error('Invalid email or password');
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  const register = async (name: string, email: string, password: string) => {
    setIsLoading(true);
    try {
      const user = await authService.register(name, email, password);
      setUser(user);
      localStorage.setItem(USER_KEY, JSON.stringify(user));
      toast.success('Registration successful');
    } catch (error) {
      console.error('Registration error:', error);
      toast.error('Registration failed. Email might be taken.');
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  const logout = async () => {
    try {
      await authService.logout();
      setUser(null);
      localStorage.removeItem(USER_KEY);
      await queryClient.cancelQueries();
      queryClient.clear();
      toast.success('Logged out');
    } catch (error) {
      console.error('Logout error', error);
      // Even if API fails, clear local state
      setUser(null);
      localStorage.removeItem(USER_KEY);
    }
  };

  return (
    <AuthContext.Provider value={{
      user,
      isLoading,
      login,
      register,
      logout,
      isAuthenticated: !!user
    }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
