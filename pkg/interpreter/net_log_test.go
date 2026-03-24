package interpreter

import (
        "fmt"
        "testing"
)

// ============================================================
// NetProvider Tests
// ============================================================

func TestMockNetProvider_DefaultResponse(t *testing.T) {
        np := NewMockNetProvider()
        resp, err := np.Get("http://example.com", nil)
        if err != nil {
                t.Fatalf("unexpected error: %v", err)
        }
        if resp.Status != 200 {
                t.Errorf("expected status 200, got %d", resp.Status)
        }
        if resp.StatusText != "200 OK" {
                t.Errorf("expected status text '200 OK', got '%s'", resp.StatusText)
        }
}

func TestMockNetProvider_AddResponse(t *testing.T) {
        np := NewMockNetProvider()
        np.AddResponse("http://api.example.com/users", &NetResponse{
                Status:     201,
                StatusText: "201 Created",
                Body:       `{"id": 1, "name": "Alice"}`,
                Headers:    map[string]string{"Content-Type": "application/json"},
        })

        resp, err := np.Get("http://api.example.com/users", nil)
        if err != nil {
                t.Fatalf("unexpected error: %v", err)
        }
        if resp.Status != 201 {
                t.Errorf("expected status 201, got %d", resp.Status)
        }
        if resp.Body != `{"id": 1, "name": "Alice"}` {
                t.Errorf("unexpected body: %s", resp.Body)
        }
}

func TestMockNetProvider_SetError(t *testing.T) {
        np := NewMockNetProvider()
        np.SetError("http://fail.example.com", fmt.Errorf("connection refused"))

        _, err := np.Get("http://fail.example.com", nil)
        if err == nil {
                t.Fatal("expected error, got nil")
        }
        if err.Error() != "connection refused" {
                t.Errorf("unexpected error: %v", err)
        }
}

func TestMockNetProvider_RequestLog(t *testing.T) {
        np := NewMockNetProvider()
        np.Get("http://example.com/a", map[string]string{"Auth": "Bearer x"})
        np.Post("http://example.com/b", "body", nil)
        np.Put("http://example.com/c", "data", nil)
        np.Delete("http://example.com/d", nil)

        log := np.GetRequestLog()
        if len(log) != 4 {
                t.Fatalf("expected 4 requests, got %d", len(log))
        }

        if log[0].Method != "GET" || log[0].URL != "http://example.com/a" {
                t.Errorf("request 0: %+v", log[0])
        }
        if log[0].Headers["Auth"] != "Bearer x" {
                t.Errorf("expected Auth header, got %+v", log[0].Headers)
        }
        if log[1].Method != "POST" || log[1].Body != "body" {
                t.Errorf("request 1: %+v", log[1])
        }
        if log[2].Method != "PUT" {
                t.Errorf("request 2 method: %s", log[2].Method)
        }
        if log[3].Method != "DELETE" {
                t.Errorf("request 3 method: %s", log[3].Method)
        }
}

func TestMockNetProvider_RequestCount(t *testing.T) {
        np := NewMockNetProvider()
        if np.RequestCount() != 0 {
                t.Errorf("expected 0, got %d", np.RequestCount())
        }
        np.Get("http://example.com", nil)
        np.Post("http://example.com", "", nil)
        if np.RequestCount() != 2 {
                t.Errorf("expected 2, got %d", np.RequestCount())
        }
}

func TestMockNetProvider_SetDefaultResponse(t *testing.T) {
        np := NewMockNetProvider()
        np.SetDefaultResponse(&NetResponse{
                Status:     404,
                StatusText: "404 Not Found",
                Body:       "not found",
                Headers:    map[string]string{},
        })

        resp, err := np.Get("http://any-url.com", nil)
        if err != nil {
                t.Fatalf("unexpected error: %v", err)
        }
        if resp.Status != 404 {
                t.Errorf("expected 404, got %d", resp.Status)
        }
        if resp.Body != "not found" {
                t.Errorf("expected 'not found', got '%s'", resp.Body)
        }
}

func TestMockNetProvider_CustomRequest(t *testing.T) {
        np := NewMockNetProvider()
        np.AddResponse("http://example.com/patch", &NetResponse{
                Status:     204,
                StatusText: "204 No Content",
                Body:       "",
                Headers:    map[string]string{},
        })

        resp, err := np.Request("PATCH", "http://example.com/patch", `{"field":"value"}`, nil, 5000)
        if err != nil {
                t.Fatalf("unexpected error: %v", err)
        }
        if resp.Status != 204 {
                t.Errorf("expected 204, got %d", resp.Status)
        }

        log := np.GetRequestLog()
        if log[0].Method != "PATCH" {
                t.Errorf("expected PATCH, got %s", log[0].Method)
        }
        if log[0].Body != `{"field":"value"}` {
                t.Errorf("unexpected body in log: %s", log[0].Body)
        }
}

