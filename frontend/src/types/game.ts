export type Cell = "" | "X" | "O";

export type TicTacToeBoard = [
    [Cell, Cell, Cell],
    [Cell, Cell, Cell],
    [Cell, Cell, Cell],
];

export interface TicTacToeState {
    board: TicTacToeBoard;
    currentTurn: string; // User ID
    winner?: string; // User ID, "draw", or "X", "O"
    last_move?: {
        row: number;
        col: number;
    }
}

export interface Game {
    id: string;
    player1Id: string;
    player2Id: string;
    initiatorId: string;
    gameType: string; // "tic-tac-toe"
    status: "pending" | "active" | "finished" | "declined";
    state: string; // JSON string of TicTacToeState
    createdAt: string;
    updatedAt: string;
}
