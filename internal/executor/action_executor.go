package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

type ActionResult struct {
	Allow   bool
	User    map[string]interface{}
	Error   error
	Message string
}

type ActionExecutor struct {
	actionRepo repository.ActionRepository
	httpClient *http.Client
}

type actionAPI struct {
	allowCalled   bool
	denyCalled    bool
	allow         bool
	user          map[string]interface{}
	error         error
	message       string
	loggingCalled bool
}

func (a *actionAPI) Allow() {
	a.allowCalled = true
	a.allow = true
}

func (a *actionAPI) Deny(message string) {
	a.denyCalled = true
	a.allow = false
	a.message = message
}

func (a *actionAPI) SetUser(user map[string]interface{}) {
	a.user = user
}

func (a *actionAPI) Log(message string) {
	a.loggingCalled = true
	logger.GetLogger().Info(message)
}

func NewActionExecutor(actionRepo repository.ActionRepository) *ActionExecutor {
	return &ActionExecutor{actionRepo: actionRepo, httpClient: &http.Client{}}
}

func (ae *ActionExecutor) ExecuteActions(ctx context.Context, actionType string, ac *ActionContext) (*ActionResult, error) {
	actions, err := ae.actionRepo.GetActionsByType(ctx, actionType)
	if err != nil {
		return nil, fmt.Errorf("failed to get actions: %w", err)
	}

	result := &ActionResult{
		Allow: true,
		User:  ac.User,
	}

	for _, action := range actions {
		actionResult, err := ae.executeAction(action, ac)
		if err != nil {
			return nil, fmt.Errorf("failed to execute action %s: %w", action.Name, err)
		}

		result.Allow = result.Allow && actionResult.Allow
		result.User = actionResult.User
		if actionResult.Error != nil {
			result.Error = actionResult.Error
		}
		if actionResult.Message != "" {
			result.Message = actionResult.Message
		}

		if !result.Allow {
			break
		}
	}

	return result, nil
}

func (ae *ActionExecutor) executeAction(action *models.Action, ac *ActionContext) (*ActionResult, error) {
	vm := goja.New()
	registry := new(require.Registry)
	registry.Enable(vm)
	errs := ae.setupEnvironment(vm)
	if errs != nil {
		return nil, fmt.Errorf("failed to setup environment: %w", errs)
	}

	// Set up the action context
	err := vm.Set("context", map[string]interface{}{
		"user":         ac.User,
		"connection":   ac.Connection,
		"request_info": ac.RequestInfo,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set context in JS runtime: %w", err)
	}

	actionAPI := &actionAPI{
		allowCalled:   false,
		denyCalled:    false,
		allow:         true,
		user:          ac.User,
		error:         nil,
		message:       "",
		loggingCalled: false,
	}

	err = vm.Set("yaas", map[string]interface{}{
		"allow":   actionAPI.Allow,
		"deny":    actionAPI.Deny,
		"setUser": actionAPI.SetUser,
		"log":     actionAPI.Log,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set action API in JS runtime: %w", err)
	}

	// Execute the action code
	_, err = vm.RunString(action.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to execute action: %w", err)
	}

	// Check if allow or deny was called
	if !actionAPI.allowCalled && !actionAPI.denyCalled {
		return nil, errors.New("action must call either yaas.allow() or yaas.deny()")
	}

	return &ActionResult{
		Allow:   actionAPI.allow,
		User:    actionAPI.user,
		Error:   actionAPI.error,
		Message: actionAPI.message,
	}, nil
}

func (ae *ActionExecutor) setupEnvironment(vm *goja.Runtime) error {
	console := vm.NewObject()
	if err := console.Set("log", func(call goja.FunctionCall) goja.Value {
		for _, arg := range call.Arguments {
			fmt.Println(arg.String())
		}
		return goja.Undefined()
	}); err != nil {
		return fmt.Errorf("failed to set console.log: %w", err)
	}
	if err := vm.Set("console", console); err != nil {
		return fmt.Errorf("failed to set console: %w", err)
	}

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
