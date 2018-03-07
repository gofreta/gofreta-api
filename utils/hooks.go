package utils

import (
	"encoding/json"
)

const (
	// HookTypeCollection specify the `collection` type hook.
	HookTypeCollection = "collection"

	// HookTypeEntity specify the `entity` type hook.
	HookTypeEntity = "entity"

	// HookActionCreate specify the `create` action hook.
	HookActionCreate = "create"

	// HookActionUpdate specify the `update` action hook.
	HookActionUpdate = "update"

	// HookActionDelete specify the `delete` action hook.
	HookActionDelete = "delete"
)

// Hook defines the general hooks format structure.
type Hook struct {
	Type   string      `json:"type" bson:"type"`
	Action string      `json:"action" bson:"action"`
	Data   interface{} `json:"data" bson:"data"`
}

// NewHook creates and returns new Hook instance.
func NewHook(hookType string, hookAction string, hookData interface{}) *Hook {
	return &Hook{
		Type:   hookType,
		Action: hookAction,
		Data:   hookData,
	}
}

// SendHook creates new hook instance and sends a POST request to the specified url.
func SendHook(url string, hookType string, hookAction string, hookData interface{}) error {
	if url == "" {
		return nil
	}

	hook := NewHook(hookType, hookAction, hookData)

	bytes, err := json.Marshal(hook)
	if err != nil {
		return err
	}

	return SendJsonPostData(url, bytes)
}
