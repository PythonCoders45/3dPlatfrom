package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // Example driver for PostgreSQL
)

/*
   DATABASE HANDLER STRUCTURE
*/

type Database struct {
	conn *sql.DB
}

func NewDatabase(dsn string) (*Database, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &Database{conn: db}, nil
}

/*
   1. USER MANAGEMENT
*/

// AddUser creates a new user and initializes their profile and balance
func (db *Database) AddUser(ctx context.Context, userID int64, username, passwordHash, email string) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert into Users
	_, err = tx.ExecContext(ctx,
		`INSERT INTO Users (UserID, Username, PasswordHash, Email) VALUES ($1, $2, $3, $4)`,
		userID, username, passwordHash, email)
	if err != nil {
		return err
	}

	// Initialize Profile
	_, err = tx.ExecContext(ctx,
		`INSERT INTO UserProfiles (UserID, Blurb, AvatarData) VALUES ($1, $2, $3)`,
		userID, "Hello! I am new here.", "{}")
	if err != nil {
		return err
	}

	// Initialize Balance
	_, err = tx.ExecContext(ctx,
		`INSERT INTO RobuxBalances (UserID, CurrentBalance) VALUES ($1, $2)`,
		userID, 0)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// RemoveUser deletes a user and related data (assuming manual cleanup for complexity)
func (db *Database) RemoveUser(ctx context.Context, userID int64) error {
	// Simple delete - database FK constraints must allow this or be handled
	_, err := db.conn.ExecContext(ctx, `DELETE FROM Users WHERE UserID = $1`, userID)
	return err
}

/*
   2. ASSETS & INVENTORY
*/

// AddAsset adds a new creator-defined asset
func (db *Database) AddAsset(ctx context.Context, assetID, creatorID int64, name, assetType string, price int, hashID string) error {
	query := `INSERT INTO Assets (AssetID, AssetName, CreatorID, AssetType, RobuxPrice, HashID) 
              VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.conn.ExecContext(ctx, query, assetID, name, creatorID, assetType, price, hashID)
	return err
}

// AddToInventory grants an asset instance to a user
func (db *Database) AddToInventory(ctx context.Context, inventoryID, userID, assetID int64, serial *int) error {
	query := `INSERT INTO Inventories (InventoryID, UserID, AssetID, SerialNumber) VALUES ($1, $2, $3, $4)`
	_, err := db.conn.ExecContext(ctx, query, inventoryID, userID, assetID, serial)
	return err
}

/*
   3. SOCIAL & GROUPS
*/

// AddFriendship creates a friendship record
func (db *Database) AddFriendship(ctx context.Context, u1, u2 int64, actionUserID int64) error {
	// Ensure UserID_1 is always the smaller ID for consistent lookup
	if u1 > u2 {
		u1, u2 = u2, u1
	}
	query := `INSERT INTO Friendships (UserID_1, UserID_2, Status, ActionUserID) VALUES ($1, $2, 'Pending', $3)`
	_, err := db.conn.ExecContext(ctx, query, u1, u2, actionUserID)
	return err
}

// RemoveFriendship deletes the link
func (db *Database) RemoveFriendship(ctx context.Context, u1, u2 int64) error {
	if u1 > u2 {
		u1, u2 = u2, u1
	}
	_, err := db.conn.ExecContext(ctx, `DELETE FROM Friendships WHERE UserID_1 = $1 AND UserID_2 = $2`, u1, u2)
	return err
}

/*
   4. ECONOMY (ROBUX & TRANSACTIONS)
*/

// ProcessPurchase transfers Robux and adds the item to inventory
func (db *Database) ProcessPurchase(ctx context.Context, buyerID, sellerID, assetID int64, price int) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Deduct Buyer
	_, err = tx.ExecContext(ctx, "UPDATE RobuxBalances SET CurrentBalance = CurrentBalance - $1 WHERE UserID = $2 AND CurrentBalance >= $1", price, buyerID)
	if err != nil {
		return fmt.Errorf("insufficient funds")
	}

	// 2. Credit Seller (minus 30% fee example)
	fee := int(float64(price) * 0.3)
	net := price - fee
	_, err = tx.ExecContext(ctx, "UPDATE RobuxBalances SET CurrentBalance = CurrentBalance + $1 WHERE UserID = $2", net, sellerID)
	if err != nil {
		return err
	}

	// 3. Record Transaction
	txID := time.Now().UnixNano()
	_, err = tx.ExecContext(ctx, `INSERT INTO MarketTransactions (TransactionID, BuyerID, SellerID, AssetID, RobuxAmount, MarketplaceFee) 
                                 VALUES ($1, $2, $3, $4, $5, $6)`, txID, buyerID, sellerID, assetID, price, fee)
	if err != nil {
		return err
	}

	return tx.Commit()
}

/*
   5. GAME SERVERS & DATASTORES
*/

// AddActiveServer registers a new running game instance
func (db *Database) AddActiveServer(ctx context.Context, jobID uuid.UUID, placeID int64, region string) error {
	query := `INSERT INTO ActiveServers (ServerJobID, PlaceID, CurrentPlayerCount, ServerRegion) VALUES ($1, $2, 0, $3)`
	_, err := db.conn.ExecContext(ctx, query, jobID, placeID, region)
	return err
}

// SetDataStoreValue saves developer-specific game data
func (db *Database) SetDataStoreValue(ctx context.Context, placeID int64, key, scope string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	query := `INSERT INTO DeveloperDataStores (PlaceID, DataKey, Scope, JsonData) 
              VALUES ($1, $2, $3, $4) 
              ON CONFLICT (PlaceID, DataKey, Scope) DO UPDATE SET JsonData = $4, UpdatedAt = CURRENT_TIMESTAMP`
	_, err = db.conn.ExecContext(ctx, query, placeID, key, scope, jsonData)
	return err
}

// RemoveDataStoreValue deletes a key from a game's datastore
func (db *Database) RemoveDataStoreValue(ctx context.Context, placeID int64, key, scope string) error {
	_, err := db.conn.ExecContext(ctx, `DELETE FROM DeveloperDataStores WHERE PlaceID = $1 AND DataKey = $2 AND Scope = $3`, placeID, key, scope)
	return err
}

/*
   6. MODERATION
*/

// AddModerationAction logs a ban or warning
func (db *Database) AddModerationAction(ctx context.Context, targetID, modID int64, actionType, reason string, duration time.Duration) error {
	id := time.Now().UnixNano()
	expiry := time.Now().Add(duration)
	query := `INSERT INTO ModerationActions (ActionID, TargetUserID, ModeratorID, ActionType, ReasonText, ExpiresAt) 
              VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.conn.ExecContext(ctx, query, id, targetID, modID, actionType, reason, expiry)
	return err
}

func main() {
	fmt.Println("Database logic file generated. This file provides CRUD operations for the provided schema.")
}