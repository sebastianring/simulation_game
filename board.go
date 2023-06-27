package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

// -------------------------------------------------- //
// -------------------------------------------------- //
// INITS AND STRUCTS -------------------------------- //
// -------------------------------------------------- //
// -------------------------------------------------- //

var initialCreature1 int
var initialFoods int

var allFoodsObjects []Pos
var allCreatureObjects []Pos

type Board struct {
	rows int
	cols int
	// displayBoard [][]int // 0 = empty, 1 = food, 10-20 = creatures
	gamelog     *Gamelog
	objectBoard [][]BoardObject
	time        int
}

type Pos struct {
	x int
	y int
}

func InitNewBoard(rows int, cols int) *Board {
	if rows < 2 || cols < 2 {
		fmt.Printf("Too few rows or cols: %v, rows: %v \n", rows, cols)
		os.Exit(1)
	}

	newBoard := Board{
		rows,
		cols,
		InitTextInfo(rows),
		*createEmptyObjectsArray(rows, cols),
		0,
	}

	initialCreature1 = 30
	initialFoods = 250

	newBoard.spawnCreature1OnBoard(initialCreature1)
	newBoard.spawnFoodOnBoard(initialFoods)

	addMessageToCurrentGamelog("Board added", 2)
	addMessageToCurrentGamelog("Welcome to the simulation game where you can simulate creatures and how they evolve!", 1)

	return &newBoard
}

// creates the initial array for all objects inside the board
func createObjectArray(rows int, cols int) *[][]BoardObject {
	arr := make([][]BoardObject, rows)
	edgeSpawnPoints := (rows*2 + cols*2 - 4)
	createSpawnChance := edgeSpawnPoints / 10 // 10% chance that a creature spawns at the edge

	for i := 0; i < rows; i++ {
		arr[i] = make([]BoardObject, cols)
		for j := 0; j < cols; j++ {
			// check if we are at the edge of the board, then roll the dice if a creature should be spawned
			if i == 0 || i == rows-1 || j == 0 || j == cols-1 {
				rng := rand.Intn(edgeSpawnPoints)
				if rng < createSpawnChance {
					arr[i][j] = newCreature1Object()
				} else {
					arr[i][j] = newEmptyObject()
				}
				// else, lets see if we can spawn some food, 2.5% chance to spawn
			} else {
				rng := rand.Intn(1000)
				if rng < 25 {
					arr[i][j] = newFoodObject()
				} else {
					arr[i][j] = newEmptyObject()
				}
			}
		}
	}

	return &arr
}

func createEmptyObjectsArray(rows int, cols int) *[][]BoardObject {
	arr := make([][]BoardObject, rows)

	for i := 0; i < rows; i++ {
		arr[i] = make([]BoardObject, cols)
		for j := 0; j < cols; j++ {
			arr[i][j] = newEmptyObject()
		}
	}

	return &arr
}

// -------------------------------------------------- //
// -------------------------------------------------- //
// BOARD FUNCTIONS ---------------------------------- //
// -------------------------------------------------- //
// -------------------------------------------------- //

func (b *Board) spawnCreature1OnBoard(qty int) {
	spawns := make([]Pos, 0)
	for len(spawns) < qty {
		newPos := b.randomPosAtEdgeOfMap()
		if !checkIfPosExistsInSlice(newPos, spawns) {
			spawns = append(spawns, newPos)
		}
	}

	for _, pos := range spawns {
		b.objectBoard[pos.y][pos.x] = newCreature1Object()
		allCreatureObjects = append(allCreatureObjects, pos)
	}
}

func (b *Board) spawnFoodOnBoard(qty int) {
	spawns := make([]Pos, 0)

	for len(spawns) < qty {
		newPos := b.randomPosWithinMap()
		if !checkIfPosExistsInSlice(newPos, spawns) && b.isSpotEmpty(newPos) {
			spawns = append(spawns, newPos)
		}
	}

	for _, pos := range spawns {
		b.objectBoard[pos.y][pos.x] = newFoodObject()
		allFoodsObjects = append(allFoodsObjects, pos)
	}
}

func (b *Board) isSpotEmpty(pos Pos) bool {
	if b.objectBoard[pos.y][pos.x].getType() == "empty" {
		return true
	}

	return false
}

func (b *Board) randomPosAtEdgeOfMap() Pos {
	// top = 0, right = 1, left = 2, bottom = 3
	edge := rand.Intn(4)
	var x int
	var y int

	if edge == 0 {
		y = 0
		x = rand.Intn(b.cols - 1)
	} else if edge == 1 {
		x = b.cols - 1
		y = rand.Intn(b.rows - 1)
	} else if edge == 2 {
		x = 0
		y = rand.Intn(b.rows - 1)
	} else {
		x = rand.Intn(b.cols - 1)
		y = b.rows - 1
	}

	return Pos{x, y}
}

