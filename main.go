/*******************************************************************************************
*
*   raylib [models] example - first person maze
*
*   This example has been created using raylib-go v0.0.0-20220104071325-2f072dc2d259 (https://github.com/gen2brain/raylib-go)
*   raylib-go is licensed under an unmodified zlib/libpng license (https://github.com/gen2brain/raylib-go/blob/master/LICENSE)
*
*   Original C version for Raylib 2.5 Copyright (c) 2019 Ramon Santamaria (@raysan5)
*   Converted to Go by Michael Redman January 4, 2022
*
********************************************************************************************/

package main

import (
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	// Initialization
	//--------------------------------------------------------------------------------------
	var screenWidth int32 = 800
	var screenHeight int32 = 450

	rl.InitWindow(screenWidth, screenHeight, "raylib [models] example - first person maze")

	// Define the camera to look into our 3d world
	camera := rl.Camera{}
	camera.Position = rl.NewVector3(0, 1.2, 0)
	camera.Target = rl.NewVector3(-0.8, 0.5, -0.8)
	camera.Up = rl.NewVector3(0.0, 1.0, 0.0)
	camera.Fovy = 75.0
	camera.Projection = rl.CameraPerspective

	imMap := rl.LoadImage("cubicmap.png")      // Load cubicmap image (RAM)
	cubicmap := rl.LoadTextureFromImage(imMap) // Convert image to texture to display (VRAM)
	mesh := rl.GenMeshCubicmap(*imMap, rl.NewVector3(1.0, 3.0, 1.0))
	model := rl.LoadModelFromMesh(mesh)

	// NOTE: By default each cube is mapped to one part of texture atlas
	texture := rl.LoadTexture("cubicmap_atlas.png")         // load map texture
	model.Materials.GetMap(rl.MapDiffuse).Texture = texture // Set map diffuse texture
	// Get map image data to be used for collision detectio
	mapPixels := rl.LoadImageColors(imMap)
	rl.UnloadImage(imMap) // Unload image from RAM

	mapPosition := rl.NewVector3(-16.0, 0.0, -8.0) // Set model position

	rl.SetTargetFPS(60) // Set our game to run at 60 frames-per-second

	rl.DisableCursor() // Locking the cursor to enable camera control with mouse

	playerModel := rl.LoadModel("robot.glb")
	playerInitialRot := float32(rl.Deg2rad * 90)

	monsterModel := rl.LoadModel("robot.glb")
	monsterPos := rl.NewVector3(-14, 0, -6)
	monster := Monster{
		cubicmap:    cubicmap,
		mapPosition: mapPosition,
		mapPixels:   mapPixels,
	}
	monsterInitialRot := float32(rl.Deg2rad * 90)

	playerAnimIndex := 0
	playerAnimCurrentFrame := 0

	robotAnims := rl.LoadModelAnimations("robot.glb")

	monsterAnimIndex := 0 // Dancing
	monsterAnimCurrentFrame := 0
	monsterDead := false

	//--------------------------------------------------------------------------------------

	// Main game loop
	for !rl.WindowShouldClose() { // Detect window close button or ESC key
		// Update
		//----------------------------------------------------------------------------------
		oldCamPos := camera.Position  // Store old camera position
		oldCamTarget := camera.Target // Store old camera target position

		// TODO: use a custom camera mode; the ThirdPerson is good for a poc,
		// but lacks customization (invert Y-axis as example)
		rl.UpdateCamera(&camera, rl.CameraThirdPerson)

		playerPos := camera.Target
		playerPos.Y = playerPos.Y - 0.5 // mid-size of the player

		// Check player collision (we simplify to 2D collision detection)
		playerPos2 := rl.NewVector2(playerPos.X, playerPos.Z)
		playerRadius := 0.3 // Collision radius (player is modelled as a cylinder for collision)

		playerRot := playerInitialRot + rl.Vector2LineAngle(rl.Vector2{X: camera.Position.X, Y: camera.Position.Z}, playerPos2)
		playerModel.Transform = rl.MatrixRotateY(playerRot)

		playerCellX := (int)(playerPos2.X - mapPosition.X + 0.5)
		playerCellY := (int)(playerPos2.Y - mapPosition.Z + 0.5)

		// Out-of-limits security check
		if playerCellX < 0 {
			playerCellX = 0
		} else if playerCellX >= int(cubicmap.Width) {
			playerCellX = int(cubicmap.Width) - 1
		}

		if playerCellY < 0 {
			playerCellY = 0
		} else if playerCellY >= int(cubicmap.Height) {
			playerCellY = int(cubicmap.Height) - 1
		}

		// Check map collisions using image data and player position
		// (just check player surrounding cells for collision)
		minYsurrounding := max(playerCellY-2, 0)
		maxYsurrounding := min(playerCellY+2, int(cubicmap.Height))
		minXsurrounding := max(playerCellX-2, 0)
		maxXsurrounding := min(playerCellX+2, int(cubicmap.Width))
		for y := minYsurrounding; y < maxYsurrounding; y++ {
			for x := minXsurrounding; x < maxXsurrounding; x++ {
				// Collision: white pixel, only check R channel
				if mapPixels[y*int(cubicmap.Width)+x].R == 255 && (rl.CheckCollisionCircleRec(playerPos2, float32(playerRadius), rl.NewRectangle(float32(mapPosition.X-0.5+float32(x)), float32(mapPosition.Z-0.5+float32(y)), 1.0, 1.0))) {
					// Collision detected, reset camera position
					camera.Position = oldCamPos
					camera.Target = oldCamTarget
				}
			}
		}

		// Monster IA
		if !monsterDead {
			monster.Position = rl.Vector2{X: monsterPos.X, Y: monsterPos.Z}
			pathToPlayer := astar(monster.Position, playerPos2, &monster)
			if len(pathToPlayer) > 1 {
				// Rotating into wanted direction
				monsterRot := monsterInitialRot + rl.Vector2LineAngle(monster.Position, pathToPlayer[1])
				monsterModel.Transform = rl.MatrixRotateY(monsterRot)
				// Moving to direction
				monster.Position = monster.Position.MoveTowards(pathToPlayer[1], 0.1)
				monsterPos = rl.Vector3{X: monster.Position.X, Y: monsterPos.Y, Z: monster.Position.Y}
				// Running
				monsterAnimIndex = 6 // Running
			} else {
				monsterAnimIndex = 0 // Dancing
			}
		}

		// Monster collision detection
		monsterRadius := 0.3
		if math.Abs(float64(monsterPos.X-playerPos.X)) < playerRadius+monsterRadius && math.Abs(float64(monsterPos.Z-playerPos.Z)) < playerRadius+monsterRadius {
			// Collision detected, reset camera position
			camera.Position = oldCamPos
			camera.Target = oldCamTarget
		}

		// Player <-> monster interactions
		if math.Abs(float64(monsterPos.X-playerPos.X)) < playerRadius*2+monsterRadius && math.Abs(float64(monsterPos.Z-playerPos.Z)) < playerRadius*2+monsterRadius {
			// if player is punching the monster, it dies (if not already dead)
			if rl.IsKeyDown(rl.KeyR) && !monsterDead && isPointTargeted(camera, monsterPos) {
				monsterDead = true
				monsterAnimCurrentFrame = 0
				monsterAnimIndex = 1 // Die
			}
		}

		switch {
		case rl.IsKeyDown(rl.KeyW):
			playerAnimIndex = 6 // Index of animation: Running
		case rl.IsKeyDown(rl.KeyR):
			playerAnimIndex = 5 // Index of animation: Punch
		default:
			playerAnimIndex = 2 // Index of animation: Idle
		}

		playerAnimPlaying := robotAnims[playerAnimIndex]
		playerAnimCurrentFrame = (playerAnimCurrentFrame + 1) % int(playerAnimPlaying.KeyframeCount)
		rl.UpdateModelAnimation(playerModel, playerAnimPlaying, float32(playerAnimCurrentFrame))

		monsterAnimPlaying := robotAnims[monsterAnimIndex]
		switch {
		case monsterDead:
			monsterAnimCurrentFrame = min(monsterAnimCurrentFrame+1, int(monsterAnimPlaying.KeyframeCount)-1)
		default:
			monsterAnimCurrentFrame = (monsterAnimCurrentFrame + 1) % int(monsterAnimPlaying.KeyframeCount)
		}
		rl.UpdateModelAnimation(monsterModel, monsterAnimPlaying, float32(monsterAnimCurrentFrame))

		//----------------------------------------------------------------------------------
		// Draw
		//----------------------------------------------------------------------------------
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		rl.BeginMode3D(camera)
		rl.DrawModel(model, mapPosition, 1.0, rl.White) // Draw maze map

		rl.DrawModel(playerModel, playerPos, 0.2, rl.White) // Draw robot player

		rl.DrawModel(monsterModel, monsterPos, 0.2, rl.White) // Draw robot monster

		rl.EndMode3D()
		rl.DrawTextureEx(cubicmap, rl.NewVector2(float32(rl.GetScreenWidth())-float32(cubicmap.Width)*4.0-20, 20.0), 0.0, 4.0, rl.White)
		rl.DrawRectangleLines(int32(rl.GetScreenWidth())-cubicmap.Width*4-20, 20, cubicmap.Width*4, cubicmap.Height*4, rl.Green)
		// Draw player position radar
		rl.DrawRectangle(int32(rl.GetScreenWidth()-int(cubicmap.Width*4)-20+(playerCellX*4)), int32(20+playerCellY*4), 4, 4, rl.Red)

		rl.DrawFPS(10, 10)

		rl.EndDrawing()
		//----------------------------------------------------------------------------------
	}

	// De-Initialization
	//--------------------------------------------------------------------------------------
	rl.UnloadTexture(cubicmap) // Unload cubicmap texture
	rl.UnloadTexture(texture)  // Unload map texture
	rl.UnloadModel(model)      // Unload map model
	rl.CloseWindow()           // Close window and OpenGL context
	//--------------------------------------------------------------------------------------
}

