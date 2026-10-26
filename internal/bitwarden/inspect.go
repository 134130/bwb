package bitwarden

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type ItemField struct {
	Key       string  `json:"key"`
	Label     string  `json:"label"`
	Sensitive bool    `json:"sensitive"`
	State     string  `json:"state"`
	Value     *string `json:"value"`
}

type ItemDetails struct {
	Fields      []ItemField `json:"fields"`
	Restriction string      `json:"restriction,omitempty"`
}

func (s *Session) Inspect(ctx context.Context, snapshot *Snapshot, id, selected string, reveal bool) (ItemDetails, error) {
	fields, err := s.Describe(ctx, snapshot, id)
	if err != nil {
		if selected == "" && (errors.Is(err, ErrReprompt) || errors.Is(err, ErrPermission)) {
			return ItemDetails{Fields: []ItemField{}, Restriction: err.Error()}, nil
		}
		if selected != "name" || !(errors.Is(err, ErrReprompt) || errors.Is(err, ErrPermission)) {
			return ItemDetails{}, err
		}
	}
	fields = append([]FieldInfo{{Path: "name", Key: "name", Label: "Name"}}, fields...)
	if selected != "" {
		var found *FieldInfo
		available := make([]string, 0, len(fields))
		for _, field := range fields {
			available = append(available, field.Key)
			if field.Key == selected {
				found = &field
			}
		}
		if found == nil {
			return ItemDetails{}, fmt.Errorf("unknown field %q; available fields: %s", selected, strings.Join(available, ", "))
		}
		fields = []FieldInfo{*found}
	}
	result := ItemDetails{Fields: make([]ItemField, 0, len(fields))}
	var decrypted *Vault
	for _, field := range fields {
		if err := ctx.Err(); err != nil {
			return ItemDetails{}, err
		}
		value := ItemField{Key: field.Key, Label: field.Label, Sensitive: field.Sensitive, State: "hidden"}
		if selected != "" || reveal || !field.Sensitive {
			if decrypted == nil {
				item, err := s.readItem(snapshot, id)
				if err != nil {
					return ItemDetails{}, err
				}
				decrypted = &Vault{Items: []Item{item}}
			}
			plain, err := decrypted.Read(id, field.Path)
			if err != nil {
				return ItemDetails{}, err
			}
			value.Value, value.State = &plain, "visible"
			if plain == "" {
				value.State = "empty"
			}
		}
		result.Fields = append(result.Fields, value)
	}
	return result, nil
}
