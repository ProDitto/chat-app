CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    profile_picture_url TEXT,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY,
    type VARCHAR(20) NOT NULL, -- 'one-on-one' or 'group'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS conversation_participants (
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_read_timestamp TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (conversation_id, user_id)
);

CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL CHECK (length(content) <= 500),
    server_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS friendships (
    id UUID PRIMARY KEY,
    user_id1 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- requester
    user_id2 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- recipient
    status VARCHAR(20) NOT NULL, -- 'pending', 'accepted', 'declined'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id1, user_id2)
);

CREATE TABLE IF NOT EXISTS groups (
    id UUID PRIMARY KEY REFERENCES conversations(id) ON DELETE CASCADE, -- Group ID is also its conversation ID
    name VARCHAR(20) NOT NULL,
    slug VARCHAR(20) UNIQUE NOT NULL CHECK (slug ~ '^[a-z0-9_]+$'), -- lowercase, numbers, underscore
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT, -- Prevent deleting user if they own a group
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- group_members table is implicitly handled by conversation_participants where conversation.type = 'group'

CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY,
    player1_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    player2_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    initiator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    game_type VARCHAR(50) NOT NULL, -- e.g., 'tic-tac-toe'
    status VARCHAR(20) NOT NULL, -- 'pending', 'active', 'finished', 'declined'
    state JSONB NOT NULL, -- Stores game-specific state (e.g., TicTacToeState)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- User to whom this event is relevant
    event_type VARCHAR(50) NOT NULL, -- e.g., 'new_message', 'friend_request', 'game_invite', 'game_update'
    payload JSONB NOT NULL, -- The actual data of the event
    server_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE INDEX IF NOT EXISTS idx_messages_conversation_timestamp ON messages (conversation_id, server_timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_friendships_user1 ON friendships(user_id1);
CREATE INDEX IF NOT EXISTS idx_friendships_user2 ON friendships(user_id2);
CREATE INDEX IF NOT EXISTS idx_groups_owner ON groups(owner_id);
CREATE INDEX IF NOT EXISTS idx_games_player1 ON games(player1_id);
CREATE INDEX IF NOT EXISTS idx_games_player2 ON games(player2_id);
CREATE INDEX IF NOT EXISTS idx_games_initiator ON games(initiator_id);
CREATE INDEX IF NOT EXISTS idx_events_user_timestamp ON events (user_id, server_timestamp DESC);