func isPointTargeted(camera rl.Camera, pos rl.Vector3) bool {
	camPos := rl.Vector2{X: camera.Position.X, Y: camera.Position.Z}
	camTarget := rl.Vector2{X: camera.Target.X, Y: camera.Target.Z}
	pos2 := rl.Vector2{X: pos.X, Y: pos.Z}

	// Calculate angle between two vectors, considering a common origin (camTarget)
	v1Normal := rl.Vector2Normalize(rl.Vector2Subtract(camPos, camTarget))
	v2Normal := rl.Vector2Normalize(rl.Vector2Subtract(pos2, camTarget))
	angle := rl.Vector2Angle(v1Normal, v2Normal) * rl.Rad2deg

	return angle+30 >= 180 || angle-30 <= -180 // Aligned +- 30°
}

type Monster struct {
	cubicmap    rl.Texture2D
	mapPosition rl.Vector3
	mapPixels   []color.RGBA
	Position    rl.Vector2
}

func (m *Monster) CanSee(pos rl.Vector2) bool {
	x := (int)(pos.X - m.mapPosition.X + 0.5)
	y := (int)(pos.Y - m.mapPosition.Z + 0.5)

	// Out-of-limits security check
	if x < 0 || x >= int(m.cubicmap.Width) {
		return false
	}

	if y < 0 || y >= int(m.cubicmap.Height) {
		return false
	}

	return m.mapPixels[y*int(m.cubicmap.Width)+x].R != 255
}
