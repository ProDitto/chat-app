import { api } from './client';
import type { FriendshipRequest } from '../types/friendship';
import type { User } from '../types/user';

export const sendFriendRequest = async (username: string): Promise<void> => {
  await api.post('/friends/requests', { username });
};

export const getFriendRequests = async (): Promise<FriendshipRequest[]> => {
  const response = await api.get('/friends/requests');
  return response.data;
};

export const respondToFriendRequest = async (requestId: string, status: 'accepted' | 'declined'): Promise<void> => {
  await api.put(`/friends/requests/${requestId}`, { status });
};

export const getFriends = async (): Promise<User[]> => {
  const response = await api.get('/friends');
  return response.data;
};