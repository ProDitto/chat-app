import type { Game, TicTacToeState } from '../../types/game';
import { useAuthStore } from '../../store/authStore';
import { useGameStore } from '../../store/gameStore';
import { toast } from '../../hooks/use-toast';
import { cn } from '../../lib/utils';
import { Loader2 } from 'lucide-react';
import { useEffect, useState } from 'react';

interface TicTacToeGameProps {
    game: Game;
}

export const TicTacToeGame = ({ game }: TicTacToeGameProps) => {
    const { user } = useAuthStore();
    const { makeMove } = useGameStore();
    const [isLoadingMove, setIsLoadingMove] = useState(false);
    const [gameState, setGameState] = useState<TicTacToeState | null>(null);

    useEffect(() => {
        try {
            const parsedState: TicTacToeState = JSON.parse(game.state as string);
            setGameState(parsedState);
        } catch (error) {
            console.error("Failed to parse game state:", error);
            toast({ title: "Error", description: "Failed to load game state.", variant: "destructive" });
        }
    }, [game.state]);

    if (!user || !gameState) {
        return <div className="text-center py-4">Loading game...</div>;
    }

    const isPlayer1 = user.id === game.player1Id;
    // const isPlayer2 = user.id === game.player2Id;
    const playerMark = isPlayer1 ? "X" : "O";
    const opponentMark = isPlayer1 ? "O" : "X";
    const isMyTurn = gameState.currentTurn === user.id;

    const handleCellClick = async (row: number, col: number) => {
        if (isLoadingMove || !isMyTurn || gameState.board[row][col] !== "" || gameState.winner) {
            return;
        }

        setIsLoadingMove(true);
        try {
            await makeMove(game.id, row, col);
        } catch (error: any) {
            toast({ title: "Error", description: error.response?.data?.error || "Failed to make move.", variant: "destructive" });
        } finally {
            setIsLoadingMove(false);
        }
    };

    let statusMessage = "Game Active";
    if (gameState.winner) {
        if (gameState.winner === user.id) {
            statusMessage = "You Win!";
            toast({ title: "Game Over!", description: "You won the game!", variant: "success" });
        } else if (gameState.winner === "draw") {
            statusMessage = "It's a Draw!";
            toast({ title: "Game Over!", description: "The game is a draw!", variant: "default" });
        } else {
            statusMessage = "You Lose!";
            toast({ title: "Game Over!", description: "You lost the game.", variant: "destructive" });
        }
    } else if (isMyTurn) {
        statusMessage = "Your Turn ( " + playerMark + " )";
    } else {
        statusMessage = "Opponent's Turn ( " + opponentMark + " )";
    }

    return (
        <div className="p-4 bg-background-primary rounded-lg shadow-md">
            <h3 className="text-lg font-bold mb-2">Tic-Tac-Toe</h3>
            <p className="text-sm text-text-secondary mb-4">{statusMessage}</p>

            <div className="grid grid-cols-3 gap-2 w-full max-w-xs mx-auto">
                {gameState.board.map((row, rowIndex) => (
                    row.map((cell, colIndex) => (
                        <button
                            key={`${rowIndex}-${colIndex}`}
                            className={cn(
                                "w-20 h-20 flex items-center justify-center text-4xl font-bold rounded-md",
                                "bg-background-secondary border border-border",
                                !gameState.winner && isMyTurn && cell === "" && "hover:bg-primary-accent/20 cursor-pointer",
                                cell === "X" && "text-blue-500",
                                cell === "O" && "text-red-500",
                                isLoadingMove && "opacity-70 cursor-not-allowed"
                            )}
                            onClick={() => handleCellClick(rowIndex, colIndex)}
                            disabled={isLoadingMove || !isMyTurn || cell !== "" || !!gameState.winner}
                        >
                            {cell}
                        </button>
                    ))
                ))}
            </div>

            {isLoadingMove && (
                <div className="flex items-center justify-center mt-4 text-primary-accent">
                    <Loader2 className="h-5 w-5 animate-spin mr-2" /> Making move...
                </div>
            )}
        </div>
    );
};
