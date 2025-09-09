import type { User } from "./user";

export interface Message {
  id: string;
  conversationId: string;
  sender: User;
  content: string;
  serverTimestamp: string;
}