func TestMockNetProvider_MultipleURLResponses(t *testing.T) {
        np := NewMockNetProvider()
        np.AddResponse("http://a.com", &NetResponse{Status: 200, Body: "A"})
        np.AddResponse("http://b.com", &NetResponse{Status: 201, Body: "B"})

        respA, _ := np.Get("http://a.com", nil)
        respB, _ := np.Get("http://b.com", nil)

        if respA.Body != "A" || respB.Body != "B" {
                t.Errorf("unexpected responses: A=%s, B=%s", respA.Body, respB.Body)
        }
}

func TestMockNetProvider_ErrorOverridesResponse(t *testing.T) {
        np := NewMockNetProvider()
        np.AddResponse("http://err.com", &NetResponse{Status: 200, Body: "ok"})
        np.SetError("http://err.com", fmt.Errorf("network down"))

        _, err := np.Get("http://err.com", nil)
        if err == nil || err.Error() != "network down" {
                t.Errorf("expected 'network down' error, got %v", err)
        }
}

// ============================================================
// LogProvider Tests
// ============================================================

func TestMockLogProvider_Basic(t *testing.T) {
        lp := NewMockLogProvider()
        lp.Info("hello", nil)
        lp.Warn("caution", nil)
        lp.Error("failure", nil)
        lp.Debug("trace", nil)

        if lp.LogCount() != 4 {
                t.Fatalf("expected 4 logs, got %d", lp.LogCount())
        }

        logs := lp.GetLogs()
        if logs[0].Level != "INFO" || logs[0].Message != "hello" {
                t.Errorf("log 0: %+v", logs[0])
        }
        if logs[1].Level != "WARN" || logs[1].Message != "caution" {
                t.Errorf("log 1: %+v", logs[1])
        }
        if logs[2].Level != "ERROR" || logs[2].Message != "failure" {
                t.Errorf("log 2: %+v", logs[2])
        }
        if logs[3].Level != "DEBUG" || logs[3].Message != "trace" {
                t.Errorf("log 3: %+v", logs[3])
        }
}

func TestMockLogProvider_WithContext(t *testing.T) {
        lp := NewMockLogProvider()
        ctx := map[string]interface{}{
                "user_id": int64(42),
                "action":  "login",
        }
        lp.Info("user action", ctx)

        logs := lp.GetLogs()
        if len(logs) != 1 {
                t.Fatalf("expected 1 log, got %d", len(logs))
        }
        if logs[0].Context["user_id"] != int64(42) {
                t.Errorf("expected user_id 42, got %v", logs[0].Context["user_id"])
        }
        if logs[0].Context["action"] != "login" {
                t.Errorf("expected action 'login', got %v", logs[0].Context["action"])
        }
}

func TestMockLogProvider_HasLog(t *testing.T) {
        lp := NewMockLogProvider()
        lp.Info("found it", nil)
        lp.Warn("be careful", nil)

        if !lp.HasLog("INFO", "found it") {
                t.Error("expected to find INFO 'found it'")
        }
        if !lp.HasLog("WARN", "be careful") {
                t.Error("expected to find WARN 'be careful'")
        }
        if lp.HasLog("ERROR", "found it") {
                t.Error("should not find ERROR 'found it'")
        }
        if lp.HasLog("INFO", "missing") {
                t.Error("should not find INFO 'missing'")
        }
}

func TestMockLogProvider_GetLogsByLevel(t *testing.T) {
        lp := NewMockLogProvider()
        lp.Info("a", nil)
        lp.Info("b", nil)
        lp.Warn("c", nil)
        lp.Error("d", nil)

        infos := lp.GetLogsByLevel("INFO")
        if len(infos) != 2 {
                t.Errorf("expected 2 INFO logs, got %d", len(infos))
        }
        warns := lp.GetLogsByLevel("WARN")
        if len(warns) != 1 {
                t.Errorf("expected 1 WARN log, got %d", len(warns))
        }
        debugs := lp.GetLogsByLevel("DEBUG")
        if len(debugs) != 0 {
                t.Errorf("expected 0 DEBUG logs, got %d", len(debugs))
        }
}

func TestMockLogProvider_Clear(t *testing.T) {
        lp := NewMockLogProvider()
        lp.Info("test", nil)
        lp.Info("test2", nil)
        if lp.LogCount() != 2 {
                t.Errorf("expected 2, got %d", lp.LogCount())
        }
        lp.Clear()
        if lp.LogCount() != 0 {
                t.Errorf("expected 0 after clear, got %d", lp.LogCount())
        }
}

func TestMockLogProvider_ContextIsolation(t *testing.T) {
        lp := NewMockLogProvider()
        ctx := map[string]interface{}{"key": "value"}
        lp.Info("test", ctx)

        // Modify original context
        ctx["key"] = "modified"

        // Log should have original value
        logs := lp.GetLogs()
        if logs[0].Context["key"] != "value" {
                t.Errorf("context was not copied: got %v", logs[0].Context["key"])
        }
}

