package main

import (
	"math"
	"sort"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const astarStep float32 = 0.5

type PriorityPos struct {
	rl.Vector2
	Priority int
}

type PriorityArray []PriorityPos

func (p PriorityArray) Len() int {
	return len(p)
}
func (p PriorityArray) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p PriorityArray) Less(i, j int) bool {
	return p[i].Priority < p[j].Priority
}

func getNeighbors(pos rl.Vector2, v Viewer) []rl.Vector2 {
	neighbors := make([]rl.Vector2, 0, 4)
	left := rl.Vector2{X: pos.X - astarStep, Y: pos.Y}
	if v.CanSee(left) {
		neighbors = append(neighbors, left)
	}
	right := rl.Vector2{X: pos.X + astarStep, Y: pos.Y}
	if v.CanSee(right) {
		neighbors = append(neighbors, right)
	}
	up := rl.Vector2{X: pos.X, Y: pos.Y - astarStep}
	if v.CanSee(up) {
		neighbors = append(neighbors, up)
	}
	down := rl.Vector2{X: pos.X, Y: pos.Y + astarStep}
	if v.CanSee(down) {
		neighbors = append(neighbors, down)
	}

	return neighbors
}

type Viewer interface {
	CanSee(pos rl.Vector2) bool
}

func astar(start rl.Vector2, goal rl.Vector2, v Viewer) []rl.Vector2 {
	var result []rl.Vector2
	frontier := make(PriorityArray, 0, 8)
	frontier = append(frontier, PriorityPos{start, 1})
	cameFrom := make(map[rl.Vector2]rl.Vector2)
	cameFrom[start] = start
	costSoFar := make(map[rl.Vector2]int)
	costSoFar[start] = 0

	for len(frontier) > 0 {
		sort.Stable(frontier)
		current := frontier[0]
		if astarVectorEquals(current.Vector2, goal) {
			p := current.Vector2
			result = append([]rl.Vector2{p}, result...)
			for p != start {
				p = cameFrom[p]
				result = append([]rl.Vector2{p}, result...)
			}
			break
		}
		frontier = frontier[1:]
		for _, next := range getNeighbors(current.Vector2, v) {
			newCost := costSoFar[current.Vector2] + 1
			_, exists := costSoFar[next]
			if !exists || newCost < costSoFar[next] {
				costSoFar[next] = newCost
				xDist := int(math.Abs(float64(goal.X - next.X)))
				yDist := int(math.Abs(float64(goal.Y - next.Y)))
				priority := newCost + xDist + yDist
				frontier = append(frontier, PriorityPos{next, priority})
				cameFrom[next] = current.Vector2
			}
		}
	}
	return result
}

func astarVectorEquals(v1 rl.Vector2, v2 rl.Vector2) bool {
	return math.Abs(float64(v1.X-v2.X)) < 1 && math.Abs(float64(v1.Y-v2.Y)) < 1
}
