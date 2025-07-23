# Balance Configuration System

This document explains how to modify Balatro CLI's gameplay balance without recompiling the game. All balance values are stored in CSV files that are loaded at runtime.

## 📁 Configuration Files

### `ante_requirements.csv` - Blind Score Requirements
Controls how many points are needed to beat each blind in each ante.

**Format:**
```csv
small,big,boss
300,450,600
375,562,750
...
```

- Each row represents one ante (row 1 = Ante 1, row 2 = Ante 2, etc.)
- Columns: `small` (Small Blind), `big` (Big Blind), `boss` (Boss Blind)
- Values: Points required to complete that blind

### `hand_scores.csv` - Poker Hand Values
Controls the base score and multiplier for each poker hand type.

**Format:**
```csv
hand,base,mult
High Card,5,1
Pair,10,2
Two Pair,20,2
...
```

- Each row represents one poker hand type
- Columns: `hand` (exact name), `base` (base score), `mult` (multiplier)
- Final score = (base + card values) × mult

## 🎯 Current Default Values

### Ante Requirements
| Ante | Small Blind | Big Blind | Boss Blind |
|------|-------------|-----------|------------|
| 1    | 300         | 450       | 600        |
| 2    | 375         | 562       | 750        |
| 3    | 450         | 675       | 900        |
| 4    | 525         | 787       | 1050       |
| 5    | 600         | 900       | 1200       |
| 6    | 675         | 1012      | 1350       |
| 7    | 750         | 1125      | 1500       |
| 8    | 825         | 1237      | 1650       |

### Hand Scores
| Hand Type | Base Score | Multiplier |
|-----------|------------|------------|
| High Card | 5 | 1x |
| Pair | 10 | 2x |
| Two Pair | 20 | 2x |
| Three of a Kind | 30 | 3x |
| Straight | 30 | 4x |
| Flush | 35 | 4x |
| Full House | 40 | 4x |
| Four of a Kind | 60 | 7x |
| Straight Flush | 100 | 8x |
| Royal Flush | 100 | 8x |

## 🔧 Making Changes

### Example 1: Easier Early Game
Make Ante 1 more forgiving for new players:

```csv
small,big,boss
200,300,400
375,562,750
450,675,900
...
```

**Impact**: Ante 1 becomes much easier to complete

### Example 2: Buff Weak Hands
Make high card and pair more viable:

```csv
hand,base,mult
High Card,10,2
Pair,15,3
Two Pair,20,2
...
```

**Impact**: High Card now scores (10 + cards) × 2 instead of (5 + cards) × 1

### Example 3: Rebalance Hand Hierarchy
Make straights and flushes more rewarding:

```csv
hand,base,mult
...
Straight,40,5
Flush,45,5
Full House,40,4
...
```

**Impact**: Straights/flushes become more valuable than full house

## 🧪 Testing Your Changes

### Quick Test
```bash
# Test a specific hand with your changes
echo "play 1 2 3" | go run . -seed 42

# Check the score output matches your expectations
```

### Full Validation
```bash
# Run all tests to ensure nothing breaks
go test

# Test multiple scenarios
echo "play 1 2 3 4 5" | go run . -seed 123
```

### Verification Checklist
- [ ] Hand scores calculate correctly with new values
- [ ] Ante progression uses your modified requirements  
- [ ] No parsing errors on game startup
- [ ] All hand types are present in CSV
- [ ] All 8 antes are defined in CSV

## ⚠️ Important Rules

### CSV Format Requirements
- **Headers must match exactly**: `small,big,boss` and `hand,base,mult`
- **No extra spaces**: `Pair,10,2` not `Pair, 10, 2`
- **Exact hand names**: Must match exactly what the game expects
- **Integer values only**: No decimals or fractions

### Hand Name Requirements
Must use these exact names in `hand_scores.csv`:
- `High Card`
- `Pair` 
- `Two Pair`
- `Three of a Kind`
- `Straight`
- `Flush`
- `Full House`
- `Four of a Kind`
- `Straight Flush`
- `Royal Flush`

### Ante Requirements
- Must have exactly 8 rows (one per ante)
- Big Blind should typically be harder than Small Blind
- Boss Blind should typically be hardest

## 🛡️ Fallback Behavior

If CSV files are missing or invalid:
- **Game will not crash** - uses hardcoded defaults
- **Warning message** displays at startup
- **All functionality preserved** with original values

Example warning:
```
Warning: Could not load hand_scores.csv, using defaults: file not found
```

## 💡 Balance Design Tips

### Progressive Difficulty
- Each ante should be meaningfully harder than the last
- Boss Blinds should feel like major challenges
- Consider how money/joker systems interact with your values

### Hand Value Relationships
- **Base vs Multiplier**: High base = reliable, high mult = explosive
- **Rarity vs Power**: Rare hands should feel appropriately rewarding
- **Strategic Diversity**: Multiple viable approaches at each ante level

### Testing Different Strategies
- **Efficiency builds**: Lower requirements reward unused hands/discards
- **Power builds**: Higher hand values enable big score runs
- **Economic builds**: Balance scoring with joker affordability

## 📊 Advanced Examples

### "High Stakes" Mode (50% Harder)
Multiply all ante requirements by 1.5:
```csv
small,big,boss
450,675,900
562,843,1125
675,1012,1350
...
```

### "Pair Paradise" (Pair-Focused Meta)
```csv
hand,base,mult
High Card,5,1
Pair,20,3
Two Pair,35,4
Three of a Kind,25,3
...
```

### "Quick Game" (Lower Requirements)
```csv
small,big,boss
200,300,400
250,375,500
300,450,600
350,525,700
400,600,800
450,675,900
500,750,1000
550,825,1100
```

## 🔄 Live Editing Workflow

1. **Make CSV changes** while game is not running
2. **Start game** - changes load automatically
3. **Test changes** with known seed for consistency
4. **Iterate** - exit game, modify CSV, restart
5. **Validate** - run tests when satisfied

The configuration system makes balance iteration fast and safe - experiment freely!