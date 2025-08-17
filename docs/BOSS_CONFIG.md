# YAML Boss Configuration

Bosses are defined in `bosses.yaml` and loaded at runtime.

## `bosses.yaml` Structure

```yaml
bosses:
  - name: "Skull King"
    effect: "DoubleChips"
  - name: "The Void"
    effect: "HalveMoney"
    final: true
```

- `name`: Display name of the boss.
- `effect`: Identifier for the boss effect.
- `final`: Optional flag. Final bosses only appear on antes divisible by 8.

## Available Effects

- `DoubleChips` – Doubles the chip target needed to defeat the blind.
- `HalveMoney` – Halves the player's money when the blind starts.
