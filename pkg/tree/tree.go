package provider

import (
	"database/sql"
	"errors"
	"fmt"
)

// Tree represents a row in the Trees table
type Tree struct {
	TreeId                int64
	TreeState             string
	TreeType              string
	HashStrategy          string
	HashAlgorithm         string
	SignatureAlgorithm    string
	DisplayName           sql.NullString
	Description           sql.NullString
	CreateTimeMillis      int64
	UpdateTimeMillis      int64
	MaxRootDurationMillis int64
	PrivateKey            []byte
	PublicKey             []byte
	Deleted               sql.NullBool
	DeleteTimeMillis      sql.NullInt64
}

// GetTreeByDescription fetches a row from the Trees table based on the description
func (p *DBProvider) GetTreeByDescription(description string) (*Tree, error) {
	query := `SELECT * FROM Trees WHERE Description = ?`
	row := p.DB.QueryRow(query, description)

	// Scan the result into a Tree struct
	var tree Tree
	err := row.Scan(
		&tree.TreeId,
		&tree.TreeState,
		&tree.TreeType,
		&tree.HashStrategy,
		&tree.HashAlgorithm,
		&tree.SignatureAlgorithm,
		&tree.DisplayName,
		&tree.Description,
		&tree.CreateTimeMillis,
		&tree.UpdateTimeMillis,
		&tree.MaxRootDurationMillis,
		&tree.PrivateKey,
		&tree.PublicKey,
		&tree.Deleted,
		&tree.DeleteTimeMillis,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no tree found with description: %s", description)
		}
		return nil, fmt.Errorf("error querying tree: %v", err)
	}

	return &tree, nil
}
