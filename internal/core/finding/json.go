package finding

import (
	"encoding/json"
	"fmt"
)

func WriteJSON(findings []*Finding) error {
	if len(findings) == 0 {
		fmt.Println("[]")
	}

	b, err := json.MarshalIndent(findings, "", "    ")
	if err != nil {
		return err
	}
	c := string(b)
	if c == "null" {
		fmt.Println("[]")
	} else {
		fmt.Println(c)
	}

	return nil
}
