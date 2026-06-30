CREATE TABLE Users (
    UserID BIGINT PRIMARY KEY,
    Username VARCHAR(20) UNIQUE NOT NULL,
    PasswordHash CHAR(60) NOT NULL,
    Email VARCHAR(255) UNIQUE,
    AccountStatus VARCHAR(20) DEFAULT 'Active',
    CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE UserProfiles (
    UserID BIGINT PRIMARY KEY,
    Blurb TEXT,
    AvatarData JSONB, -- Stores scaling, colors, and asset attachments
    FOREIGN KEY (UserID) REFERENCES Users(UserID)
);

CREATE TABLE RobuxBalances (
    UserID BIGINT PRIMARY KEY,
    CurrentBalance INT DEFAULT 0 CHECK (CurrentBalance >= 0),
    PremiumExpiryDate DATE,
    FOREIGN KEY (UserID) REFERENCES Users(UserID)
);

CREATE TABLE MarketTransactions (
    TransactionID BIGINT PRIMARY KEY,
    BuyerID BIGINT,
    SellerID BIGINT,
    AssetID BIGINT,
    RobuxAmount INT NOT NULL,
    MarketplaceFee INT NOT NULL,
    TransactionTime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (BuyerID) REFERENCES Users(UserID),
    FOREIGN KEY (SellerID) REFERENCES Users(UserID)
);

CREATE TABLE Assets (
    AssetID BIGINT PRIMARY KEY,
    AssetName VARCHAR(100) NOT NULL,
    CreatorID BIGINT NOT NULL,
    AssetType VARCHAR(30) NOT NULL, -- Hat, Gear, Shirt, Gamepass
    RobuxPrice INT DEFAULT 0,
    IsLimited BOOLEAN DEFAULT FALSE,
    TotalStock INT,
    HashID VARCHAR(64) NOT NULL, -- Points to the content delivery network (CDN) file
    FOREIGN KEY (CreatorID) REFERENCES Users(UserID)
);

CREATE TABLE Inventories (
    InventoryID BIGINT PRIMARY KEY,
    UserID BIGINT NOT NULL,
    AssetID BIGINT NOT NULL,
    AcquiredAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    SerialNumber INT, -- Used if the asset is a Limited item
    FOREIGN KEY (UserID) REFERENCES Users(UserID),
    FOREIGN KEY (AssetID) REFERENCES Assets(AssetID)
);

CREATE TABLE Places (
    PlaceID BIGINT PRIMARY KEY,
    UniverseID BIGINT NOT NULL,
    CreatorID BIGINT NOT NULL,
    Name VARCHAR(100) NOT NULL,
    MaxPlayers INT DEFAULT 50,
    FOREIGN KEY (CreatorID) REFERENCES Users(UserID)
);

CREATE TABLE ActiveServers (
    ServerJobID UUID PRIMARY KEY,
    PlaceID BIGINT NOT NULL,
    CurrentPlayerCount INT DEFAULT 0,
    ServerRegion VARCHAR(50),
    FOREIGN KEY (PlaceID) REFERENCES Places(PlaceID)
);

CREATE TABLE DeveloperDataStores (
    PlaceID BIGINT NOT NULL,
    DataKey VARCHAR(255) NOT NULL,
    Scope VARCHAR(50) DEFAULT 'global',
    JsonData JSONB NOT NULL, -- Stores custom game data (e.g., tycoon cash, simulator levels)
    UpdatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (PlaceID, DataKey, Scope),
    FOREIGN KEY (PlaceID) REFERENCES Places(PlaceID)
);

CREATE TABLE Friendships (
    UserID_1 BIGINT,
    UserID_2 BIGINT,
    Status VARCHAR(20) CHECK (Status IN ('Pending', 'Friends')),
    ActionUserID BIGINT NOT NULL, -- Who sent the request
    PRIMARY KEY (UserID_1, UserID_2),
    FOREIGN KEY (UserID_1) REFERENCES Users(UserID),
    FOREIGN KEY (UserID_2) REFERENCES Users(UserID)
);

CREATE TABLE Groups (
    GroupID BIGINT PRIMARY KEY,
    GroupName VARCHAR(50) UNIQUE NOT NULL,
    OwnerID BIGINT NOT NULL,
    Description TEXT,
    GroupRobuxBalance INT DEFAULT 0 CHECK (GroupRobuxBalance >= 0),
    CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (OwnerID) REFERENCES Users(UserID)
);

CREATE TABLE GroupRoles (
    RoleID BIGINT PRIMARY KEY,
    GroupID BIGINT NOT NULL,
    RankValue INT CHECK (RankValue BETWEEN 0 AND 255), -- 255 is always Owner
    RoleName VARCHAR(40) NOT NULL,
    CanPostOnWall BOOLEAN DEFAULT TRUE,
    CanSpendFunds BOOLEAN DEFAULT FALSE,
    CanKickMembers BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (GroupID) REFERENCES Groups(GroupID),
    UNIQUE (GroupID, RankValue)
);

CREATE TABLE GroupMembers (
    GroupID BIGINT NOT NULL,
    UserID BIGINT NOT NULL,
    RoleID BIGINT NOT NULL,
    JoinedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (GroupID, UserID),
    FOREIGN KEY (GroupID) REFERENCES Groups(GroupID),
    FOREIGN KEY (UserID) REFERENCES Users(UserID),
    FOREIGN KEY (RoleID) REFERENCES GroupRoles(RoleID)
);

CREATE TABLE TradeOffers (
    TradeID BIGINT PRIMARY KEY,
    SenderID BIGINT NOT NULL,
    ReceiverID BIGINT NOT NULL,
    TradeStatus VARCHAR(20) DEFAULT 'Pending', -- Pending, Accepted, Declined, Countered, Expired
    SenderRobuxSweetener INT DEFAULT 0,
    ReceiverRobuxSweetener INT DEFAULT 0,
    ExpiresAt TIMESTAMP NOT NULL,
    CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (SenderID) REFERENCES Users(UserID),
    FOREIGN KEY (ReceiverID) REFERENCES Users(UserID)
);

CREATE TABLE TradeItems (
    TradeItemID BIGINT PRIMARY KEY,
    TradeID BIGINT NOT NULL,
    InventoryID BIGINT NOT NULL, -- References the specific serial-numbered asset
    IsSenderItem BOOLEAN NOT NULL, -- TRUE if offered by sender, FALSE if requested from receiver
    FOREIGN KEY (TradeID) REFERENCES TradeOffers(TradeID),
    FOREIGN KEY (InventoryID) REFERENCES Inventories(InventoryID)
);

CREATE TABLE Reports (
    ReportID BIGINT PRIMARY KEY,
    ReporterID BIGINT NOT NULL,
    OffenderID BIGINT, -- Can be NULL if reporting a Place or Asset instead of a user
    TargetType VARCHAR(20) NOT NULL, -- User, Chat, Asset, Place
    TargetID BIGINT NOT NULL, -- ID of the specific item/player being reported
    ReasonCategory VARCHAR(50) NOT NULL, -- Bullying, Dating, Scamming, etc.
    Details TEXT,
    ReportStatus VARCHAR(20) DEFAULT 'Pending', -- Pending, Reviewed, Actioned, Dismissed
    CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ReporterID) REFERENCES Users(UserID)
);

