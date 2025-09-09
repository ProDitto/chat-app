import type { User } from './user';

export interface FriendshipRequest {
  id: string;
  sender: User;
  status: string;
  created_at: string;
}