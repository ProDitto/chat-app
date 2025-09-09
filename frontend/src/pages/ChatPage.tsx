import { useEffect } from 'react';
import ChatWindow from '../components/layout/ChatWindow';
import Sidebar from '../components/layout/Sidebar';
import { useSocketStore } from '../store/socketStore';
import { useAuthStore } from '../store/authStore';
import { toast } from '../hooks/use-toast';
import { useChatStore } from '../store/chatStore'; // Import chat store to handle message events
import { useGameStore } from '../store/gameStore'; // Import game store to handle game events
import type { Event } from '../types/event';
import { useLongPolling } from '../hooks/use-long-polling';

const ChatPage = () => {
  const { connect, disconnect, isConnected: isWsConnected } = useSocketStore();
  const accessToken = useAuthStore((state) => state.accessToken);
  const { addMessage, fetchConversations, deleteOneToOneChat } = useChatStore();
  const { handleGameEvent } = useGameStore();

  // Initialize long polling for event fallback
  useLongPolling(accessToken, isWsConnected, (event: Event) => {
    console.log("Long polling event received:", event);
    // Handle specific event types
    switch (event.event_type) {
      case 'new_message':
        addMessage(event.payload, true); // Treat as a new message from socket
        break;
      case 'friend_request':
        toast({
          title: "New Friend Request!",
          description: `${event.payload.sender.username} wants to be friends.`,
          duration: 5000,
        });
        break;
      case 'friend_accepted':
        toast({
          title: "Friend Request Accepted!",
          description: `${event.payload.user1.username || event.payload.user2.username} is now your friend.`,
          duration: 5000,
        });
        fetchConversations(); // Refresh conversations to show new 1-on-1 chat
        break;
      case 'game_invite':
      case 'game_update':
        handleGameEvent(event.payload);
        toast({
          title: `Game ${event.event_type === 'game_invite' ? 'Invitation' : 'Update'}`,
          description: `Tic-Tac-Toe: ${event.payload.status}`,
          duration: 5000
        });
        break;
      case 'group_created':
        toast({
          title: "New Group!",
          description: `You've been added to group "${event.payload.group.name}".`,
          duration: 5000,
        });
        fetchConversations();
        break;
      case 'group_joined':
        toast({
          title: "Group Joined!",
          description: `You joined group "${event.payload.group.name}".`,
          duration: 5000,
        });
        fetchConversations();
        break;
      case 'group_left':
        toast({
          title: "Group Update",
          description: `You left or were removed from group "${event.payload.group?.name || event.payload.groupId}".`,
          duration: 5000,
        });
        fetchConversations();
        break;
      case 'conversation_deleted':
        toast({
          title: "Conversation Deleted",
          description: `A conversation (${event.payload.groupId || event.payload.conversationId}) has been deleted.`,
          duration: 5000,
        });
        // This needs careful handling. If it's a 1-1 chat that user deleted for self,
        // the backend delete should already be handled. If it's a group deletion,
        // it needs to remove from local state.
        if (event.payload.conversationId) {
          // Attempt to remove from chat store if it's an external deletion
          deleteOneToOneChat(event.payload.conversationId); // This also handles local DB
        } else if (event.payload.groupId) {
          // If it's a group being deleted
          fetchConversations(); // Refresh conversations
        }
        break;
      default:
        toast({
          title: "New Event",
          description: `Type: ${event.event_type}, Payload: ${JSON.stringify(event.payload)}`,
          duration: 3000,
        });
    }
  });


  useEffect(() => {
    if (accessToken) {
      connect(accessToken);
    }

    // Disconnect only when component unmounts
    return () => {
      disconnect();
    };
  }, [connect, disconnect, accessToken]);


  return (
    <div className="flex h-screen overflow-hidden bg-background-secondary">
      <Sidebar />
      <ChatWindow />
    </div>
  );
};

export default ChatPage;