func (b *Board) randomPosWithinMap() Pos {
	minDistanceFromBorder := 3
	x := rand.Intn(b.cols-minDistanceFromBorder*2) + minDistanceFromBorder
	y := rand.Intn(b.rows-minDistanceFromBorder*2) + minDistanceFromBorder

	return Pos{x, y}
}

func checkIfPosExistsInSlice(pos Pos, slice []Pos) bool {
	for _, slicePos := range slice {
		if pos.x == slicePos.x && pos.y == slicePos.y {
			return true
		}
	}

	return false
}

// func checkIfValExistsInSlice(val []int, slice [][]int) bool {
// 	for _, val2 := range slice {
// 		if len(val) == len(val2) {
// 			for i := 0; i < len(val); i++ {
// 				if val[i] == val2[i] {
// 					return false
// 				}
// 			}
// 		}
// 	}
//
// 	return false
// }

func (b *Board) tickFrame() {
	b.time++
	b.creatureUpdatesPerTick()

	// ----------- debugging ---------- //

	res := make([]string, 1)

	for _, pos := range allCreatureObjects {
		speed := b.objectBoard[pos.y][pos.x].getIntData("speed")
		id := b.objectBoard[pos.y][pos.x].getIntData("id")
		res = append(res, strconv.Itoa(speed)+":"+strconv.Itoa(id))
	}

	addMessageToCurrentGamelog(strings.Join(res, ", "), 2)

	// ----------- end debugging ------ //

	DrawFrame(b)
}

func (b *Board) creatureUpdatesPerTick() {
	// var updatedAllCreatureObjects Pos
	updatedAllCreatureObjects := make([]Pos, 0)
	deadCreatures := make([]Pos, 0)

	for i, pos := range allCreatureObjects {
		addMessageToCurrentGamelog(strconv.Itoa(b.objectBoard[pos.y][pos.x].getIntData("id"))+" "+strconv.Itoa(i)+" "+strconv.Itoa(pos.x)+" "+strconv.Itoa(pos.y), 2)
		action := b.objectBoard[pos.y][pos.x].updateTick()

		if action == "move" {
			// oldPos := pos
			newPos, moveType := b.newPosAndMove(pos)
			// tempObject := b.objectBoard[newPos.y][newPos.x]
			b.objectBoard[newPos.y][newPos.x] = b.objectBoard[pos.y][pos.x]

			if moveType == "food" {
				addMessageToCurrentGamelog("Food eaten by "+strconv.Itoa(b.objectBoard[pos.y][pos.x].getIntData("id")), 2)
				b.objectBoard[newPos.y][newPos.x].updateVal("heal")
				b.objectBoard[pos.y][pos.x] = newEmptyObject()
				deleteFood(newPos)
			} else {
				b.objectBoard[pos.y][pos.x] = newEmptyObject()
			}

			updatedAllCreatureObjects = append(updatedAllCreatureObjects, newPos)

			// addMessageToCurrentGamelog("New POS: "+strconv.Itoa(pos.x)+" "+strconv.Itoa(pos.y), 1)
		} else if action == "dead" {
			deadCreatures = append(deadCreatures, pos)
		} else {
			updatedAllCreatureObjects = append(updatedAllCreatureObjects, pos)
		}
	}

	// delete dead creatures after tick is complete
	for _, pos := range deadCreatures {
		b.objectBoard[pos.y][pos.x] = newEmptyObject()
		deleteCreature(pos)
	}

	// update all creatures from last tick
	allCreatureObjects = updatedAllCreatureObjects

	if b.checkIfCreaturesAreInactive() == true {
		if b.checkIfCreaturesAreDead() {
			gameOn = false
			addMessageToCurrentGamelog("All creatures are dead, end the game", 1)
		}
		// NEXT ROUND
		b.newRound()
	}
}

