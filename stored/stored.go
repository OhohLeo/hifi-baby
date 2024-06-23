package stored

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/OhohLeo/hifi-baby/audio"
)

type Stored struct {
	Audio audio.StoredConfig `json:"audio"`

	path string
}

func NewStored(path string) (*Stored, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var stored Stored

	// Decode the JSON config file into the storedConfig field
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&stored); err != nil {
		return nil, err
	}

	stored.path = path

	return &stored, nil
}

func (s *Stored) Update(newStored *Stored) error {
	updatedStoredData, err := json.Marshal(newStored)
	if err != nil {
		return err
	}

	// Write the updated configuration back to file
	err = os.WriteFile(s.path, updatedStoredData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to file at path %q: %w", s.path, err)
	}

	return nil
}