func TestRealLogProvider_Basic(t *testing.T) {
        // RealLogProvider also stores logs for GetLogs()
        lp := NewRealLogProvider()
        lp.Info("real info", nil)
        lp.Debug("real debug", map[string]interface{}{"x": int64(1)})

        logs := lp.GetLogs()
        if len(logs) != 2 {
                t.Fatalf("expected 2 logs, got %d", len(logs))
        }
        if logs[0].Level != "INFO" || logs[0].Message != "real info" {
                t.Errorf("log 0: %+v", logs[0])
        }
        if logs[1].Level != "DEBUG" || logs[1].Message != "real debug" {
                t.Errorf("log 1: %+v", logs[1])
        }
}

// ============================================================
// EffectContext Net/Log Integration Tests
// ============================================================

func TestEffectContext_HasNetAndLog(t *testing.T) {
        ctx := NewEffectContext()
        if ctx.Net() == nil {
                t.Error("real context should have NetProvider")
        }
        if ctx.Log() == nil {
                t.Error("real context should have LogProvider")
        }
}

func TestMockEffectContext_HasNetAndLog(t *testing.T) {
        ctx := NewMockEffectContext()
        if ctx.Net() == nil {
                t.Error("mock context should have NetProvider")
        }
        if ctx.Log() == nil {
                t.Error("mock context should have LogProvider")
        }
        // Check they are actually mock implementations
        if GetMockNetProvider(ctx) == nil {
                t.Error("expected MockNetProvider")
        }
        if GetMockLogProvider(ctx) == nil {
                t.Error("expected MockLogProvider")
        }
}

func TestEffectContext_WithNet(t *testing.T) {
        ctx := NewMockEffectContext()
        np := NewMockNetProvider()
        np.AddResponse("http://test.com", &NetResponse{Status: 418, Body: "teapot"})

        newCtx := ctx.WithNet(np)
        resp, _ := newCtx.Net().Get("http://test.com", nil)
        if resp.Status != 418 {
                t.Errorf("expected 418, got %d", resp.Status)
        }
        // Original should be unaffected
        resp2, _ := ctx.Net().Get("http://test.com", nil)
        if resp2.Status != 200 {
                t.Errorf("original should have default 200, got %d", resp2.Status)
        }
}

func TestEffectContext_WithLog(t *testing.T) {
        ctx := NewMockEffectContext()
        lp := NewMockLogProvider()

        newCtx := ctx.WithLog(lp)
        newCtx.Log().Info("test", nil)

        if GetMockLogProvider(newCtx).LogCount() != 1 {
                t.Error("expected 1 log in new context")
        }
        if GetMockLogProvider(ctx).LogCount() != 0 {
                t.Error("expected 0 logs in original context")
        }
}

func TestEffectContext_Clone_IncludesNetLog(t *testing.T) {
        ctx := NewMockEffectContext()
        GetMockNetProvider(ctx).AddResponse("http://x.com", &NetResponse{Status: 202, Body: "ok"})
        ctx.Log().Info("pre-clone", nil)

        cloned := ctx.Clone()
        // Cloned shares the same providers
        resp, _ := cloned.Net().Get("http://x.com", nil)
        if resp.Status != 202 {
                t.Errorf("expected 202, got %d", resp.Status)
        }
        if GetMockLogProvider(cloned).LogCount() != 1 {
                t.Error("expected 1 log in cloned context")
        }
}

func TestEffectContext_DeriveWithNetLog(t *testing.T) {
        ctx := NewMockEffectContext()
        np := NewMockNetProvider()
        lp := NewMockLogProvider()

        derived := ctx.DeriveWithNetLog(np, lp)
        if GetMockNetProvider(derived) != np {
                t.Error("net provider should be overridden")
        }
        if GetMockLogProvider(derived) != lp {
                t.Error("log provider should be overridden")
        }
        // File should be inherited
        if derived.File() != ctx.File() {
                t.Error("file provider should be inherited")
        }
}

func TestEffectContext_DeriveWithNetLog_NilKeepsOriginal(t *testing.T) {
        ctx := NewMockEffectContext()
        derived := ctx.DeriveWithNetLog(nil, nil)
        if derived.Net() != ctx.Net() {
                t.Error("nil net should keep original")
        }
        if derived.Log() != ctx.Log() {
                t.Error("nil log should keep original")
        }
}

func TestMockBuilder_WithNetProvider(t *testing.T) {
        np := NewMockNetProvider()
        np.AddResponse("http://builder.com", &NetResponse{Status: 301})

        ctx := NewMockBuilder().WithNetProvider(np).Build()
        resp, _ := ctx.Net().Get("http://builder.com", nil)
        if resp.Status != 301 {
                t.Errorf("expected 301, got %d", resp.Status)
        }
}

func TestMockBuilder_WithLogProvider(t *testing.T) {
        lp := NewMockLogProvider()
        ctx := NewMockBuilder().WithLogProvider(lp).Build()
        ctx.Log().Info("builder test", nil)
        if lp.LogCount() != 1 {
                t.Error("expected 1 log")
        }
}

