import { create } from 'zustand';
import { useChatStore } from './chatStore';
import { toast } from '../hooks/use-toast';
import { useGameStore } from './gameStore';
import { type Event } from '../types/event'; // Import the Event type
import { useFriendStore } from './friendStore';

// interface WebSocketMessage {
//     type: string;
//     payload: any;
// }

interface SocketState {
  socket: WebSocket | null;
  isConnected: boolean;
  isConnecting: boolean;
  connect: (token: string) => void;
  disconnect: () => void;
  sendMessage: (type: string, payload: any) => void;
  reconnectAttempt: number;
}

const WEBSOCKET_URL = import.meta.env.VITE_WEBSOCKET_URL || 'ws://localhost:8080/ws';
const RECONNECT_MAX_ATTEMPTS = 5;
const RECONNECT_INTERVAL_MS = 3000; // 3 seconds
const WS_TIMEOUT_FALLBACK_MS = 60000; // 60 seconds

export const useSocketStore = create<SocketState>((set, get) => ({
  socket: null,
  isConnected: false,
  isConnecting: false,
  reconnectAttempt: 0,
  
  connect: (token) => {
    if (get().isConnected || get().isConnecting) {
      return;
    }
    
    set({ isConnecting: true });
    
    const ws = new WebSocket(`${WEBSOCKET_URL}?token=${token}`);
    let wsFallbackTimeout: ReturnType<typeof setTimeout>;

    ws.onopen = () => {
      console.log('WebSocket connected');
      set({ isConnected: true, isConnecting: false, socket: ws, reconnectAttempt: 0 });
      toast({ title: 'Connected', description: 'Real-time communication established.' });
      clearTimeout(wsFallbackTimeout); // Clear any pending fallback
    };

    ws.onmessage = (event) => {
      try {
        const parsedEvent: Event = JSON.parse(event.data); // Events now come wrapped as a general Event type
        // console.log('WebSocket event received:', parsedEvent);

        switch (parsedEvent.event_type) { // Use event_type from the Event object
            case 'new_message':
                useChatStore.getState().addMessage(parsedEvent.payload, true);
                break;
            case 'friend_request':
                toast({ 
                    title: "New Friend Request!", 
                    description: `${parsedEvent.payload.sender.username} wants to be friends.`, 
                    duration: 5000 
                });
                useFriendStore.getState().fetchRequests(); // Refresh requests
                break;
            case 'friend_accepted':
                toast({
                    title: "Friend Request Accepted!",
                    description: `${parsedEvent.payload.user1.username || parsedEvent.payload.user2.username} is now your friend.`,
                    duration: 5000,
                });
                useFriendStore.getState().fetchFriends(); // Refresh friends
                useChatStore.getState().fetchConversations(); // Refresh conversations to show new 1-on-1 chat
                break;
            case 'game_invite':
            case 'game_update':
                useGameStore.getState().handleGameEvent(parsedEvent.payload);
                if (parsedEvent.event_type === 'game_invite') {
                    toast({ 
                        title: "Game Invitation!", 
                        description: `You've been invited to a game by ${parsedEvent.payload.initiatorId}!`,
                        duration: 5000 
                    });
                }
                break;
            case 'group_created':
                toast({
                    title: "New Group!",
                    description: `You've been added to group "${parsedEvent.payload.group.name}".`,
                    duration: 5000,
                });
                useChatStore.getState().fetchConversations();
                break;
            case 'group_joined':
                toast({
                    title: "Group Joined!",
                    description: `You joined group "${parsedEvent.payload.group.name}".`,
                    duration: 5000,
                });
                useChatStore.getState().fetchConversations();
                break;
            case 'group_left':
                toast({
                    title: "Group Update",
                    description: `You left or were removed from group "${parsedEvent.payload.group?.name || parsedEvent.payload.groupId}".`,
                    duration: 5000,
                });
                useChatStore.getState().fetchConversations();
                break;
            case 'conversation_deleted':
                 toast({
                    title: "Conversation Deleted",
                    description: `A conversation has been deleted.`,
                    duration: 5000,
                });
                // This will be handled by the chatStore itself if local, otherwise refresh
                useChatStore.getState().fetchConversations();
                break;
            default:
                console.log("Unhandled WebSocket event type:", parsedEvent.event_type, parsedEvent.payload);
                toast({
                    title: "New WebSocket Event",
                    description: `Type: ${parsedEvent.event_type}`,
                    duration: 3000,
                });
        }
      } catch (e) {
        console.error("Error parsing WebSocket event:", e);
      }
    };

    ws.onclose = () => {
      console.log('WebSocket disconnected');
      set({ isConnected: false, isConnecting: false, socket: null });
      
      const currentAttempt = get().reconnectAttempt;
      if (currentAttempt < RECONNECT_MAX_ATTEMPTS) {
        set((state) => ({ reconnectAttempt: state.reconnectAttempt + 1 }));
        setTimeout(() => get().connect(token), RECONNECT_INTERVAL_MS);
      } else {
        toast({ title: 'Connection Failed', description: 'Max reconnection attempts reached. Falling back to long polling.', variant: 'destructive' });
        // The useLongPolling hook handles this now based on isWsConnected state
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      toast({ title: 'Connection Error', description: 'WebSocket error occurred.', variant: 'destructive' });
      ws.close(); // Force close to trigger onclose handler for retry logic
    };

    // Set a timeout to fallback to long polling if WebSocket doesn't connect within 60 seconds
    wsFallbackTimeout = setTimeout(() => {
        if (!get().isConnected) {
            console.warn("WebSocket connection timed out, falling back to long polling.");
            ws.close(); // This will trigger onclose and then the long polling fallback
        }
    }, WS_TIMEOUT_FALLBACK_MS);

    set({ socket: ws });
  },

  disconnect: () => {
    get().socket?.close();
    set({ socket: null, isConnected: false, isConnecting: false, reconnectAttempt: 0 });
  },

  sendMessage: (type, payload) => {
    const ws = get().socket;
    if (ws && ws.readyState === WebSocket.OPEN) {
      const message = { type, payload }; // Send as a raw WebSocketMessage for backend to process
      ws.send(JSON.stringify(message));
    } else {
      console.error('WebSocket is not connected. Message not sent:', { type, payload });
      toast({ title: 'Cannot Send Message', description: 'You are not connected. Please refresh or check connection.', variant: 'destructive' });
    }
  },
}));
