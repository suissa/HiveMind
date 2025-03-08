package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	v8 "rogchap.com/v8go"
)

// V8Executor implementa a interface JSExecutor usando V8
type V8Executor struct {
	isolate    *v8.Isolate
	global     *v8.ObjectTemplate
	modules    map[string]*v8.Object
	modulePath string
	mu         sync.Mutex
}

// NewV8Executor cria uma nova instância do V8Executor
func NewV8Executor() (*V8Executor, error) {
	// Criar isolate V8
	isolate := v8.NewIsolate()
	global := v8.NewObjectTemplate(isolate)

	// Configurar funções globais
	if err := setupGlobalFunctions(global); err != nil {
		return nil, err
	}

	executor := &V8Executor{
		isolate:    isolate,
		global:     global,
		modules:    make(map[string]*v8.Object),
		modulePath: "./modules",
	}

	// Criar diretório de módulos se não existir
	if err := os.MkdirAll(executor.modulePath, 0755); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de módulos: %v", err)
	}

	return executor, nil
}

// Execute executa um script JavaScript
func (e *V8Executor) Execute(options JSExecutionOptions) (*JSResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	startTime := time.Now()
	ctx := v8.NewContext(e.isolate, e.global)
	defer ctx.Close()

	// Configurar contexto
	if options.Context != nil {
		if err := e.setupContext(ctx, options.Context); err != nil {
			return nil, err
		}
	}

	// Executar com timeout se especificado
	var (
		result *JSResult
		err    error
	)

	if options.Context != nil && options.Context.Timeout > 0 {
		done := make(chan bool)
		go func() {
			result, err = e.executeScript(ctx, options)
			done <- true
		}()

		select {
		case <-done:
			// Script completou normalmente
		case <-time.After(time.Duration(options.Context.Timeout) * time.Second):
			return nil, fmt.Errorf("timeout após %d segundos", options.Context.Timeout)
		}
	} else {
		result, err = e.executeScript(ctx, options)
	}

	if err != nil {
		return nil, err
	}

	result.Duration = time.Since(startTime).String()
	result.MemoryUsed = e.getMemoryUsage()

	return result, nil
}

// EvaluateExpression avalia uma expressão JavaScript
func (e *V8Executor) EvaluateExpression(expression string) (*JSValue, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	ctx := v8.NewContext(e.isolate, e.global)
	defer ctx.Close()

	val, err := ctx.RunScript(expression, "expression")
	if err != nil {
		return nil, fmt.Errorf("erro ao avaliar expressão: %v", err)
	}

	return convertV8ValueToJSValue(val)
}

// CreateContext cria um novo contexto de execução
func (e *V8Executor) CreateContext(globals map[string]interface{}) (*JSContext, error) {
	context := &JSContext{
		Globals:     globals,
		Timeout:     30, // 30 segundos por padrão
		MemoryLimit: 128 * 1024 * 1024, // 128MB por padrão
	}

	return context, nil
}

// LoadModule carrega um módulo JavaScript
func (e *V8Executor) LoadModule(name, source string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Salvar módulo em arquivo
	filename := filepath.Join(e.modulePath, name+".js")
	if err := os.WriteFile(filename, []byte(source), 0644); err != nil {
		return fmt.Errorf("erro ao salvar módulo: %v", err)
	}

	// Carregar e compilar módulo
	ctx := v8.NewContext(e.isolate, e.global)
	defer ctx.Close()

	obj, err := ctx.RunScript(source, name)
	if err != nil {
		return fmt.Errorf("erro ao carregar módulo: %v", err)
	}

	if obj.IsObject() {
		e.modules[name] = obj.Object()
	}

	return nil
}

// GetAvailableModules retorna os módulos disponíveis
func (e *V8Executor) GetAvailableModules() []string {
	modules := make([]string, 0, len(e.modules))
	for name := range e.modules {
		modules = append(modules, name)
	}
	return modules
}

// Funções auxiliares

func (e *V8Executor) executeScript(ctx *v8.Context, options JSExecutionOptions) (*JSResult, error) {
	// Preparar resultado
	result := &JSResult{}

	// Executar script
	val, err := ctx.RunScript(options.Script, "script")
	if err != nil {
		if err, ok := err.(*v8.JSError); ok {
			result.Error = &JSError{
				Message:    err.Message,
				LineNumber: err.Location.LineNumber,
				Column:     err.Location.ColumnNumber,
				Stack:     err.StackTrace,
				Source:    options.Script,
			}
			return result, nil
		}
		return nil, err
	}

	// Converter resultado
	jsValue, err := convertV8ValueToJSValue(val)
	if err != nil {
		return nil, err
	}

	result.Value = jsValue
	return result, nil
}

