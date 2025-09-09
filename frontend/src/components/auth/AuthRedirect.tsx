import type React from 'react';
import { useAuthStore } from '../../store/authStore';
import { Navigate } from 'react-router-dom';

const AuthRedirect = ({ children }: { children: React.ReactNode }) => {
  const { isAuthenticated } = useAuthStore();

  if (isAuthenticated) {
    return <Navigate to="/" replace />;
  }

  return children;
};

export default AuthRedirect;
