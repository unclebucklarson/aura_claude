package interpreter

import "fmt"

// createStdFileExports creates the exports for the std.file module.
// The file provider is captured via closure, enabling effect mocking.
func createStdFileExports(fp FileProvider) map[string]Value {
	exports := make(map[string]Value)

	// read(path) -> Result[String, String]
	exports["read"] = &BuiltinFnVal{
		Name: "file.read",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "file.read() requires exactly 1 argument (path)"})
			}
			path, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("file.read() path must be a String, got %s", valueTypeNames[args[0].Type()])})
			}
			content, err := fp.ReadFile(path.Val)
			if err != nil {
				return &ResultVal{IsOk: false, Val: &StringVal{Val: err.Error()}}
			}
			return &ResultVal{IsOk: true, Val: &StringVal{Val: content}}
		},
	}

	// write(path, content) -> Result[None, String]
	exports["write"] = &BuiltinFnVal{
		Name: "file.write",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "file.write() requires exactly 2 arguments (path, content)"})
			}
			path, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("file.write() path must be a String, got %s", valueTypeNames[args[0].Type()])})
			}
			content, ok := args[1].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("file.write() content must be a String, got %s", valueTypeNames[args[1].Type()])})
			}
			err := fp.WriteFile(path.Val, content.Val)
			if err != nil {
				return &ResultVal{IsOk: false, Val: &StringVal{Val: err.Error()}}
			}
			return &ResultVal{IsOk: true, Val: &NoneVal{}}
		},
	}

	// append(path, content) -> Result[None, String]
	exports["append"] = &BuiltinFnVal{
		Name: "file.append",
		Fn: func(args []Value) Value {
			if len(args) != 2 {
				panic(&RuntimeError{Message: "file.append() requires exactly 2 arguments (path, content)"})
			}
			path, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("file.append() path must be a String, got %s", valueTypeNames[args[0].Type()])})
			}
			content, ok := args[1].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("file.append() content must be a String, got %s", valueTypeNames[args[1].Type()])})
			}
			err := fp.AppendFile(path.Val, content.Val)
			if err != nil {
				return &ResultVal{IsOk: false, Val: &StringVal{Val: err.Error()}}
			}
			return &ResultVal{IsOk: true, Val: &NoneVal{}}
		},
	}

	// exists(path) -> Bool
	exports["exists"] = &BuiltinFnVal{
		Name: "file.exists",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "file.exists() requires exactly 1 argument (path)"})
			}
			path, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("file.exists() path must be a String, got %s", valueTypeNames[args[0].Type()])})
			}
			return &BoolVal{Val: fp.Exists(path.Val)}
		},
	}

	// delete(path) -> Result[None, String]
	exports["delete"] = &BuiltinFnVal{
		Name: "file.delete",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "file.delete() requires exactly 1 argument (path)"})
			}
			path, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("file.delete() path must be a String, got %s", valueTypeNames[args[0].Type()])})
			}
			err := fp.Delete(path.Val)
			if err != nil {
				return &ResultVal{IsOk: false, Val: &StringVal{Val: err.Error()}}
			}
			return &ResultVal{IsOk: true, Val: &NoneVal{}}
		},
	}

	// list_dir(path) -> Result[List[String], String]
	exports["list_dir"] = &BuiltinFnVal{
		Name: "file.list_dir",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "file.list_dir() requires exactly 1 argument (path)"})
			}
			path, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("file.list_dir() path must be a String, got %s", valueTypeNames[args[0].Type()])})
			}
			entries, err := fp.ListDir(path.Val)
			if err != nil {
				return &ResultVal{IsOk: false, Val: &StringVal{Val: err.Error()}}
			}
			elements := make([]Value, len(entries))
			for i, e := range entries {
				elements[i] = &StringVal{Val: e}
			}
			return &ResultVal{IsOk: true, Val: &ListVal{Elements: elements}}
		},
	}

	// create_dir(path) -> Result[None, String]
	exports["create_dir"] = &BuiltinFnVal{
		Name: "file.create_dir",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "file.create_dir() requires exactly 1 argument (path)"})
			}
			path, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("file.create_dir() path must be a String, got %s", valueTypeNames[args[0].Type()])})
			}
			err := fp.CreateDir(path.Val)
			if err != nil {
				return &ResultVal{IsOk: false, Val: &StringVal{Val: err.Error()}}
			}
			return &ResultVal{IsOk: true, Val: &NoneVal{}}
		},
	}

	// is_file(path) -> Bool
	exports["is_file"] = &BuiltinFnVal{
		Name: "file.is_file",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "file.is_file() requires exactly 1 argument (path)"})
			}
			path, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("file.is_file() path must be a String, got %s", valueTypeNames[args[0].Type()])})
			}
			return &BoolVal{Val: fp.IsFile(path.Val)}
		},
	}

	// is_dir(path) -> Bool
	exports["is_dir"] = &BuiltinFnVal{
		Name: "file.is_dir",
		Fn: func(args []Value) Value {
			if len(args) != 1 {
				panic(&RuntimeError{Message: "file.is_dir() requires exactly 1 argument (path)"})
			}
			path, ok := args[0].(*StringVal)
			if !ok {
				panic(&RuntimeError{Message: fmt.Sprintf("file.is_dir() path must be a String, got %s", valueTypeNames[args[0].Type()])})
			}
			return &BoolVal{Val: fp.IsDir(path.Val)}
		},
	}

	return exports
}
