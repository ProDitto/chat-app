import { api } from './client';
import { type Event } from '../types/event';

// This is the long-polling fallback endpoint
export const fetchEvents = async (sinceEventId?: string, limit: number = 50): Promise<Event[]> => {
    const params = new URLSearchParams();
    if (sinceEventId) {
        params.append('since', sinceEventId);
    }
    params.append('limit', String(limit));
    // Set a client-side timeout slightly less than the server's long polling timeout
    // to ensure the client can re-initiate a new request before the server closes the connection.
    const response = await api.get('/events', { params, timeout: 59000 });
    return response.data;
};
// -- frontend/src/types/user.ts
export interface User {
    id: string;
    username: string;
    profilePictureUrl?: string; // Optional now
    isVerified?: boolean;
}
