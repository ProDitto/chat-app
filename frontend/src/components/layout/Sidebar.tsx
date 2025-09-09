import { useChatStore } from '../../store/chatStore';
import { useAuthStore } from '../../store/authStore';
import UserAvatar from '../common/UserAvatar';
import { LogOut, Settings, Users } from 'lucide-react';
import { FriendshipManager } from '../friends/FriendshipManager';
import { Link } from 'react-router-dom';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '../../components/ui/Dialog';
import { useState } from 'react';
import { createGroup } from '../../api/group';
import { toast } from '../../hooks/use-toast';
// import { useTheme } from '../../components/theme/ThemeProvider';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';

const Sidebar = () => {
  const { conversations, setActiveConversationId, activeConversationId } = useChatStore();
  const { user, logout } = useAuthStore();
  const [showCreateGroupDialog, setShowCreateGroupDialog] = useState(false);
  const [newGroupName, setNewGroupName] = useState('');
  const [newGroupSlug, setNewGroupSlug] = useState('');
  // const { theme, setTheme } = useTheme();

  const handleCreateGroup = async () => {
    if (!newGroupName.trim()) {
      toast({ title: 'Error', description: 'Group name cannot be empty.', variant: 'destructive' });
      return;
    }
    if (!newGroupSlug.trim()) {
      toast({ title: 'Error', description: 'Group slug cannot be empty.', variant: 'destructive' });
      return;
    }
    // Basic slug validation, stricter validation happens on backend
    if (!/^[a-z0-9_]+$/.test(newGroupSlug)) {
      toast({ title: 'Error', description: 'Group slug can only contain lowercase alphabets, numbers, and underscore.', variant: 'destructive' });
      return;
    }
    if (newGroupName.length > 20) {
      toast({ title: 'Error', description: 'Group name must be at most 20 characters.', variant: 'destructive' });
      return;
    }
    if (newGroupSlug.length > 20) {
      toast({ title: 'Error', description: 'Group slug must be at most 20 characters.', variant: 'destructive' });
      return;
    }

    try {
      await createGroup(newGroupName, newGroupSlug, []); // For MVP, no initial members beyond owner
      toast({ title: 'Success', description: `Group "${newGroupName}" created.` });
      setNewGroupName('');
      setNewGroupSlug('');
      setShowCreateGroupDialog(false);
      useChatStore.getState().fetchConversations(); // Refresh conversations
    } catch (error: any) {
      toast({ title: 'Error', description: error.response?.data?.error || 'Failed to create group.', variant: 'destructive' });
    }
  };

  return (
    <div className="flex flex-col w-80 border-r border-border bg-background-primary">
      <div className="p-4 border-b border-border flex items-center justify-between">
        <Link to="/settings" className="flex items-center gap-2 group">
          <UserAvatar user={user || {}} />
          <span className="font-semibold group-hover:text-primary-accent transition-colors">{user?.username || 'Guest'}</span>
        </Link>
        <div className="flex items-center space-x-2">
          <Dialog open={showCreateGroupDialog} onOpenChange={setShowCreateGroupDialog}>
            <DialogTrigger asChild>
              <Button variant="ghost" size="icon" title="Create Group">
                <Users className="w-5 h-5" />
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Create New Group</DialogTitle>
                <DialogDescription>
                  Enter a name and a unique slug for your new group.
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-4 py-2">
                <Input
                  placeholder="Group Name"
                  value={newGroupName}
                  onChange={(e) => setNewGroupName(e.target.value)}
                  maxLength={20}
                />
                <Input
                  placeholder="Group Slug (unique, lowercase, numbers, _)"
                  value={newGroupSlug}
                  onChange={(e) => setNewGroupSlug(e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, ''))}
                  maxLength={20}
                />
              </div>
              <Button onClick={handleCreateGroup}>Create Group</Button>
            </DialogContent>
          </Dialog>

          <Link to="/settings">
            <Button variant="ghost" size="icon" title="Settings">
              <Settings className="w-5 h-5" />
            </Button>
          </Link>
          <Button variant="ghost" size="icon" onClick={logout} title="Log out">
            <LogOut className="w-5 h-5" />
          </Button>
        </div>
      </div>

      <FriendshipManager />

      <div className="flex-1 overflow-y-auto p-2">
        <h2 className="p-2 text-xs font-semibold text-text-secondary uppercase">Conversations</h2>
        <nav className="space-y-1">
          {conversations.map((convo) => (
            <button
              key={convo.id}
              onClick={() => setActiveConversationId(convo.id)}
              className={`w-full text-left flex items-center p-2 rounded-md transition-colors ${activeConversationId === convo.id
                  ? 'bg-primary-accent/10 text-primary-accent'
                  : 'hover:bg-background-secondary'
                }`}
            >
              <UserAvatar user={{ username: convo.name, profilePictureUrl: convo.group?.id ? '' : (convo.participants && convo.participants.length > 0 ? convo.participants[0].profilePictureUrl : '') }} />
              <div className="ml-3 flex-1 overflow-hidden">
                <p className="font-semibold">{convo.name}</p>
                {convo.lastMessage && (
                  <p className="text-sm text-text-secondary truncate">{convo.lastMessage.content}</p>
                )}
              </div>
              {convo.unreadCount > 0 && (
                <span className="ml-2 bg-primary-accent text-white text-xs font-bold rounded-full h-5 w-5 flex items-center justify-center">
                  {convo.unreadCount}
                </span>
              )}
            </button>
          ))}
        </nav>
      </div>
    </div>
  );
};

export default Sidebar;
