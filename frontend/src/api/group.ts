import { api } from './client';
import type { User } from '../types/user';
import type { Group } from '../types/group';

export const createGroup = async (name: string, slug: string, initialMembers: string[]): Promise<Group> => {
  const response = await api.post('/groups', { name, slug, initial_members: initialMembers });
  return response.data;
};

export const getGroupDetails = async (groupId: string): Promise<{ group: Group, members: User[] }> => {
  const response = await api.get(`/groups/${groupId}`);
  return response.data;
};

export const joinGroup = async (groupId: string): Promise<void> => {
  await api.post(`/groups/${groupId}/join`);
};

export const leaveGroup = async (groupId: string): Promise<void> => {
  await api.post(`/groups/${groupId}/leave`);
};

export const removeGroupMember = async (groupId: string, memberId: string): Promise<void> => {
  await api.delete(`/groups/${groupId}/members/${memberId}`);
};

export const deleteGroup = async (groupId: string): Promise<void> => {
  await api.delete(`/groups/${groupId}`);
};