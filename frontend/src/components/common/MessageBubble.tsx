import { useAuthStore } from '../../store/authStore';
import type { Message } from '../../types/message';
import UserAvatar from './UserAvatar';

interface MessageBubbleProps {
  message: Message;
}

const MessageBubble = ({ message }: MessageBubbleProps) => {
  const { user } = useAuthStore();
  const isCurrentUser = message.sender.id === user?.id;

  return (
    <div className={`flex items-start gap-3 my-4 ${isCurrentUser ? 'justify-end' : ''}`}>
      {!isCurrentUser && <UserAvatar user={message.sender} size="sm" />}
      <div
        className={`max-w-xs md:max-w-md lg:max-w-lg p-3 rounded-lg ${
          isCurrentUser
            ? 'bg-primary-accent text-white rounded-br-none'
            : 'bg-background-primary border border-border rounded-bl-none'
        }`}
      >
        {!isCurrentUser && (
            <p className="text-sm font-semibold mb-1 text-secondary-accent">{message.sender.username}</p>
        )}
        <p className="text-base">{message.content}</p>
        <p className="text-xs mt-1 text-right opacity-70">
          {new Date(message.serverTimestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
        </p>
      </div>
    </div>
  );
};

export default MessageBubble;