func TestMockBuilder_WithMockResponse(t *testing.T) {
        ctx := NewMockBuilder().
                WithMockResponse("http://api.com/v1", &NetResponse{
                        Status: 200,
                        Body:   `{"result": "success"}`,
                }).
                Build()

        resp, _ := ctx.Net().Get("http://api.com/v1", nil)
        if resp.Body != `{"result": "success"}` {
                t.Errorf("unexpected body: %s", resp.Body)
        }
}

func TestGetMockNetProvider_NonMock(t *testing.T) {
        ctx := NewEffectContext()
        if GetMockNetProvider(ctx) != nil {
                t.Error("real context should return nil for GetMockNetProvider")
        }
}

func TestGetMockLogProvider_NonMock(t *testing.T) {
        ctx := NewEffectContext()
        if GetMockLogProvider(ctx) != nil {
                t.Error("real context should return nil for GetMockLogProvider")
        }
}

// ============================================================
// std.net Function Tests
// ============================================================

func TestStdNet_Get(t *testing.T) {
        np := NewMockNetProvider()
        np.AddResponse("http://api.test/data", &NetResponse{
                Status:     200,
                StatusText: "200 OK",
                Body:       `{"key":"value"}`,
                Headers:    map[string]string{"Content-Type": "application/json"},
        })
        exports := createStdNetExports(np)

        getFn := exports["get"].(*BuiltinFnVal)
        result := getFn.Fn([]Value{&StringVal{Val: "http://api.test/data"}})

        res, ok := result.(*ResultVal)
        if !ok || !res.IsOk {
                t.Fatalf("expected Ok result, got %v", result)
        }
        respMap := res.Val.(*MapVal)
        // Check status
        for i, k := range respMap.Keys {
                if ks, ok := k.(*StringVal); ok && ks.Val == "status" {
                        if sv, ok := respMap.Values[i].(*IntVal); ok {
                                if sv.Val != 200 {
                                        t.Errorf("expected status 200, got %d", sv.Val)
                                }
                        }
                }
                if ks, ok := k.(*StringVal); ok && ks.Val == "body" {
                        if sv, ok := respMap.Values[i].(*StringVal); ok {
                                if sv.Val != `{"key":"value"}` {
                                        t.Errorf("unexpected body: %s", sv.Val)
                                }
                        }
                }
        }
}

func TestStdNet_Get_WithHeaders(t *testing.T) {
        np := NewMockNetProvider()
        exports := createStdNetExports(np)

        getFn := exports["get"].(*BuiltinFnVal)
        headers := &MapVal{
                Keys:   []Value{&StringVal{Val: "Authorization"}},
                Values: []Value{&StringVal{Val: "Bearer token123"}},
        }
        getFn.Fn([]Value{&StringVal{Val: "http://api.test"}, headers})

        log := np.GetRequestLog()
        if log[0].Headers["Authorization"] != "Bearer token123" {
                t.Errorf("expected auth header, got %+v", log[0].Headers)
        }
}

func TestStdNet_Get_Error(t *testing.T) {
        np := NewMockNetProvider()
        np.SetError("http://fail.test", fmt.Errorf("timeout"))
        exports := createStdNetExports(np)

        getFn := exports["get"].(*BuiltinFnVal)
        result := getFn.Fn([]Value{&StringVal{Val: "http://fail.test"}})

        res := result.(*ResultVal)
        if res.IsOk {
                t.Fatal("expected Err result")
        }
        if errMsg := res.Val.(*StringVal).Val; errMsg != "timeout" {
                t.Errorf("expected 'timeout', got '%s'", errMsg)
        }
}

func TestStdNet_Post(t *testing.T) {
        np := NewMockNetProvider()
        np.AddResponse("http://api.test/create", &NetResponse{Status: 201, Body: "created"})
        exports := createStdNetExports(np)

        postFn := exports["post"].(*BuiltinFnVal)
        result := postFn.Fn([]Value{
                &StringVal{Val: "http://api.test/create"},
                &StringVal{Val: `{"name":"test"}`},
        })

        res := result.(*ResultVal)
        if !res.IsOk {
                t.Fatal("expected Ok result")
        }

        log := np.GetRequestLog()
        if log[0].Method != "POST" || log[0].Body != `{"name":"test"}` {
                t.Errorf("unexpected request: %+v", log[0])
        }
}

func TestStdNet_Put(t *testing.T) {
        np := NewMockNetProvider()
        np.AddResponse("http://api.test/update", &NetResponse{Status: 200, Body: "updated"})
        exports := createStdNetExports(np)

        putFn := exports["put"].(*BuiltinFnVal)
        result := putFn.Fn([]Value{
                &StringVal{Val: "http://api.test/update"},
                &StringVal{Val: `{"name":"updated"}`},
        })

        res := result.(*ResultVal)
        if !res.IsOk {
                t.Fatal("expected Ok result")
        }

        log := np.GetRequestLog()
        if log[0].Method != "PUT" {
                t.Errorf("expected PUT, got %s", log[0].Method)
        }
}