CREATE TABLE ModerationActions (
    ActionID BIGINT PRIMARY KEY,
    TargetUserID BIGINT NOT NULL,
    ModeratorID BIGINT, -- NULL if issued by an automated filter system
    ActionType VARCHAR(20) NOT NULL, -- Warn, 1DayBan, 7DayBan, DeleteAccount
    ReasonText TEXT NOT NULL,
    ExpiresAt TIMESTAMP, -- NULL for permanent account deletion
    CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (TargetUserID) REFERENCES Users(UserID)
);

CREATE TABLE ChatLogs (
    MessageID BIGINT PRIMARY KEY,
    SenderID BIGINT NOT NULL,
    TargetServerID UUID, -- NULL if using website private messages instead of in-game chat
    RawMessage TEXT NOT NULL,
    FilteredMessage TEXT NOT NULL, -- Text after the Roblox hashtag filtering engine processes it
    SentAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (SenderID) REFERENCES Users(UserID)
);

CREATE TABLE Badges (
    BadgeID BIGINT PRIMARY KEY,
    PlaceID BIGINT NOT NULL,
    BadgeName VARCHAR(100) NOT NULL,
    Description TEXT,
    IconAssetID BIGINT NOT NULL,
    IsEnabled BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (PlaceID) REFERENCES Places(PlaceID),
    FOREIGN KEY (IconAssetID) REFERENCES Assets(AssetID)
);

