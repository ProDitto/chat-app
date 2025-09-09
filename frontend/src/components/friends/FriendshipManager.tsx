import { useEffect, useState } from 'react';
import { useFriendStore } from '../../store/friendStore';
import { UserPlus, Check, X, Bell } from 'lucide-react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '../ui/Dialog';
import UserAvatar from '../common/UserAvatar';
import { useChatStore } from '../../store/chatStore';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { toast } from '../../hooks/use-toast';

export function FriendshipManager() {
  const { requests, friends, fetchFriends, fetchRequests, sendRequest, respondToRequest } = useFriendStore();
  const [username, setUsername] = useState('');
  const { fetchConversations } = useChatStore();

  useEffect(() => {
    fetchFriends();
    fetchRequests();
  }, [fetchFriends, fetchRequests]);

  const handleSendRequest = async () => {
    if (!username.trim()) return;
    try {
      await sendRequest(username);
      toast({ title: 'Success', description: 'Friend request sent.' });
      setUsername('');
    } catch (error: any) {
      toast({ title: 'Error', description: error.response?.data?.error || 'Failed to send request.', variant: 'destructive' });
    }
  };

  const handleRespond = async (requestId: string, status: 'accepted' | 'declined') => {
    try {
      await respondToRequest(requestId, status);
      toast({ title: 'Success', description: `Request ${status}.` });
      fetchConversations(); // Refresh conversations to show new 1-on-1 chat
    } catch (error: any) {
      toast({ title: 'Error', description: error.response?.data?.error || 'Failed to respond.', variant: 'destructive' });
    }
  };

  return (
    <div className="p-4 border-b border-border">
      <Dialog>
        <DialogTrigger asChild>
          <Button variant="outline" className="w-full justify-start">
            <UserPlus className="mr-2 h-4 w-4" /> Add Friend
          </Button>
        </DialogTrigger>
        <DialogContent>
          <DialogHeader><DialogTitle>Add a Friend</DialogTitle></DialogHeader>
          <div className="flex items-center space-x-2">
            <Input
              placeholder="Enter username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
            />
            <Button onClick={handleSendRequest}>Send</Button>
          </div>
        </DialogContent>
      </Dialog>
      
      {requests.length > 0 && (
        <Dialog>
            <DialogTrigger asChild>
                <Button variant="ghost" className="w-full justify-start mt-2 relative">
                    <Bell className="mr-2 h-4 w-4" />
                    Pending Requests
                    <span className="ml-auto bg-status-error text-white text-xs font-bold rounded-full h-5 w-5 flex items-center justify-center">
                        {requests.length}
                    </span>
                </Button>
            </DialogTrigger>
            <DialogContent>
                <DialogHeader><DialogTitle>Pending Friend Requests</DialogTitle></DialogHeader>
                 {requests.map(req => (
                    <div key={req.id} className="flex items-center justify-between p-2 rounded hover:bg-background-secondary">
                        <div className="flex items-center gap-2">
                            <UserAvatar user={req.sender} size="sm" />
                            <span>{req.sender.username}</span>
                        </div>
                        <div>
                            <Button variant="ghost" size="icon" onClick={() => handleRespond(req.id, 'accepted')} title="Accept">
                                <Check className="h-4 w-4 text-status-success" />
                            </Button>
                            <Button variant="ghost" size="icon" onClick={() => handleRespond(req.id, 'declined')} title="Decline">
                                <X className="h-4 w-4 text-status-error" />
                            </Button>
                        </div>
                    </div>
                ))}
            </DialogContent>
        </Dialog>
      )}

      {friends.length > 0 && (
        <div className="mt-4">
          <h3 className="text-sm font-semibold mb-2">Friends ({friends.length})</h3>
          <div className="space-y-1">
            {friends.map(friend => (
              <div key={friend.id} className="flex items-center p-2 rounded hover:bg-background-secondary">
                <UserAvatar user={friend} size="sm" />
                <span className="ml-2">{friend.username}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
