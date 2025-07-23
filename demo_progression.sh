#!/bin/bash

echo "🎮 Balatro CLI - Blind Progression Demo"
echo "======================================="
echo "This demo shows progression through the first ante's three blinds"
echo ""

# Demo function that plays a sequence of hands
demo_blind() {
    local seed=$1
    local commands=$2
    local blind_name=$3

    echo "🎯 Playing $blind_name..."
    echo "Commands: $commands"
    echo ""

    echo "$commands" | go run . -seed $seed
    echo ""
    echo "----------------------------------------"
    echo ""
}

# Build the game first
echo "Building game..."
go build -o balatro_demo .

if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

echo "✅ Build successful!"
echo ""

# Demo Small Blind progression (seed 123 gives us good cards)
echo "📍 DEMO: Small Blind (Target: 300 points)"
echo "play 1 2 3 4 5
play 1 2 3
quit" | ./balatro_demo -seed 123

echo ""
echo "🎊 That was the Small Blind! In a real game, this would advance to Big Blind"
echo "   Big Blind target would be 450 points (1.5x harder)"
echo "   Boss Blind target would be 600 points (2x harder)"
echo ""

# Show the blind requirements table
echo "📊 BLIND REQUIREMENTS REFERENCE:"
echo "================================="
echo "Ante | Small | Big   | Boss"
echo "-----|-------|-------|-------"
for ante in {1..8}; do
    small=$((300 + (ante-1)*75))
    big=$((small * 3 / 2))
    boss=$((small * 2))
    printf "%-4d | %-5d | %-5d | %-5d\n" $ante $small $big $boss
done

echo ""
echo "🎮 To play the full game, run: go run ."
echo "💡 Use -seed <number> for reproducible gameplay"

# Clean up
rm -f balatro_demo
