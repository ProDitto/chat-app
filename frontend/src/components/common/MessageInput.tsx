import { useState } from 'react';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { SendHorizonal } from 'lucide-react';
import { useSocketStore } from '../../store/socketStore';
import { useAuthStore } from '../../store/authStore';

interface MessageInputProps {
  conversationId: string;
}

const MessageInput = ({ conversationId }: MessageInputProps) => {
  const [content, setContent] = useState('');
  const { sendMessage } = useSocketStore();
  const user = useAuthStore(state => state.user);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (content.trim() && user) {
      const payload = {
        conversation_id: conversationId,
        content: content.trim(),
        sender_id: user.id
      };
      sendMessage('send_message', payload);
      setContent('');
    }
  };

  return (
    <form onSubmit={handleSubmit} className="flex items-center gap-2">
      <Input
        type="text"
        placeholder="Type a message..."
        value={content}
        onChange={(e) => setContent(e.target.value)}
        className="flex-1"
        maxLength={500}
      />
      <Button type="submit" size="icon" disabled={!content.trim()}>
        <SendHorizonal className="w-5 h-5" />
      </Button>
    </form>
  );
};

export default MessageInput;
