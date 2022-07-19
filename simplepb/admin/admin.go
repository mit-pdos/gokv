package admin

import (
)

func EnterNewConfig() {
	// Get new epoch number from config service.
	// Read from config service, fenced with that epoch.

	// Enter new epoch on one of the old servers.
	// Get a copy of the state from that old server.
	// Set the state of all the new servers.
	// Write to config service saying the new servers have up-to-date state.
	// Tell one of the servers to become primary.
}
