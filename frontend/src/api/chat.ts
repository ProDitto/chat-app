import { api } from './client';
import { type Conversation } from '../types/conversation';
import { type Message } from '../types/message';

export const getConversations = async (): Promise<Conversation[]> => {
  const response = await api.get('/conversations');
  return response.data;
};

export const getMessages = async (conversationId: string, before?: string): Promise<Message[]> => {
  const params = new URLSearchParams();
  if (before) {
    params.append('before', before);
  }
  params.append('limit', '20'); // Page size 20 messages
  
  const response = await api.get(`/conversations/${conversationId}/messages`, { params });
  return response.data;
};

export const markAsRead = async (conversationId: string): Promise<void> => {
  await api.post(`/conversations/${conversationId}/read`);
};

export const deleteConversation = async (conversationId: string): Promise<void> => {
  await api.delete(`/conversations/${conversationId}`);
};