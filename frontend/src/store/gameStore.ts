import { create } from 'zustand';
import type { Game } from '../types/game';
import { toast } from '../hooks/use-toast';
import { makeGameMove, getGame } from '../api/game';

interface GameState {
    activeGameId: string | null;
    games: Game[]; // All games the user is involved in
    setActiveGameId: (gameId: string | null) => void;
    addGame: (game: Game) => void;
    updateGame: (updatedGame: Game) => void;
    makeMove: (gameId: string, row: number, col: number) => Promise<void>;
    handleGameEvent: (gameData: any) => void; // For WebSocket/Long Polling events
    clearActiveGame: () => void;
    fetchGame: (gameId: string) => Promise<void>;
}

export const useGameStore = create<GameState>((set, get) => ({
    activeGameId: null,
    games: [],

    setActiveGameId: (gameId) => {
        set({ activeGameId: gameId });
        if (gameId && !get().games.find(g => g.id === gameId)) {
            get().fetchGame(gameId); // Fetch if not already in store
        }
    },

    addGame: (game) => {
        set((state) => {
            if (!state.games.some(g => g.id === game.id)) {
                return { games: [...state.games, game] };
            }
            return state;
        });
    },

    updateGame: (updatedGame) => {
        set((state) => ({
            games: state.games.map(game => 
                game.id === updatedGame.id ? updatedGame : game
            ),
        }));
    },
    
    fetchGame: async (gameId) => {
        try {
            const game = await getGame(gameId);
            get().addGame(game);
        } catch (error: any) {
            console.error("Failed to fetch game:", error);
            toast({ title: "Error", description: error.response?.data?.error || "Failed to load game.", variant: "destructive" });
        }
    },

    makeMove: async (gameId, row, col) => {
        try {
            const updatedGame = await makeGameMove(gameId, row, col);
            get().updateGame(updatedGame);
        } catch (error: any) {
            console.error("Failed to make game move:", error);
            throw error; // Re-throw for UI to handle specific error messages
        }
    },

    handleGameEvent: (gameData) => {
        const incomingGame: Game = gameData;
        set((state) => {
            const existingGameIndex = state.games.findIndex(g => g.id === incomingGame.id);
            if (existingGameIndex > -1) {
                // Update existing game
                const updatedGames = [...state.games];
                updatedGames[existingGameIndex] = incomingGame;
                return { games: updatedGames };
            } else {
                // Add new game (e.g., a new invitation)
                return { games: [...state.games, incomingGame] };
            }
        });
        // If the updated game is the active one, ensure UI updates
        if (get().activeGameId === incomingGame.id) {
            // No explicit re-setting needed as `updateGame` already handles state immutability,
            // but if the component relies on `activeGameId` changing to trigger effects,
            // this could be `set({ activeGameId: incomingGame.id });` for force refresh,
            // though generally direct state updates should suffice.
        }
    },
    
    clearActiveGame: () => set({ activeGameId: null }),
}));
