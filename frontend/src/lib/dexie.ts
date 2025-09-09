import Dexie, { type Table } from 'dexie';
import type { Message } from '../types/message';
import type { User } from '../types/user';

class ChatDB extends Dexie {
  messages!: Table<Message>;
  users!: Table<User>;

  constructor() {
    super('ChatDatabase');
    this.version(1).stores({
      messages: '++id, conversationId, serverTimestamp',
      users: '&id, username',
    });
  }
}

export const db = new ChatDB();

// Function to clear history while preserving the last 'n' messages
export const clearHistory = async (conversationId: string, messagesToKeep: number) => {
    const messageCount = await db.messages.where({ conversationId }).count();
    if (messageCount > messagesToKeep) {
        const messagesToDelete = await db.messages
            .where({ conversationId })
            .sortBy('serverTimestamp')
            // .limit(messageCount - messagesToKeep)
            // .toArray();

        const idsToDelete = messagesToDelete.map(m => m.id);
        await db.messages.bulkDelete(idsToDelete);
    }
};
