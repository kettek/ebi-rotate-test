package game

import (
	"fmt"
	"math"
	"math/rand"
	"rotate-test/internal/res"
)

// TileState is a bitfield of potential tile states.
type TileState uint8

// TileState constants, wow.
const (
	TileStateSolid TileState = 1 << iota
	TileStateLiquid
	TileStateHurts
	TileStatePoisons
)

// Tile represents a single tile within a chunk.
type Tile struct {
	State TileState
	// Some sorta image or ID reference.
	Details []TileDetail // loaded details.
}

type Tiles [][]Tile

type TileDetail struct {
	RID    res.RID
	State  interface{} // I guess for now
	detail *res.Detail
	visual Visual
}

type World struct {
	sneed         int64
	loadingChunks []*Chunk
	loadedChunks  []*Chunk
}

func NewWorld() *World {
	return &World{
		sneed: rand.Int63(),
	}
}

func (w *World) Update() error {
	// Check on our loading chunks.
	chunks := w.loadingChunks[:0]
	for _, chunk := range w.loadingChunks {
		select {
		case err := <-chunk.loadChan:
			if err != nil {
				panic(fmt.Errorf("failed to load chunk: %w", err))
			}
			fmt.Println("loaded chunk", chunk.X, chunk.Y)
			chunk.loaded = true
		default:
		}
		if chunk.loaded {
			w.loadedChunks = append(w.loadedChunks, chunk)
		} else {
			chunks = append(chunks, chunk)
		}
	}
	w.loadingChunks = chunks

	// Update our loaded chunks.
	var chunkUpdateRequests []ChunkUpdateRequests
	chunks = w.loadedChunks[:0]
	for _, chunk := range w.loadedChunks {
		chunkUpdateRequests = append(chunkUpdateRequests, ChunkUpdateRequests{
			Chunk:    chunk,
			Requests: chunk.Update(w),
		})
		// TODO: Check if chunk should be unloaded (probably distance from player).
		chunks = append(chunks, chunk)
	}
	w.loadedChunks = chunks

	// Process chunk updates.
	for _, chunkUpdate := range chunkUpdateRequests {
		for _, chunkRequest := range chunkUpdate.Requests {
			switch chunkRequest := chunkRequest.(type) {
			case ChunkUpdateThingRequest:
				for _, thingRequest := range chunkRequest.Requests {
					switch thingRequest := thingRequest.(type) {
					case RequestMove:
						chunkRequest.Thing.HandleRequest(thingRequest, true)
						cx, cy := int(math.Floor(thingRequest.To.X()/ChunkPixelSize/ChunkTileSize)), int(math.Floor(thingRequest.To.Y()/ChunkPixelSize/ChunkTileSize))
						if cx != chunkRequest.Thing.Chunk().X || cy != chunkRequest.Thing.Chunk().Y {
							targetChunk := w.LoadChunk(cx, cy)
							chunkUpdate.Chunk.RemoveThing(chunkRequest.Thing)
							targetChunk.AddThing(chunkRequest.Thing, VisualLayerWorld)
						}
					case RequestRotate:
						chunkRequest.Thing.HandleRequest(thingRequest, true)
					}
					/*case ChunkUpdateMoveThing:
					switch thing := result.Thing.(type) {
					case *Mover:
						if result.To.X() < 0 || result.To.Y() < 0 || result.To.X() >= chunkUpdate.Chunk.Width() || result.To.Y() >= chunkUpdate.Chunk.Height() {
							x := 0
							y := 0
							if result.To.X() < 0 {
								x = -1
							} else if result.To.X() >= chunkUpdate.Chunk.Width() {
								x = 1
							}
							if result.To.Y() < 0 {
								y = -1
							} else if result.To.Y() >= chunkUpdate.Chunk.Height() {
								y = 1
							}
							// Get/load target chunk.
							chunk := w.LoadChunk(thing.Chunk().X+x, thing.Chunk().Y+y)

							// Calculate position in new chunk and assign.
							var rx, ry float64
							if x < 0 {
								rx = chunk.Width() - 1
							} else if x > 0 {
								rx = 0
							}
							if y < 0 {
								ry = chunk.Height() - 1
							} else if y > 0 {
								ry = 0
							}
							rx += result.To.X()
							ry += result.To.Y()

							// TODO: See if rx, ry is open in target chunk.
							if true {
								// Remove thing from current chunk and add it to target chunk.
								chunkUpdate.Chunk.Things.Remove(thing)
								chunk.Things.Add(thing)
								thing.SetChunk(chunk)
								thing.Vec2.Assign(Vec2{rx, ry})
							}
						}*/
				}
			}
		}
	}

	return nil
}

func (w *World) ChunksAround(x, y int) (chunks []*Chunk) {
	for x1 := x - 1; x1 <= x+1; x1++ {
		for y1 := y - 1; y1 <= y+1; y1++ {
			chunks = append(chunks, w.LoadChunk(x1, y1))
		}
	}
	return chunks
}

func (w *World) LoadChunk(x, y int) *Chunk {
	for _, chunk := range w.loadingChunks {
		if chunk.X == x && chunk.Y == y {
			return chunk
		}
	}
	for _, chunk := range w.loadedChunks {
		if chunk.X == x && chunk.Y == y {
			return chunk
		}
	}
	chunk := NewChunk()
	chunk.X = x
	chunk.Y = y
	w.loadingChunks = append(w.loadingChunks, chunk)
	go chunk.Load(w.sneed)
	return chunk
}