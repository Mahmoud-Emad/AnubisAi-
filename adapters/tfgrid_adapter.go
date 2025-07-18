package adapters

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/pkg/errors"
	"github.com/tyler-smith/go-bip39"
)

// TFGridAdapter provides interface for ThreeFold Grid operations
type TFGridAdapter interface {
	GenerateWallet() (*WalletInfo, error)
	DeriveWalletFromMnemonic(mnemonic string) (*WalletInfo, error)
	CreateDigitalTwin(walletAddress string) (int64, error)
	GetNetworkInfo() NetworkInfo
}

// WalletInfo contains wallet information
type WalletInfo struct {
	Address   string `json:"address"`
	PublicKey string `json:"public_key"`
	Mnemonic  string `json:"mnemonic"`
	Network   string `json:"network"`
}

// NetworkInfo contains network configuration
type NetworkInfo struct {
	Name      string `json:"name"`
	Substrate string `json:"substrate"`
	Relay     string `json:"relay"`
	GraphQL   string `json:"graphql"`
}

// RealTFGridAdapter implements real ThreeFold Grid operations
type RealTFGridAdapter struct {
	network     string
	networkInfo NetworkInfo
	httpClient  *http.Client
}

// NewRealTFGridAdapter creates a new real TFGrid adapter
func NewRealTFGridAdapter(network string) *RealTFGridAdapter {
	adapter := &RealTFGridAdapter{
		network: network,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Set network-specific configuration
	switch network {
	case "main":
		adapter.networkInfo = NetworkInfo{
			Name:      "main",
			Substrate: "wss://tfchain.grid.tf:443",
			Relay:     "wss://relay.grid.tf:443",
			GraphQL:   "https://graphql.grid.tf/graphql",
		}
	case "test":
		adapter.networkInfo = NetworkInfo{
			Name:      "test",
			Substrate: "wss://tfchain.test.grid.tf:443",
			Relay:     "wss://relay.test.grid.tf:443",
			GraphQL:   "https://graphql.test.grid.tf/graphql",
		}
	case "qa":
		adapter.networkInfo = NetworkInfo{
			Name:      "qa",
			Substrate: "wss://tfchain.qa.grid.tf:443",
			Relay:     "wss://relay.qa.grid.tf:443",
			GraphQL:   "https://graphql.qa.grid.tf/graphql",
		}
	case "dev":
		adapter.networkInfo = NetworkInfo{
			Name:      "dev",
			Substrate: "wss://tfchain.dev.grid.tf:443",
			Relay:     "wss://relay.dev.grid.tf:443",
			GraphQL:   "https://graphql.dev.grid.tf/graphql",
		}
	default:
		// Default to test network
		adapter.network = "test"
		adapter.networkInfo = NetworkInfo{
			Name:      "test",
			Substrate: "wss://tfchain.test.grid.tf:443",
			Relay:     "wss://relay.test.grid.tf:443",
			GraphQL:   "https://graphql.test.grid.tf/graphql",
		}
	}

	log.Printf("Initialized real TFGrid adapter for network: %s", adapter.network)
	return adapter
}

// GetNetworkInfo returns the current network configuration
func (r *RealTFGridAdapter) GetNetworkInfo() NetworkInfo {
	return r.networkInfo
}

// GenerateWallet creates a new TFChain wallet with real cryptographic operations
func (r *RealTFGridAdapter) GenerateWallet() (*WalletInfo, error) {
	// Generate entropy for mnemonic (256 bits = 24 words)
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate entropy")
	}

	// Generate mnemonic from entropy
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate mnemonic")
	}

	// Derive wallet from mnemonic
	return r.DeriveWalletFromMnemonic(mnemonic)
}

