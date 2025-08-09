# Balatno Architecture: Clean Separation of Game Logic and User Interface

## Overview

Balatno demonstrates a well-factored architecture that cleanly separates game logic from user interface concerns. This separation enables multiple UI implementations (console and TUI) while maintaining a single, pure game engine. The architecture follows event-driven design patterns with clear boundaries between components.

## Architectural Components

### Core Game Engine (`game.go`)
The heart of the system is the `Game` struct, which contains all game state and logic:
- Game state (score, hands, discards, money, cards, jokers)
- Game rules and mechanics (hand evaluation, blind progression, shop logic)
- Pure business logic with no UI dependencies
- Communicates only through events via `EventEmitter`

### Event System (`game_events.go`)
The event system provides the communication backbone:
- **Events**: Immutable data structures representing game state changes
- **EventEmitter**: Publishes events to registered handlers
- **EventHandler**: Interface for processing events and requesting player input

### UI Implementations
Two complete UI implementations share the same game engine:

#### Console Mode (`logger_event_handler.go`)
- Text-based interface using stdin/stdout
- Handles events by printing formatted output
- Gets player input through `bufio.Scanner`
- Ideal for debugging and simple gameplay

#### TUI Mode (`tui.go`, `tui_event_handler.go`)
- Rich terminal interface using Bubbletea framework
- Real-time updates, visual card selection, timeout handling
- Mode-based UI system for different game states
- Asynchronous communication through channels

## Clean Separation of Concerns

### Game Logic Independence
The `Game` struct is completely UI-agnostic:
```go
type Game struct {
    totalScore        int
    handsPlayed       int
    discardsUsed      int
    deck              []Card
    playerCards       []Card
    currentAnte       int
    eventEmitter      *SimpleEventEmitter
    // ... other pure game state
}
```

The game engine:
- Contains no `fmt.Print` statements or UI code
- Makes no assumptions about how events are presented
- Requests player input through abstract interfaces
- Maintains game state without UI concerns

### Event-Driven Communication
All communication between game logic and UI flows through events:

**Game → UI (Events)**
- `GameStateChangedEvent`: Score, hands, money updates
- `HandPlayedEvent`: Results of playing cards
- `ShopOpenedEvent`: Shop items and state
- `BlindDefeatedEvent`: Progress notifications

**UI → Game (Player Actions)**
- `PlayerActionPlay`: Play selected cards
- `PlayerActionDiscard`: Discard and draw new cards
- `PlayerActionBuy`: Purchase shop items
- Communicated through channels in TUI, direct calls in console

### Interface Abstraction
The `EventHandler` interface defines the contract between game and UI:
```go
type EventHandler interface {
    HandleEvent(event Event)
    GetPlayerAction(canDiscard bool) (action PlayerAction, params []string, quit bool)
    GetShopAction() (action PlayerAction, params []string, quit bool)
    Close()
}
```

This abstraction allows:
- Multiple UI implementations without changing game code
- Testability through mock handlers
- Clean dependency inversion

## TUI Mode System

### Mode-Based UI Architecture
The TUI implements a sophisticated mode system that handles different game states:

```go
type Mode interface {
    renderContent(m TUIModel) string
    toggleHelp() Mode
    handleKeyPress(m *TUIModel, msg string) (tea.Model, tea.Cmd)
    getControls() string
}
```

#### Game Modes
- **GameMode**: Main gameplay interface
  - Card selection with number keys (1-7)
  - Play/discard actions
  - Visual progress indicators
- **ShoppingMode**: Shop interface
  - Item browsing and selection
  - Purchase and reroll actions
- **Help Modes**: Context-sensitive help screens

### Asynchronous Communication
Player actions in TUI mode use channel-based communication:

1. **Action Request**: Game requests player input through `PlayerActionRequest`
2. **UI Processing**: TUI handles user input and prepares response
3. **Response Channel**: Action result sent back through `PlayerActionResponse`
4. **Game Continuation**: Game processes action and continues

```go
type PlayerActionRequest struct {
    CanDiscard   bool
    ResponseChan chan PlayerActionResponse
}
```

This design prevents blocking and allows for:
- Timeout handling
- Real-time UI updates during input
- Graceful shutdown

## Player Action Flow

### From UI to Game Engine

1. **Input Capture**: UI captures user input (keypress, command)
2. **Action Translation**: UI translates input to `PlayerAction` with parameters
3. **Validation**: Basic validation at UI level (card selection limits)
4. **Communication**: Action sent to game through appropriate channel/method
5. **Game Processing**: Game validates action and updates state
6. **Event Emission**: Game emits events describing state changes
7. **UI Update**: Event handler updates UI to reflect new state

### Example: Playing Cards in TUI

1. User presses number keys to select cards
2. `GameMode.handleKeyPress` updates `selectedCards` state
3. User presses Enter/P to play
4. `handlePlay` converts selection to `PlayerActionPlay` with card indices
5. Action sent through `PlayerActionResponse` channel
6. Game validates hand, calculates score, updates state
7. Game emits `HandPlayedEvent` with results
8. TUI receives event and updates display

## Benefits of This Architecture

### Maintainability
- Clear boundaries between components
- Single responsibility for each module
- Easy to modify UI without touching game logic

### Testability
- Game logic can be tested independently with mock handlers
- UI components can be tested with fake game events
- Integration tests can verify event flow

### Extensibility
- New UI implementations require only implementing `EventHandler`
- New game features only need to emit appropriate events
- Additional game modes can be added without UI changes

### Flexibility
- Multiple UIs can run simultaneously
- Game can be embedded in other applications
- Different presentation styles for different contexts

## Code Organization

The architecture is reflected in the file organization:

**Core Game Logic:**
- `game.go` - Main game engine and state
- `hands.go`, `deck.go`, `jokers.go` - Game mechanics
- `game_events.go` - Event system definition

**UI Implementations:**
- `logger_event_handler.go` - Console mode
- `tui.go`, `tui_event_handler.go` - TUI framework
- `tui_game.go`, `tui_shop.go` - TUI mode implementations
- `tui_styles.go` - Visual styling

**Application Entry:**
- `main.go` - Orchestrates game creation and UI selection

This architecture demonstrates how to build maintainable, testable applications with clean separation between business logic and presentation concerns. The event-driven design and interface abstractions make the codebase flexible and extensible while keeping complexity manageable.