package main

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// --- STRUCTURES FOR AVATAR INTERFACE ---

type AvatarConfig struct {
	UserID      int64  `json:"user_id"`
	SkinColorR  uint8  `json:"skin_r"`
	SkinColorG  uint8  `json:"skin_g"`
	SkinColorB  uint8  `json:"skin_b"`
	ShirtAsset  string `json:"shirt_url"` // URL to download shirt texture
}

type SceneMap struct {
	XMLName xml.Name `xml:"Scene"`
	Parts   []MapPart `xml:"Part"`
}

type MapPart struct {
	X, Y, Z             float32 `xml:"X"`
	Width, Height, Depth float32 `xml:"Width"`
	R, G, B             uint8   `xml:"R"`
}

// Player state tracking structure
type Character struct {
	Position    rl.Vector3
	Velocity    rl.Vector3
	Speed       float32
	IsGrounded  bool
	SkinColor   rl.Color
	HasShirt    bool
	ShirtTex    rl.Texture2D
}

// Fetch avatar profile data from our web server
func fetchAvatarData(serverURL, userID string) (*AvatarConfig, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/avatar/%s", serverURL, userID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var config AvatarConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

// --- MAIN ENGINE RUNTIME LOOP WITH MOVEMENT ---

func StartEngineWithMovement(downloadURL, serverURL, userID string) {
	// 1. Download scene layout
	fmt.Println("[Engine] Fetching scene topology map...")
	scene, err := downloadAndParseMap(downloadURL) // Uses the zip extractor from previous file
	if err != nil {
		fmt.Printf("[Error] Map initialization failed: %v\n", err)
		return
	}

	// 2. Initialize Hardware Display Window
	rl.InitWindow(800, 600, "Noisogic Player - 3D Simulation")
	defer rl.CloseWindow()

	// 3. Instantiate local character settings
	player := Character{
		Position:  rl.NewVector3(0.0, 3.0, 0.0), // Spawn location
		Speed:     12.0,
		SkinColor: rl.NewColor(253, 234, 141, 255), // Default yellow avatar color
	}

	// Try fetching custom avatar configuration if logged in
	if userID != "" && userID != "0" {
		cfg, err := fetchAvatarData(serverURL, userID)
		if err == nil {
			player.SkinColor = rl.NewColor(cfg.SkinColorR, cfg.SkinColorG, cfg.SkinColorB, 255)
			// Optional: If a shirt texture is specified, you would download and call rl.LoadTextureFromImage() here
		}
	}

	// Setup look-at camera tracking target
	camera := rl.Camera3D{}
	camera.Position = rl.NewVector3(0.0, 12.0, 25.0)
	camera.Target = player.Position
	camera.Up = rl.NewVector3(0.0, 1.0, 0.0)
	camera.Fovy = 60.0
	camera.Projection = rl.CameraPerspective

	rl.SetTargetFPS(60)

	// Main physics & rendering frame tick
	for !rl.WindowShouldClose() {
		deltaTime := rl.GetFrameTime()

		// --- INPUT POLLING & PHYSICS MOVEMENT ENGINE ---
		
		direction := rl.NewVector3(0, 0, 0)
		
		// Map keyboard strings directly to vector directions
		if rl.IsKeyDown(rl.KeyW) { direction.Z -= 1.0 }
		if rl.IsKeyDown(rl.KeyS) { direction.Z += 1.0 }
		if rl.IsKeyDown(rl.KeyA) { direction.X -= 1.0 }
		if rl.IsKeyDown(rl.KeyD) { direction.X += 1.0 }

		if rl.Vector3Length(direction) > 0 {
			direction = rl.Vector3Normalize(direction)
			// Multiply vector components by our configured velocity multipliers
			player.Velocity.X = direction.X * player.Speed
			player.Velocity.Z = direction.Z * player.Speed
		} else {
			// Apply classic vintage friction deceleration drop-off
			player.Velocity.X *= 0.8
			player.Velocity.Z *= 0.8
		}

		// Simple gravity simulation physics sequence
		const Gravity float32 = -28.0
		if !player.IsGrounded {
			player.Velocity.Y += Gravity * deltaTime
		} else {
			player.Velocity.Y = 0
			// Jump constraint check via Space bar trigger
			if rl.IsKeyPressed(rl.KeySpace) {
				player.Velocity.Y = 12.0 // Instant upwards velocity spike
				player.IsGrounded = false
			}
		}

		// Apply final computed velocities to the 3D translation matrix components
		player.Position.X += player.Velocity.X * deltaTime
		player.Position.Y += player.Velocity.Y * deltaTime
		player.Position.Z += player.Velocity.Z * deltaTime

		// Basic bounds logic check (Floor plane boundary safeguard at Y = 1.0)
		if player.Position.Y < 1.0 {
			player.Position.Y = 1.0
			player.IsGrounded = true
		}

		// Keep camera locked behind our moving 3D avatar coordinate space
		camera.Target = player.Position
		camera.Position.X = player.Position.X + 0.0
		camera.Position.Y = player.Position.Y + 10.0
		camera.Position.Z = player.Position.Z + 20.0

		// --- GRAPHICS HARDWARE GENERATION FRAME STEP ---
		rl.BeginDrawing()
		rl.ClearBackground(rl.NewColor(192, 208, 255, 255)) // Classic sky blue

		rl.BeginMode3D(camera)
		rl.DrawGrid(100, 2.0) // Render bottom baseplate reference grid

		// Draw world static environment elements loaded from .pxgf package
		for _, part := range scene.Parts {
			pos := rl.NewVector3(part.X, part.Y, part.Z)
			col := rl.NewColor(part.R, part.G, part.B, 255)
			rl.DrawCube(pos, part.Width, part.Height, part.Depth, col)
			rl.DrawCubeWires(pos, part.Width, part.Height, part.Depth, rl.DarkGray)
		}

		// --- DRAW PLAYER AVATAR PRIMITIVE ---
		// Compose character out of distinct primitive cubes matching early character scales
		// Head block
		rl.DrawCube(rl.NewVector3(player.Position.X, player.Position.Y+2.2, player.Position.Z), 1.0, 1.0, 1.0, player.SkinColor)
		rl.DrawCubeWires(rl.NewVector3(player.Position.X, player.Position.Y+2.2, player.Position.Z), 1.0, 1.0, 1.0, rl.Black)
		// Torso block
		rl.DrawCube(rl.NewVector3(player.Position.X, player.Position.Y+1.0, player.Position.Z), 2.0, 2.0, 1.0, rl.Blue) // Classic blue torso default
		rl.DrawCubeWires(rl.NewVector3(player.Position.X, player.Position.Y+1.0, player.Position.Z), 2.0, 2.0, 1.0, rl.Black)

		rl.EndMode3D()

		// Overlay text diagnostic metrics interface
		rl.DrawText("CONTROLS: WASD to Move | SPACEBAR to Jump", 10, 10, 18, rl.Black)
		rl.DrawText(fmt.Sprintf("Position: X: %.2f, Y: %.2f, Z: %.2f", player.Position.X, player.Position.Y, player.Position.Z), 10, 35, 14, rl.DarkGray)
		rl.EndDrawing()
	}
}