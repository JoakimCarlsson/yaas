package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dop251/goja"
	"github.com/joakimcarlsson/yaas/internal/models"
	"github.com/joakimcarlsson/yaas/internal/repository"
)

type ActionContext struct {
	User        map[string]interface{}
	Connection  string
	RequestInfo map[string]interface{}
}

type ActionExecutor struct {
	actionRepo repository.ActionRepository
}

func NewActionExecutor(actionRepo repository.ActionRepository) *ActionExecutor {
	return &ActionExecutor{actionRepo: actionRepo}
}

func (ae *ActionExecutor) ExecuteActions(ctx context.Context, actionType string, ac *ActionContext) error {
	actions, err := ae.actionRepo.GetActionsByType(ctx, actionType)
	if err != nil {
		return fmt.Errorf("failed to get actions: %w", err)
	}

	for _, action := range actions {
		if err := ae.executeAction(ctx, action, ac); err != nil {
			return fmt.Errorf("failed to execute action %s: %w", action.Name, err)
		}
	}

	return nil
}

func (ae *ActionExecutor) executeAction(ctx context.Context, action *models.Action, ac *ActionContext) error {
	vm := goja.New()

	jsContext, err := json.Marshal(ac)
	if err != nil {
		return err
	}

	vmCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	//vm.SetParserOptions(goja.WithDisableSourceMaps)

	err = vm.Set("context", string(jsContext))
	if err != nil {
		return fmt.Errorf("failed to set context: %w", err)
	}

	console := vm.NewObject()
	err = console.Set("log", func(call goja.FunctionCall) goja.Value {
		fmt.Printf("Action log: %s\n", call.Argument(0).String())
		return goja.Undefined()
	})
	if err != nil {
		return fmt.Errorf("failed to set console.log: %w", err)
	}
	err = vm.Set("console", console)
	if err != nil {
		return fmt.Errorf("failed to set console: %w", err)
	}

	_, err = vm.RunString(action.Code)
	if err != nil {
		return fmt.Errorf("failed to execute action: %w", err)
	}

	select {
	case <-vmCtx.Done():
		return fmt.Errorf("action execution cancelled: %w", vmCtx.Err())
	default:
		// Context not cancelled, continue
	}

	return nil
}
