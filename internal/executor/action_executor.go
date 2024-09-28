package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/joakimcarlsson/yaas/internal/logger"
	"github.com/joakimcarlsson/yaas/internal/models"
	"github.com/joakimcarlsson/yaas/internal/repository"
	"io"
	"net/http"
)

type ActionContext struct {
	User        map[string]interface{}
	Connection  string
	RequestInfo map[string]interface{}
}

type ActionExecutor struct {
	actionRepo repository.ActionRepository
	httpClient *http.Client
}

func NewActionExecutor(actionRepo repository.ActionRepository) *ActionExecutor {
	return &ActionExecutor{actionRepo: actionRepo, httpClient: &http.Client{}}
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

	registry := new(require.Registry)
	registry.Enable(vm)

	consoleLogger := func(call goja.FunctionCall) goja.Value {
		logger.WithFields(logger.Fields{
			"action": action.Name,
			"params": call.Arguments,
		})
		return goja.Undefined()
	}

	if err := vm.Set("console", map[string]interface{}{
		"log":   consoleLogger,
		"error": consoleLogger,
	}); err != nil {
		return fmt.Errorf("failed to set custom console logger: %w", err)
	}

	if err := ae.setupEnvironment(vm, ctx); err != nil {
		return fmt.Errorf("failed to setup environment: %w", err)
	}

	err := vm.Set("context", map[string]interface{}{
		"user":         ac.User,
		"connection":   ac.Connection,
		"request_info": ac.RequestInfo,
	})
	if err != nil {
		return fmt.Errorf("failed to set context in JS runtime: %w", err)
	}

	wrappedCode := fmt.Sprintf(`
        const run = async () => {
            try {
                %s
            } catch (error) {
                console.error('Error in action execution:', error);
                throw error;
            }
        };
        run().catch(error => {
            throw error;
        });
    `, action.Code)

	_, err = vm.RunString(wrappedCode)
	if err != nil {
		if gojaErr, ok := err.(*goja.Exception); ok {
			stackTrace := gojaErr.Value().ToString()
			fmt.Printf("Goja error: %s\nStack trace: %s\n", err.Error(), stackTrace)
		}
		return fmt.Errorf("failed to execute action: %w", err)
	}

	return nil
}

func (ae *ActionExecutor) setupEnvironment(vm *goja.Runtime, ctx context.Context) error {
	if err := vm.Set("fetch", func(call goja.FunctionCall) goja.Value {
		return ae.fetchPolyfill(vm, call)
	}); err != nil {
		return fmt.Errorf("failed to set fetch: %w", err)
	}

	return nil
}

func (ae *ActionExecutor) fetchPolyfill(vm *goja.Runtime, call goja.FunctionCall) goja.Value {
	url := call.Argument(0).String()
	options := call.Argument(1).Export()

	method := "GET"
	headers := make(http.Header)
	var body []byte

	if options != nil {
		if opt, ok := options.(map[string]interface{}); ok {
			if m, ok := opt["method"].(string); ok {
				method = m
			}
			if h, ok := opt["headers"].(map[string]interface{}); ok {
				for k, v := range h {
					headers.Set(k, fmt.Sprint(v))
				}
			}
			if b, ok := opt["body"].(string); ok {
				body = []byte(b)
			}
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return vm.ToValue(map[string]interface{}{
			"error": err.Error(),
		})
	}

	req.Header = headers

	resp, err := ae.httpClient.Do(req)
	if err != nil {
		return vm.ToValue(map[string]interface{}{
			"error": err.Error(),
		})
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return vm.ToValue(map[string]interface{}{
			"error": err.Error(),
		})
	}

	return vm.ToValue(map[string]interface{}{
		"ok":     resp.StatusCode >= 200 && resp.StatusCode < 300,
		"status": resp.StatusCode,
		"json": func() goja.Value {
			var result interface{}
			err := json.Unmarshal(respBody, &result)
			if err != nil {
				return vm.ToValue(map[string]interface{}{
					"error": err.Error(),
				})
			}
			return vm.ToValue(result)
		},
		"text": func() goja.Value {
			return vm.ToValue(string(respBody))
		},
	})
}
