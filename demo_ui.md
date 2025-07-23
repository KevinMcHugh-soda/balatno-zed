# Balatro CLI - Terminal UI Demo

## Static Layout Design

The terminal UI uses a completely static layout with fixed regions that never move. Data is rendered in-place rather than being added to new parts of the screen.

## Layout Structure

```
Row 0:  🃏 BALATRO CLI 🃏                                    [TITLE - Fixed]
Row 1:  
Row 2:  🔸 Ante: 1/8 | Blind: Small Blind                   [GAME INFO - Fixed]
Row 3:  🎯 Score: 150/300 [████████████░░░░░░] (50.0%)      [SCORE - Fixed]
Row 4:  💰 Money: $25                                        [MONEY - Fixed]
Row 5:  🎴 Hands Left: 3/4 | 🗑️ Discards Left: 2/3         [HANDS/DISCARDS - Fixed]
Row 6:  🃏 Jokers: None                                      [JOKERS - Fixed]
Row 7:  
Row 8:  Your Cards (sorted by rank):                         [CARDS HEADER - Fixed]
Row 9:  A:2♠  B:7♥  C:8♣  D:9♦  E:J♠  F:K♥  G:A♣  H:A♦     [CARDS ROW 1 - Fixed]
Row 10: 
Row 11: 
Row 12: 
...
Row -5: Commands: (p)lay <cards>, (d)iscard <cards>, (r)esort, (q)uit | Example: 'play A B'  [COMMANDS - Fixed from bottom]
Row -4: > play A B_                                          [INPUT - Fixed from bottom]
Row -3: 
Row -2: ✨ Scored 15 points! (+$3)                          [MESSAGE - Fixed from bottom]
Row -1: 
```

## Key Features

### Fixed Regions
- **Title Bar**: Always at row 0
- **Game Status**: Rows 2-6, never moves
- **Cards Area**: Rows 8-12, always in same position
- **Input Section**: Always 4-5 rows from bottom
- **Message Area**: Always 2 rows from bottom

### Static Data Display
- Ante and blind information always in same spot
- Score progress bar updates in-place
- Money counter updates in-place
- Cards are displayed in a fixed grid layout
- Command prompt always at same position

### Card Selection
- Cards are labeled A, B, C, etc. for easy selection
- Players can use either letters (A B C) or numbers (1 2 3)
- Cards are sorted by rank or suit (toggle with 'r' command)

## Usage Examples

### Playing Cards
```
> play A B C
✨ Scored 45 points! (+$8)
```

### Discarding Cards
```
> discard D E
🗑️ Discarded 2 cards
```

### Resorting Cards
```
> r
🃏 Cards sorted by suit
```

## Benefits of Static Layout

1. **Predictable Interface**: Players always know where to look for information
2. **Clean Updates**: No scrolling text or moving elements
3. **Easy Scanning**: Important info always in same location
4. **Professional Look**: Like a real terminal application
5. **Consistent Experience**: Interface doesn't change as game progresses

## Running the Demo

```bash
cd balatno
go build
./balatno
```

Use Escape or Ctrl+C to quit the game.