func TestStdNet_Delete(t *testing.T) {
        np := NewMockNetProvider()
        np.AddResponse("http://api.test/item/1", &NetResponse{Status: 204})
        exports := createStdNetExports(np)

        deleteFn := exports["delete"].(*BuiltinFnVal)
        result := deleteFn.Fn([]Value{&StringVal{Val: "http://api.test/item/1"}})

        res := result.(*ResultVal)
        if !res.IsOk {
                t.Fatal("expected Ok result")
        }

        log := np.GetRequestLog()
        if log[0].Method != "DELETE" {
                t.Errorf("expected DELETE, got %s", log[0].Method)
        }
}

func TestStdNet_Request(t *testing.T) {
        np := NewMockNetProvider()
        np.AddResponse("http://api.test/custom", &NetResponse{Status: 200, Body: "ok"})
        exports := createStdNetExports(np)

        requestFn := exports["request"].(*BuiltinFnVal)
        config := &MapVal{
                Keys: []Value{
                        &StringVal{Val: "method"},
                        &StringVal{Val: "url"},
                        &StringVal{Val: "body"},
                        &StringVal{Val: "timeout"},
                },
                Values: []Value{
                        &StringVal{Val: "PATCH"},
                        &StringVal{Val: "http://api.test/custom"},
                        &StringVal{Val: `{"patch":"data"}`},
                        &IntVal{Val: 5000},
                },
        }
        result := requestFn.Fn([]Value{config})

        res := result.(*ResultVal)
        if !res.IsOk {
                t.Fatal("expected Ok result")
        }

        log := np.GetRequestLog()
        if log[0].Method != "PATCH" {
                t.Errorf("expected PATCH, got %s", log[0].Method)
        }
}

func TestStdNet_Request_WithHeaders(t *testing.T) {
        np := NewMockNetProvider()
        exports := createStdNetExports(np)

        requestFn := exports["request"].(*BuiltinFnVal)
        headerMap := &MapVal{
                Keys:   []Value{&StringVal{Val: "X-Custom"}},
                Values: []Value{&StringVal{Val: "test-value"}},
        }
        config := &MapVal{
                Keys: []Value{
                        &StringVal{Val: "method"},
                        &StringVal{Val: "url"},
                        &StringVal{Val: "headers"},
                },
                Values: []Value{
                        &StringVal{Val: "GET"},
                        &StringVal{Val: "http://test.com"},
                        headerMap,
                },
        }
        requestFn.Fn([]Value{config})

        log := np.GetRequestLog()
        if log[0].Headers["X-Custom"] != "test-value" {
                t.Errorf("expected X-Custom header, got %+v", log[0].Headers)
        }
}

func TestStdNet_Request_MissingMethod(t *testing.T) {
        np := NewMockNetProvider()
        exports := createStdNetExports(np)

        requestFn := exports["request"].(*BuiltinFnVal)
        config := &MapVal{
                Keys:   []Value{&StringVal{Val: "url"}},
                Values: []Value{&StringVal{Val: "http://test.com"}},
        }

        defer func() {
                r := recover()
                if r == nil {
                        t.Fatal("expected panic for missing method")
                }
                if re, ok := r.(*RuntimeError); ok {
                        if re.Message != "net.request() config must include 'method'" {
                                t.Errorf("unexpected error: %s", re.Message)
                        }
                }
        }()
        requestFn.Fn([]Value{config})
}

func TestStdNet_Request_MissingURL(t *testing.T) {
        np := NewMockNetProvider()
        exports := createStdNetExports(np)

        requestFn := exports["request"].(*BuiltinFnVal)
        config := &MapVal{
                Keys:   []Value{&StringVal{Val: "method"}},
                Values: []Value{&StringVal{Val: "GET"}},
        }

        defer func() {
                r := recover()
                if r == nil {
                        t.Fatal("expected panic for missing url")
                }
        }()
        requestFn.Fn([]Value{config})
}

func TestStdNet_Get_InvalidArgs(t *testing.T) {
        np := NewMockNetProvider()
        exports := createStdNetExports(np)
        getFn := exports["get"].(*BuiltinFnVal)

        // No args
        func() {
                defer func() {
                        if r := recover(); r == nil {
                                t.Error("expected panic with no args")
                        }
                }()
                getFn.Fn([]Value{})
        }()

        // Non-string arg
        func() {
                defer func() {
                        if r := recover(); r == nil {
                                t.Error("expected panic with non-string arg")
                        }
                }()
                getFn.Fn([]Value{&IntVal{Val: 42}})
        }()
}

func TestStdNet_Post_InvalidArgs(t *testing.T) {
        np := NewMockNetProvider()
        exports := createStdNetExports(np)
        postFn := exports["post"].(*BuiltinFnVal)

        // Too few args
        func() {
                defer func() {
                        if r := recover(); r == nil {
                                t.Error("expected panic with 1 arg")
                        }
                }()
                postFn.Fn([]Value{&StringVal{Val: "url"}})
        }()

        // Non-string body
        func() {
                defer func() {
                        if r := recover(); r == nil {
                                t.Error("expected panic with non-string body")
                        }
                }()
                postFn.Fn([]Value{&StringVal{Val: "url"}, &IntVal{Val: 123}})
        }()
}

