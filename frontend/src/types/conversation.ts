import type { Message } from "./message";

export const ConversationType = {
    OneToOne: 'one-on-one',
    Group: 'group',
} as const;

export interface Conversation {
    id: string;
    name: string; // User's name for 1-on-1, group name for group
    unreadCount: number;
    lastMessage?: Message;
}

export type ConversationType = typeof ConversationType[keyof typeof ConversationType];