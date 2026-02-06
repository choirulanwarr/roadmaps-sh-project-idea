package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type Difficulty struct {
	Name    string
	Chances int
}

type GameState struct {
	secretNumber int
	maxChances   int
	currentGuess int
	attempts     int
	startTime    time.Time
}

func main() {
	fmt.Println("Welcome to the Number Guessing Game!")

	for {
		playRound()

		fmt.Print("\nWould you like to play again? (y/n): ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		response := strings.ToLower(strings.TrimSpace(scanner.Text()))

		if response != "y" && response != "yes" {
			fmt.Println("Thanks for playing! Goodbye!")
			break
		}
		fmt.Println()
	}
}

func playRound() {
	// Display welcome message and rules
	fmt.Println("I'm thinking of a number between 1 and 100.")

	// Select difficulty
	difficulty := selectDifficulty()

	// Initialize game state
	gameState := GameState{
		secretNumber: generateRandomNumber(1, 100),
		maxChances:   difficulty.Chances,
		attempts:     0,
		startTime:    time.Now(),
	}

	fmt.Printf("Great! You have selected the %s difficulty level.\n", difficulty.Name)
	fmt.Printf("You have %d chances to guess the correct number.\n", difficulty.Chances)
	fmt.Println("Let's start the game!")
	fmt.Println()

	// Game loop
	scanner := bufio.NewScanner(os.Stdin)

	for gameState.attempts < gameState.maxChances {
		fmt.Printf("Enter your guess: ")

		if !scanner.Scan() {
			fmt.Println("Error reading input. Exiting game.")
			return
		}

		input := strings.TrimSpace(scanner.Text())
		guess, err := strconv.Atoi(input)

		if err != nil {
			fmt.Println("Please enter a valid number.")
			continue
		}

		if guess < 1 || guess > 100 {
			fmt.Println("Please enter a number between 1 and 100.")
			continue
		}

		gameState.attempts++
		gameState.currentGuess = guess

		if guess == gameState.secretNumber {
			endTime := time.Since(gameState.startTime)
			fmt.Printf("Congratulations! You guessed the correct number in %d attempts.\n", gameState.attempts)
			fmt.Printf("Time taken: %.2f seconds\n", endTime.Seconds())
			return
		} else if guess < gameState.secretNumber {
			fmt.Printf("Incorrect! The number is greater than %d.\n", guess)
		} else {
			fmt.Printf("Incorrect! The number is less than %d.\n", guess)
		}

		remaining := gameState.maxChances - gameState.attempts
		if remaining > 0 {
			fmt.Printf("You have %d chance(s) left.\n", remaining)

			// Provide hints based on proximity
			distance := abs(guess - gameState.secretNumber)
			if distance <= 5 {
				fmt.Println("🔥 Very close! You're within 5 numbers of the secret number!")
			} else if distance <= 10 {
				fmt.Println("😊 Getting warmer! You're within 10 numbers of the secret number!")
			} else if distance <= 20 {
				fmt.Println("😐 Getting warm! You're within 20 numbers of the secret number!")
			} else {
				fmt.Println("🧊 Cold! Try a different range!")
			}
		}
		fmt.Println()
	}

	// Player ran out of chances
	fmt.Printf("Game Over! You've run out of chances.\n")
	fmt.Printf("The correct number was: %d\n", gameState.secretNumber)
	fmt.Printf("Time taken: %.2f seconds\n", time.Since(gameState.startTime).Seconds())
}

func selectDifficulty() Difficulty {
	fmt.Println("Please select the difficulty level:")
	fmt.Println("1. Easy (10 chances)")
	fmt.Println("2. Medium (5 chances)")
	fmt.Println("3. Hard (3 chances)")

	difficulties := map[int]Difficulty{
		1: {"Easy", 10},
		2: {"Medium", 5},
		3: {"Hard", 3},
	}

	for {
		fmt.Print("Enter your choice: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		choice, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Please enter a valid number (1, 2, or 3).")
			continue
		}

		if difficulty, exists := difficulties[choice]; exists {
			return difficulty
		} else {
			fmt.Println("Please enter a valid choice (1, 2, or 3).")
		}
	}
}

func generateRandomNumber(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
