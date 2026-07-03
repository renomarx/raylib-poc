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
	"log"

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
	camera.Position = rl.NewVector3(0.2, 2.0, 0.2)
	camera.Target = rl.NewVector3(1.0, 0.5, 1.0)
	camera.Up = rl.NewVector3(0.0, 1.0, 0.0)
	camera.Fovy = 45.0
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

	// animIndex := 0
	// animCurrentFrame := 0

	// modelAnims := rl.LoadModelAnimations("robot.glb")

	//--------------------------------------------------------------------------------------

	var playerRot float32 = 0.0
	// Main game loop
	for !rl.WindowShouldClose() { // Detect window close button or ESC key
		// Update
		//----------------------------------------------------------------------------------
		oldCamPos := camera.Position // Store old camera position

		playerPos := camera.Target
		playerPos.Y = playerPos.Y - 0.5 // size of the player
		oldPlayerPos := playerPos

		// TODO: use a custom camera mode; the ThirdPerson is good for a poc,
		// but lacks customization (invert Y-axis as example)
		rl.UpdateCamera(&camera, rl.CameraThirdPerson)

		// Check player collision (we simplify to 2D collision detection)
		playerPos2 := rl.NewVector2(playerPos.X, playerPos.Z)
		playerRadius := 0.1 // Collision radius (player is modelled as a cylinder for collision)

		playerRot += rl.Vector3Angle(oldCamPos, camera.Position) * 100
		log.Printf("Camera rotation: %f", playerRot)
		playerModel.Transform = rl.MatrixRotateXYZ(rl.Vector3{X: 0, Y: rl.Deg2rad * playerRot, Z: 0})

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
		// Improvement: Just check player surrounding cells for collision
		minYsurrounding := max(playerCellY-10, 0)
		maxYsurrounding := min(playerCellY+10, int(cubicmap.Height))
		minXsurrounding := max(playerCellX-10, 0)
		maxXsurrounding := min(playerCellX+10, int(cubicmap.Width))
		for y := minYsurrounding; y < maxYsurrounding; y++ {
			for x := minXsurrounding; x < maxXsurrounding; x++ {
				// Collision: white pixel, only check R channel
				if mapPixels[y*int(cubicmap.Width)+x].R == 255 && (rl.CheckCollisionCircleRec(playerPos2, float32(playerRadius), rl.NewRectangle(float32(mapPosition.X-0.5+float32(x)), float32(mapPosition.Z-0.5+float32(y)), 1.0, 1.0))) {
					// Collision detected, reset camera position
					camera.Position = oldCamPos
					playerPos = oldPlayerPos
				}
			}
		}
		//----------------------------------------------------------------------------------
		// Draw
		//----------------------------------------------------------------------------------
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		rl.BeginMode3D(camera)
		rl.DrawModel(model, mapPosition, 1.0, rl.White) // Draw maze map

		rl.DrawModel(playerModel, playerPos, 0.2, rl.White) // Draw robot player

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
