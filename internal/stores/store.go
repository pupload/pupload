package stores

import (
	"bytes"
	"encoding/json"
	"fmt"
	"pupload/internal/models"
	locals3 "pupload/internal/stores/local_s3"
)

func UnmarshalStore(input models.StoreInput) (models.Store, error) {

	switch input.Type {

	case "s3":
		return nil, nil

	case "s3local":
		var params locals3.LocalS3StoreInput
		if err := decodeParams(input.Params, &params); err != nil {
			return nil, fmt.Errorf("decode params for s3local: %w", err)
		}

		store, err := locals3.NewLocalS3Store(params)
		if err != nil {
			return nil, fmt.Errorf("Unable to create locals3 store: %w", err)
		}
		return store, nil

	default:
		return nil, nil

	}

}

func decodeParams(raw json.RawMessage, out any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return err
	}

	if dec.More() {
		return fmt.Errorf("Unexpected extra data in params")
	}

	return nil
}
