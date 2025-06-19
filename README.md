# anyone

anyone is a package that provides a function to run multiple functions in parallel and return the first successful result or return the last error result if all the functions return an error or panic.

# usage
```go

package main

import "https://github.com/theanarkh/anyone"

type Dummy struct {}

func main() {  
    result, err := anyone.Run(
		context.Background(),
		[]Worker[*Dummy]{
			func() (*Dummy, error) {
				return &Dummy{}, nil
			},
			func() (*Dummy, error) {
				return errors.New("error")
			},
		},
	)
    fmt.Println(result)
}
```