func TestStdNet_Request_InvalidArgs(t *testing.T) {
        np := NewMockNetProvider()
        exports := createStdNetExports(np)
        requestFn := exports["request"].(*BuiltinFnVal)

        // Non-map arg
        func() {
                defer func() {
                        if r := recover(); r == nil {
                                t.Error("expected panic with non-map arg")
                        }
                }()
                requestFn.Fn([]Value{&StringVal{Val: "not a map"}})
        }()
}

// ============================================================
// std.log Function Tests
// ============================================================

func TestStdLog_Info(t *testing.T) {
        lp := NewMockLogProvider()
        exports := createStdLogExports(lp)

        infoFn := exports["info"].(*BuiltinFnVal)
        result := infoFn.Fn([]Value{&StringVal{Val: "hello world"}})

        if _, ok := result.(*NoneVal); !ok {
                t.Error("expected NoneVal")
        }
        if !lp.HasLog("INFO", "hello world") {
                t.Error("expected INFO 'hello world'")
        }
}

func TestStdLog_Warn(t *testing.T) {
        lp := NewMockLogProvider()
        exports := createStdLogExports(lp)

        warnFn := exports["warn"].(*BuiltinFnVal)
        warnFn.Fn([]Value{&StringVal{Val: "be careful"}})

        if !lp.HasLog("WARN", "be careful") {
                t.Error("expected WARN 'be careful'")
        }
}

func TestStdLog_Error(t *testing.T) {
        lp := NewMockLogProvider()
        exports := createStdLogExports(lp)

        errorFn := exports["error"].(*BuiltinFnVal)
        errorFn.Fn([]Value{&StringVal{Val: "something broke"}})

        if !lp.HasLog("ERROR", "something broke") {
                t.Error("expected ERROR 'something broke'")
        }
}

func TestStdLog_Debug(t *testing.T) {
        lp := NewMockLogProvider()
        exports := createStdLogExports(lp)

        debugFn := exports["debug"].(*BuiltinFnVal)
        debugFn.Fn([]Value{&StringVal{Val: "trace info"}})

        if !lp.HasLog("DEBUG", "trace info") {
                t.Error("expected DEBUG 'trace info'")
        }
}

func TestStdLog_InfoWithContext(t *testing.T) {
        lp := NewMockLogProvider()
        exports := createStdLogExports(lp)

        infoFn := exports["info"].(*BuiltinFnVal)
        ctx := &MapVal{
                Keys:   []Value{&StringVal{Val: "user"}, &StringVal{Val: "count"}},
                Values: []Value{&StringVal{Val: "alice"}, &IntVal{Val: 5}},
        }
        infoFn.Fn([]Value{&StringVal{Val: "action"}, ctx})

        logs := lp.GetLogs()
        if len(logs) != 1 {
                t.Fatalf("expected 1 log, got %d", len(logs))
        }
        if logs[0].Context["user"] != "alice" {
                t.Errorf("expected user 'alice', got %v", logs[0].Context["user"])
        }
        if logs[0].Context["count"] != int64(5) {
                t.Errorf("expected count 5, got %v", logs[0].Context["count"])
        }
}

func TestStdLog_GetLogs(t *testing.T) {
        lp := NewMockLogProvider()
        exports := createStdLogExports(lp)

        infoFn := exports["info"].(*BuiltinFnVal)
        warnFn := exports["warn"].(*BuiltinFnVal)
        infoFn.Fn([]Value{&StringVal{Val: "msg1"}})
        warnFn.Fn([]Value{&StringVal{Val: "msg2"}})

        getLogsFn := exports["get_logs"].(*BuiltinFnVal)
        result := getLogsFn.Fn([]Value{})

        list, ok := result.(*ListVal)
        if !ok {
                t.Fatalf("expected ListVal, got %T", result)
        }
        if len(list.Elements) != 2 {
                t.Fatalf("expected 2 entries, got %d", len(list.Elements))
        }

        // Check first entry
        entry := list.Elements[0].(*MapVal)
        for i, k := range entry.Keys {
                if ks, ok := k.(*StringVal); ok && ks.Val == "level" {
                        if vs, ok := entry.Values[i].(*StringVal); ok && vs.Val != "INFO" {
                                t.Errorf("expected INFO, got %s", vs.Val)
                        }
                }
                if ks, ok := k.(*StringVal); ok && ks.Val == "message" {
                        if vs, ok := entry.Values[i].(*StringVal); ok && vs.Val != "msg1" {
                                t.Errorf("expected msg1, got %s", vs.Val)
                        }
                }
        }
}

