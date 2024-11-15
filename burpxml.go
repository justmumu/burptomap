package burptomap

import (
	"encoding/xml"
	"os"
)

func UnmarshalFile(fp string) (*Items, error) {
	fBytes, err := os.ReadFile(fp)
	if err != nil {
		return nil, err
	}

	root := Items{}

	if err := xml.Unmarshal(fBytes, &root); err != nil {
		return nil, err
	}

	return &root, nil
}
