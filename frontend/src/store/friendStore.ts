import { create } from 'zustand';
import { getFriends as apiGetFriends, getFriendRequests, respondToFriendRequest, sendFriendRequest } from '../api/friends';
import { type User } from '../types/user';
import { type FriendshipRequest } from '../types/friendship';
import { toast } from '../hooks/use-toast';

interface FriendState {
  friends: User[];
  requests: FriendshipRequest[];
  fetchFriends: () => Promise<void>;
  fetchRequests: () => Promise<void>;
  sendRequest: (username: string) => Promise<void>;
  respondToRequest: (requestId: string, status: 'accepted' | 'declined') => Promise<void>;
}

export const useFriendStore = create<FriendState>((set, get) => ({
  friends: [],
  requests: [],

  fetchFriends: async () => {
    try {
      const friends = await apiGetFriends();
      set({ friends });
    } catch (error) {
      console.error("Failed to fetch friends:", error);
      toast({ title: "Error", description: "Failed to load friends.", variant: "destructive" });
    }
  },

  fetchRequests: async () => {
    try {
      const requests = await getFriendRequests();
      set({ requests });
    } catch (error) {
      console.error("Failed to fetch friend requests:", error);
      toast({ title: "Error", description: "Failed to load friend requests.", variant: "destructive" });
    }
  },

  sendRequest: async (username: string) => {
    try {
      await sendFriendRequest(username);
    } catch (error: any) {
      console.error("Failed to send friend request:", error);
      throw error; // Re-throw for UI to handle specific error messages
    }
  },

  respondToRequest: async (requestId, status) => {
    try {
      await respondToFriendRequest(requestId, status);
      // Refresh both requests and friends list
      get().fetchRequests();
      get().fetchFriends();
    } catch (error: any) {
      console.error("Failed to respond to request:", error);
      throw error; // Re-throw for UI to handle specific error messages
    }
  },
}));
