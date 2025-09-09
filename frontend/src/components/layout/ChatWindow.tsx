import { useChatStore } from '../../store/chatStore';
import { motion, AnimatePresence } from 'framer-motion';
import { MessagesSquare, Trash2, Loader2, Gamepad, PhoneCall, Video } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { useInView } from 'framer-motion';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter
} from '../../components/ui/Dialog';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { toast } from '../../hooks/use-toast';
import { useAuthStore } from '../../store/authStore';
import { TicTacToeGame } from '../game/TicTacToeGame';
import { useGameStore } from '../../store/gameStore';
import { inviteToGame } from '../../api/game';
import { ConversationType } from '../../types/conversation'; // Import ConversationType
import MessageInput from '../common/MessageInput';
import MessageBubble from '../common/MessageBubble';

const ChatWindow = () => {
  const { activeConversationId, messages, conversations, fetchOlderMessages, hasMoreMessages, deleteOneToOneChat, clearLocalHistory } = useChatStore();
  const activeConversation = conversations.find(c => c.id === activeConversationId);
  const topMessageRef = useRef<HTMLDivElement>(null);
  const endOfMessagesRef = useRef<HTMLDivElement>(null); // For auto-scrolling
  const isInView = useInView(topMessageRef);
  const [isLoadingOlderMessages, setIsLoadingOlderMessages] = useState(false);
  const [showClearHistoryDialog, setShowClearHistoryDialog] = useState(false);
  const [messagesToKeep, setMessagesToKeep] = useState(100); // Default N
  const { user } = useAuthStore();
  const { activeGameId, games, setActiveGameId } = useGameStore();

  const conversationMessages = messages[activeConversationId || ''] || [];

  // Auto-scroll to bottom on new messages
  useEffect(() => {
    if (endOfMessagesRef.current) {
      endOfMessagesRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [conversationMessages.length, activeConversationId]);

  // Fetch older messages when top is in view
  useEffect(() => {
    const loadMessages = async () => {
      if (
        isInView &&
        hasMoreMessages(activeConversationId || '') &&
        !isLoadingOlderMessages &&
        activeConversationId
      ) {
        setIsLoadingOlderMessages(true);
        await fetchOlderMessages(activeConversationId);
        setIsLoadingOlderMessages(false);
      }
    };
    loadMessages();
  }, [isInView, hasMoreMessages, fetchOlderMessages, activeConversationId, isLoadingOlderMessages]);

  const handleDeleteOneToOneChat = async () => {
    if (!activeConversationId) return;
    try {
      await deleteOneToOneChat(activeConversationId);
      toast({ title: 'Success', description: 'Chat deleted.' });
      useChatStore.getState().setActiveConversationId(null); // Clear active chat
      useChatStore.getState().fetchConversations(); // Refresh sidebar
    } catch (error: any) {
      toast({ title: 'Error', description: error.response?.data?.error || 'Failed to delete chat.', variant: 'destructive' });
    }
  };

  const handleClearLocalHistory = async () => {
    if (!activeConversationId) return;
    try {
      await clearLocalHistory(activeConversationId, messagesToKeep);
      toast({ title: 'Success', description: `Local history cleared, preserving ${messagesToKeep} messages.` });
      setShowClearHistoryDialog(false);
    } catch (error: any) {
      toast({ title: 'Error', description: error.response?.data?.error || 'Failed to clear local history.', variant: 'destructive' });
    }
  };

  const handleInviteToGame = async () => {
    if (!activeConversation || activeConversation.type !== ConversationType.OneToOne || !user) {
        toast({ title: "Error", description: "You can only invite a friend in a one-on-one chat.", variant: "destructive" });
        return;
    }
    const opponent = activeConversation.participants?.find(p => p.id !== user.id);
    if (!opponent) {
        toast({ title: "Error", description: "Could not find opponent for game invitation.", variant: "destructive" });
        return;
    }

    try {
        const game = await inviteToGame(opponent.username, "tic-tac-toe");
        toast({ title: "Game Invitation Sent", description: `Invited ${opponent.username} to Tic-Tac-Toe!` });
        setActiveGameId(game.id);
    } catch (error: any) {
        toast({ title: "Error", description: error.response?.data?.error || "Failed to send game invitation.", variant: "destructive" });
    }
  };

  const handleStartCall = (callType: 'audio' | 'video') => {
    if (!activeConversation || activeConversation.type !== ConversationType.OneToOne || !user) {
        toast({ title: "Error", description: `You can only start a ${callType} call with a friend in a one-on-one chat.`, variant: "destructive" });
        return;
    }
    const opponent = activeConversation.participants?.find(p => p.id !== user.id);
    if (!opponent) {
        toast({ title: "Error", description: "Could not find opponent for call.", variant: "destructive" });
        return;
    }

    // Placeholder for actual call initiation logic
    toast({ title: "Call Initiated", description: `Starting a ${callType} call with ${opponent.username}... (SFU integration needed)`, duration: 3000 });
    console.log(`Simulating ${callType} call to ${opponent.username} (ID: ${opponent.id})`);
    // In a real app, this would send a WebSocket message or API call to initiate WebRTC signaling.
  };


  if (!activeConversationId || !activeConversation) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center text-text-secondary">
        <MessagesSquare className="w-16 h-16 mb-4" />
        <h2 className="text-xl font-semibold">Select a conversation</h2>
        <p>Start chatting with your friends and groups.</p>
      </div>
    );
  }

  const currentActiveGame = games.find(g => g.id === activeGameId); // Use activeGameId from store

  return (
    <div className="flex-1 flex flex-col">
      <header className="p-4 border-b border-border bg-background-primary z-10 flex items-center justify-between">
        <h2 className="text-xl font-bold">{activeConversation.name}</h2>
        <div className="flex space-x-2">
            {activeConversation.type === ConversationType.OneToOne && (
                <>
                    <Button variant="ghost" size="icon" title="Start Audio Call" onClick={() => handleStartCall('audio')}>
                        <PhoneCall className="w-5 h-5" />
                    </Button>
                    <Button variant="ghost" size="icon" title="Start Video Call" onClick={() => handleStartCall('video')}>
                        <Video className="w-5 h-5" />
                    </Button>
                    <Button variant="ghost" size="icon" title="Invite to Game" onClick={handleInviteToGame}>
                        <Gamepad className="w-5 h-5" />
                    </Button>
                    <Dialog open={showClearHistoryDialog} onOpenChange={setShowClearHistoryDialog}>
                        <DialogTrigger asChild>
                            <Button variant="ghost" size="icon" title="Clear Local History">
                                <Trash2 className="w-5 h-5" />
                            </Button>
                        </DialogTrigger>
                        <DialogContent>
                            <DialogHeader>
                                <DialogTitle>Clear Local Chat History</DialogTitle>
                                <DialogDescription>
                                    Are you sure you want to clear the local history for this conversation?
                                    You can choose to preserve the last N messages.
                                </DialogDescription>
                            </DialogHeader>
                            <div className="flex items-center space-x-2 py-4">
                                <Input
                                    type="number"
                                    min="0"
                                    value={messagesToKeep}
                                    onChange={(e) => setMessagesToKeep(Number(e.target.value))}
                                    className="w-24"
                                />
                                <label className="text-sm text-text-secondary">messages to keep</label>
                            </div>
                            <DialogFooter>
                                <Button variant="outline" onClick={() => setShowClearHistoryDialog(false)}>Cancel</Button>
                                <Button variant="destructive" onClick={handleClearLocalHistory}>Clear History</Button>
                            </DialogFooter>
                        </DialogContent>
                    </Dialog>
                    {/* Dialog for deleting 1-1 chat */}
                    <Dialog>
                        <DialogTrigger asChild>
                            <Button variant="ghost" size="icon" title="Delete Chat">
                                <Trash2 className="w-5 h-5" />
                            </Button>
                        </DialogTrigger>
                        <DialogContent>
                            <DialogHeader>
                                <DialogTitle>Delete One-on-One Chat</DialogTitle>
                                <DialogDescription>
                                    Are you sure you want to delete this conversation for yourself? This action cannot be undone.
                                </DialogDescription>
                            </DialogHeader>
                            <DialogFooter>
                                <Button variant="outline">Cancel</Button>
                                <Button variant="destructive" onClick={handleDeleteOneToOneChat}>Delete Chat</Button>
                            </DialogFooter>
                        </DialogContent>
                    </Dialog>
                </>
            )}
            {activeConversation.type === ConversationType.Group && (
                <>
                    <Button variant="ghost" size="icon" title="Start Group Audio Call" onClick={() => handleStartCall('audio')}>
                        <PhoneCall className="w-5 h-5" />
                    </Button>
                    <Button variant="ghost" size="icon" title="Start Group Video Call" onClick={() => handleStartCall('video')}>
                        <Video className="w-5 h-5" />
                    </Button>
                     {/* Potentially other group actions like leave/remove members via another dialog */}
                </>
            )}
        </div>
      </header>
      <main className="flex-1 overflow-y-auto p-4 flex flex-col-reverse">
        <div ref={endOfMessagesRef} /> {/* For auto-scrolling to bottom */}
        <div className="flex flex-col">
           <AnimatePresence>
            {conversationMessages.map((msg, index) => (
              <motion.div
                key={msg.id}
                layout
                initial={{ opacity: 0, y: 50 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0 }}
                ref={index === 0 ? topMessageRef : null}
              >
                <MessageBubble message={msg} />
              </motion.div>
            ))}
          </AnimatePresence>
           {isLoadingOlderMessages && (
             <div className="text-center p-4 flex items-center justify-center text-primary-accent">
               <Loader2 className="h-5 w-5 animate-spin mr-2" /> Loading older messages...
             </div>
           )}
        </div>
      </main>
      <footer className="p-4 border-t border-border bg-background-primary">
        {currentActiveGame && currentActiveGame.status === 'active' ? (
            <TicTacToeGame game={currentActiveGame} />
        ) : (
            <MessageInput conversationId={activeConversationId} />
        )}
      </footer>
    </div>
  );
};

export default ChatWindow;
