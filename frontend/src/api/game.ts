import { api } from './client';
import type { Game } from '../types/game';

export const inviteToGame = async (opponentUsername: string, gameType: string): Promise<Game> => {
    const response = await api.post('/games/invite', { opponent_username: opponentUsername, game_type: gameType });
    return response.data;
};

export const getGame = async (gameId: string): Promise<Game> => {
    const response = await api.get(`/games/${gameId}`);
    return response.data;
};

export const makeGameMove = async (gameId: string, row: number, col: number): Promise<Game> => {
    const response = await api.post(`/games/${gameId}/move`, { row, col });
    return response.data;
};

export const respondToGameInvite = async (gameId: string, accept: boolean): Promise<Game> => {
    const response = await api.post(`/games/${gameId}/respond`, { accept });
    return response.data;
};
