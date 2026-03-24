package interpreter

// stdlib_net.go implements the std.net module for HTTP network operations.
// All functions use the NetProvider from the effect system, enabling mockable HTTP.

// createStdNetExports creates the exports for the std.net module.
func createStdNetExports(provider NetProvider) map[string]Value {
	exports := make(map[string]Value)

	// Helper to convert NetResponse to an Aura MapVal
	responseToMap := func(resp *NetResponse) *MapVal {
		headers := &MapVal{
			Keys:   make([]Value, 0, len(resp.Headers)),
			Values: make([]Value, 0, len(resp.Headers)),
		}
		for k, v := range resp.Headers {
			headers.Keys = append(headers.Keys, &StringVal{Val: k})
			headers.Values = append(headers.Values, &StringVal{Val: v})
		}

		return &MapVal{
			Keys: []Value{
				&StringVal{Val: "status"},
				&StringVal{Val: "status_text"},
				&StringVal{Val: "body"},
				&StringVal{Val: "headers"},
			},
			Values: []Value{
				&IntVal{Val: int64(resp.Status)},
				&StringVal{Val: resp.StatusText},
				&StringVal{Val: resp.Body},
				headers,
			},
		}
	}

	// Helper to extract headers from an optional map argument
	extractHeaders := func(args []Value, idx int) map[string]string {
		headers := make(map[string]string)
		if idx < len(args) {
			if m, ok := args[idx].(*MapVal); ok {
				for i, k := range m.Keys {
					if ks, ok := k.(*StringVal); ok {
						if vs, ok := m.Values[i].(*StringVal); ok {
							headers[ks.Val] = vs.Val
						}
					}
				}
			}
		}
		return headers
	}

	// get(url) -> Result[Response, String]
	// get(url, headers) -> Result[Response, String]
	exports["get"] = &BuiltinFnVal{
		Name: "net.get",
		Fn: func(args []Value) Value {
			if len(args) < 1 || len(args) > 2 {
				panic(&RuntimeError{Message: "net.get() requires 1-2 arguments (url, headers?)"})
			}
			url, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: "net.get() first argument must be a string"})
			}
			headers := extractHeaders(args, 1)

			resp, err := provider.Get(url.Val, headers)
			if err != nil {
				return &ResultVal{IsOk: false, Val: &StringVal{Val: err.Error()}}
			}
			return &ResultVal{IsOk: true, Val: responseToMap(resp)}
		},
	}

	// post(url, body) -> Result[Response, String]
	// post(url, body, headers) -> Result[Response, String]
	exports["post"] = &BuiltinFnVal{
		Name: "net.post",
		Fn: func(args []Value) Value {
			if len(args) < 2 || len(args) > 3 {
				panic(&RuntimeError{Message: "net.post() requires 2-3 arguments (url, body, headers?)"})
			}
			url, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: "net.post() first argument must be a string"})
			}
			body, ok := args[1].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: "net.post() second argument must be a string"})
			}
			headers := extractHeaders(args, 2)

			resp, err := provider.Post(url.Val, body.Val, headers)
			if err != nil {
				return &ResultVal{IsOk: false, Val: &StringVal{Val: err.Error()}}
			}
			return &ResultVal{IsOk: true, Val: responseToMap(resp)}
		},
	}

	// put(url, body) -> Result[Response, String]
	// put(url, body, headers) -> Result[Response, String]
	exports["put"] = &BuiltinFnVal{
		Name: "net.put",
		Fn: func(args []Value) Value {
			if len(args) < 2 || len(args) > 3 {
				panic(&RuntimeError{Message: "net.put() requires 2-3 arguments (url, body, headers?)"})
			}
			url, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: "net.put() first argument must be a string"})
			}
			body, ok := args[1].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: "net.put() second argument must be a string"})
			}
			headers := extractHeaders(args, 2)

			resp, err := provider.Put(url.Val, body.Val, headers)
			if err != nil {
				return &ResultVal{IsOk: false, Val: &StringVal{Val: err.Error()}}
			}
			return &ResultVal{IsOk: true, Val: responseToMap(resp)}
		},
	}

	// delete(url) -> Result[Response, String]
	// delete(url, headers) -> Result[Response, String]
	exports["delete"] = &BuiltinFnVal{
		Name: "net.delete",
		Fn: func(args []Value) Value {
			if len(args) < 1 || len(args) > 2 {
				panic(&RuntimeError{Message: "net.delete() requires 1-2 arguments (url, headers?)"})
			}
			url, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: "net.delete() first argument must be a string"})
			}
			headers := extractHeaders(args, 1)

			resp, err := provider.Delete(url.Val, headers)
			if err != nil {
				return &ResultVal{IsOk: false, Val: &StringVal{Val: err.Error()}}
			}
			return &ResultVal{IsOk: true, Val: responseToMap(resp)}
		},
	}

	// request(config) -> Result[Response, String]
	// config is a Map with keys: method, url, body?, headers?, timeout?
	exports["request"] = &BuiltinFnVal{
		Name: "net.request",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "net.request() requires exactly 1 argument (config map)"})
			}
			config, ok := args[0].(*MapVal)
			if !ok {
				panic(&RuntimeError{Message: "net.request() argument must be a map"})
			}

			// Extract config fields
			var method, url, body string
			headers := make(map[string]string)
			timeoutMs := 0

			for i, k := range config.Keys {
				ks, ok := k.(*StringVal)
				if !ok {
					continue
				}
				switch ks.Val {
				case "method":
					if vs, ok := config.Values[i].(*StringVal); ok {
						method = vs.Val
					}
				case "url":
					if vs, ok := config.Values[i].(*StringVal); ok {
						url = vs.Val
					}
				case "body":
					if vs, ok := config.Values[i].(*StringVal); ok {
						body = vs.Val
					}
				case "headers":
					if m, ok := config.Values[i].(*MapVal); ok {
						for j, hk := range m.Keys {
							if hks, ok := hk.(*StringVal); ok {
								if hvs, ok := m.Values[j].(*StringVal); ok {
									headers[hks.Val] = hvs.Val
								}
							}
						}
					}
				case "timeout":
					if vs, ok := config.Values[i].(*IntVal); ok {
						timeoutMs = int(vs.Val)
					}
				}
			}

			if method == "" {
				panic(&RuntimeError{Message: "net.request() config must include 'method'"})
			}
			if url == "" {
				panic(&RuntimeError{Message: "net.request() config must include 'url'"})
			}

			resp, err := provider.Request(method, url, body, headers, timeoutMs)
			if err != nil {
				return &ResultVal{IsOk: false, Val: &StringVal{Val: err.Error()}}
			}
			return &ResultVal{IsOk: true, Val: responseToMap(resp)}
		},
	}

	return exports
}