CREATE TABLE AwardedBadges (
    UserID BIGINT NOT NULL,
    BadgeID BIGINT NOT NULL,
    AwardedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (UserID, BadgeID),
    FOREIGN KEY (UserID) REFERENCES Users(UserID),
    FOREIGN KEY (BadgeID) REFERENCES Badges(BadgeID)
);

CREATE TABLE AdCampaigns (
    CampaignID BIGINT PRIMARY KEY,
    CreatorID BIGINT NOT NULL,
    TargetAssetID BIGINT, -- Can point to an asset (shirt) or a place (game)
    CampaignType VARCHAR(20) CHECK (CampaignType IN ('Banner', 'Skyscraper', 'SponsoredGame')),
    TotalBidRobux INT NOT NULL CHECK (TotalBidRobux >= 10),
    DailyBudgetRobux INT NOT NULL,
    StartDate TIMESTAMP NOT NULL,
    EndDate TIMESTAMP NOT NULL,
    Status VARCHAR(20) DEFAULT 'Pending', -- Pending, Running, Completed, Paused
    FOREIGN KEY (CreatorID) REFERENCES Users(UserID)
);

CREATE TABLE AdPerformanceDaily (
    CampaignID BIGINT NOT NULL,
    LogDate DATE NOT NULL,
    Impressions BIGINT DEFAULT 0,
    Clicks BIGINT DEFAULT 0,
    RobuxSpent INT DEFAULT 0,
    PRIMARY KEY (CampaignID, LogDate),
    FOREIGN KEY (CampaignID) REFERENCES AdCampaigns(CampaignID)
);

CREATE TABLE DeveloperProducts (
    ProductID BIGINT PRIMARY KEY,
    PlaceID BIGINT NOT NULL,
    ProductName VARCHAR(100) NOT NULL,
    RobuxPrice INT NOT NULL CHECK (RobuxPrice >= 0),
    FOREIGN KEY (PlaceID) REFERENCES Places(PlaceID)
);

CREATE TABLE ExperienceSubscriptions (
    SubscriptionID BIGINT PRIMARY KEY,
    UniverseID BIGINT NOT NULL,
    Name VARCHAR(100) NOT NULL,
    RobuxPricePerMonth INT NOT NULL,
    IsActive BOOLEAN DEFAULT TRUE
);

CREATE TABLE UserSubscriptions (
    UserSubscriptionID BIGINT PRIMARY KEY,
    UserID BIGINT NOT NULL,
    SubscriptionID BIGINT NOT NULL,
    Status VARCHAR(20) DEFAULT 'Active', -- Active, Cancelled, Lapsed
    NextBillingDate TIMESTAMP NOT NULL,
    FOREIGN KEY (UserID) REFERENCES Users(UserID),
    FOREIGN KEY (SubscriptionID) REFERENCES ExperienceSubscriptions(SubscriptionID)
);

CREATE TABLE MatchmakingQueues (
    UserID BIGINT PRIMARY KEY,
    TargetPlaceID BIGINT NOT NULL,
    PartyID UUID, -- Groups friends together during matchmaking
    JoinedQueueAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (UserID) REFERENCES Users(UserID),
    FOREIGN KEY (TargetPlaceID) REFERENCES Places(PlaceID)
);

CREATE TABLE TeleportReservations (
    ReservationToken UUID PRIMARY KEY,
    SourceServerID UUID NOT NULL,
    TargetPlaceID BIGINT NOT NULL,
    TargetServerID UUID, -- Can be NULL if matchmaking picks the server dynamically
    ReservedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (TargetPlaceID) REFERENCES Places(PlaceID)
);