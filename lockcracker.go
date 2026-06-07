package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
)

var (
	user32              = syscall.NewLazyDLL("user32.dll")
	procKeybdEvent      = user32.NewProc("keybd_event")
	procGetAsyncKeyState = user32.NewProc("GetAsyncKeyState")
)

const (
	KEYEVENTF_KEYUP = 0x0002
	VK_W            = 0x57
	VK_A            = 0x41
	VK_S            = 0x53
	VK_D            = 0x44
	VK_F10          = 0x79
	VK_ESC          = 0x1B
)

type PlateState []int

type Dependency struct {
	AffectedPlate int
	Direction     int
}

type Move struct {
	Plate     int
	Direction int
	Count     int
}

type KeyAction struct {
	KeyCode int
	Delay   time.Duration
	Name    string
}

func pressKey(keyCode int) {
	procKeybdEvent.Call(uintptr(keyCode), 0, 0, 0)
	time.Sleep(15 * time.Millisecond)
	procKeybdEvent.Call(uintptr(keyCode), 0, KEYEVENTF_KEYUP, 0)
}

func isKeyPressed(keyCode int) bool {
	ret, _, _ := procGetAsyncKeyState.Call(uintptr(keyCode))
	return ret&0x8000 != 0
}

func movePlate(state PlateState, plate int, direction int, dependencies [][]Dependency) PlateState {
	newState := make(PlateState, len(state))
	copy(newState, state)

	currentPos := newState[plate]
	newPos := currentPos + direction

	if newPos < 1 || newPos > 7 {
		return nil
	}

	newState[plate] = newPos

	for _, dep := range dependencies[plate] {
		actualDirection := direction * dep.Direction
		currentDepPos := newState[dep.AffectedPlate]
		newDepPos := currentDepPos + actualDirection

		if newDepPos < 1 || newDepPos > 7 {
			return nil
		}

		newState[dep.AffectedPlate] = newDepPos
	}

	return newState
}

func parseDependenciesInteractive(nPlates int) [][]Dependency {
	dependencies := make([][]Dependency, nPlates)
	for i := range dependencies {
		dependencies[i] = []Dependency{}
	}

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("\nEnter dependencies for each plate:")
	fmt.Println("Format: 1N, 2S 3A, 4S2S5A, 2S-4A")
	fmt.Println("S = Synced (moves together), A = Asynced (opposite direction), N = None")
	fmt.Println("Examples: '2S 4A' or '2S-4A' or '2S4A' - all work the same")
	fmt.Println(strings.Repeat("-", 50))

	for i := 0; i < nPlates; i++ {
		for {
			fmt.Printf("Plate %d affects: ", i+1)
			scanner.Scan()
			depStr := scanner.Text()
			
			depStr = strings.TrimSpace(depStr)
			depStr = strings.ToUpper(depStr)
			
			// Remove all spaces and hyphens
			cleanStr := strings.ReplaceAll(depStr, " ", "")
			cleanStr = strings.ReplaceAll(cleanStr, "-", "")

			if cleanStr == "" || cleanStr == "N" {
				dependencies[i] = []Dependency{}
				break
			}
			
			var tempDeps []Dependency
			valid := true
			currentNum := ""
			
			for _, char := range cleanStr {
				if char >= '0' && char <= '9' {
					currentNum += string(char)
				} else if char == 'S' || char == 'A' {
					if currentNum == "" {
						fmt.Printf("  Error: '%c' found without plate number in '%s'. Try again.\n", char, depStr)
						valid = false
						break
					}
					
					affectedPlate := 0
					fmt.Sscanf(currentNum, "%d", &affectedPlate)
					affectedPlate--
					
					if affectedPlate < 0 || affectedPlate >= nPlates {
						fmt.Printf("  Error: plate %d does not exist (total: %d). Try again.\n", affectedPlate+1, nPlates)
						valid = false
						break
					}
					
					if affectedPlate == i {
						fmt.Println("  Error: plate cannot affect itself. Try again.")
						valid = false
						break
					}
					
					direction := 1
					if char == 'A' {
						direction = -1
					}
					
					tempDeps = append(tempDeps, Dependency{
						AffectedPlate: affectedPlate,
						Direction:     direction,
					})
					
					currentNum = ""
				} else {
					fmt.Printf("  Error: invalid character '%c'. Use only digits, S, A, N.\n", char)
					valid = false
					break
				}
			}
			
			if valid && currentNum != "" {
				fmt.Printf("  Error: number '%s' without S or A type. Try again.\n", currentNum)
				valid = false
			}

			if valid && len(tempDeps) == 0 {
				fmt.Println("  Error: no valid dependencies found. Try again or enter N for none.")
				valid = false
			}

			if valid {
				dependencies[i] = tempDeps
				break
			}
		}
	}

	return dependencies
}