func (b *Board) newRound() {
	addMessageToCurrentGamelog("All creatures inactive, starting new round", 1)
	// addMessageToCurrentGamelog(strconv.Itoa(len(allCreatureObjects)), 1)

	for i, creaturePos := range allCreatureObjects {
		addMessageToCurrentGamelog(strconv.Itoa(i), 2)

		findNewPos := false
		for !findNewPos {
			newPos := b.randomPosAtEdgeOfMap()
			if b.isSpotEmpty(newPos) {
				b.objectBoard[newPos.y][newPos.x] = b.objectBoard[creaturePos.y][creaturePos.x]
				b.objectBoard[newPos.y][newPos.x].resetValues()
				b.objectBoard[creaturePos.y][creaturePos.x] = newEmptyObject()

				// addMessageToCurrentGamelog("old creature pos: x: "+
				// 	strconv.Itoa(creaturePos.x)+
				// 	" y: "+strconv.Itoa(creaturePos.y), 2)
				//
				// allCreatureObjects[i] = newPos
				//
				// addMessageToCurrentGamelog("new creature pos: x: "+
				// 	strconv.Itoa(allCreatureObjects[i].x)+
				// 	" y: "+strconv.Itoa(allCreatureObjects[i].y), 2)

				findNewPos = true
			}
		}
	}
}

func (b *Board) checkIfCreatureSpawnsOffspring() {
	for _, pos := range allCreatureObjects {
		b.objectBoard[pos.y][pos.x].getIntData("hp")

	}
}

func (b *Board) checkIfCreaturesAreDead() bool {
	for _, pos := range allCreatureObjects {
		dead := b.objectBoard[pos.y][pos.x].isDead()
		// moving := b.objectBoard[pos.y][pos.x].isMoving()
		// addMessageToCurrentGamelog("DEAD:" + strconv.FormatBool(dead) + " MOVING: " + strconv.FormatBool(moving))

		if !dead {
			return false
		}
	}

	return true
}

func (b *Board) checkIfCreaturesAreInactive() bool {
	for _, pos := range allCreatureObjects {
		dead := b.objectBoard[pos.y][pos.x].isDead()
		moving := b.objectBoard[pos.y][pos.x].isMoving()

		// addMessageToCurrentGamelog("Current counter: " + strconv.Itoa(i) + "total length: " + strconv.Itoa(len(allCreatureObjects)))
		// addMessageToCurrentGamelog("DEAD:" + strconv.FormatBool(dead) + " MOVING: " + strconv.FormatBool(moving))

		if !dead && moving || dead {
			return false
		}
	}

	return true
}

func (b *Board) newPosAndMove(currentPos Pos) (Pos, string) {
	newPos := Pos{-1, -1}

	// HOW TO MAKE THE CREATURES MOVE INWARDS TO LOOK FOR FOOD?
	// The closer they are to one edge, the more probable they are to move towards the other edge?
	// Example: x = 99, y = 40
	// Width-x = the probability to move the left
	// Height-y = the probability to move upwards?

	for newPos.x == -1 || newPos.y == -1 {
		direction := rand.Intn(2) // 0 = x movement, 1 = y-movement
		var x int
		var y int
		if direction == 0 {
			xdirection := rand.Intn(b.cols)
			xprobability := b.cols - 1 - currentPos.x
			if xdirection < xprobability {
				x = currentPos.x + 1
			} else {
				x = currentPos.x - 1
			}
			y = currentPos.y
		} else {
			ydirection := rand.Intn(b.rows)
			yprobability := b.rows - 1 - currentPos.y
			if ydirection < yprobability {
				y = currentPos.y + 1
			} else {
				y = currentPos.y - 1
			}
			x = currentPos.x
		}

		valid, moveType := b.checkIfNewPosIsValid(x, y)

		if valid {
			newPos.x = x
			newPos.y = y
			return newPos, moveType
		}
	}

	return newPos, "empty"
}

func (b *Board) checkIfNewPosIsValid(x int, y int) (bool, string) {
	if x < 0 || x >= b.cols || y < 0 || y >= b.rows {
		return false, ""
	}
	objectType := b.objectBoard[y][x].getType()
	if objectType == "food" {
		return true, "food"
	} else if objectType == "empty" {
		return true, "empty"
	}

	return false, ""
}

func deleteFood(pos Pos) {
	var element int
	for i, val := range allFoodsObjects {
		if val.x == pos.x && val.y == pos.y {
			element = i
			break
		}
	}

	allFoodsObjects = deleteIndexInPosSlice(allFoodsObjects, element)
}

func deleteCreature(pos Pos) {
	var element int
	for i, val := range allCreatureObjects {
		if val.x == pos.x && val.y == pos.y {
			element = i
			break
		}
	}

	allCreatureObjects = deleteIndexInPosSlice(allCreatureObjects, element)
}

func deleteIndexInPosSlice(posSlice []Pos, index int) []Pos {
	posSlice[index] = posSlice[len(posSlice)-1]
	return posSlice[:len(posSlice)-1]
}
