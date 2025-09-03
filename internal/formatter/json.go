package formatter

import (
	"encoding/json"

	"github.com/spandigital/mcp-server-dump/internal/model"
)

// FormatJSON formats server info as JSON
func FormatJSON(info *model.ServerInfo) ([]byte, error) {
	return json.MarshalIndent(info, "", "  ")
}
