import { BrowserRouter, Routes, Route } from 'react-router-dom';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import ChatPage from './pages/ChatPage';
import ProtectedRoute from './components/auth/ProtectedRoute';
import AuthRedirect from './components/auth/AuthRedirect';
import { useAuthStore } from './store/authStore';
import { useEffect } from 'react';
import { useChatStore } from './store/chatStore';
import VerifyEmailPage from './pages/VerifyEmailPage.tsx';
import RequestPasswordResetPage from './pages/RequestPasswordResetPage.tsx';
import ResetPasswordPage from './pages/ResetPasswordPage.tsx';
import SettingsPage from './pages/SettingsPage.tsx';

const App = () => {
  const { initializeAuth, isAuthenticated } = useAuthStore();
  const { fetchConversations, clearChatData } = useChatStore();

  useEffect(() => {
    initializeAuth();
  }, [initializeAuth]);

  useEffect(() => {
    if (isAuthenticated) {
      fetchConversations();
    } else {
      clearChatData();
    }
  }, [isAuthenticated, fetchConversations, clearChatData]);


  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={
          <AuthRedirect>
            <LoginPage />
          </AuthRedirect>
        } />
        <Route path="/register" element={
          <AuthRedirect>
            <RegisterPage />
          </AuthRedirect>
        } />
        <Route path="/verify-email" element={<VerifyEmailPage />} />
        <Route path="/request-password-reset" element={<RequestPasswordResetPage />} />
        <Route path="/reset-password" element={<ResetPasswordPage />} />
        
        <Route path="/" element={
          <ProtectedRoute>
            <ChatPage />
          </ProtectedRoute>
        } />
        <Route path="/settings" element={
          <ProtectedRoute>
            <SettingsPage />
          </ProtectedRoute>
        } />
      </Routes>
    </BrowserRouter>
  );
};

export default App;