func TestStdLog_GetLogs_Empty(t *testing.T) {
        lp := NewMockLogProvider()
        exports := createStdLogExports(lp)

        getLogsFn := exports["get_logs"].(*BuiltinFnVal)
        result := getLogsFn.Fn([]Value{})

        list := result.(*ListVal)
        if len(list.Elements) != 0 {
                t.Errorf("expected empty list, got %d elements", len(list.Elements))
        }
}

func TestStdLog_Info_InvalidArgs(t *testing.T) {
        lp := NewMockLogProvider()
        exports := createStdLogExports(lp)
        infoFn := exports["info"].(*BuiltinFnVal)

        // No args
        func() {
                defer func() {
                        if r := recover(); r == nil {
                                t.Error("expected panic with no args")
                        }
                }()
                infoFn.Fn([]Value{})
        }()

        // Non-string arg
        func() {
                defer func() {
                        if r := recover(); r == nil {
                                t.Error("expected panic with non-string arg")
                        }
                }()
                infoFn.Fn([]Value{&IntVal{Val: 42}})
        }()

        // Too many args
        func() {
                defer func() {
                        if r := recover(); r == nil {
                                t.Error("expected panic with 3 args")
                        }
                }()
                infoFn.Fn([]Value{&StringVal{Val: "a"}, &MapVal{}, &StringVal{Val: "extra"}})
        }()
}

func TestStdLog_GetLogs_InvalidArgs(t *testing.T) {
        lp := NewMockLogProvider()
        exports := createStdLogExports(lp)
        getLogsFn := exports["get_logs"].(*BuiltinFnVal)

        defer func() {
                if r := recover(); r == nil {
                        t.Error("expected panic with args")
                }
        }()
        getLogsFn.Fn([]Value{&StringVal{Val: "extra"}})
}

func TestStdLog_AllLevels(t *testing.T) {
        lp := NewMockLogProvider()
        exports := createStdLogExports(lp)

        exports["info"].(*BuiltinFnVal).Fn([]Value{&StringVal{Val: "i"}})
        exports["warn"].(*BuiltinFnVal).Fn([]Value{&StringVal{Val: "w"}})
        exports["error"].(*BuiltinFnVal).Fn([]Value{&StringVal{Val: "e"}})
        exports["debug"].(*BuiltinFnVal).Fn([]Value{&StringVal{Val: "d"}})

        if lp.LogCount() != 4 {
                t.Errorf("expected 4 logs, got %d", lp.LogCount())
        }

        levels := []string{"INFO", "WARN", "ERROR", "DEBUG"}
        msgs := []string{"i", "w", "e", "d"}
        for i, level := range levels {
                if !lp.HasLog(level, msgs[i]) {
                        t.Errorf("missing log: %s %s", level, msgs[i])
                }
        }
}

// ============================================================
// Conversion Helper Tests
// ============================================================

func TestAuraValueToGo(t *testing.T) {
        tests := []struct {
                input    Value
                expected interface{}
        }{
                {&IntVal{Val: 42}, int64(42)},
                {&FloatVal{Val: 3.14}, 3.14},
                {&StringVal{Val: "hello"}, "hello"},
                {&BoolVal{Val: true}, true},
                {&BoolVal{Val: false}, false},
                {&NoneVal{}, nil},
        }

        for _, tt := range tests {
                result := auraValueToGo(tt.input)
                if result != tt.expected {
                        t.Errorf("auraValueToGo(%v) = %v, want %v", tt.input, result, tt.expected)
                }
        }
}

func TestGoValueToAura(t *testing.T) {
        tests := []struct {
                input    interface{}
                expected string // use String() representation
        }{
                {int64(42), "42"},
                {3.14, "3.14"},
                {"hello", "hello"},
                {true, "true"},
                {false, "false"},
                {nil, "none"},
        }

        for _, tt := range tests {
                result := goValueToAura(tt.input)
                if result.String() != tt.expected {
                        t.Errorf("goValueToAura(%v) = %s, want %s", tt.input, result.String(), tt.expected)
                }
        }
}

func TestGoValueToAura_Unknown(t *testing.T) {
        result := goValueToAura(struct{}{})
        if s, ok := result.(*StringVal); !ok || s.Val != "<unknown>" {
                t.Errorf("expected <unknown>, got %v", result)
        }
}

// ============================================================
// Integration Tests
// ============================================================

func TestIntegration_NetAndLog_Together(t *testing.T) {
        ctx := NewMockEffectContext()

        // Set up mock net response
        np := GetMockNetProvider(ctx)
        np.AddResponse("http://api.test/data", &NetResponse{
                Status: 200,
                Body:   `{"result":"ok"}`,
        })

        // Use net
        resp, err := ctx.Net().Get("http://api.test/data", nil)
        if err != nil {
                t.Fatalf("unexpected error: %v", err)
        }
        if resp.Status != 200 {
                t.Errorf("expected 200, got %d", resp.Status)
        }

        // Log the result
        ctx.Log().Info("fetched data", map[string]interface{}{
                "status": int64(resp.Status),
                "url":    "http://api.test/data",
        })

        // Verify
        lp := GetMockLogProvider(ctx)
        if !lp.HasLog("INFO", "fetched data") {
                t.Error("expected log entry")
        }
        logs := lp.GetLogs()
        if logs[0].Context["status"] != int64(200) {
                t.Errorf("expected status 200 in context, got %v", logs[0].Context["status"])
        }
}

