import { create } from 'zustand';
import type { Conversation } from '../types/conversation';
import { type Message } from '../types/message';
import { db } from '../lib/dexie';
import { getConversations, getMessages, markAsRead, deleteConversation as apiDeleteConversation } from '../api/chat';
import { toast } from '../hooks/use-toast';
import { useAuthStore } from './authStore';

interface ChatState {
  conversations: Conversation[];
  messages: Record<string, Message[]>;
  hasMore: Record<string, boolean>; // Indicates if there are more messages to fetch for a conversation
  activeConversationId: string | null;
  fetchConversations: () => Promise<void>;
  addMessage: (message: Message, fromSocket?: boolean) => void;
  setActiveConversationId: (id: string | null) => Promise<void>;
  loadMessagesFromDB: (conversationId: string) => Promise<void>;
  fetchOlderMessages: (conversationId: string) => Promise<void>;
  hasMoreMessages: (conversationId: string) => boolean;
  clearLocalHistory: (conversationId: string, messagesToKeep: number) => Promise<void>;
  deleteOneToOneChat: (conversationId: string) => Promise<void>;
  clearChatData: () => void; // Clears all chat-related state on logout
}

const PAGE_SIZE = 20; // Number of messages to fetch per page

export const useChatStore = create<ChatState>((set, get) => ({
  conversations: [],
  messages: {},
  hasMore: {},
  activeConversationId: null,

  fetchConversations: async () => {
    try {
      const convos = await getConversations();
      // Sort conversations by last message timestamp
      const sortedConvos = convos.sort((a, b) => {
        const timeA = new Date(a.lastMessage?.serverTimestamp || 0).getTime();
        const timeB = new Date(b.lastMessage?.serverTimestamp || 0).getTime();
        return timeB - timeA;
      });
      set({ conversations: sortedConvos });
    } catch (error) {
      console.error("Failed to fetch conversations:", error);
      toast({ title: "Error", description: "Failed to load conversations.", variant: "destructive" });
    }
  },
  
  addMessage: async (message, fromSocket = false) => {
    const currentUserId = useAuthStore.getState().user?.id;
    // Check if the message is from the current user. If so, don't increment unread count
    const isSenderCurrentUser = message.sender.id === currentUserId;

    if (fromSocket && message.conversationId === get().activeConversationId && !isSenderCurrentUser) {
      // If message is for the active chat and not from current user, mark it as read immediately
      await markAsRead(message.conversationId);
    } else if (fromSocket && !isSenderCurrentUser) {
      // Increment unread count for other conversations
      set(state => ({
        conversations: state.conversations.map(c => 
          c.id === message.conversationId ? { ...c, unreadCount: (c.unreadCount || 0) + 1 } : c
        )
      }));
    }

    await db.messages.put(message); // Save to Dexie

    set((state) => {
      const updatedMessages = {
        ...state.messages,
        [message.conversationId]: [
          ...(state.messages[message.conversationId] || []),
          message,
        ].sort((a, b) => new Date(a.serverTimestamp).getTime() - new Date(b.serverTimestamp).getTime()),
      };

      // Update last message and re-sort conversations
      const updatedConversations = state.conversations
        .map(c => 
          c.id === message.conversationId ? { ...c, lastMessage: message } : c
        )
        .sort((a,b) => new Date(b.lastMessage?.serverTimestamp || 0).getTime() - new Date(a.lastMessage?.serverTimestamp || 0).getTime());
      
      return {
        messages: updatedMessages,
        conversations: updatedConversations,
      };
    });
  },

  setActiveConversationId: async (id) => {
    if (!id) {
      set({ activeConversationId: null });
      return;
    }
    set({ activeConversationId: id });
    
    // Mark as read and reset unread count
    await markAsRead(id);
    set(state => ({
      conversations: state.conversations.map(c => c.id === id ? { ...c, unreadCount: 0 } : c)
    }));
    
    // Load messages from IndexedDB first
    await get().loadMessagesFromDB(id);

    // If no messages in DB or need more, fetch from API
    // Check hasMoreMessages explicitly for the initial fetch too
    if (!(get().messages[id] || []).length || get().hasMore[id] === undefined || get().hasMoreMessages(id)) {
      await get().fetchOlderMessages(id);
    }
  },

  loadMessagesFromDB: async (conversationId) => {
    const messages = await db.messages
      .where('conversationId')
      .equals(conversationId)
      .sortBy('serverTimestamp');
    set((state) => ({
      messages: { ...state.messages, [conversationId]: messages },
      hasMore: { ...state.hasMore, [conversationId]: true }, // Assume more until proven otherwise by API
    }));
  },

  fetchOlderMessages: async (conversationId) => {
    const currentMessages = get().messages[conversationId] || [];
    const oldestTimestamp = currentMessages.length > 0
      ? currentMessages[0].serverTimestamp
      : new Date().toISOString(); // If no messages, fetch from now

    try {
      const olderMessages = await getMessages(conversationId, oldestTimestamp);
      if (olderMessages.length > 0) {
        // Ensure new messages have 'id' before bulkAdd.
        // Dexie's put method will update if ID exists, or add if new.
        await db.messages.bulkPut(olderMessages);
        set(state => ({
          messages: {
            ...state.messages,
            [conversationId]: [...olderMessages, ...(state.messages[conversationId] || [])]
              .sort((a, b) => new Date(a.serverTimestamp).getTime() - new Date(b.serverTimestamp).getTime()),
          },
        }));
      }
      
      // Update hasMore state
      if (olderMessages.length < PAGE_SIZE) {
        set(state => ({ hasMore: { ...state.hasMore, [conversationId]: false } }));
      } else {
         set(state => ({ hasMore: { ...state.hasMore, [conversationId]: true } }));
      }

    } catch (error) {
      console.error("Failed to fetch older messages:", error);
      toast({ title: "Error", description: "Failed to load older messages.", variant: "destructive" });
    }
  },
  
  hasMoreMessages: (conversationId) => get().hasMore[conversationId] !== false,

  clearLocalHistory: async (conversationId, messagesToKeep) => {
    try {
        // First, fetch all messages to sort and identify which ones to keep/delete
        const allMessages = await db.messages.where({ conversationId }).sortBy('serverTimestamp');
        
        if (allMessages.length > messagesToKeep) {
            // Messages are already sorted oldest to newest
            const messagesToDelete = allMessages.slice(0, allMessages.length - messagesToKeep);
            
            await db.messages.bulkDelete(messagesToDelete.map(m => m.id));
        }
        await get().loadMessagesFromDB(conversationId); // Reload messages from DB to update UI
        toast({ title: "Success", description: `Cleared local history for conversation ${conversationId}, preserving ${messagesToKeep} messages.` });
    } catch (error) {
        console.error("Failed to clear local history:", error);
        toast({ title: "Error", description: "Failed to clear local chat history.", variant: "destructive" });
    }
  },

  deleteOneToOneChat: async (conversationId) => {
    try {
      await apiDeleteConversation(conversationId);
      await db.messages.where({ conversationId }).delete(); // Delete from local DB
      set(state => {
        const newMessages = { ...state.messages };
        delete newMessages[conversationId];
        const newConversations = state.conversations.filter(c => c.id !== conversationId);
        return {
          messages: newMessages,
          conversations: newConversations,
          activeConversationId: state.activeConversationId === conversationId ? null : state.activeConversationId,
        };
      });
    } catch (error) {
      console.error("Failed to delete 1-1 chat:", error);
      throw error; // Re-throw for UI to handle
    }
  },
  
  clearChatData: () => set({ conversations: [], messages: {}, activeConversationId: null, hasMore: {} })
}));