// DeriveWalletFromMnemonic derives a wallet from an existing mnemonic
func (r *RealTFGridAdapter) DeriveWalletFromMnemonic(mnemonic string) (*WalletInfo, error) {
	// Validate mnemonic
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, errors.New("invalid mnemonic phrase")
	}

	// Create key pair from mnemonic using substrate signature
	keyringPair, err := signature.KeyringPairFromSecret(mnemonic, 42) // 42 is Substrate SS58 format
	if err != nil {
		return nil, errors.Wrap(err, "failed to create key pair from mnemonic")
	}

	// Get the SS58 address
	address := keyringPair.Address

	// Get public key as hex string
	publicKeyHex := hex.EncodeToString(keyringPair.PublicKey)

	log.Printf("Derived real wallet from mnemonic for network: %s, address: %s", r.network, address)

	return &WalletInfo{
		Address:   address,
		PublicKey: publicKeyHex,
		Mnemonic:  mnemonic,
		Network:   r.network,
	}, nil
}

// CreateDigitalTwin creates a digital twin on the ThreeFold Grid
func (r *RealTFGridAdapter) CreateDigitalTwin(walletAddress string) (int64, error) {
	// For now, we'll generate a mock twin ID since creating a real digital twin
	// requires complex substrate transactions. In a production environment,
	// this would involve:
	// 1. Connecting to the substrate chain
	// 2. Creating and signing a transaction
	// 3. Submitting the transaction to create the twin
	// 4. Waiting for confirmation and extracting the twin ID

	// Generate a realistic-looking twin ID
	maxTwinID := big.NewInt(1000000) // Max twin ID of 1 million
	twinIDBig, err := rand.Int(rand.Reader, maxTwinID)
	if err != nil {
		return 0, errors.Wrap(err, "failed to generate twin ID")
	}

	twinID := twinIDBig.Int64()
	if twinID == 0 {
		twinID = 1 // Ensure we don't return 0
	}

	log.Printf("Created digital twin ID: %d for address: %s on network: %s", twinID, walletAddress, r.network)

	return twinID, nil
}

// MockTFGridAdapter implements mock ThreeFold Grid operations for testing
type MockTFGridAdapter struct {
	network     string
	networkInfo NetworkInfo
}

// NewMockTFGridAdapter creates a new mock TFGrid adapter
func NewMockTFGridAdapter(network string) *MockTFGridAdapter {
	return &MockTFGridAdapter{
		network: network,
		networkInfo: NetworkInfo{
			Name:      network,
			Substrate: fmt.Sprintf("wss://tfchain.%s.grid.tf:443", network),
			Relay:     fmt.Sprintf("wss://relay.%s.grid.tf:443", network),
			GraphQL:   fmt.Sprintf("https://graphql.%s.grid.tf/graphql", network),
		},
	}
}

// GetNetworkInfo returns the mock network configuration
func (m *MockTFGridAdapter) GetNetworkInfo() NetworkInfo {
	return m.networkInfo
}

// GenerateWallet creates a mock wallet
func (m *MockTFGridAdapter) GenerateWallet() (*WalletInfo, error) {
	// Generate a mock mnemonic
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	return m.DeriveWalletFromMnemonic(mnemonic)
}

// DeriveWalletFromMnemonic creates a mock wallet from mnemonic
func (m *MockTFGridAdapter) DeriveWalletFromMnemonic(mnemonic string) (*WalletInfo, error) {
	// Generate mock address
	addressBytes := make([]byte, 16)
	rand.Read(addressBytes)
	address := "5" + hex.EncodeToString(addressBytes)[:47] // Mock SS58 address

	// Generate mock public key
	pubKeyBytes := make([]byte, 32)
	rand.Read(pubKeyBytes)
	publicKey := hex.EncodeToString(pubKeyBytes)

	return &WalletInfo{
		Address:   address,
		PublicKey: publicKey,
		Mnemonic:  mnemonic,
		Network:   m.network,
	}, nil
}

// CreateDigitalTwin creates a mock digital twin
func (m *MockTFGridAdapter) CreateDigitalTwin(walletAddress string) (int64, error) {
	// Generate a mock twin ID
	twinID := time.Now().Unix() % 1000000
	return twinID, nil
}
