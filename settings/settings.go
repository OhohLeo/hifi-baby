package settings

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/OhohLeo/hifi-baby/audio"
)

type Settings struct {
	Audio audio.Settings `json:"audio"`

	path string
}

func NewSettings(path string) (*Settings, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var settings Settings

	// Decode the JSON config file into the settings field
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&settings); err != nil {
		return nil, err
	}

	settings.path = path

	return &settings, nil
}

func (s *Settings) Update(newSettings *Settings) error {
	updatedSettingsData, err := json.Marshal(newSettings)
	if err != nil {
		return err
	}

	// Write the updated configuration back to file
	err = os.WriteFile(s.path, updatedSettingsData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to file at path %q: %w", s.path, err)
	}

	return nil
}
