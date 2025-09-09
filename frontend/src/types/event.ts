// Mirrors backend Event structure
export type EventType = 
    "new_message" | 
    "friend_request" | 
    "friend_accepted" | 
    "game_invite" | 
    "game_update" |
    "group_created" |
    "group_joined" |
    "group_left" |
    "conversation_deleted";

export interface Event {
    id: string;
    user_id: string; // The user this event is primarily relevant to
    event_type: EventType;
    payload: any; // The actual data of the event, type depends on event_type
    server_timestamp: string; // ISO 8601 date string
}