func getPossibleMoves(state PlateState, dependencies [][]Dependency, nPlates int) []Move {
	var moves []Move
	for plate := 0; plate < nPlates; plate++ {
		for _, direction := range []int{1, -1} {
			newState := movePlate(state, plate, direction, dependencies)
			if newState != nil {
				moves = append(moves, Move{Plate: plate, Direction: direction})
			}
		}
	}
	return moves
}

func stateToKey(state PlateState) string {
	key := make([]byte, len(state))
	for i, v := range state {
		key[i] = byte(v) + '0'
	}
	return string(key)
}

type bfsNode struct {
	state PlateState
	path  []Move
}

func bfsSolve(initialState, targetState PlateState, dependencies [][]Dependency, nPlates int, maxMoves int) []Move {
	queue := []bfsNode{{state: initialState, path: []Move{}}}
	visited := make(map[string]bool)
	visited[stateToKey(initialState)] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if stateToKey(current.state) == stateToKey(targetState) {
			return current.path
		}

		if len(current.path) >= maxMoves {
			continue
		}

		possibleMoves := getPossibleMoves(current.state, dependencies, nPlates)

		for _, move := range possibleMoves {
			newState := movePlate(current.state, move.Plate, move.Direction, dependencies)
			key := stateToKey(newState)
			if !visited[key] {
				visited[key] = true
				newPath := make([]Move, len(current.path)+1)
				copy(newPath, current.path)
				newPath[len(current.path)] = move
				queue = append(queue, bfsNode{state: newState, path: newPath})
			}
		}
	}

	return nil
}

func compressPath(path []Move) []Move {
	if len(path) == 0 {
		return []Move{}
	}

	var compressed []Move
	currentPlate := path[0].Plate
	currentDir := path[0].Direction
	count := 1

	for i := 1; i < len(path); i++ {
		if path[i].Plate == currentPlate && path[i].Direction == currentDir {
			count++
		} else {
			compressed = append(compressed, Move{Plate: currentPlate, Direction: currentDir, Count: count})
			currentPlate = path[i].Plate
			currentDir = path[i].Direction
			count = 1
		}
	}

	compressed = append(compressed, Move{Plate: currentPlate, Direction: currentDir, Count: count})
	return compressed
}

func formatSolutionMatrix(compressedPath []Move, itemsPerRow int) string {
	var parts []string
	for _, move := range compressedPath {
		arrow := "<"
		if move.Direction == -1 {
		arrow = ">"
	}
		arrows := strings.Repeat(arrow, move.Count)
		parts = append(parts, fmt.Sprintf("[ %d%s ]", move.Plate+1, arrows))
	}

	var resultLines []string
	for i := 0; i < len(parts); i += itemsPerRow {
		end := i + itemsPerRow
		if end > len(parts) {
			end = len(parts)
		}
		row := parts[i:end]
		resultLines = append(resultLines, strings.Join(row, " "))
	}

	return strings.Join(resultLines, "\n")
}

func generateKeySequence(compressedPath []Move, delayMs int) []KeyAction {
	var actions []KeyAction
	currentPlate := 0

	for _, move := range compressedPath {
		plateDiff := move.Plate - currentPlate
		if plateDiff > 0 {
			for i := 0; i < plateDiff; i++ {
				actions = append(actions, KeyAction{
					KeyCode: VK_W,
					Delay:   time.Duration(delayMs) * time.Millisecond,
					Name:    "W",
				})
			}
		} else if plateDiff < 0 {
			for i := 0; i < -plateDiff; i++ {
				actions = append(actions, KeyAction{
					KeyCode: VK_S,
					Delay:   time.Duration(delayMs) * time.Millisecond,
					Name:    "S",
				})
			}
		}

		keyCode := VK_A
		keyName := "A"
		if move.Direction == -1 {
			keyCode = VK_D
			keyName = "D"
		}

		for i := 0; i < move.Count; i++ {
			actions = append(actions, KeyAction{
				KeyCode: keyCode,
				Delay:   time.Duration(delayMs) * time.Millisecond,
				Name:    keyName,
			})
		}

		currentPlate = move.Plate
	}

	return actions
}

func waitForHotkey() bool {
	fmt.Println("\nSwitch to the game window!")
	fmt.Println("Press F10 to execute, or ESC to cancel...")
	
	time.Sleep(500 * time.Millisecond)
	
	for {
		if isKeyPressed(VK_F10) {
			for isKeyPressed(VK_F10) {
				time.Sleep(10 * time.Millisecond)
			}
			return true
		}
		
		if isKeyPressed(VK_ESC) {
			for isKeyPressed(VK_ESC) {
				time.Sleep(10 * time.Millisecond)
			}
			return false
		}
		
		time.Sleep(50 * time.Millisecond)
	}
}