func (e *V8Executor) setupContext(ctx *v8.Context, jsCtx *JSContext) error {
	// Adicionar variáveis globais
	for key, value := range jsCtx.Globals {
		if err := setGlobalValue(ctx, key, value); err != nil {
			return err
		}
	}

	// Carregar módulos
	for _, module := range jsCtx.Modules {
		if obj, ok := e.modules[module]; ok {
			global := ctx.Global()
			if err := global.Set(module, obj); err != nil {
				return fmt.Errorf("erro ao carregar módulo %s: %v", module, err)
			}
		}
	}

	return nil
}

func setupGlobalFunctions(global *v8.ObjectTemplate) error {
	// Função console.log
	console := v8.NewObjectTemplate(global.Isolate())
	console.Set("log", v8.NewFunctionTemplate(global.Isolate(), func(info *v8.FunctionCallbackInfo) *v8.Value {
		args := make([]interface{}, info.Length())
		for i := 0; i < info.Length(); i++ {
			args[i] = info.Get(i).String()
		}
		fmt.Println(args...)
		return nil
	}))

	if err := global.Set("console", console); err != nil {
		return fmt.Errorf("erro ao configurar console: %v", err)
	}

	return nil
}

func setGlobalValue(ctx *v8.Context, key string, value interface{}) error {
	v8val, err := convertGoValueToV8(ctx.Isolate(), value)
	if err != nil {
		return err
	}

	return ctx.Global().Set(key, v8val)
}

func convertV8ValueToJSValue(val *v8.Value) (*JSValue, error) {
	if val == nil {
		return &JSValue{Type: "undefined"}, nil
	}

	jsVal := &JSValue{}

	switch {
	case val.IsBoolean():
		jsVal.Type = "boolean"
		jsVal.Value = val.Boolean()
	case val.IsNumber():
		jsVal.Type = "number"
		jsVal.Value = val.Number()
	case val.IsString():
		jsVal.Type = "string"
		jsVal.Value = val.String()
	case val.IsObject():
		jsVal.Type = "object"
		if obj := val.Object(); obj != nil {
			jsVal.Value = convertV8ObjectToMap(obj)
		}
	case val.IsArray():
		jsVal.Type = "array"
		if arr := val.Object(); arr != nil {
			jsVal.Value = convertV8ArrayToSlice(arr)
		}
	case val.IsFunction():
		jsVal.Type = "function"
		jsVal.Value = "[Function]"
	case val.IsUndefined():
		jsVal.Type = "undefined"
	case val.IsNull():
		jsVal.Type = "null"
	default:
		return nil, fmt.Errorf("tipo não suportado: %v", val)
	}

	return jsVal, nil
}

func convertGoValueToV8(isolate *v8.Isolate, value interface{}) (*v8.Value, error) {
	switch v := value.(type) {
	case bool:
		return v8.NewValue(isolate, v)
	case int:
		return v8.NewValue(isolate, float64(v))
	case float64:
		return v8.NewValue(isolate, v)
	case string:
		return v8.NewValue(isolate, v)
	case []interface{}:
		arr := v8.NewArray(isolate)
		for i, item := range v {
			val, err := convertGoValueToV8(isolate, item)
			if err != nil {
				return nil, err
			}
			if err := arr.Set(uint32(i), val); err != nil {
				return nil, err
			}
		}
		return v8.NewValue(isolate, arr)
	case map[string]interface{}:
		obj := v8.NewObjectTemplate(isolate)
		for key, val := range v {
			v8val, err := convertGoValueToV8(isolate, val)
			if err != nil {
				return nil, err
			}
			if err := obj.Set(key, v8val); err != nil {
				return nil, err
			}
		}
		return v8.NewValue(isolate, obj)
	default:
		return v8.NewValue(isolate, nil)
	}
}

func convertV8ObjectToMap(obj *v8.Object) map[string]interface{} {
	result := make(map[string]interface{})
	names := obj.GetPropertyNames()
	for _, name := range names {
		if val, err := obj.Get(name); err == nil {
			if jsVal, err := convertV8ValueToJSValue(val); err == nil {
				result[name] = jsVal.Value
			}
		}
	}
	return result
}

func convertV8ArrayToSlice(arr *v8.Object) []interface{} {
	length := arr.GetPropertyNames()
	result := make([]interface{}, len(length))
	for i := 0; i < len(length); i++ {
		if val, err := arr.GetIndex(uint32(i)); err == nil {
			if jsVal, err := convertV8ValueToJSValue(val); err == nil {
				result[i] = jsVal.Value
			}
		}
	}
	return result
}

func (e *V8Executor) getMemoryUsage() int64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.Alloc)
} 