func TestIntegration_MockBuilder_FullSetup(t *testing.T) {
        ctx := NewMockBuilder().
                WithFile("/app/config.json", `{"debug":true}`).
                WithTime(1700000000).
                WithEnvVar("APP_ENV", "test").
                WithMockResponse("http://api.com/health", &NetResponse{
                        Status: 200,
                        Body:   "healthy",
                }).
                Build()

        // Verify all providers work
        content, err := ctx.File().ReadFile("/app/config.json")
        if err != nil || content != `{"debug":true}` {
                t.Error("file provider failed")
        }
        if ctx.Time().Now() != 1700000000 {
                t.Error("time provider failed")
        }
        val, ok := ctx.Env().Get("APP_ENV")
        if !ok || val != "test" {
                t.Error("env provider failed")
        }
        resp, err := ctx.Net().Get("http://api.com/health", nil)
        if err != nil || resp.Body != "healthy" {
                t.Error("net provider failed")
        }
        ctx.Log().Info("integration test", nil)
        if GetMockLogProvider(ctx).LogCount() != 1 {
                t.Error("log provider failed")
        }
}

func TestIntegration_StdNetWithInterpreter(t *testing.T) {
        np := NewMockNetProvider()
        np.AddResponse("http://test.api/users", &NetResponse{
                Status:     200,
                StatusText: "200 OK",
                Body:       `[{"name":"Alice"}]`,
                Headers:    map[string]string{"Content-Type": "application/json"},
        })

        exports := createStdNetExports(np)

        // Verify all 5 functions exist
        expectedFns := []string{"get", "post", "put", "delete", "request"}
        for _, name := range expectedFns {
                if _, ok := exports[name]; !ok {
                        t.Errorf("missing export: %s", name)
                }
        }
}

func TestIntegration_StdLogWithInterpreter(t *testing.T) {
        lp := NewMockLogProvider()
        exports := createStdLogExports(lp)

        // Verify all 6 functions exist
        expectedFns := []string{"info", "warn", "error", "debug", "with_context", "get_logs"}
        for _, name := range expectedFns {
                if _, ok := exports[name]; !ok {
                        t.Errorf("missing export: %s", name)
                }
        }
}

func TestStdNet_ResponseMapStructure(t *testing.T) {
        np := NewMockNetProvider()
        np.AddResponse("http://test.com", &NetResponse{
                Status:     200,
                StatusText: "200 OK",
                Body:       "hello",
                Headers:    map[string]string{"X-Test": "value"},
        })
        exports := createStdNetExports(np)

        getFn := exports["get"].(*BuiltinFnVal)
        result := getFn.Fn([]Value{&StringVal{Val: "http://test.com"}})
        res := result.(*ResultVal)
        respMap := res.Val.(*MapVal)

        // Should have 4 keys: status, status_text, body, headers
        if len(respMap.Keys) != 4 {
                t.Fatalf("expected 4 keys, got %d", len(respMap.Keys))
        }

        found := map[string]bool{}
        for _, k := range respMap.Keys {
                if ks, ok := k.(*StringVal); ok {
                        found[ks.Val] = true
                }
        }
        for _, key := range []string{"status", "status_text", "body", "headers"} {
                if !found[key] {
                        t.Errorf("missing key in response map: %s", key)
                }
        }
}

func TestStdLog_ContextWithMixedTypes(t *testing.T) {
        lp := NewMockLogProvider()
        exports := createStdLogExports(lp)

        infoFn := exports["info"].(*BuiltinFnVal)
        ctx := &MapVal{
                Keys: []Value{
                        &StringVal{Val: "str_val"},
                        &StringVal{Val: "int_val"},
                        &StringVal{Val: "float_val"},
                        &StringVal{Val: "bool_val"},
                },
                Values: []Value{
                        &StringVal{Val: "hello"},
                        &IntVal{Val: 42},
                        &FloatVal{Val: 3.14},
                        &BoolVal{Val: true},
                },
        }
        infoFn.Fn([]Value{&StringVal{Val: "mixed"}, ctx})

        logs := lp.GetLogs()
        if logs[0].Context["str_val"] != "hello" {
                t.Errorf("str_val: %v", logs[0].Context["str_val"])
        }
        if logs[0].Context["int_val"] != int64(42) {
                t.Errorf("int_val: %v", logs[0].Context["int_val"])
        }
        if logs[0].Context["float_val"] != 3.14 {
                t.Errorf("float_val: %v", logs[0].Context["float_val"])
        }
        if logs[0].Context["bool_val"] != true {
                t.Errorf("bool_val: %v", logs[0].Context["bool_val"])
        }
}