func executeKeySequence(actions []KeyAction) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("EXECUTING...")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("Starting in 3 seconds...")
	fmt.Println("DO NOT touch keyboard or mouse!")
	
	for i := 3; i > 0; i-- {
		fmt.Printf("%d...\n", i)
		time.Sleep(1 * time.Second)
	}
	
	fmt.Println("GO!")

	for i, action := range actions {
		pressKey(action.KeyCode)
		
		if (i+1)%10 == 0 {
			fmt.Printf("Progress: %d/%d\n", i+1, len(actions))
		}
		
		time.Sleep(action.Delay)
	}
	
	fmt.Println("\n✓ Complete!")
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	
	fmt.Println("\n=== Lock Cracker ===")
// Input number of plates

var nPlates int
for {
    fmt.Print("\nNumber of plates: ")
    scanner.Scan()
    _, err := fmt.Sscanf(scanner.Text(), "%d", &nPlates)
    if err != nil || nPlates <= 0 {
        fmt.Println("Error: enter a positive integer!")
        continue
    }
    break
}

	// Input current pin positions
	var currentStr string
	for {
		fmt.Printf("\nCurrent code\n(%d digits 1-7): ", nPlates)
		scanner.Scan()
		currentStr = strings.TrimSpace(scanner.Text())
		if len(currentStr) == nPlates {
			valid := true
			for _, c := range currentStr {
				if c < '1' || c > '7' {
					valid = false
					break
				}
			}
			if valid {
				break
			}
		}
		fmt.Printf("Error: exactly %d digits from 1 to 7 required!\n", nPlates)
	}

	initialState := make(PlateState, nPlates)
	for i, c := range currentStr {
		initialState[i] = int(c - '0')
	}

	// Input target pin positions (default: all 4's)
	var targetStr string
	defaultTarget := strings.Repeat("4", nPlates)
	fmt.Printf("\nTarget code (default: %s)\n", defaultTarget)
	fmt.Printf("Press Enter for default or enter %d digits 1-7: ", nPlates)
	scanner.Scan()
	targetStr = strings.TrimSpace(scanner.Text())
	
	if targetStr == "" {
		targetStr = defaultTarget
		fmt.Printf("Using default: %s\n", targetStr)
	}
	
	for len(targetStr) != nPlates {
		fmt.Printf("Error: exactly %d digits from 1 to 7 required!\n", nPlates)
		fmt.Printf("Target code: ")
		scanner.Scan()
		targetStr = strings.TrimSpace(scanner.Text())
		if targetStr == "" {
			targetStr = defaultTarget
			fmt.Printf("Using default: %s\n", targetStr)
		}
	}

	targetState := make(PlateState, nPlates)
	for i, c := range targetStr {
		targetState[i] = int(c - '0')
	}

	// Interactive dependency input
	dependencies := parseDependenciesInteractive(nPlates)

	// Check if lock is already solved
	if stateToKey(initialState) == stateToKey(targetState) {
		fmt.Println("\n✓ Lock is already in the correct position!")
		return
	}

	fmt.Println("\nSolving...")

	solution := bfsSolve(initialState, targetState, dependencies, nPlates, 200)

	if solution == nil {
		fmt.Println("\n✗ No solution found!")
		fmt.Println("Possible reasons:")
		fmt.Println("  - Dependencies entered incorrectly")
		fmt.Println("  - Combination too complex")
		fmt.Println("  - This lock state cannot be solved")
	} else {
		compressed := compressPath(solution)
		formatted := formatSolutionMatrix(compressed, 5)

		fmt.Printf("\n✓ SOLUTION FOUND! (%d moves)\n", len(solution))
		fmt.Println(strings.Repeat("-", 50))
		fmt.Printf("\n%s\n", formatted)
		fmt.Println(strings.Repeat("-", 50))

		// Get delay
		var delayMs int
		fmt.Print("\nDelay between keystrokes in ms (default 350): ")
		scanner.Scan()
		delayInput := strings.TrimSpace(scanner.Text())
		if delayInput == "" {
			delayMs = 350
		} else {
			fmt.Sscanf(delayInput, "%d", &delayMs)
			if delayMs < 30 {
				delayMs = 30
			}
		}

		// Generate key sequence
		keySequence := generateKeySequence(compressed, delayMs)
		fmt.Printf("Total keystrokes to execute: %d\n", len(keySequence))

		// Ask if user wants to execute
		fmt.Print("\nExecute automatically? (y/n): ")
		scanner.Scan()
		execute := strings.ToLower(strings.TrimSpace(scanner.Text()))

		if execute == "y" {
			fmt.Println("\nPosition yourself on PLATE 1 in the game!")
			
			if waitForHotkey() {
				executeKeySequence(keySequence)
			} else {
				fmt.Println("\nCancelled.")
			}
		} else {
			fmt.Println("\nFollow the sequence manually.")
		}
	}

	fmt.Println("\nPress Enter to exit...")
	scanner.Scan()
}
